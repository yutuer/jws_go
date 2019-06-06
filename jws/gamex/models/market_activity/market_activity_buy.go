package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *PlayerMarketActivitys) OnBuy(acid string, buyType int, now_t int64) {
	ma.UpdateMarketActivity(acid, now_t)
	typ := uint32(gamedata.ActBuy)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Buy_Resource) {
		return
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

		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		buyTypReq := subCfg[1].GetFCValue2()
		if buyType != int(buyTypReq) {
			continue
		}

		pa.Value += 1

		// 整理客户端显示state
		for i, c := range subCfg {
			if uint32(pa.Value) >= c.GetFCValue1() && pa.State[i-1] == MA_ST_INIT {
				pa.State[i-1] = MA_ST_ACT
			}
		}
		logs.Debug("MarketActivitys OnGameMode %d %v", buyType, pa)
	}
}
func (ma *PlayerMarketActivitys) UpdateOnBy(acid string, now_t int64) {
	typ := uint32(gamedata.ActBuy)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Total_Buy_Resource) {
		return
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

		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		if subCfg == nil {
			continue
		}
		// 整理客户端显示state
		for i, c := range subCfg {
			if uint32(pa.Value) >= c.GetFCValue1() && pa.State[i-1] == MA_ST_INIT {
				pa.State[i-1] = MA_ST_ACT
			}
		}

	}

}
