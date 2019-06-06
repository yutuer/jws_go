package herogacharace

import (
	"fmt"

	"sync"

	"math"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//限时神将系统, 跨服排行榜系统
// 直接最接对接独立Redis数据库
// 数据库配置从etcd中gameid:
//  a4k/gid/HeroGachaRace/jsonmap(rankgroup:name dbID: addr, port, auth)
func GetEtcdInfoKey() string {
	return fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, game.Cfg.Gid, ETCD_SERVICE_NAME)
}

//myzset: 命名规则： HGR:rankgroup:activityid
//member: 命名规则： acid:playername
//分数: 命名规则：满足相同分数下, 先来的占据高分位 100.(1.0/201610101212600) (分数.(1.0/时间自然序纳秒自然序-开服事件纳秒时间))

//这个清理只需要定时运行或者太长了再运行就可以 只保留列表中 100位 ZREMRANGEBYRANK myzset 0 -101
//获取前100位排名者  ZREVRANGE myzset 0 101 [WITHSCORES]
//
//添加我的积分参加排名 zadd myzset score member
//获取我当前的排名 zrevrank
//（acID, 积分，排名， 区服, 名字）

type HGRankMember struct {
	AccountID  string
	PlayerName string
}

func (m *HGRankMember) String() string {
	return fmt.Sprintf("%s|%s", m.AccountID, m.PlayerName)
}

type HeroGachaRankItem struct {
	sid uint

	Member HGRankMember
	Score  uint64
	Rank   uint64

	ShardDisplayName string
	PlayerName       string
}

// GetRedisKey
// 参考 member: 命名规则
// 参考 分数: 命名规则
func (i *HeroGachaRankItem) SetByRedisValue(rank uint64, member string, score float64) error {
	//acid:playername
	acids := strings.Split(member, "|")
	if len(acids) != 2 {
		return fmt.Errorf("HeroGachaRankItem.SetByRedisValue split member failed: %s", member)
	}

	acid := acids[0]
	i.Member = HGRankMember{
		AccountID:  acid,
		PlayerName: acids[1],
	}
	i.PlayerName = acids[1]

	if account, err := db.ParseAccount(acid); err != nil {
		return err
	} else {
		i.sid = account.ShardId
	}

	//score 先来的占据高分位 100.(1.0/201610101212600) 这里只需要取得 整数部分,全面舍去小数
	i.Score = uint64(math.Floor(score))
	i.Rank = rank
	return nil
}

type HeroGachaRace struct {
	sid uint

	locker   sync.RWMutex
	Items    [MAXRANK]HeroGachaRankItem
	NumItems int
	//最小Score是多少, 根据自己是否超过最小Score,决定是否应该插入数据库
	//最小Score只增不减,
	// 当本次插入的值超过当前最小Score,则需要立即从数据库更新排行
	// 当插入的新分数改变了自己的排名,则需要立即从数据库更新排行
	MinScore uint64

	redisCfg    RedisDBSetting
	redis       redis.Conn
	BalanceChan <-chan time.Time
	ResetChan   <-chan time.Time
	CheatChan   chan struct{}

	curActivity *HGRActivity

	starterWg sync.WaitGroup
	stopChan  chan bool

	getScoreTS int64
}

func NewHeroGachaRace(sid uint) *HeroGachaRace {
	m := new(HeroGachaRace)
	m.sid = sid
	return m
}

func (hgr *HeroGachaRace) Start() {
	if !game.Cfg.IsRunModeDev() {
		defaultm, err := getEtcdCfg()
		if err != nil {
			panic("HeroGachaRace Start" + err.Error())
		}
		if _, ok := defaultm["default"]; !ok {
			panic("HeroGachaRace do not found etcd key, eg./a6k/200/HeroGachaRace.")
		}
	}

	// init
	actId := gamedata.GetHGRCurrValidActivityId()
	if actId > 0 {
		cfg := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(uint32(actId))
		hgr.InitCurActivity(&HGRActivity{
			GroupID:    gamedata.GetHotDatas().Activity.GetShardGroup(uint32(hgr.sid)),
			ActivityId: uint32(actId),
			StartTime:  cfg.StartTime,
			EndTime:    cfg.EndTime,
		})
	}
	logs.Debug("HGR config: %v, %v", hgr.curActivity, hgr.redisCfg)
}

func (hgr *HeroGachaRace) Reset() {
	logs.Debug("HeroGachaRace over %d", hgr.curActivity)
	hgr.MinScore = 0
	hgr.redisCfg = RedisDBSetting{}
	hgr.redis = nil
	hgr.curActivity = nil
	hgr.BalanceChan = nil
	hgr.CheatChan = nil
	hgr.NumItems = 0
	hgr.Items = [MAXRANK]HeroGachaRankItem{}
	hgr.getScoreTS = 0
}

func (hgr *HeroGachaRace) AfterStart(g *gin.Engine) {

}

func (hgr *HeroGachaRace) BeforeStop() {

}

func (hgr *HeroGachaRace) Stop() {
	if hgr.stopChan != nil {
		close(hgr.stopChan)
	}
	hgr.starterWg.Wait()
}
