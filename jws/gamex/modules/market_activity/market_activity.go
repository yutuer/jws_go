package market_activity

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genMarketActivityModule(sid uint) *MarketActivityModule {
	ma := &MarketActivityModule{
		sid: sid,

		command_chan: make(chan *marketCommand, 1024),

		maRank: &MarketRank{},

		maTimeSet: &MarketTimeSet{},
	}

	ma.maTimeSet.ma = ma
	ma.maTimeSet.setStatus = make(map[uint32]MarketTimeStatus)

	ma.maRank.Init(ma)

	return ma
}

type MarketActivityModule struct {
	sid uint

	waitter util.WaitGroupWrapper

	tc chan bool //timer

	command_chan chan *marketCommand

	maRank    *MarketRank
	maTimeSet *MarketTimeSet
}

func (ma *MarketActivityModule) Start() {
	// external reg
	gamedata.AddHotDataNotify(modules.Module_MarketActivity, ma.NotifyHotDataUpdate)

	logs.Debug("[MarketActivityModule] Start")

	ma.maRank.ReloadAll()
	logs.Debug("[MarketActivityModule] Load Over")
	// goroutine for command queue
	ma.waitter.Wrap(func() {
		for cc := range ma.command_chan {
			logs.Debug("[MarketActivityModule], command [%v]", cc)

			func() {
				defer logs.PanicCatcherWithInfo("[MarketActivityModule] command process panic")
				ma.dispatch(cc)
			}()
		}
		logs.Warn("[MarketActivityModule] command_chan close")
	})
}

func (ma *MarketActivityModule) AfterStart(g *gin.Engine) {
	ma.maTimeSet.reloadHotData()
}

func (ma *MarketActivityModule) BeforeStop() {
	gamedata.DelHotDataNotify(modules.Module_MarketActivity)
}

func (ma *MarketActivityModule) Stop() {
	logs.Debug("[MarketActivityModule] Stop")

	close(ma.command_chan)
	ma.waitter.Wait()
}

func (ma *MarketActivityModule) commandChan() chan *marketCommand {
	return ma.command_chan
}

func (ma *MarketActivityModule) commandExecAsyn(cmd *marketCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case ma.commandChan() <- cmd:
	case <-ctx.Done():
		logs.Error("[commandExecAsyn] chann full, cmd put timeout[%s]", util.ASyncCmdTimeOut.String())
	}
}

// --- Debug For Cheat

func (ma *MarketActivityModule) DebugSnapShootAndReward(actType uint32, actID uint32) {
	ma.notifyMakeSnapShoot(actType, actID)
	ma.notifySendReward(actType, actID)
}

func (ma *MarketActivityModule) DebugClearSnapShoot(actType uint32, actID uint32) {
	ma.maRank.clear(actType)
	ma.maRank.db.debugClearSnap(actType, actID)

	ma.maRank.record.RewardRecord[fmt.Sprint(actID)] = 0
	ma.maRank.db.setRankRecord(&ma.maRank.record)
}
