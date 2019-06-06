package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildWorshipInfo struct {
	PersionTakeNum   int64
	TakeId           int64
	TakeSign         int64
	HasRewards       []int64
	Reward           []string
	OneTime          []int64
	DoubleTime       []int64
	WorshipAccoundID string
	LastResetTime    int64 // 上次重置时间
}

func (e *GuildWorshipInfo) GetWorshipInfo() (int64, int64) {
	return e.PersionTakeNum, e.TakeId
}

func (e *GuildWorshipInfo) SetWorshipInfo(mem int64) {
	e.PersionTakeNum += 1
	e.TakeId = mem
}

func (e *GuildWorshipInfo) UpdateWorshipHasReward(boxId int64) {
	e.HasRewards = append(e.HasRewards, boxId)
}

func (e *GuildWorshipInfo) CheckDailyReset(now int64) bool {
	logs.Debug("guild Worship CheckDailyReset, %d, %d", now, e.LastResetTime)
	if !util.IsSameUnixByStartTime(e.LastResetTime, now,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypGuildWorshipReset)) {

		e.DailyReset(now)
		logs.Debug("guild Worship Profile daily reset")
		return true
	}
	return false
}

func (e *GuildWorshipInfo) DailyReset(now int64) {
	e.PersionTakeNum = 0
	e.TakeId = 0
	e.TakeSign = 0
	e.HasRewards = nil
	e.Reward = nil
	e.OneTime = nil
	e.DoubleTime = nil
	e.WorshipAccoundID = ""
	e.LastResetTime = now
}
