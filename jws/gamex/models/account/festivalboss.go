package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type FestivalBossInfo struct {
	FbShopRewardTime      []int64 `json:"fb_shop_reward_time"`
	FesstivalBossKillTime int64   `json:"fesstival_boss_kill_time"`
	FestivalBossActId     uint32  `json:"festtival_boss_act_id"`
}

func (e *FestivalBossInfo) GetFbShopRewardTime() []int64 {
	return e.FbShopRewardTime
}

func (e *FestivalBossInfo) UpdateFestivalActId(actId uint32) {
	e.FestivalBossActId = actId
}

func (e *FestivalBossInfo) UpdateFbShopRewardTime(festivalid uint32, goodsid int64) {
	if len(e.FbShopRewardTime) == 0 {
		e.FbShopRewardTime = make([]int64, gamedata.GetFestivalShopGoodsCount(festivalid))
	}
	e.FbShopRewardTime[goodsid] += 1
}

func (e *FestivalBossInfo) SetFbShopRewardTime2zero() {
	e.FbShopRewardTime = e.FbShopRewardTime[:0]
}

func (e *FestivalBossInfo) GetBossKillTime() int64 {
	return e.FesstivalBossKillTime
}

func (e *FestivalBossInfo) UpdataFbFestivalBossKillTime() {
	e.FesstivalBossKillTime += 1
}

func (e *FestivalBossInfo) SetFbKillTime2zero() {
	e.FesstivalBossKillTime = 0
}
