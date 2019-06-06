package csrob

import (
	"fmt"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type rankGuild struct {
	trigQueue  chan string
	preMap     map[string]uint32
	batchStr   string
	batchStamp int64

	close chan struct{}
	sync.WaitGroup

	res *resources

	mWeekRewardTime    int64
	mWeekRewardShardID uint
	mWeekReward        map[string]bool
	mWeekRewardLock    sync.RWMutex

	ticker     *time.Ticker
	tickerList []*tickerElem
}

func newRankGuild(res *resources) *rankGuild {
	return &rankGuild{
		trigQueue: make(chan string, 256),
		preMap:    make(map[string]uint32),

		close: make(chan struct{}, 1),
		res:   res,

		mWeekReward: map[string]bool{},
	}
}

func (r *rankGuild) Start() {
	go func() {
		r.Add(1)
		r.doRank()
		r.Done()
	}()
	logs.Info("[CSRob] rankGuild Start")
}

func (r *rankGuild) Stop() {
	r.close <- struct{}{}
	r.Wait()
	close(r.close)
	close(r.trigQueue)
	logs.Info("[CSRob] rankGuild Stop")
}

func (r *rankGuild) doRank() {
	r.tickerList = []*tickerElem{
		&tickerElem{maxInterval: 1, currInterval: 1, action: r.updateBatch},
		&tickerElem{maxInterval: 5, currInterval: 5, action: r.tickerDoGuildWeekReward},
	}

	r.ticker = time.NewTicker(intervalCommonTicker)
	after := time.NewTicker(delayRankDo)

	r.doTickerList(time.Now())

	bClose := false
	for !bClose {
		select {
		case <-r.close:
			bClose = true
		case trig := <-r.trigQueue:
			logs.Debug("[CSRob] receive trig")
			r.preMap[trig] = r.preMap[trig] + 1
		case <-after.C:
			func() {
				defer logs.PanicCatcherWithInfo("[CSRob] rankGuild doRank refreshTriggers Panic")
				r.refreshTriggers()
			}()
		case now := <-r.ticker.C:
			logs.Debug("[CSRob] doCommon, now [%s](%d)", now.String(), now.Unix())
			r.doTickerList(now)
		}
	}

	logs.Info("[CSRob] doRank close")
}

type tickerElem struct {
	maxInterval  int
	currInterval int
	action       func(time.Time, *tickerElem)
	lastAct      int64
}

func (r *rankGuild) doTickerList(now time.Time) {
	defer logs.PanicCatcherWithInfo("[CSRob] rankGuild doTickerList Panic")

	for _, elem := range r.tickerList {
		elem.currInterval++
		if elem.currInterval < elem.maxInterval {
			continue
		}
		elem.currInterval = 0
		elem.action(now, elem)
	}
}

func (r *rankGuild) refreshTriggers() {
	for guid, count := range r.preMap {
		logs.Debug("[CSRob] refreshTriggers guild [%s], count [%d]", guid, count)

		if true == r.res.poolName.GetGuildCache(guid).Dismissed {
			logs.Warn("[CSRob] refreshTriggers guild [%s] is dismissed", guid)
			continue
		}

		count, robTime, err := r.res.GuildDB.getRobTimes(guid, r.batchStr)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}

		if err := r.pushToRobRank(guid, count, robTime); nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}
	}

	r.preMap = map[string]uint32{}
}

func (r *rankGuild) addTrig(guid string) {
	r.trigQueue <- guid
}

func (r *rankGuild) pushToRobRank(guid string, count uint32, robTime int64) error {
	return r.res.GuildDB.pushGuildToRobRank(guid, count, robTime, r.batchStr)
}

func (r *rankGuild) removeFromRobRank(guid string) error {
	return r.res.GuildDB.removeGuildFromRobRank(guid, r.batchStr)
}

func (r *rankGuild) rangeFromRobRank(num int) ([]GuildRankElem, error) {
	return r.res.GuildDB.rangeFromRobRank(num, r.batchStr)
}

func (r *rankGuild) getRankFromRobRank(guid string) (uint32, error) {
	return r.res.GuildDB.getRankFromRobRank(guid, r.batchStr)
}

