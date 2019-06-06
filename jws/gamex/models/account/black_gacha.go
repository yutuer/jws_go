package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

// 黑盒宝箱相关信息
type BlackGachaInfo struct {
	BlackGachaHeroInfo   BlackGachaActivity
	BlackGachaWeaponInfo BlackGachaActivity
	LastResetTime        int64
}

type BlackGachaActivity struct {
	SubActivies []BlackGachaSubInfo
	ActivityId  uint32
}

type BlackGachaSubInfo struct {
	TodayFreeUsedCount    int     // 今天
	GachaCount            int     // 累计抽奖次数
	HasClaimedExtraReward []int64 // 已经领取的额外奖励
	SubId                 uint32
}

func (bga *BlackGachaInfo) GetSubActivity(activityId, subId uint32) *BlackGachaSubInfo {
	if bga.BlackGachaHeroInfo.ActivityId == activityId {
		return bga.BlackGachaHeroInfo.Get(subId)
	} else if bga.BlackGachaWeaponInfo.ActivityId == activityId {
		return bga.BlackGachaWeaponInfo.Get(subId)
	}
	return nil
}

func (bga *BlackGachaInfo) TryDailyReset(nowTime int64) {
	if !gamedata.IsSameDayCommon(nowTime, bga.LastResetTime) {
		bga.LastResetTime = nowTime
		bga.BlackGachaHeroInfo.DailyReset()
		bga.BlackGachaWeaponInfo.DailyReset()
	}
}

func (bga *BlackGachaActivity) DailyReset() {
	for i := range bga.SubActivies {
		bga.SubActivies[i].TodayFreeUsedCount = 0
	}
}

func (bga *BlackGachaActivity) Clear() {
	bga.SubActivies = nil
	bga.ActivityId = 0
}

func (bga *BlackGachaActivity) Reset(newActId uint32) {
	bga.Clear()
	bga.ActivityId = newActId
}

func (bga *BlackGachaActivity) Get(subId uint32) *BlackGachaSubInfo {
	for i := range bga.SubActivies {
		if bga.SubActivies[i].SubId == subId {
			return &bga.SubActivies[i]
		}
	}
	return nil
}

func (bga *BlackGachaSubInfo) GetHasClaimedReward() []int64 {
	if bga.HasClaimedExtraReward == nil {
		bga.HasClaimedExtraReward = make([]int64, 0)
	}
	return bga.HasClaimedExtraReward
}

func (bga *BlackGachaSubInfo) AddHasClaimedReward(rewardId int64) {
	if bga.HasClaimedExtraReward == nil {
		bga.HasClaimedExtraReward = make([]int64, 0)
	}
	bga.HasClaimedExtraReward = append(bga.HasClaimedExtraReward, rewardId)
}

func (bga *BlackGachaSubInfo) ContainsReward(rewardId int64) bool {
	for _, reward := range bga.HasClaimedExtraReward {
		if rewardId == reward {
			return true
		}
	}
	return false
}
