package counter

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

type PlayerCounter struct {
	Counts              [CounterTypeCountMax]int   `json:"c"`
	CountLastUpdateTime [CounterTypeCountMax]int64 `json:"lt"`
	LastTime            [CounterTypeCountMax]int64 `json:"last"`
	S                   [CounterTypeCountMax]int64 `json:"s"`
	needSync            bool
}

// Count API
func (p *PlayerCounter) Use(typ int, ud UpdateData) bool {
	if !p.Has(typ, ud) {
		return false
	}
	if p.Counts[typ] > 0 {
		p.Counts[typ] -= 1
		return true
	} else {
		return false
	}
}

func (p *PlayerCounter) UseN(typ int, times int, ud UpdateData) bool {
	lc, nt := p.Get(typ, ud)
	if nt < 0 || lc < times {
		return false
	}
	p.Counts[typ] -= times
	return true
}

// TODO
// 手动增加调用,自动增加切勿调用(由于增加了teampvp特殊判断,自动增加将会也失去限制)
func (p *PlayerCounter) Add(typ int, times int, ud UpdateData) bool {
	_, nt := p.Get(typ, ud)
	if nt < 0 {
		return false
	}

	p.Counts[typ] += times
	// not good!!!!!!!!!!!
	if typ == CounterTypeTeamPvp || typ == CounterTypeWspvpChallenge {
		return true
	}

	data := gamedata.GetGameModeControlData(typ)

	// not science, 凭什么type为2就不可以不受限制?
	if data.GetType != 2 {
		if p.Counts[typ] > data.GetCount {
			p.Counts[typ] = data.GetCount
		}

	}

	return true
}

func (p *PlayerCounter) UseJustDayBegin(typ int, data UpdateData) bool {
	return p.Has(typ, data) // 这里面会更新
}

func (p *PlayerCounter) UseJustDayEnd(typ int, data UpdateData) bool {
	// 如果跨天则默认扣除的是前一天的
	// 这个逻辑意味着扣次数的过程是一个时间段, 当最终扣除次数时,可能已经跨天了
	// 所以逻辑是扣次数开始时,只检查次数是否满足,不实际扣除
	// 扣次数结束时,如果跨天了就算前一天的(之前已经满足条件)
	d := gamedata.GetGameModeControlData(typ)
	if d.GetType != 2 && p.updateByTyp(typ, data) {
		return true
	}
	return p.Use(typ, data)
}

func (p *PlayerCounter) Has(typ int, ud UpdateData) bool {
	if typ < 0 || typ >= len(p.Counts) {
		return false
	}

	data := gamedata.GetGameModeControlData(typ)
	if data == nil {
		return false
	}

	if data.OpenLevel > 0 && ud.GetCorpLv() < data.OpenLevel {
		// 开启级别
		return false
	}

	p.updateByTyp(typ, ud)
	return p.Counts[typ] > 0
}

func (p *PlayerCounter) Get(typ int, ud UpdateData) (int, int64) {
	if typ < 0 || typ >= len(p.Counts) {
		return 0, -1
	}

	data := gamedata.GetGameModeControlData(typ)
	if data == nil {
		return 0, -1
	}

	if data.OpenLevel > 0 && ud.GetCorpLv() < data.OpenLevel {
		// 开启级别
		return 0, -1
	}

	p.updateByTyp(typ, ud)
	return p.Counts[typ], p.CountLastUpdateTime[typ] + util.DaySec
}

func (p *PlayerCounter) GetWithTimeValue(typ int, ud UpdateData) (uint32, int64, int64, int64) {
	count, nextRefershTime := p.Get(typ, ud)
	return uint32(count), nextRefershTime,
		p.LastTime[typ], p.S[typ]
}

func (p *PlayerCounter) CheckUpdate(ud UpdateData) bool {
	re := false
	for i := 0; i < len(p.Counts); i++ {
		re = re || p.updateByTyp(i, ud)
	}
	return re
}

func (p *PlayerCounter) SetNeedSync(is bool) {
	p.needSync = is
}

func (p *PlayerCounter) IsNeedSync() bool {
	return p.needSync
}

// 对于每日刷新的返回一天的最大值
func (p *PlayerCounter) GetDailyMax(typ int) int {
	data := gamedata.GetGameModeControlData(typ)
	if data == nil {
		return -1
	}
	return data.GetCount
}
