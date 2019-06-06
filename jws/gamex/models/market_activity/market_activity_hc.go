package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *PlayerMarketActivitys) OnPay(acid string, hcBuy, now_t int64) {
	ma.onPay(acid, hcBuy, now_t)
	ma.onPayPreDay(acid, hcBuy, now_t)
}

func (ma *PlayerMarketActivitys) onPayPreDay(acid string, hcBuy, now_t int64) {
	_pa := ma._preCheck(acid, gamedata.ActPayPreDay, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Query_day) {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity
		simpleCfg := activityCfg.GetActivitySimpleInfoById(pa.ActivityId)
		if simpleCfg == nil {
			continue
		}
		diffDay := gamedata.GetCommonDayDiff(simpleCfg.StartTime, now_t)
		pa._TmpValueCheck(int(diffDay))
		pa.TmpValue[int(diffDay)] += int64(hcBuy)
		pa.LastUpdateTime = now_t

		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		hcBuyReq := subCfg[1].GetFCValue1() // 这里默认充值累计多少天的FCValue1的参数都一样
		if pa.TmpValue[int(diffDay)] >= int64(hcBuyReq) {
			var value int64
			for _, v := range pa.TmpValue {
				if v >= int64(hcBuyReq) {
					value++
				}
			}
			if value != pa.Value {
				pa.Value = value
			}
			// 整理客户端显示state
			for i, s := range pa.State {
				if i < int(value) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}
		logs.Debug("MarketActivitys onPayPreDay %d %v", hcBuy, pa)
	}

}

func (ma *PlayerMarketActivitys) UpdateonPayPreDay(acid string, now_t int64) {
	_pa := ma._preCheck(acid, gamedata.ActPayPreDay, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Query_day) {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity
		simpleCfg := activityCfg.GetActivitySimpleInfoById(pa.ActivityId)
		if simpleCfg == nil {
			continue
		}
		diffDay := gamedata.GetCommonDayDiff(simpleCfg.StartTime, now_t)
		pa._TmpValueCheck(int(diffDay))

		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		if subCfg == nil {
			continue
		}
		hcBuyReq := subCfg[1].GetFCValue1() // 这里默认充值累计多少天的FCValue1的参数都一样
		if pa.TmpValue[int(diffDay)] >= int64(hcBuyReq) {
			var value int64
			for _, v := range pa.TmpValue {
				if v >= int64(hcBuyReq) {
					value++
				}
			}
			if value != pa.Value {
				pa.Value = value
			}
			// 整理客户端显示state
			for i, s := range pa.State {
				if i < int(value) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}

	}

}

// 累计充值和日累计充值是一样的逻辑,所以合并处理
func (ma *PlayerMarketActivitys) onPay(acid string, hcBuy, now_t int64) {
	_pa1 := ma._preCheck(acid, gamedata.ActPay, now_t)
	_pa2 := ma._preCheck(acid, gamedata.ActDayPay, now_t)
	_pa := append(_pa2, _pa1[:]...)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity

		pa.Value += hcBuy
		// 整理客户端显示state
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		for i, s := range pa.State {
			if _cfg, ok := subCfg[uint32(i+1)]; ok {
				if pa.Value >= int64(_cfg.GetFCValue1()) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}
		logs.Debug("MarketActivitys onPay %d %v", hcBuy, pa)
	}

}
func (ma *PlayerMarketActivitys) UpdateOnPay(acid string, now_t int64) {
	_pa1 := ma._preCheck(acid, gamedata.ActPay, now_t)
	_pa2 := ma._preCheck(acid, gamedata.ActDayPay, now_t)
	_pa := append(_pa2, _pa1[:]...)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity

		// 整理客户端显示state
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		if subCfg == nil {
			continue
		}
		for i, s := range pa.State {
			if _cfg, ok := subCfg[uint32(i+1)]; ok {
				if pa.Value >= int64(_cfg.GetFCValue1()) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}

	}

}

// 日累计消费X钻和累计消费X钻逻辑相同,一块处理
func (ma *PlayerMarketActivitys) OnHcCost(acid string, hc, now_t int64) bool {
	_pa1 := ma._preCheck(acid, gamedata.ActHcCost, now_t)
	_pa2 := ma._preCheck(acid, gamedata.ActDayHcCost, now_t)
	_pa := append(_pa1, _pa2[:]...)
	if _pa == nil || len(_pa) <= 0 {
		return false
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity

		pa.Value += hc
		// 整理客户端显示state
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		for i, s := range pa.State {
			if _cfg, ok := subCfg[uint32(i+1)]; ok {
				if pa.Value >= int64(_cfg.GetFCValue1()) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}

		logs.Debug("MarketActivitys OnHcCost %d %v", hc, pa)
	}
	return true
}

func (ma *PlayerMarketActivitys) UpdateOnHcCost(acid string, now_t int64) {
	_pa1 := ma._preCheck(acid, gamedata.ActHcCost, now_t)
	_pa2 := ma._preCheck(acid, gamedata.ActDayHcCost, now_t)
	_pa := append(_pa1, _pa2[:]...)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity
		// 整理客户端显示state
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		if subCfg == nil {
			continue
		}
		for i, s := range pa.State {
			if _cfg, ok := subCfg[uint32(i+1)]; ok {
				if pa.Value >= int64(_cfg.GetFCValue1()) {
					if s == MA_ST_INIT {
						pa.State[i] = MA_ST_ACT
					}
				} else {
					break
				}
			}
		}

	}
	return

}

func (ma *PlayerMarketActivitys) _preCheck(acid string, typ int, now_t int64) []*PlayerMarketActivity {
	ma.UpdateMarketActivity(acid, now_t)
	_pa := ma.getActByTyp(uint32(typ))
	ret := make([]*PlayerMarketActivity, 0, 4)
	if _pa == nil || len(_pa) <= 0 {
		return nil
	}
	//gmtoos 检查活动是否开启
	if !ma._hotValueCheck(acid, typ) {
		return nil
	}
	for _, pa := range _pa {
		if pa.Balanced {
			continue
		}

		activityCfg := gamedata.GetHotDatas().Activity
		simpleCfg := activityCfg.GetActivitySimpleInfoById(pa.ActivityId)
		if simpleCfg == nil {
			continue
		}
		if now_t >= simpleCfg.EndTime || now_t < simpleCfg.StartTime {
			continue
		}
		ret = append(ret, pa)
	}

	return ret
}

func (ma *PlayerMarketActivitys) GetCurDayHcPay(acid string, now_t int64) int64 {
	_pa := ma._preCheck(acid, gamedata.ActPayPreDay, now_t)
	if _pa != nil && len(_pa) > 0 {
		for _, pa := range _pa {
			diffDay := gamedata.GetCommonDayDiff(pa.LastUpdateTime, now_t)
			if diffDay == 0 {
				activityCfg := gamedata.GetHotDatas().Activity
				simpleCfg := activityCfg.GetActivitySimpleInfoById(pa.ActivityId)
				if simpleCfg == nil {
					continue
				}
				if now_t >= simpleCfg.EndTime || now_t < simpleCfg.StartTime {
					continue
				}
				diffDay := gamedata.GetCommonDayDiff(simpleCfg.StartTime, now_t)
				return pa._getTmpValue(int(diffDay))
			}
		}
	}
	return 0
}

func (ma *PlayerMarketActivitys) _hotValueCheck(acid string, typ int) bool {
	account, _ := db.ParseAccount(acid)
	if typ == gamedata.ActPay {
		if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Query) {
			return false
		}
	}
	if typ == gamedata.ActHcCost {
		if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Buy) {
			return false
		}
	}
	if typ == gamedata.ActDayPay {
		if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Day_Total_Query) {
			return false
		}
	}
	if typ == gamedata.ActDayHcCost {
		if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Day_Total_Buy) {
			return false
		}
	}
	return true

}
