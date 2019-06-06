package gve_notify

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"

	"vcs.taiyouxi.net/platform/planx/util/timingwheel"

	"sync"

	"encoding/json"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify/post_data"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	multiHelper "vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	cmdTypNull = iota
	cmdTypGameStart
	cmdTypGameStop
	cmdTypAccountData
)

func genModule(sid uint) *module {
	return &module{
		sid: sid,
	}
}

type cmd struct {
	Typ           int
	GameID        string
	GameSecret    string
	GameServerUrl string
	IsHasReward   bool
	IsSuccess     bool
	IsDouble      bool
	IsUseHc       bool
	Reward        []string
	Count         []uint32
	AccountID     string
	AccountData   *helper.Avatar2ClientByJson
	ResChan       chan cmd
	IsBot         bool
}

type module struct {
	sid            uint
	listenPostAddr string
	gveStartUrl    string
	gveStopUrl     string
	cmdChann       chan cmd
	wg             sync.WaitGroup
	quitChan       chan bool
	tWheel         *timingwheel.TimingWheel
	dataWaitting   map[string]chan cmd

	LastGveMatchBroadCastTime  int64 // gve广播上次的时间
	LastGveMatchBroadCastMutex sync.RWMutex
}

func (m *module) AfterStart(g *gin.Engine) {
	g.POST("/gamex/v1/api/user/gvestart", func(c *gin.Context) {
		s := multiHelper.GameStartInfo{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, err.Error())
			return
		}

		logs.Trace("gvestart %s, %s", s.AcIDs)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resChann := make(chan cmd, 1)
		m.cmdChann <- cmd{
			Typ:           cmdTypGameStart,
			GameID:        s.GameID,
			GameSecret:    s.Secret,
			GameServerUrl: s.MServerUrl,
			AccountID:     s.AcIDs,
			ResChan:       resChann,
			IsBot:         s.FightCount == 1, // 你只有一个人在战斗,所以你将会与两个机器人并肩战斗
		}

		select {
		case res := <-resChann:
			startGVEPostResData := post_data.StartGVEPostResData{
				Data:     *res.AccountData,
				IsDouble: res.IsDouble,
				IsUseHc:  res.IsUseHc,
				Reward:   res.Reward,
				Count:    res.Count,
			}

			if res.IsBot {
				robotData, err := GenGVERobotCompanion(m.sid, int64(res.AccountData.CorpGs), res.AccountID)
				if err != nil {
					logs.Debug("GenRobotCompanion error by %v", err)
				} else {
					startGVEPostResData.RobotData = robotData
				}
			}

			data, err := json.Marshal(startGVEPostResData)
			if err == nil {
				c.String(200, string(data))
			} else {
				c.String(401, err.Error())
			}
		case <-ctx.Done():
			// 主动发个空的data 取消掉dataWaitting中等待的chan
			m.cmdChann <- cmd{
				Typ:       cmdTypAccountData,
				GameID:    s.GameID,
				AccountID: s.AcIDs,
			}
			c.String(401, "TimeOut")
			return
		}
	})

	g.POST("/gamex/v1/api/user/gvestop", func(c *gin.Context) {
		s := multiHelper.GameStopInfo{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, err.Error())
			return
		}
		logs.Trace("gvestop %s, %s", s.AcIDs)
		m.cmdChann <- cmd{
			Typ:         cmdTypGameStop,
			GameID:      s.GameID,
			AccountID:   s.AcIDs,
			IsHasReward: s.IsHasReward,
			IsSuccess:   s.IsSuccess,
		}

		if err == nil {
			c.String(200, string("ok"))
		} else {
			c.String(401, err.Error())
		}
	})
}

func (m *module) BeforeStop() {
}

func (m *module) Start() {
	m.quitChan = make(chan bool, 1)
	m.tWheel = timingwheel.NewTimingWheel(time.Second, 300)
	m.cmdChann = make(chan cmd, 1024)
	m.dataWaitting = make(map[string]chan cmd, 1024)
	for i, sid := range game.Cfg.ShardId {
		if m.sid == uint(sid) {
			m.listenPostAddr = game.Cfg.ListenPostAddr[i]
			m.gveStartUrl = fmt.Sprintf("http://%s/%s", m.listenPostAddr, uutil.JwsCfg.GVEStartUrl)
			m.gveStopUrl = fmt.Sprintf("http://%s/%s", m.listenPostAddr, uutil.JwsCfg.GVEStopUrl)
			break
		}
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		//tinc := 10 * time.Second
		//timerChan := m.tWheel.After(tinc)
		postService.RegServices(m.gveStartUrl, m.gveStopUrl, 0)
		for {
			defer func() {
				if err := recover(); err != nil {
					logs.Error("GVE Panic, Err %v", err)
				}
			}()
			select {
			case command, _ := <-m.cmdChann:
				m.wg.Add(1)
				func() {
					defer m.wg.Done()
					m.processCmd(&command)
				}()
			//case <-timerChan:
			//	timerChan = m.tWheel.After(tinc)
			//	m.wg.Add(1)
			//	func() {
			//		defer m.wg.Done()
			//		postService.RegServices(m.gveStartUrl, m.gveStopUrl, tinc)
			//	}()
			case <-m.quitChan:
				return
			}
		}
	}()
}

func (m *module) processCmd(c *cmd) {
	switch c.Typ {
	case cmdTypGameStart:
		m.onGVEStart(c)
	case cmdTypGameStop:
		m.onGVEStop(c)
	case cmdTypAccountData:
		m.onAccountData(c)
	}
}

func (m *module) Stop() {
	m.quitChan <- true
	m.tWheel.Stop()
	m.wg.Wait()
	postService.UnRegServices()
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func SendAccountData(acID, gameID, gameUrl string,
	data *helper.Avatar2ClientByJson,
	rewards []string, count []uint32,
	isDouble, isUseHc, isBot bool) {

	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	a, _ := db.ParseAccount(acID)

	select {
	case GetModule(a.ShardId).cmdChann <- cmd{
		Typ:           cmdTypAccountData,
		AccountID:     acID,
		GameID:        gameID,
		GameServerUrl: gameUrl,
		AccountData:   data,
		Reward:        rewards,
		IsDouble:      isDouble,
		IsUseHc:       isUseHc,
		Count:         count,
		IsBot:         isBot,
	}:
	case <-ctx.Done():
		logs.Error("Gve SendAccountData put timeout")
	}
}

func (m *module) TryBroadCastMatchMsg() bool {
	t := time.Now().Unix()
	ct := int64(gamedata.GetCommonCfg().GetGVEADTime())
	m.LastGveMatchBroadCastMutex.RLock()
	if t-m.LastGveMatchBroadCastTime < ct {
		m.LastGveMatchBroadCastMutex.RUnlock()
		return false
	}
	m.LastGveMatchBroadCastMutex.RUnlock()
	m.LastGveMatchBroadCastMutex.Lock()
	defer m.LastGveMatchBroadCastMutex.Unlock()
	if t-m.LastGveMatchBroadCastTime < ct {
		return false
	}
	m.LastGveMatchBroadCastTime = t
	return true
}