func (r *rankGuild) tickerDoGuildWeekReward(now time.Time, e *tickerElem) {
	lastRewardTime := r.getLastRankRewardTime(now)
	if e.lastAct <= lastRewardTime && lastRewardTime <= now.Unix() {
		r.res.CommandMod.notifyRewardGuildWeek(lastRewardTime)
		e.lastAct = now.Unix()
	}
}

func (r *rankGuild) updateBatch(now time.Time, e *tickerElem) {
	nst := r.getNextRankStartTime(now)
	if nst == r.batchStamp {
		return
	}
	e.lastAct = now.Unix()

	r.batchStamp = nst
	bt := time.Unix(r.batchStamp, 0).In(util.ServerTimeLocal)
	logs.Debug("[CSRob] updateBatch, with [%s](%d)", bt.String(), bt.Unix())
	if game.Cfg.IsRunModeProd() {
		r.batchStr = fmt.Sprintf("%04d%02d%02d", bt.Year(), int(bt.Month()), bt.Day())
	} else {
		r.batchStr = fmt.Sprintf("%04d%02d%02d%02d%02d", bt.Year(), int(bt.Month()), bt.Day(), bt.Hour(), bt.Minute())
	}
}

func (r *rankGuild) getLastRankRewardTime(now time.Time) int64 {
	weekEnd := gamedata.CSRobThisWeekRankTime(now)
	if weekEnd > now.Unix() {
		weekEnd -= util.WeekSec
	}
	logs.Debug("[CSRob] getLastRankRewardTime, [%d]", weekEnd)
	return weekEnd
}

func (r *rankGuild) getNextRankStartTime(now time.Time) int64 {
	weekEnd := gamedata.CSRobNextWeekStartTime(now)
	if weekEnd < now.Unix() {
		weekEnd += util.WeekSec
	}
	logs.Debug("[CSRob] getNextRankStartTime, [%d]", weekEnd)
	return weekEnd
}

func (r *rankGuild) loadWeekReward() {
	reward, err := r.res.RewardDB.getRewardWeek()
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] rankGuild loadWeekReward getRewardWeek failed, %v", err))
		return
	}
	logs.Debug("[CSRob] rankGuild loadWeekReward getRewardWeek [%v]", reward)
	if nil == reward {
		return
	}

	r.mWeekRewardTime = reward.Time
	r.mWeekRewardShardID = reward.Sid
	r.addWeekReward(reward.Members)
}

func (r *rankGuild) checkInWeekReward(acid string) bool {
	if r.res.sid != r.mWeekRewardShardID {
		return false
	}
	r.mWeekRewardLock.RLock()
	defer r.mWeekRewardLock.RUnlock()
	return r.mWeekReward[acid]
}

func (r *rankGuild) clearWeekReward() {
	r.mWeekRewardLock.Lock()
	defer r.mWeekRewardLock.Unlock()
	r.mWeekReward = map[string]bool{}
}

func (r *rankGuild) addWeekReward(list []string) {
	r.mWeekRewardLock.Lock()
	defer r.mWeekRewardLock.Unlock()
	for _, acid := range list {
		r.mWeekReward[acid] = true
	}
}
func (r *rankGuild) getWeekRewardList() []string {
	r.mWeekRewardLock.RLock()
	defer r.mWeekRewardLock.RUnlock()
	list := []string{}

	for acid, b := range r.mWeekReward {
		if true == b {
			list = append(list, acid)
		}
	}

	return list
}

func (r *rankGuild) CheckMeHasTitle(acid string) (bool, int64) {
	if r.mWeekRewardShardID != r.res.sid {
		return false, r.mWeekRewardTime
	}
	return r.checkInWeekReward(acid), r.mWeekRewardTime
}

//----debug:cheat
func (r *rankGuild) DebugDoWeekReward() {
	logs.Debug("[CSRob] DebugDoWeekReward start")

	status, err := r.res.GuildDB.getCommonStatus()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}

	status.WeekRewardTime = 0
	if err := r.res.GuildDB.setCommonStatus(status); nil != err {
		logs.Error(fmt.Sprintf("[CSRob] DebugDoWeekReward setCommonStatus failed, %v", err))
		return
	}

	r.res.CommandMod.notifyRewardGuildWeek(r.getLastRankRewardTime(time.Now()))
	logs.Debug("[CSRob] DebugDoWeekReward end")
}
