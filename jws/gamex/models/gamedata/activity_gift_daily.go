package gamedata

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
ActiveType	周期类别	0=根据配置统一开始和结束；1=账号首次进入游戏时激活
DateRestrict	日期约束	0=累计签到；1=连续签到

日期约束
连续签到 在首日日期确定时，即确定每条记录所绑定的具体日期。玩家每日登录时只能领取同日期记录下的奖励。
累计签到 每天登陆都领取下一个未领取的奖励，不存在半路上错过的说法。但每缺席一日，玩家等于失去一个最靠后的奖励。

周期类别
根据配置统一开始和结束（预设周期）每个玩家的节日连续礼包的周期的起止日期都是全服甚至跨服统一的。
账号首次进入游戏时激活（个人激活）每个玩家的新手七日奖励的周期的起止日期都是不同的

*/

// 七日连登活动ID 1
const ActivityID7DayLogin = 1

type activityGiftDaily struct {
	CurrActivityId    uint32       `json:"id"`
	StartActivityTime int64        `json:"st"`      // 开始时间
	CurrGiftIdx       int          `json:"gidx"`    // 表示当前领取到第几个了 从0开始
	CurrGiftStat      GiftStat_t   `json:"s"`       // 领奖状态
	AllGiftStats      []GiftStat_t `json:"all_g_s"` // 所有奖励的状态
}

func (a *activityGiftDaily) update(now_time int64) (is_need_remove bool) {
	// 外层会在跨天时调用
	is_need_remove = false
	data := a.GetData()
	if data == nil {
		logs.Error("activityGiftDaily %d data nil by %v", a.CurrActivityId, *a)
		is_need_remove = true
		return
	}
	is_in_time := util.IsTimeBetweenUnix(
		data.StartTime,
		data.EndTime,
		now_time)

	if !is_in_time {
		is_need_remove = true
		return
	}

	giftLen := len(a.GetData().Gift)
	if data.GetTyp == ActivityGift_Get_Typ_Pass {
		// 活动被激活才开始检查
		if a.StartActivityTime > 0 {
			// 连续签到 在首日日期确定时，即确定每条记录所绑定的具体日期。
			// 玩家每日登录时只能领取同日期记录下的奖励。
			day_bef := GetCommonDayDiff(a.StartActivityTime, now_time)
			a.CurrGiftIdx = int(day_bef)
			if a.CurrGiftIdx >= giftLen {
				//				is_need_remove = true
				return
			}
			a.CurrGiftStat = GiftStat_NoGet
		}
	} else {
		if a.CurrGiftStat != GiftStat_NoGet {
			a.CurrGiftIdx++
		}
		if a.CurrGiftIdx >= giftLen {
			//			is_need_remove = true
			return
		}
		a.CurrGiftStat = GiftStat_NoGet
	}

	return
}

func (a *activityGiftDaily) GetData() *activityGiftData {
	re, ok := GetActivityGiftMapData(a.CurrActivityId)
	if ok {
		return re
	} else {
		return nil
	}
}

// 用来判断活动的奖发没发完，如果返回false则说明活动已经结束了，没有奖励了也留着，防止重复加活动
func (a *activityGiftDaily) IsHasReward() (bool, int) {
	data := a.GetData()
	return a.CurrGiftIdx < len(data.Gift), len(data.Gift)
}

func (a *activityGiftDaily) getGiftNowDay() *giftDailyData {
	//a.Update(now_time) 外层需要Update啊
	act := a.GetData()
	if act == nil {
		logs.Error("No activity Now")
		return nil
	}
	gift := act.Gift
	if gift == nil {
		return nil
	}
	if a.CurrGiftIdx < 0 || a.CurrGiftIdx >= len(gift) {
		return nil
	}
	return &gift[a.CurrGiftIdx]
}

// 这里在需要时顺便返回data
func (a *activityGiftDaily) GetGiftToGet(vip_lv uint32) (has_gift bool, data *PriceDatas) {
	//a.Update(now_time) 外层需要Update啊
	has_gift = false
	data = nil

	if a.CurrGiftStat == GiftStat_HasGetVip {
		// vip都领了，自然不能领了
		return
	}

	gift_data := a.getGiftNowDay()
	if gift_data == nil {
		return
	}

	if a.CurrGiftStat == GiftStat_NoGet {
		// 还没领呢，自然可以领（正常情况下每天都应该有）
		has_gift = true
		if gift_data.VipNeed > 0 && vip_lv >= gift_data.VipNeed {
			data = &gift_data.Vip
		} else {
			data = &gift_data.Base
		}
		return
	} else {
		if gift_data.VipNeed > 0 && vip_lv >= gift_data.VipNeed {
			has_gift = true
			data = &gift_data.VipAddon
			return
		}
	}
	return
}

