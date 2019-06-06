package gvg

import (
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	mulutil "vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

type GVGGamesManager struct {
	waitter sync.WaitGroup
	games   *util.LockMap

	quitChan    chan bool
	tWheel      *timingwheel.TimingWheel
	lastRegTime int64
	rd          util.Kiss64Rng
}

func (g *GVGGamesManager) Start(ginEngine *gin.Engine) {
	g.games = util.NewLockMap(1024 * 4)

	g.quitChan = make(chan bool, 1)
	//30 seconds long, precision is 10 ms
	g.tWheel = mulutil.GetQuickTimeWheel()

	g.rd.Seed(time.Now().Unix())

	g.registerToEtcd()

}

func (g *GVGGamesManager) Stop() {

	close(g.quitChan)
	postService.UnRegService(
		//etcdRoot
		postService.GetByGID(
			postService.GVGServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.GVGToken,
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr)
}

func (g *GVGGamesManager) registerToEtcd() {
	postService.RegService(
		//etcdRoot
		postService.GetByGID(
			postService.GVGServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.GVGToken,
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr,
		//serviceUrl which others use to find me
		fmt.Sprintf("http://%s%s",
			config.Cfg.ListenNotifyAddr,
			helper.OnGVGSuccessPostUrl),
		g.games.GetLen(), 0)
}

func (g *GVGGamesManager) loop() {
	nowT := time.Now().Unix()
	if nowT-g.lastRegTime > helper.RegTimePreSeconds {
		g.lastRegTime = nowT
		g.registerToEtcd()
	}
}

func (g *GVGGamesManager) StartHttp(ginEngine *gin.Engine) {
	ginEngine.POST(helper.OnGVGSuccessPostUrl,
		func(c *gin.Context) {
			s := helper.GVGStartFightData{}
			err := c.Bind(&s)
			if err != nil {
				logs.Error("c.bind err: %v", err.Error())
				c.String(400, err.Error())
				return
			}
			logs.Debug("get post req from cs service, info: %v", s)

			gameID := uuid.NewV4().String()
			err = g.GVGCreateGame(gameID, false, &s)

			if err == nil {
				url := fmt.Sprintf("ws://%s/gvg", multConfig.Cfg.PublicIP)
				//大数据埋点
				//avaId := make([]int, 0)
				//vip := make([]uint32, 0)
				//compressGS := make([]int, 0)
				//for _, info := range s.Info {
				//	avaId = append(avaId, info.AvatarId)
				//	vip = append(vip, info.VipLv)
				//	compressGS = append(compressGS, info.Gs)
				//}
				//isTick := false
				//if s.BoxStatus != 0 {
				//	isTick = true
				//}
				//r := logiclog2.LogicInfo_GVGossBattleStart{
				//	BattleID:      gameID,
				//	AccounID:      s.AcID,
				//	AvatarID:      avaId,
				//	CompressGS:    compressGS,
				//	BossID:        s.BossID,
				//	VIP:           vip,
				//	IsTickRedBox:  isTick,
				//	WhoTickRedBox: s.CostID,
				//}
				//TypeInfo := logiclog2.LogicTag_GVGossBattleStart
				//format := logiclog2.BITag
				//logiclog.MultiInfo(TypeInfo, r, format)
				fc := helper.GVGStartFigntRetData{
					WebsktUrl: url,
					RoomID:    gameID,
				}
				c.JSON(200, fc)
			} else {
				logs.Error("GVGCreateGame err %v %s",
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

// CreateGame 外部接口调用，产生创建游戏的消息GVGGameManagerCommandMsgCreateGame
func (g *GVGGamesManager) GVGCreateGame(gameID string, isHard bool, data *helper.GVGStartFightData) error {
	logs.Info("OnMatchSuccessPostUrl game %s, hard %v",
		gameID, isHard)

	n := NewGVGGame(gameID, uint(data.Sid))
	//n.SceneID = data.SceneID
	//n.GroupID = data.GroupID
	//n.BossID = data.BossID
	//n.Level = data.Level
	//n.BoxStatus = data.BoxStatus
	//n.CostID = data.CostID
	//stageData := gamedata.GetStageData(n.SceneID)
	//if stageData == nil {
	//	return fmt.Errorf("gamedata err, no stage info")
	//}

	n.Stat.Init(
		data,
		0, 0,
		1,
		g.rd.Int63())
	n.Stat.Rng = g.rd
	n.AcIDs = []string{data.Acid1, data.Acid2}
	g.games.Set(gameID, n)

	return n.Start(data)
}

// GetGame 外部接口调用，获取当前游戏列表的消息GVGGameManagerCommandMsgGetGame
func (g *GVGGamesManager) GVGGetGame(gameID string, s string) *GVGGame {
	game, ok := g.games.Get(gameID)
	if ok {
		return game.(*GVGGame)
	}
	return nil
}

// GetGame 外部接口调用，获取当前游戏列表的消息GVGGameManagerCommandMsgGetGame
func (g *GVGGamesManager) GVGGameOver(gameID string) {
	_, ok := g.games.Get(gameID)
	if ok {
		g.games.Delete(gameID)
	}
}

var GVGGamesMgr GVGGamesManager
