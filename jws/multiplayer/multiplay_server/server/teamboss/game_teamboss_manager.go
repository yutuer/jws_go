package teamboss

import (
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"fmt"

	"github.com/gin-gonic/gin"
	gm "github.com/rcrowley/go-metrics"
	tb_helper "vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	logiclog2 "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/logiclog"
	mulutil "vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

type TBGamesManager struct {
	waitter sync.WaitGroup
	games   *util.LockMap

	quitChan    chan bool
	tWheel      *timingwheel.TimingWheel
	lastRegTime int64
	rd          util.Kiss64Rng
	counter     gm.Counter
}

func (g *TBGamesManager) Start(ginEngine *gin.Engine) {
	g.games = util.NewLockMap(1024 * 4)

	g.quitChan = make(chan bool, 1)
	//30 seconds long, precision is 10 ms
	g.tWheel = mulutil.GetQuickTimeWheel()

	g.rd.Seed(time.Now().Unix())

	g.registerToEtcd()
	g.counter = metrics.NewCounter("teamboss_c")

}

func (g *TBGamesManager) Stop() {
	close(g.quitChan)
	postService.UnRegService(
		//etcdRoot
		postService.GetByGID(
			postService.TBServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.TeamBossToken,
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr)
}

func (g *TBGamesManager) registerToEtcd() {
	postService.RegService(
		//etcdRoot
		postService.GetByGID(
			postService.TBServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.TeamBossToken,
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr,
		//serviceUrl which others use to find me
		fmt.Sprintf("http://%s%s",
			config.Cfg.ListenNotifyAddr,
			helper.OnTBSuccessPostUrl),
		g.games.GetLen(), 0)
}

func (g *TBGamesManager) loop() {
	nowT := time.Now().Unix()
	if nowT-g.lastRegTime > helper.RegTimePreSeconds {
		g.lastRegTime = nowT
		g.registerToEtcd()
	}
}

func (g *TBGamesManager) StartHttp(ginEngine *gin.Engine) {
	ginEngine.POST(helper.OnTBSuccessPostUrl,
		func(c *gin.Context) {
			s := helper.TBStartFightData{}
			err := c.Bind(&s)
			if err != nil {
				logs.Error("c.bind err: %v", err.Error())
				c.String(400, err.Error())
				return
			}
			logs.Debug("get post req from cs service, info: %v", s)

			if err := s.Init(); err != nil {
				c.String(400, err.Error())
				return
			}
			gameID := tb_helper.RoomID2Global(s.RoomID, s.GID, s.GroupID, time.Now().Unix())
			err = g.TBCreateGame(gameID, false, &s)

			if err == nil {
				url := fmt.Sprintf("ws://%s/teamboss", multConfig.Cfg.PublicIP)

				fc := helper.TeamBossCreateinfo{
					WebsktUrl:    url,
					GlobalRoomID: gameID,
				}
				//大数据埋点
				avaId := make([]int, 0)
				vip := make([]uint32, 0)
				compressGS := make([]int, 0)
				for _, info := range s.Info {
					avaId = append(avaId, info.AvatarId)
					vip = append(vip, info.VipLv)
					compressGS = append(compressGS, info.Gs)
				}
				isTick := false
				if s.BoxStatus != 0 {
					isTick = true
				}
				r := logiclog2.LogicInfo_TBossBattleStart{
					BattleID:      gameID,
					AccounID:      s.AcID,
					AvatarID:      avaId,
					CompressGS:    compressGS,
					BossID:        s.BossID,
					VIP:           vip,
					IsTickRedBox:  isTick,
					WhoTickRedBox: s.CostID,
				}
				TypeInfo := logiclog2.LogicTag_TBossBattleStart
				format := logiclog2.BITag
				logiclog.MultiInfo(TypeInfo, r, format)
				c.JSON(200, fc)
			} else {
				logs.Error("TBCreateGame err %v %s",
					false, err.Error())
				c.String(400, err.Error())
			}

		})
	//go func() {
	//	err := ginEngine.Run(config.Cfg.ListenNotifyAddr)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
}

// CreateGame 外部接口调用，产生创建游戏的消息TBGameManagerCommandMsgCreateGame
func (g *TBGamesManager) TBCreateGame(gameID string, isHard bool, data *helper.TBStartFightData) error {
	logs.Info("OnMatchSuccessPostUrl game %s, hard %v",
		gameID, isHard)

	n := NewTBGame(gameID, uint(data.GID))
	n.SceneID = data.SceneID
	n.GroupID = data.GroupID
	n.BossID = data.BossID
	n.Level = data.Level
	n.BoxStatus = data.BoxStatus
	n.CostID = data.CostID
	stageData := gamedata.GetStageData(n.SceneID)
	if stageData == nil {
		return fmt.Errorf("gamedata err, no stage info")
	}

	n.Stat.Init(
		isHard,
		0, 0,
		data.AcID,
		1,
		g.rd.Int63())
	n.Stat.Rng = g.rd
	n.AcIDs = data.AcID
	g.games.Set(gameID, n)
	g.counter.Inc(1)
	return n.Start(data)
}

// GetGame 外部接口调用，获取当前游戏列表的消息TBGameManagerCommandMsgGetGame
func (g *TBGamesManager) TBGetGame(gameID string, s string) *TBGame {
	game, ok := g.games.Get(gameID)
	if ok {
		return game.(*TBGame)
	}
	return nil
}

// GetGame 外部接口调用，获取当前游戏列表的消息TBGameManagerCommandMsgGetGame
func (g *TBGamesManager) TBGameOver(gameID string) {
	_, ok := g.games.Get(gameID)
	if ok {
		g.games.Delete(gameID)
		g.counter.Dec(1)

	}
}

var TBGamesMgr TBGamesManager