func (a *activityGiftDaily) SetHasGet(vip_lv uint32) (isRemove bool) {
	gift_data := a.getGiftNowDay()
	if gift_data == nil {
		// 这其实是配置出错了
		a.CurrGiftStat = GiftStat_HasGetVip
		a.AllGiftStats[a.CurrGiftIdx] = GiftStat_HasGetVip
		return
	}
	if a.CurrGiftStat == GiftStat_NoGet {
		if gift_data.VipNeed > 0 && vip_lv >= gift_data.VipNeed {
			a.CurrGiftStat = GiftStat_HasGetVip
			a.AllGiftStats[a.CurrGiftIdx] = GiftStat_HasGetVip
		} else {
			a.CurrGiftStat = GiftStat_HasGetBase
			a.AllGiftStats[a.CurrGiftIdx] = GiftStat_HasGetBase
		}
	} else {
		if gift_data.VipNeed > 0 && vip_lv >= gift_data.VipNeed {
			a.CurrGiftStat = GiftStat_HasGetVip
			a.AllGiftStats[a.CurrGiftIdx] = GiftStat_HasGetVip
		}
	}
	giftLen := len(a.GetData().Gift)
	if ((gift_data.VipNeed <= 0 && a.CurrGiftStat == GiftStat_HasGetBase) || a.CurrGiftStat == GiftStat_HasGetVip) &&
		a.CurrGiftIdx >= giftLen-1 {
		//		isRemove = true
	}
	return
}

type ActivityGiftDailys struct {
	Gifts []activityGiftDaily `json:"gift"`
	// 上次更新触发修改的时间，
	// 如果和今天是一天，则表示今天的数据已经更新过了
	Last_get_gift_time int64 `json:"lt"`
}

func (a *ActivityGiftDailys) add(act_start_time int64, data *activityGiftData) {
	a.Gifts = append(a.Gifts, activityGiftDaily{
		CurrActivityId:    data.Id,
		StartActivityTime: act_start_time,
		CurrGiftIdx:       0,
		CurrGiftStat:      GiftStat_NoGet,
		AllGiftStats:      make([]GiftStat_t, len(data.Gift)),
	})
}

func (a *ActivityGiftDailys) isHasAdd(id uint32) bool {
	// 同时出现的活动不会很多
	//logs.Trace("isHasAdd %d %v", id, a.Gifts)
	for i := 0; i < len(a.Gifts); i++ {
		if a.Gifts[i].CurrActivityId == id {
			return true
		}
	}
	return false
}

func (a *ActivityGiftDailys) getAct(id uint32) (idx int, act *activityGiftDaily) {
	for i := 0; i < len(a.Gifts); i++ {
		if a.Gifts[i].CurrActivityId == id {
			return i, &a.Gifts[i]
		}
	}
	return 0, nil
}

func (a *ActivityGiftDailys) Update(now_time int64) {
	// 每天只更新一次
	if !IsSameDayCommon(now_time, a.Last_get_gift_time) {
		a.Last_get_gift_time = now_time
		// 添加新的
		gs := GetActivityGiftData()
		for i := 0; i < len(gs); i++ {
			if !util.IsTimeBetweenUnix(gs[i].StartTime, gs[i].EndTime, now_time) {
				continue
			}
			if a.isHasAdd(gs[i].Id) {
				continue
			}
			// 若和全局时间无效的活动，则根据玩家行为激活
			act_start_time := now_time
			if gs[i].StartTime <= 0 && gs[i].EndTime <= 0 {
				act_start_time = 0
			}
			a.add(act_start_time, gs[i])
		}

		for i := 0; i < len(a.Gifts); {
			is_remove := a.Gifts[i].update(now_time)
			if is_remove {
				a.removeGift(i)
			} else {
				i++
			}
		}
	}
}

// 需要根据玩家行为激活的使用此接口进行激活（相对的根据服务器时间激活的活动无效）
func (a *ActivityGiftDailys) ActiveGift(aid uint32, now_time int64) (bool, error) {
	_, act := a.getAct(aid)
	if act == nil {
		return false, fmt.Errorf("gift id not found")
	}
	if act.StartActivityTime <= 0 {
		act.StartActivityTime = now_time
		return true, nil
	}
	return false, nil
}

func (a *ActivityGiftDailys) GetGiftToGet(aid uint32, vip_lv uint32, now_time int64) (bool, *PriceDatas) {
	a.Update(now_time)
	_, act := a.getAct(aid)
	if act == nil {
		return false, nil
	}
	if act.StartActivityTime <= 0 {
		return false, nil
	}
	return act.GetGiftToGet(vip_lv)
}

func (a *ActivityGiftDailys) SetHasGet(aid uint32, vip_lv uint32, now_time int64) {
	a.Update(now_time)
	idx, act := a.getAct(aid)
	if act != nil && act.StartActivityTime > 0 {
		if act.SetHasGet(vip_lv) {
			a.removeGift(idx)
		}
	}
}

func (a *ActivityGiftDailys) Init() {
	a.Last_get_gift_time = 0
}

func (a *ActivityGiftDailys) removeGift(idx int) {
	if idx < len(a.Gifts)-1 {
		a.Gifts[idx] = a.Gifts[len(a.Gifts)-1]
	}
	a.Gifts = a.Gifts[:len(a.Gifts)-1]
}

func (a *activityGiftDaily) GetAllGiftData(vip_lv uint32) []*PriceDatas {
	res := make([]*PriceDatas, 0, len(a.GetData().Gift))
	for i, s := range a.AllGiftStats {
		g := a.GetData().Gift[i]
		if s == GiftStat_NoGet {
			// 还没领呢，自然可以领（正常情况下每天都应该有）
			if g.VipNeed > 0 && vip_lv >= g.VipNeed {
				res = append(res, &g.Vip)
			} else {
				res = append(res, &g.Base)
			}
		} else {
			if g.VipNeed > 0 && vip_lv >= g.VipNeed {
				res = append(res, &g.VipAddon)
			} else {
				res = append(res, &g.Base)
			}
		}
	}
	return res
}
