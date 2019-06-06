package counter

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

// Update func

func (p *PlayerCounter) updateByTyp(typ int, ud UpdateData) bool {
	data := gamedata.GetGameModeControlData(typ)
	if data == nil {
		return false
	}
	switch data.GetType {
	case 0:
		return p.updateInFlush(typ, ud.GetProfileNowTime(), data)
	case 1:
		return p.updateInAdd(typ, ud.GetProfileNowTime(), data)
	case 2:
		return p.updateInTimer(typ, ud.GetProfileNowTime(), data)
	}
	return false
}

// 刷新类型0 重设
func (p *PlayerCounter) updateInFlush(typ int, nowT int64, data *gamedata.GameModeControlData) bool {
	lastT := p.CountLastUpdateTime[typ]
	if nowT-lastT >= util.WeekSec {
		// 超过一周的话肯定会扫到刷新点
		// 如果现在已经是最大值以上的话，不更新
		if p.Counts[typ] < data.GetCount {
			p.Counts[typ] = data.GetCount
		}
		p.CountLastUpdateTime[typ] = getLastUpdateTimeBeforeNow(nowT, data.GetDailyTime)
		return true
	}
	l := lastT + util.DaySec
	for l <= nowT {
		week := util.GetWeek(l)
		if data.IsNeedGetDayInWeek[week] {
			// 通过刷新点
			// 如果现在已经是最大值以上的话，不更新
			if p.Counts[typ] < data.GetCount {
				p.Counts[typ] = data.GetCount
			}
			p.CountLastUpdateTime[typ] = getLastUpdateTimeBeforeNow(nowT, data.GetDailyTime)
			return true
		}
		l += util.DaySec
	}

	return false
}

// 刷新类型1 累加
func (p *PlayerCounter) updateInAdd(typ int, nowT int64, data *gamedata.GameModeControlData) bool {
	lastT := p.CountLastUpdateTime[typ]
	re := false
	l := lastT + util.DaySec
	for l <= nowT {
		week := util.GetWeek(l)
		if data.IsNeedGetDayInWeek[week] {
			// 通过刷新点
			p.Counts[typ] += data.GetCount
			re = true
		}
		l += util.DaySec
	}
	if p.Counts[typ] > data.CountMax {
		p.Counts[typ] = data.CountMax
	}
	p.CountLastUpdateTime[typ] = getLastUpdateTimeBeforeNow(nowT, data.GetDailyTime)
	return re
}

// 刷新类型2 时间
func (p *PlayerCounter) updateInTimer(typ int, nowT int64, data *gamedata.GameModeControlData) bool {
	max := data.GetCount
	now_unix_sec := nowT

	// 如果现在已经是最大值以上的话，不更新
	if p.Counts[typ] >= max {
		p.LastTime[typ] = now_unix_sec
		return false
	}

	one_need := data.MinPreAdd * util.MinSec

	if one_need <= 0 {
		p.LastTime[typ] = now_unix_sec
		return false
	}

	// 计算增量
	add_point, s := util.AccountTime2Point(
		now_unix_sec,
		p.LastTime[typ],
		one_need)

	// 更新值
	p.Counts[typ] += int(add_point)
	// 时间增长不会超过最大值
	if p.Counts[typ] > max {
		p.Counts[typ] = max
	}

	// 依据文档中的算法上次更新时间应该置为this
	p.LastTime[typ] = now_unix_sec - s
	p.S[typ] = s
	return true
}
