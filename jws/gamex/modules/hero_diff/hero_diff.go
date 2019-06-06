package hero_diff

import (
	"github.com/gin-gonic/gin"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type heroDiffModule struct {
	sid          uint
	resetTimer   <-chan time.Time
	quitChan     chan struct{}
	heroDiffInfo heroDiffInfo
	waitter      util.WaitGroupWrapper
}

func (hd *heroDiffModule) Start() {
	hd.quitChan = make(chan struct{}, 1)
	hd.loadFromDB()
	hdi := &hd.heroDiffInfo
	nowT := time.Now().Unix()
	if hdi.LastResetTime == 0 {
		hdi.LastResetTime = nowT
	}
	resetTime := util.GetNextDailyWeekTime(gamedata.GetHeroDiffResetBeginSec(hdi.LastResetTime), hdi.LastResetTime)
	logs.Debug("herodiff resettime : %d", resetTime)
	timeInterval := resetTime - nowT
	if timeInterval < 0 {
		timeInterval = 0
	}
	hd.resetTimer = time.After(time.Second * time.Duration(timeInterval))
	hd.waitter.Wrap(func() {
		for {
			select {
			case <-hd.quitChan:
				logs.Debug("herodiff module quit chan")
				return
			case <-hd.resetTimer:
				nowT := time.Now().Unix()
				hdi.LastResetTime = nowT
				nextResetTime := util.GetNextDailyWeekTime(gamedata.GetHeroDiffResetBeginSec(nowT), nowT)
				resetDur := nextResetTime - nowT
				hd.resetTimer = time.After(time.Second * time.Duration(resetDur))
				logs.Debug("reset dur: %d, time: $d", resetDur, nextResetTime)
				hd.deleteRank()
			}
		}
	})
}

func (hd *heroDiffModule) AfterStart(g *gin.Engine) {

}

func (hd *heroDiffModule) BeforeStop() {

}

func (hd *heroDiffModule) Stop() {
	hd.quitChan <- struct{}{}
	hd.waitter.Wait()
	hd.saveToDB()
}
