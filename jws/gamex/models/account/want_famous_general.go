package account

import (
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	RewardDiceNo = 6
)

type WantFamousGeneralInfo struct {
	NextPlayRefreshTime    int64    `json:"next_play_ref_t"`
	CanPlayCountCurr       uint32   `json:"can_play_c"`
	CanPlayCountTotal      uint32   `json:"can_play_c_tl"`
	CanFreeResetCountCurr  uint32   `json:"can_free_reset_c"`
	CanFreeResetCountTotal uint32   `json:"can_free_reset_c_tl"`
	CurrResetCount         uint32   `json:"curr_p_c"`          // 本轮改投次数
	CurrHcResetCount       uint32   `json:"curr_hc_p_c"`       // 本轮花钻改投次数
	PlayResult             []uint32 `json:"play_res"`          // 保持骰子当前点数，1-6，0为未开始
	DailyAllAward          uint32   `json:"daily_all_award_c"` // 本日本人总获得碎片数

	DayPlayTimes            uint32 `json:"day_play_t"` // 本日玩了几次
	DayPlayTimesRefreshTime int64  `json:"day_play_ref_t"`
}

func (w *WantFamousGeneralInfo) UpdateInfo(p *Account, now_time int64) {
	if now_time < w.NextPlayRefreshTime {
		return
	}
	w.NextPlayRefreshTime = util.GetNextDailyTime(
		gamedata.GetCommonDayBeginSec(now_time), now_time)
	w.clear(p)
}

// 投骰子
func (w *WantFamousGeneralInfo) CastDice(r *rand.Rand, now_time int64) {
	w.addDayPlayTimes(now_time)
	cfg := gamedata.GetWantGeneralProbablityConfig(w.DayPlayTimes)
	_rn := uint32(r.Int31n(int32(cfg.GetFirstUpperLimit() - cfg.GetFirstLowLimit() + 1)))
	n6 := cfg.GetFirstLowLimit() + _rn
	w.PlayResult = make([]uint32, gamedata.WantGeneralDiceCount)
	for i := 0; i < gamedata.WantGeneralDiceCount; i++ {
		if uint32(i) < n6 {
			w.PlayResult[i] = RewardDiceNo
		} else {
			w.PlayResult[i] = uint32(r.Int31n(RewardDiceNo-1) + 1)
		}
	}
	logs.Debug("WantGeneral CastDice %d n6 %d res %v", w.DayPlayTimes, n6, w.PlayResult)
}

// 重置
func (w *WantFamousGeneralInfo) ResetCast(r *rand.Rand) {
	cfg := gamedata.GetWantGeneralProbablityConfig(w.DayPlayTimes)
	if cfg != nil && w.CurrResetCount >= cfg.GetGuaranteeTimes() { // 保底
		for i, n := range w.PlayResult {
			if n != RewardDiceNo {
				w.PlayResult[i] = RewardDiceNo
			}
		}
		logs.Debug("WantGeneral ResetCast-Guarantee %d Reset %d HcReset %d %v",
			w.DayPlayTimes, w.CurrResetCount, w.CurrHcResetCount, w.PlayResult)
	} else { // 随机1个
		rf := r.Float32()
		if rf <= cfg.GetResetProbability() {
			for i, n := range w.PlayResult {
				if n != RewardDiceNo {
					w.PlayResult[i] = RewardDiceNo
					break
				}
			}
		}
		for i, n := range w.PlayResult {
			if n != RewardDiceNo {
				w.PlayResult[i] = uint32(r.Int31n(RewardDiceNo-1) + 1)
			}
		}

		logs.Debug("WantGeneral ResetCast-Rand %d Reset %d HcReset %d r %f %v",
			w.DayPlayTimes, w.CurrResetCount, w.CurrHcResetCount, rf, w.PlayResult)
	}
}

// 结算
func (w *WantFamousGeneralInfo) AwardClear() uint32 {
	var res uint32
	for _, s := range w.PlayResult {
		if s == RewardDiceNo {
			res++
		}
	}
	w.PlayResult = []uint32{}
	w.CurrHcResetCount = 0
	w.CurrResetCount = 0
	return res
}

func (w *WantFamousGeneralInfo) AddTodayHeroPiece(n uint32) {
	w.DailyAllAward += n
}

// 在不见好就收之前是不能重新投骰子的，此接口判断是否已经见好就收了
func (w *WantFamousGeneralInfo) IsCanCast() bool {
	if w.CurrHcResetCount > 0 || w.CurrResetCount > 0 {
		return false
	}
	for _, s := range w.PlayResult {
		if s > 0 {
			return false
		}
	}
	return true
}

func (w *WantFamousGeneralInfo) IsCanReset() bool {
	var n, n6 int
	for _, s := range w.PlayResult {
		if s > 0 {
			n++
		}
		if s == RewardDiceNo {
			n6++
		}
	}
	if n <= 0 || n6 == gamedata.WantGeneralDiceCount {
		return false
	}
	return true
}

func (w *WantFamousGeneralInfo) clear(p *Account) {
	w.CanPlayCountCurr = gamedata.GetWantGeneralCommonConfig().GetActivityTimes()
	w.CanFreeResetCountCurr = gamedata.GetWantGeneralCommonConfig().GetFreeResetTimes()
	if p.GuildProfile.InGuild() {
		bonus := guild.GetModule(p.AccountID.ShardId).GetGuildScienceBonus(
			p.GuildProfile.GuildUUID, p.AccountID.String(), gamedata.GST_WantGeneral)
		w.CanPlayCountCurr += uint32(bonus[0])
		w.CanFreeResetCountCurr += uint32(bonus[1])
	}
	w.CanPlayCountTotal = w.CanPlayCountCurr
	w.CanFreeResetCountTotal = w.CanFreeResetCountCurr
	w.DailyAllAward = 0
}

func (w *WantFamousGeneralInfo) DebugReset(p *Account) {
	w.clear(p)
}

func (w *WantFamousGeneralInfo) addDayPlayTimes(now_time int64) {
	if now_time >= w.DayPlayTimesRefreshTime {
		w.DayPlayTimesRefreshTime = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_time), now_time)
		w.DayPlayTimes = 0
	}
	w.DayPlayTimes++
}
