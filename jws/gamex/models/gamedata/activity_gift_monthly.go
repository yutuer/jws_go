package gamedata

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GiftStat_t int

const (
	GiftStat_NoGet      GiftStat_t = iota // 0表示还没领取
	GiftStat_HasGetBase                   // 1表示已经领取过基本奖励了
	GiftStat_HasGetVip                    // 2 表示已经领取过VIP奖励了
	GiftStat_ReSign                       // 3 补签
)

type ActivityGiftMonthly struct {
	Curr_activity_id uint32     `json:"id"`
	Curr_gift_idx    int        `json:"gidx"` // 表示当前领取到第几个了 从0开始
	Curr_gift_stat   GiftStat_t `json:"s"`    // 领奖状态

	// 上次更新触发修改的时间，
	// 如果和今天是一天，则表示今天的数据已经更新过了
	Last_get_gift_time int64 `json:"lt"`
}

func (a *ActivityGiftMonthly) Update(now_time int64) {
	// 每天只更新一次
	if !IsSameDayCommon(now_time, a.Last_get_gift_time) {
		a.Last_get_gift_time = now_time

		// 清空领奖状态
		if a.Curr_gift_stat != GiftStat_NoGet {
			a.Curr_gift_idx++
		}
		a.Curr_gift_stat = GiftStat_NoGet

		// 检查一下是不是跨月份活动了
		mouth_act := GetNowMonthlyGiftData(now_time)
		if mouth_act == nil {
			logs.Warn("No mouth_activity Now")
			return
		}
		if mouth_act.Id != a.Curr_activity_id {
			// 跨月了
			a.Curr_activity_id = mouth_act.Id
			a.Curr_gift_idx = 0
		}
	}
}

func (a *ActivityGiftMonthly) OnReSign() {
	// 代码顺序不能换  先切换到下一天，然后设置状态
	a.Curr_gift_idx++
	a.Curr_gift_stat = GiftStat_ReSign
}

func (a *ActivityGiftMonthly) getGiftNowDay(now_time int64, reSign bool) *giftData {
	if reSign {
		return a.getGiftNowDayForResign(now_time)
	} else {
		return a.getGiftNowDayForCommonSign(now_time)
	}
}

func (a *ActivityGiftMonthly) getGiftNowDayForCommonSign(now_time int64) *giftData {
	a.Update(now_time)
	mouth_act := GetNowMonthlyGiftData(now_time)
	if mouth_act == nil {
		logs.Warn("No mouth_activity Now")
		return nil
	}
	if a.Curr_gift_idx >= Mouth_Gift_Num_Max {
		logs.Error("a.Curr_gift_idx %d Err", a.Curr_gift_idx)
		return nil
	}

	return &mouth_act.Gift[a.Curr_gift_idx]
}

func (a *ActivityGiftMonthly) getGiftNowDayForResign(now_time int64) *giftData {
	a.Update(now_time)
	mouth_act := GetNowMonthlyGiftData(now_time)
	if mouth_act == nil {
		logs.Warn("No mouth_activity Now")
		return nil
	}
	if a.Curr_gift_idx+1 >= Mouth_Gift_Num_Max {
		logs.Error("a.Curr_gift_idx %d Err", a.Curr_gift_idx)
		return nil
	}
	logs.Debug("<month gift> get gift for resign %d", a.Curr_gift_idx+1)
	return &mouth_act.Gift[a.Curr_gift_idx+1]
}

// 这里在需要时顺便返回data
// 正常签到是签当天的
// 补签是签下一天的
func (a *ActivityGiftMonthly) GetGiftToGet(vip_lv uint32, now_time int64, reSign bool) (has_gift bool, data []*CostData) {
	a.Update(now_time)
	has_gift = false
	data = nil

	if reSign {
		// TODO 单元测试
		dayInMonth := GetNowMonthlyGiftDayth(now_time)
		if a.Curr_gift_idx >= dayInMonth-1 {
			// 补签次数不足
			logs.Warn("<MONTH SIGN> not left day to resign %d, %d", a.Curr_gift_idx, dayInMonth)
			return
		}
		if a.Curr_gift_stat == GiftStat_NoGet {
			// 必须今天签到之后才能补签
			logs.Warn("<MONTH SIGN> you need sign today first")
			return
		}
	} else {
		if a.Curr_gift_stat != GiftStat_NoGet {
			// 任何签到之后都不能再签了
			logs.Warn("<MONTH SIGN> cannot repeat sign ")
			return
		}
	}

	gift_data := a.getGiftNowDay(now_time, reSign)
	if gift_data == nil {
		// 这其实是配置出错了
		logs.Warn("<MONTH SIGN> config err")
		return
	}

	logs.Trace("GetGiftToGet %d %d %v %v %v", vip_lv, gift_data.VipNeed, gift_data.Vip, gift_data.Base, gift_data.VipNeed)
	if reSign {
		has_gift = true
		if vip_lv >= gift_data.VipNeed && gift_data.VipNeed != 0 {
			data = []*CostData{&gift_data.Base, &gift_data.VipAddon}
		} else {
			data = []*CostData{&gift_data.Base}
		}
	} else {
		has_gift = true
		if vip_lv >= gift_data.VipNeed && gift_data.VipNeed != 0 {
			data = []*CostData{&gift_data.Base, &gift_data.VipAddon}
		} else {
			data = []*CostData{&gift_data.Base}
		}
	}
	return
}

func (a *ActivityGiftMonthly) SetHasGet(vip_lv uint32, now_time int64, reSign bool) uint32 {
	gift_data := a.getGiftNowDay(now_time, reSign)
	if gift_data == nil {
		// 这其实是配置出错了
		a.Curr_gift_stat = GiftStat_HasGetVip
		return errCode.CommonInner
	}
	if reSign {
		a.OnReSign()
	} else {
		if vip_lv >= gift_data.VipNeed && gift_data.VipNeed > 0 {
			a.Curr_gift_stat = GiftStat_HasGetVip
		} else {
			a.Curr_gift_stat = GiftStat_HasGetBase
		}
	}
	return 0
}
