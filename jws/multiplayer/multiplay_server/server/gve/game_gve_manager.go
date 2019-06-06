package gve

import (
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
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

type GVEGamesManager struct {
	waitter sync.WaitGroup
	games   *util.LockMap

	quitChan    chan bool
	tWheel      *timingwheel.TimingWheel
	lastRegTime int64
	rd          util.Kiss64Rng
}

func (g *GVEGamesManager) Start(ginEngine *gin.Engine) {
	g.games = util.NewLockMap(1024 * 4)

	g.quitChan = make(chan bool, 1)
	//30 seconds long, precision is 10 ms
	g.tWheel = mulutil.GetQuickTimeWheel()

	g.rd.Seed(time.Now().Unix())

	g.registerToEtcd()
}

func (g *GVEGamesManager) Stop() {
	close(g.quitChan)
	postService.UnRegService(
		//etcdRoot
		postService.GetByGID(
			postService.MultiplayServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.FmtMatchToken(multConfig.Cfg.MatchToken),
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr)
}

func (g *GVEGamesManager) registerToEtcd() {
	postService.RegService(
		//etcdRoot
		postService.GetByGID(
			postService.MultiplayServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.FmtMatchToken(multConfig.Cfg.MatchToken),
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr,
		//serviceUrl which others use to find me
		fmt.Sprintf("http://%s%s",
			config.Cfg.ListenNotifyAddr,
			helper.OnMatchSuccessPostUrl),
		g.games.GetLen(), 0)
}

func (g *GVEGamesManager) loop() {
	nowT := time.Now().Unix()
	if nowT-g.lastRegTime > helper.RegTimePreSeconds {
		g.lastRegTime = nowT
		g.registerToEtcd()
	}
}

func (g *GVEGamesManager) StartHttp(ginEngine *gin.Engine) {
	ginEngine.POST(helper.OnMatchSuccessPostUrl,
		func(c *gin.Context) {
			s := helper.MatchGameInfo{}
			err := c.Bind(&s)

			if err != nil {
				c.String(400, err.Error())
				return
			}

			err = g.GVECreateGame(uuid.NewV4().String(), s.IsHard, s.AcIDs[:])

			if err == nil {
				c.String(200, string("ok"))
			} else {
				logs.Error("TBCreateGame err %v %v %s",
					s.AcIDs, s.IsHard, err.Error())
				c.String(401, err.Error())
			}
		})
	//go func() {
	//	err := ginEngine.Run(config.Cfg.ListenNotifyAddr)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
}

// CreateGame 外部接口调用，产生创建游戏的消息GVEGameManagerCommandMsgCreateGame
func (g *GVEGamesManager) GVECreateGame(gameID string, isHard bool, acID []string) error {
	logs.Info("OnMatchSuccessPostUrl game %s acids %v, hard %v",
		gameID, acID, isHard)

	n := NewGVEGame(gameID)

	fightTime, _ := gamedata.GetGVEGameCfg()
	n.Stat.Init(
		isHard,
		time.Now().Unix(),
		time.Now().Unix()+fightTime*60+60, //多60秒缓冲
		acID,
		1,
		g.rd.Int63())
	n.Stat.Rng = g.rd

	n.AcIDs = acID
	g.games.Set(gameID, n)

	return n.Start()
}

// GetGame 外部接口调用，获取当前游戏列表的消息GVEGameManagerCommandMsgGetGame
func (g *GVEGamesManager) GVEGetGame(gameID string, s string) *GVEGame {
	game, ok := g.games.Get(gameID)
	if ok {
		return game.(*GVEGame)
	}
	return nil
}

// GetGame 外部接口调用，获取当前游戏列表的消息GVEGameManagerCommandMsgGetGame
func (g *GVEGamesManager) GVEGameOver(gameID string) {
	_, ok := g.games.Get(gameID)
	if ok {
		g.games.Delete(gameID)
	}
}

var GVEGamesMgr GVEGamesManager
