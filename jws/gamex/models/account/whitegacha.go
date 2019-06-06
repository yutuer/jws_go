package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type WhiteGachaInfo struct {
	LastGachaTime   int64
	GachaNum        int64
	GachaBless      int64
	TodayFreeCount  int64
	WhiteGachaActId uint32
}

func (wg *WhiteGachaInfo) UpdateWhiteGacha() {
	wg.GachaNum += 1
	wg.GachaBless += 1
}

func (wg *WhiteGachaInfo) UpdateWhiteGachaActId(actId uint32) {
	wg.WhiteGachaActId = actId
}

func (wg *WhiteGachaInfo) IsCanFree(activityId uint32, now_time int64, idx int64) bool {
	data := gamedata.GetHotDatas().Activity.GetActivityGachaSeting(activityId)
	if data == nil {
		logs.Error("Gacha Data Err By %d", idx)
		return false
	}

	if data.GetFreeTime() <= 0 {
		return false
	}

	return now_time >= int64(data.GetFreeTime())+wg.LastGachaTime
}

func (wg *WhiteGachaInfo) SetUseFreeNow(now_time int64) {
	wg.LastGachaTime = now_time
	wg.TodayFreeCount += 1
}

func (wg *WhiteGachaInfo) SetWhiteGacha2Zero() {
	wg.LastGachaTime = 0
	wg.GachaNum = 0
	wg.GachaBless = 0
	wg.TodayFreeCount = 0
}
