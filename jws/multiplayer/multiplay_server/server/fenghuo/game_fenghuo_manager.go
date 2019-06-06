package fenghuo

import (
	"errors"
	"runtime/debug"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"fmt"

	"sync/atomic"

	"github.com/gin-gonic/gin"
	modelshelper "vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	mulutil "vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

type FenghuoGamesManager struct {
	waitter sync.WaitGroup
	gamesmu sync.RWMutex
	games   map[string]*FenghuoGame

	quitChan    chan bool
	tWheel      *timingwheel.TimingWheel
	lastRegTime int64
	onlyOneLoop int32
	rd          util.Kiss64Rng
}

func (g *FenghuoGamesManager) Start(ginEngine *gin.Engine) {
	g.waitter.Add(1)
	defer g.waitter.Done()
	g.games = make(map[string]*FenghuoGame, 1024)

	g.quitChan = make(chan bool, 1)
	//30 seconds long, precision is 10 ms
	g.tWheel = mulutil.GetQuickTimeWheel()

	g.rd.Seed(time.Now().Unix())

	g.registerToEtcd()

	g.waitter.Add(1)
	func() {
		defer g.waitter.Done()
		defer func() {
			if err := recover(); err != nil {
				logs.Error("GVEGamesManager Panic, session Err %v", err)
				debug.PrintStack()
			}
		}()
		timerChan := g.tWheel.After(time.Second)
		for {
			select {
			case <-timerChan:
				timerChan = g.tWheel.After(time.Second)
				g.waitter.Add(1)
				func() {
					defer g.waitter.Done()
					g.loop()
				}()
			case <-g.quitChan:
				return
			}
		}
		logs.Trace("GVEGamesManager Stoped")
	}()
	g.waitter.Wait()
}

func (g *FenghuoGamesManager) Stop() {
	close(g.quitChan)
}

func (g *FenghuoGamesManager) registerToEtcd() {
	g.gamesmu.RLock()
	lgames := len(g.games)
	g.gamesmu.RUnlock()

	postService.RegService(
		//etcdRoot
		postService.GetByGID(
			postService.FenghuoServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			helper.FmtMatchToken(multConfig.Cfg.MatchToken),
		),
		//serviceID use ip
		config.Cfg.ListenNotifyAddr,
		//serviceUrl which others use to find me
		fmt.Sprintf("http://%s%s",
			config.Cfg.ListenNotifyAddr,
			helper.OnFenghuoSuccessPostUrl),
		lgames, helper.RegTimePreSeconds*10*time.Second)
}

func (g *FenghuoGamesManager) loop() {
	nowT := time.Now().Unix()
	if nowT-g.lastRegTime > helper.RegTimePreSeconds {
		g.lastRegTime = nowT
		g.registerToEtcd()

		if atomic.CompareAndSwapInt32(&g.onlyOneLoop, 0, 1) {
			go func() {
				defer func() {
					//reset all back
					atomic.StoreInt32(&g.onlyOneLoop, 0)
				}()

				g.gamesmu.Lock()
				defer g.gamesmu.Unlock()
				for k, game := range g.games {
					if game.AllSubLevelDone() {
						game.GameOver()
						delete(g.games, k)
					}
				}

				logs.Debug("FenghuoGamesManager.loop finished.")
			}()
		} else {
			logs.Debug("FenghuoGamesManager.loop but last one not finished.")
		}
	}

}

func (g *FenghuoGamesManager) StartHttp(ginEngine *gin.Engine) {
	ginEngine.POST("/fenghuotest", func(c *gin.Context) {
		c.JSON(200, gin.H{"ret": fmt.Sprintf("http://%s%s", multConfig.Cfg.ListenNotifyAddr, helper.OnFenghuoSuccessPostUrl)})
	})
	ginEngine.POST(helper.OnFenghuoSuccessPostUrl,
		func(c *gin.Context) {
			s := helper.FenghuoValue{}
			err := c.Bind(&s)

			if err != nil {
				c.String(400, err.Error())
				return
			}
			if s.Shutdown {
				//退出房间
				//TODO by YZH secret 还没有实现
				fg := g.GetGame(s.RoomID, "")
				if fg != nil {
					fg.ForceOver()
				}
				//ok := g.GameOver(s.GlobalRoomID)
				logs.Info("OnFenghuoSuccessPostUrl Destroy Room ID %s", s.RoomID)
				c.JSON(200, gin.H{"ret": true})
				logs.Info("Fenghuo Multiplay Rooms %d", g.GameNum())
				return
			} else {
				if len(s.AcIDs[:]) != fenghuoMaxPlayer {
					logs.Error("OnFenghuoSuccessPostUrl acids is not right %v", s.AcIDs)
					c.String(400, err.Error())
					return
				}
				logs.Info("OnFenghuoSuccessPostUrl %v, avatars:%v", s.AcIDs, len(s.AvatarInfo))
				roomid := uuid.NewV4().String()
				err = g.CreateGame(roomid, s.AcIDs[:], s.AvatarInfo[:])
				if err == nil {
					url := fmt.Sprintf("ws://%s/wsfenghuo", multConfig.Cfg.PublicIP)

					fc := helper.FenghuoCreateInfo{
						WebsktUrl: url,
						RoomID:    roomid,
						CancelUrl: fmt.Sprintf("http://%s%s", multConfig.Cfg.ListenNotifyAddr, helper.OnFenghuoSuccessPostUrl),
					}

					c.JSON(200, fc)
					logs.Info("Fenghuo Multiplay Rooms %d", g.GameNum())
					return
				}
			}

			c.String(401, err.Error())

		})

}

//FIXME by YZH 如果创建后, 某个时间长度后,都没有链接上来,则主动关闭释放资源。
func (g *FenghuoGamesManager) CreateGame(gameID string, acID []string, avatars []*modelshelper.Avatar2ClientByJson) error {
	if len(acID) != fenghuoMaxPlayer || len(avatars) != fenghuoMaxPlayer {
		return errors.New("Only 2 players allowed.")
	}

	n := NewFenghuoGame(gameID)
	for i := range n.AcIDs {
		n.AcIDs[i] = acID[i]
		n.Avatars[i] = avatars[i]
	}

	g.gamesmu.Lock()
	defer g.gamesmu.Unlock()
	g.games[gameID] = n
	logs.Trace("g.games %v", g.games)
	return n.Start()
}

func (g *FenghuoGamesManager) GetGame(gameID string, secret string) *FenghuoGame {
	g.gamesmu.RLock()
	defer g.gamesmu.RUnlock()
	game, ok := g.games[gameID]
	if ok {
		return game
	}
	return nil
}

//func (g *FenghuoGamesManager) GameOver(gameID string) bool {
//	g.gamesmu.Lock()
//	defer g.gamesmu.Unlock()
//	_, ok := g.games[gameID]
//	if ok {
//		delete(g.games, gameID)
//		return true
//	}
//	return false
//}

func (g *FenghuoGamesManager) GameNum() int {
	g.gamesmu.RLock()
	defer g.gamesmu.RUnlock()
	return len(g.games)
}

var FHGamesMgr FenghuoGamesManager
