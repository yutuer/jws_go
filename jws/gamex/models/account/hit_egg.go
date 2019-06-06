package account

import (
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerHitEgg struct {
	ActivityEndTime   int64    `json:"act_st"`
	NextEggUpdateTime int64    `json:"neut"`  // 下次蛋刷新时间
	CurIdx            uint32   `json:"cidx"`  // 当前是第几次了
	EggsShow          []bool   `json:"esw"`   // 当前显示蛋情况
	EggsWeight        []uint32 `json:"ew"`    // 当前蛋权重
	EggHammerHcLeft   int64    `json:"ehhl"`  // 当前hc换锤子，溢出的部分
	TodayGotHc        int64    `json:"tghc"`  // 本日获得hc
	TodayGotHammer    int64    `json:"tghmr"` // 本日获得锤子数
	IsEnd             bool     `json:"is_end"`
}

func (he *PlayerHitEgg) UpdateHitEggActivityTime(p *Account, activityStartTime, activityEndTime, now_time int64) {
	if now_time < activityStartTime || now_time > activityEndTime {
		he.EndHitEggActivity(p, now_time)
	} else {
		he.IsEnd = false
	}
}

func (he *PlayerHitEgg) EndHitEggActivity(p *Account, now_time int64) {
	if !he.IsEnd {
		n := p.Profile.GetSC().GetSC(helper.SC_EggKey)
		if n > 0 {
			p.Profile.GetSC().UseSC(helper.SC_EggKey, n, "HitEggEnd")
		}
		he.IsEnd = true
		he._resetHitEgg(now_time)
	}
}

func (he *PlayerHitEgg) UpdateHitEgg(now_time int64) {
	if he.IsEnd || now_time < he.NextEggUpdateTime {
		return
	}
	he._resetHitEgg(now_time)
}

func (he *PlayerHitEgg) _resetHitEgg(now_time int64) {
	he.NextEggUpdateTime = util.GetNextDailyTime(
		gamedata.GetCommonDayBeginSec(now_time), now_time)
	he.CurIdx = 1
	he.EggsShow = make([]bool, gamedata.EggCountInAGame())
	for i := 0; i < gamedata.EggCountInAGame(); i++ {
		he.EggsShow[i] = true
	}
	he.EggsWeight = gamedata.HitEggInitWeight(he.CurIdx)
	he.EggHammerHcLeft = 0
	he.TodayGotHc = 0
	he.TodayGotHammer = 0
}

func (he *PlayerHitEgg) OnAddHcBug(p *Account, addHcBuy, now_time int64) int64 {
	if he.IsEnd {
		return 0
	}
	lc, nt := p.Profile.GetCounts().Get(counter.CounterTypeHitHammerDailyLimit, p)
	if nt < 0 || lc <= 0 {
		return 0
	}
	he.UpdateHitEgg(now_time)
	c := he.EggHammerHcLeft + addHcBuy
	hmr_c := c / int64(gamedata.GetCommonCfg().GetHCPerEggKey())
	if int64(lc) >= hmr_c {
		he.EggHammerHcLeft = c % int64(gamedata.GetCommonCfg().GetHCPerEggKey())
	} else {
		hmr_c = int64(lc)
		he.EggHammerHcLeft = c - hmr_c*int64(gamedata.GetCommonCfg().GetHCPerEggKey())
	}
	p.Profile.GetCounts().UseN(counter.CounterTypeHitHammerDailyLimit, int(hmr_c), p)
	he.TodayGotHammer += hmr_c
	return hmr_c
}

func (he *PlayerHitEgg) RandHitEgg(rnd *rand.Rand) (isSpec bool, loot string, weight uint32) {
	smw := he.GetSumWeight()
	r := rnd.Int31n(int32(smw))
	var ri int
	var rw uint32
	for i, w := range he.EggsWeight {
		if uint32(r) < w {
			ri = i
			rw = w
			break
		}
	}

	logs.Debug("RandHitEgg r %d ws %d %v got %d", r, smw, he.EggsWeight, ri)

	cfg := gamedata.GetHitEggReward(he.CurIdx)
	if ri == 0 { // 是大奖
		he.NextEgg()
		return true, cfg.GetSpecialLootID(), cfg.GetSpecialLootWeight()
	} else { // 不是大奖
		for i := ri; i < len(he.EggsWeight); i++ {
			if he.EggsWeight[i] > 0 {
				he.EggsWeight[i] -= rw
			}
		}

		if he.GetSumWeight() <= 0 { // 所有蛋都砸完
			he.NextEgg()
		}

		ls := cfg.GetLoots()
		return false, ls[ri-1].GetLootDataID(), ls[ri-1].GetWeight()
	}
}

func (he *PlayerHitEgg) NextEgg() {
	he.CurIdx++
	if he.CurIdx > uint32(gamedata.HitEggDailyEggCount()) {
		he.CurIdx = 1
	}

	he.EggsShow = make([]bool, gamedata.EggCountInAGame())
	for i := 0; i < gamedata.EggCountInAGame(); i++ {
		he.EggsShow[i] = true
	}
	he.EggsWeight = gamedata.HitEggInitWeight(he.CurIdx)
	logs.Debug("NextEgg %d", he.EggsWeight)
}

func (he *PlayerHitEgg) GetSumWeight() uint32 {
	var res uint32
	for _, v := range he.EggsWeight {
		if v > res {
			res = v
		}
	}
	return res
}
