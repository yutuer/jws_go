package worldboss

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//RoundStatus ..
type RoundStatus struct {
	dirty           bool
	BatchTag        string
	LastResetTime   int64
	LastNewBossTime int64
	LastRewardTime  int64
}

type tickerHolder struct {
	res *resources

	tickerList []*tickerElem

	roundStatus *RoundStatus

	isClosed bool
}

func newTickerHolder(res *resources) *tickerHolder {
	t := &tickerHolder{}
	t.res = res
	t.tickerList = []*tickerElem{}
	t.roundStatus = &RoundStatus{}

	t.regTicker(180, 0, t.trySaveRoundToDB)
	t.regTicker(60, 0, t.tryResetNewRound)
	t.regTicker(60, 50, t.tryDamageRankReward)
	return t
}

//loadRoundFromDB ..
func (t *tickerHolder) loadRoundFromDB() error {
	now := time.Now()
	todayResetTime := gamedata.GetTodayWBResetTime(now)
	if now.Unix() >= todayResetTime {
		t.roundStatus.BatchTag = fmt.Sprintf("%04d%02d%02d", now.Year(), int(now.Month()), now.Day())
	} else {
		yesterday := now.Add(-24 * time.Hour)
		t.roundStatus.BatchTag = fmt.Sprintf("%04d%02d%02d", yesterday.Year(), int(yesterday.Month()), yesterday.Day())
	}

	round, err := t.res.BossDB.getRoundStatus()
	if nil != err {
		return fmt.Errorf("getRoundStatus failed, %v", err)
	}
	if nil != round {
		t.roundStatus.LastNewBossTime = round.LastNewBossTime
		t.roundStatus.LastResetTime = round.LastResetTime
		t.roundStatus.LastRewardTime = round.LastRewardTime
	}
	logs.Trace("[WorldBoss] tickerHolder loadRoundFromDB, round %+v", t.roundStatus)

	return nil
}

func (t *tickerHolder) start() {
	go t.holding()
}

func (t *tickerHolder) stop() {
	t.isClosed = true
}

//trySaveRoundToDB ..
func (t *tickerHolder) trySaveRoundToDB(now time.Time) {
	if true == t.roundStatus.dirty {
		if err := t.saveRoundToDB(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] tickerHolder trySaveRoundToDB, %v", err))
		}
		t.roundStatus.dirty = false
	}
}

func (t *tickerHolder) saveRoundToDB() error {
	if err := t.res.BossDB.setRoundStatus(*t.roundStatus); nil != err {
		return fmt.Errorf("setRoundStatus failed, %v", err)
	}
	logs.Trace("[WorldBoss] tickerHolder saveRoundToDB, setRoundStatus %+v", t.roundStatus)
	return nil
}

//tryResetNewRound ..
func (t *tickerHolder) tryResetNewRound(now time.Time) {
	todayResetTime := gamedata.GetTodayWBResetTime(now)
	if t.roundStatus.LastResetTime < todayResetTime && now.Unix() > todayResetTime {
		logs.Trace("[WorldBoss] tickerHolder tryResetNewRound, todayResetTime [%d]", todayResetTime)
		t.res.BossMod.resetNewRoundBoss(now)
		t.res.RankDamageMod.resetNewRound(now)
		t.res.FormationRankMod.resetNewRound(now)
		t.res.PlayerMod.resetNewRound(now)

		t.roundStatus.LastResetTime = now.Unix()
		t.roundStatus.LastNewBossTime = now.Unix()
		t.roundStatus.BatchTag = fmt.Sprintf("%04d%02d%02d", now.Year(), int(now.Month()), now.Day())
		t.roundStatus.dirty = true

		clearBatchTime := now.Add(-2 * util.WeekSec * time.Second)
		clearBatch := fmt.Sprintf("%04d%02d%02d", clearBatchTime.Year(), int(clearBatchTime.Month()), clearBatchTime.Day())
		t.res.RankDB.removeRank(clearBatch)
		t.res.RankDB.removeFormationRank(clearBatch)
		t.res.BossDB.removeBossBatch(clearBatch)
		t.res.PlayerDB.removePlayerBatch(clearBatch)
	}
}

//tryDamageRankReward ..
func (t *tickerHolder) tryDamageRankReward(now time.Time) {
	todayRewardTime := gamedata.GetTodayWBRewardTime(now)
	if t.roundStatus.LastRewardTime < todayRewardTime && now.Unix() > todayRewardTime {
		if err := t.res.callback.DamageRank(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] tickerHolder tryDamageRankReward DamageRank failed, %v", err))
			return
		}
		if err := t.res.callback.Marquee(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] tickerHolder tryDamageRankReward Marquee failed, %v", err))
		}
		t.roundStatus.LastRewardTime = now.Unix()
		t.roundStatus.dirty = true
		logs.Info("[WorldBoss] tickerHolder tryDamageRankReward DamageRank")
		return
	}
	yesterdayRewardTime := todayRewardTime - util.DaySec
	if t.roundStatus.LastRewardTime < yesterdayRewardTime && now.Unix() > yesterdayRewardTime {
		if err := t.res.callback.DamageRank(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] tickerHolder tryDamageRankReward [Yesterday] DamageRank failed, %v", err))
			return
		}
		if err := t.res.callback.Marquee(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] tickerHolder tryDamageRankReward Marquee failed, %v", err))
		}
		t.roundStatus.LastRewardTime = now.Unix()
		t.roundStatus.dirty = true
		logs.Info("[WorldBoss] tickerHolder tryDamageRankReward [Yesterday] DamageRank")
		return
	}
}

//----ticker

func (t *tickerHolder) holding() {
	defer logs.PanicCatcherWithInfo("[WorldBoss], process message")

	ticker := time.NewTicker(defaultTickInterval)

	for !t.isClosed {
		select {
		case now := <-ticker.C:
			// logs.Debug("[WorldBoss] tickerHolder ticker")
			for _, elem := range t.tickerList {
				elem.doTick(now)
			}
		}
	}
}

func (t *tickerHolder) regTicker(i, c uint32, f func(time.Time)) {
	t.tickerList = append(t.tickerList, &tickerElem{
		interval: i,
		curr:     c,
		call:     f,
	})
}

type tickerElem struct {
	interval uint32
	curr     uint32
	call     func(time.Time)
}

func (e *tickerElem) doTick(now time.Time) {
	defer logs.PanicCatcherWithInfo("[WorldBoss] do ticker")

	e.curr++
	if e.curr < e.interval {
		return
	}
	e.curr = 0
	e.call(now)
}
