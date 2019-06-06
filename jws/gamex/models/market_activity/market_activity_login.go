package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *PlayerMarketActivitys) OnLogin(acid string, now_t int64) {
	_pa := ma._preCheck(acid, gamedata.ActLogin, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Enter_day) {
		return
	}
	for _, pa := range _pa {
		activityCfg := gamedata.GetHotDatas().Activity

		diffDay := gamedata.GetCommonDayDiff(pa.LastUpdateTime, now_t)
		if diffDay > 0 {
			pa.Value++
			pa.LastUpdateTime = now_t
		}
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
		logs.Debug("MarketActivitys OnLogin %v", pa)
	}

}
func (ma *PlayerMarketActivitys) UpdateOnLogin(acid string, now_t int64) {
	_pa := ma._preCheck(acid, gamedata.ActLogin, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Enter_day) {
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
		logs.Debug("MarketActivitys OnLogin %v", pa)
	}

}

// gvg每日签到,
func (ma *PlayerMarketActivitys) UpdateGVGDailySignOnLogin(acid string, now_t int64) {
	activities := ma._preCheck(acid, gamedata.ActGvgDailySign, now_t)
	if activities == nil || len(activities) == 0 {
		logs.Debug("no more activities, %s", acid)
		return
	}

	for _, activity := range activities {
		logs.Debug("gve daily sign now %s, old %s", now_t, activity.LastUpdateTime)
		isSameDay := gamedata.IsSameDayCommon(now_t, activity.LastUpdateTime)
		if !isSameDay {
			activity.LastUpdateTime = now_t
			activity.State[0] = MA_ST_ACT
		}
	}
}
