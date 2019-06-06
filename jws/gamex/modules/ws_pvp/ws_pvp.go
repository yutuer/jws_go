package ws_pvp

import (
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type WSPVPModule struct {
	sid             uint
	groupId         int
	topN            []*WsPvpRankPlayer // 排名排行榜
	topNByBest9     []*WsPvpRankPlayer // 最强9人战力排行榜
	saveQueue       []*WSPVPInfo
	waitter         util.WaitGroupWrapper
	readChan        chan wspvpReadCommand
	reloadChan      chan bool
	readQuitChan    chan struct{}
	dbQuitChan      chan struct{}
	saveChan        chan *WSPVPInfo
	lastUpdateClock int
	isUpdate        bool
}

// 角色从module里面拉取数据的相关结构
type wspvpReadCommand struct {
	commandType int
	topN        []*WsPvpRankPlayer
	resChan     chan<- wspvpReadResCommand // 获取用的回复channel
}

type wspvpReadResCommand struct {
	TopN []*WsPvpRankPlayer
}

// 定时从数据库更新相关结构

func (w *WSPVPModule) Start() {
	w.readChan = make(chan wspvpReadCommand)
	w.readQuitChan = make(chan struct{})
	w.dbQuitChan = make(chan struct{})
	w.reloadChan = make(chan bool)
	w.saveChan = make(chan *WSPVPInfo, 100)
	w.initRank(new(WsInitRobot))
	w.startMemIOGoRoutine() // 内存goroutine
	w.startDBIOGoRoutine()  // 数据库goroutine
}

func (*WSPVPModule) AfterStart(g *gin.Engine) {

}

func (*WSPVPModule) BeforeStop() {

}

func (w *WSPVPModule) Stop() {
	logs.Debug("wspvp module stop")
	w.readQuitChan <- struct{}{}
	w.dbQuitChan <- struct{}{}
	w.waitter.Wait()
}

type InitRobotInterface interface {
	InitRobt(sid uint32)
	InitTopN(module *WSPVPModule)
}

type WsInitRobot struct {
}

func (wi *WsInitRobot) InitRobt(sid uint32) {
	InitRobot(sid)
}

func (wi *WsInitRobot) InitTopN(module *WSPVPModule) {
	module.initTopN()
}

func (w *WSPVPModule) initRank(ir InitRobotInterface) {
	count := getRankSize(w.groupId)
	if count > 0 {
		w.loadTopNByInit(ir)
		logs.Info("load wspvp rank OK 1")
		return
	}

	if IsInitRobot(w.groupId) {
		ir.InitRobt(uint32(w.sid))
		ir.InitTopN(w)
		logs.Info("load wspvp rank OK 2")
		return
	} else {
		w.loadTopNByInit(ir)
		logs.Info("load wspvp rank OK 3")
	}

}

func (w *WSPVPModule) loadTopNByInit(ir InitRobotInterface) {
	timer := time.After(time.Second)
	checkTime := 0
	for {
		<-timer
		checkTime++
		count := getRankSize(w.groupId)
		//logs.Debug("sid=%d, count=%d", w.sid, count)
		if count == WS_PVP_RANK_MAX {
			ir.InitTopN(w)
			break
		} else {
			timer = time.After(time.Second)
		}
		if checkTime > 30 {
			panic(fmt.Sprintf("Init robot takes more than 30s count=%d, groupId=%d", count, w.groupId))
		}
	}
}

// 玩家从该GoRoutine读取排行榜信息
func (w *WSPVPModule) startMemIOGoRoutine() {
	w.waitter.Wrap(func() {
		defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
		for {
			select {
			case command, ok := <-w.readChan:
				if !ok {
					logs.Warn("rwspvp read chan close")
					return
				}
				func() {
					defer logs.PanicCatcherWithInfo("startMemIOGoRoutine command fatal error")
					switch command.commandType {
					case WSPVP_PLAYER_GET_TOPN:
						command.resChan <- wspvpReadResCommand{
							TopN: w.topN,
						}
					case WSPVP_UPDATE_TOPN:
						w.topN = command.topN
					case WSPVP_PLAYER_GET_BEST9_TOPN:
						command.resChan <- wspvpReadResCommand{
							TopN: w.topNByBest9,
						}
					case WSPVP_UPDATE_BEST9_TOPN:
						w.topNByBest9 = command.topN
					}
				}()
			case <-w.readQuitChan:
				logs.Warn("wspvp read goroutine quit")
				return
			}
		}
	})
}

// 定时从redis里面拉取新的排行榜信息
func (w *WSPVPModule) startDBIOGoRoutine() {
	w.waitter.Wrap(func() {
		defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
		resetTimer := time.After(time.Second * Wspvp_Time_Interval)
		saveTime := 0
		for {
			select {
			// 手动更新排行榜
			case _, ok := <-w.reloadChan:
				if !ok {
					logs.Warn("wspvp read chan close")
					return
				}
				w.reloadTopNFromRedis()
				logs.Debug("wspvp update top by trigger reload")
			// 定时更新排行榜 + 定时存储更新角色
			case _, ok := <-resetTimer:
				if !ok {
					logs.Warn("wspvp timer chan close")
					return
				}
				saveTime++
				if saveTime%3 == 0 {
					w.updateTopN()
					logs.Debug("wspvp update top by timer")
				} else if saveTime%3 == 1 {
					w.savePlayerInfo()
					logs.Debug("wspvp save player")
				} else {
					w.reloadBest9TopNFromRedis()
					logs.Debug("wspvp save player")
				}
				resetTimer = time.After(time.Second * Wspvp_Time_Interval)
			// 更新角色详细信息
			case cmd, ok := <-w.saveChan:
				if !ok {
					logs.Warn("wspvp save chan close")
					return
				}
				w.updateSaveCache(cmd)
				logs.Debug("wspvp update save player cache")
			// 退出
			case _, ok := <-w.dbQuitChan:
				if !ok {
					logs.Warn("wspvp save chan close")
					return
				}
				logs.Warn("wspvp db goroutine quit")
				w.savePlayerInfo()
				return
			}
		}
	})
}

func (w *WSPVPModule) UpdateTitle(topN []*WsPvpRankPlayer) {
	var rankAcids []string
	for i, rank := range topN {
		if i < 10 {
			rankAcids = append(rankAcids, rank.Acid)
		}
	}

	title_rank.GetModule(w.sid).SetWuShuangRank(rankAcids)
}

func (w *WSPVPModule) reloadTopNFromRedis() []*WsPvpRankPlayer {
	defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
	newTopN := loadTopN(w.groupId)
	w.readChan <- wspvpReadCommand{
		commandType: WSPVP_UPDATE_TOPN,
		topN:        newTopN,
	}
	logs.Debug("update top n")
	return newTopN
}

func (w *WSPVPModule) reloadBest9TopNFromRedis() []*WsPvpRankPlayer {
	defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
	newTopN := loadBest9TopN(w.groupId)
	w.readChan <- wspvpReadCommand{
		commandType: WSPVP_UPDATE_BEST9_TOPN,
		topN:        newTopN,
	}
	logs.Debug("update top n")
	return newTopN
}

func (w *WSPVPModule) updateSaveCache(info *WSPVPInfo) {
	defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
	for i, player := range w.saveQueue {
		if player.Acid == info.Acid {
			w.saveQueue[i] = info
			return
		}
	}
	w.saveQueue = append(w.saveQueue, info)
}

func (w *WSPVPModule) savePlayerInfo() {
	defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
	BatchSavePlayerInfo(w.groupId, w.saveQueue)
	w.saveQueue = make([]*WSPVPInfo, 0)
}

func (w *WSPVPModule) initTopN() {
	w.topN = loadTopN(w.groupId)
}

func (w *WSPVPModule) updateTopN() {
	defer logs.PanicCatcherWithInfo("startDBIOGoRoutine command fatal error")
	hour, min, _ := util.Clock(time.Now())
	nowMin := hour*60 + min
	// 锁定期间只用更新一次排行榜
	if gamedata.IsRankInWspvpRange(nowMin) {
		if !w.isUpdate {
			topN := w.reloadTopNFromRedis()
			w.isUpdate = true
			w.UpdateTitle(topN)
		}
	} else {
		w.isUpdate = false
		w.reloadTopNFromRedis()
	}
}
