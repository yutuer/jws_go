package world_boss

import (
	"fmt"
	"math"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Idle = iota
	Battle
)

type WorldBossData struct {
	BuffLevel            int           `json:"bf_lv"`
	GotRewards           []int64       `json:"got_rw"`
	WorldBossDay         string        `json:"wb_day"`
	BattleNowInfo        BattleNowInfo `json:"bt_n_info"`
	BattleTimes          int           `json:"battle_ts"`
	Damage               uint64        `json:"damage"`
	State                int           `json:"state"`
	StartBattleTime      int64         `json:"s_b_t"`
	DisableRemindBuyBuff bool          `json:"d_r_b_b"`
	LastResetTime        int64         `json:"l_r_t"`
	CurDamage            int64         `json:"c_dmg"`
	MaxHeroATK           float32       `json:"max_hero_atk"`
}

type BattleNowInfo struct {
	HadCostTimes bool `json:"had_c_t"`
}

func (wb *WorldBossData) SetHadCostTimes(flag bool) {
	wb.BattleNowInfo.HadCostTimes = flag
}

func (wb *WorldBossData) IsHadCostTimes() bool {
	return wb.BattleNowInfo.HadCostTimes
}

func (wb *WorldBossData) SetRewards(id int) {
	if wb.GotRewards == nil {
		wb.GotRewards = make([]int64, 0)
	}
	wb.GotRewards = append(wb.GotRewards, int64(id))
}

func (wb *WorldBossData) Reset() {
	wb.GotRewards = make([]int64, 0)
	wb.Damage = 0
	wb.BattleTimes = 0
}

func (wb *WorldBossData) SetDamageInDay(damage uint64, nowT time.Time) {
	wb.Damage = damage
	wb.WorldBossDay = fmt.Sprintf("%d-%d-%d", nowT.Year(), nowT.Month(), nowT.Day())
}

func (wb *WorldBossData) HadGetRewards(id int) bool {
	for _, item := range wb.GotRewards {
		if item == int64(id) {
			return true
		}
	}
	return false
}

func (wb *WorldBossData) NeedExit() bool {
	return wb.State == Battle
}

func AntiCheatAllDamage(damage int64, maxAttr float32, buffLevel int, battletTime int64) int64 {
	maxDamage := float64(maxAttr) * (1 + 0.1*float64(buffLevel)) * 1200 * (float64(battletTime) / 60)
	logs.Debug("maxDamage: %v", maxDamage)
	return int64(math.Min(float64(damage), maxDamage))
}

func AntiCheatSingleDamage(damage int64, maxAttr float32, buffLevel int) int64 {
	maxDamage := float64(maxAttr) * (1 + 0.1*float64(buffLevel)) * 80
	logs.Debug("1s maxDamage: %v", maxDamage)
	return int64(math.Min(float64(damage), maxDamage))
}
