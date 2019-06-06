package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

/**
武将多余碎片抽奖功能
*/

type HeroSurplusInfo struct {
	EndTime               int64                          `json:"end_time"`         // 抽奖宝箱的结束时间  ==0 或者 < nowTime 说明没有这个入口了
	DailyDrawCount        [helper.Hero_Surplus_Count]int `json:"daily_draw_count"` // 每日抽奖次数
	OpenCount             int                            `json:"open_count"`       // 每日活动开启次数
	DailyUsedEN           int                            `json:"daily_used_en"`    // 每日累计消耗体力
	DailyFirstOpen2Client bool                           `json:"daily_first_open"` // 用于给客户端弹提示的
	DailyResetTime        int64                          `json:"daily_reset_time"` // 每日重置时间
}

func (h *HeroSurplusInfo) AddUsedEn(val int, nowTime int64) {
	h.DailyUsedEN += val
	openEn := gamedata.GetHeroCommonConfig().GetSPStoreAppear()
	if uint32(h.DailyUsedEN) >= openEn {
		if h.OpenCount < int(gamedata.GetHeroCommonConfig().GetSPStoreTimes()) {
			h.DailyUsedEN -= int(openEn)
			h.triggerNewOpen(nowTime)
		}
	}
}

func (h *HeroSurplusInfo) triggerNewOpen(nowTime int64) {
	h.DailyFirstOpen2Client = true
	h.EndTime = nowTime + int64(gamedata.GetHeroCommonConfig().GetSPStoreDuration()*60)
	h.OpenCount++
}

func (h *HeroSurplusInfo) AddDailyDrawCount(surplusId, val int) {
	if surplusId < 0 || surplusId >= helper.Hero_Surplus_Count {
		return
	}
	h.DailyDrawCount[surplusId] += val
}

func (h *HeroSurplusInfo) TryDailyReset(nowTime int64) {
	if !gamedata.IsSameDayCommon(nowTime, h.DailyResetTime) {
		h.DailyResetTime = nowTime
		h.dailyReset()
	}
}

func (h *HeroSurplusInfo) dailyReset() {
	for i := range h.DailyDrawCount {
		h.DailyDrawCount[i] = 0
	}
	h.OpenCount = 0
	h.DailyUsedEN = 0
	h.DailyFirstOpen2Client = false
}

func (h *HeroSurplusInfo) DebugOpen(nowTime int64) {
	h.triggerNewOpen(nowTime)
}

func (h *HeroSurplusInfo) DebugReset(nowTime int64) {
	h.DailyResetTime = nowTime
	h.dailyReset()
}
