package count

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

type CountData struct {
	Count           int   `json:"c"`
	LastClearTime   int64 `json:"t"`
	DayBeginTimeTyp int   `json:"dbtyp"`
}

func NewDailyClear(dayBeginTimeTyp int) CountData {
	return CountData{
		DayBeginTimeTyp: dayBeginTimeTyp,
	}
}

func (c *CountData) Add(nowT int64, add int) int {
	c.update(nowT)
	c.Count += add
	return c.Count
}

func (c *CountData) Get(nowT int64) int {
	c.update(nowT)
	return c.Count
}

func (c *CountData) update(nowT int64) {
	if c.DayBeginTimeTyp < 0 {
		return
	}

	dayStartTime := gamedata.GetBeginTimeByTyp(c.DayBeginTimeTyp)
	if dayStartTime.IsNil() {
		return
	}

	if !util.IsSameUnixByStartTime(
		nowT,
		c.LastClearTime,
		dayStartTime) {
		c.Count = 0
		c.LastClearTime = nowT
	}
}
