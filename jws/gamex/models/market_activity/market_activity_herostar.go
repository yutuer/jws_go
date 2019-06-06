package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *PlayerMarketActivitys) OnHeroStar(acid string, heroId int, starLevel int, now_t int64) bool {
	ma.UpdateMarketActivity(acid, now_t)
	typ := uint32(gamedata.ActHeroStar)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return false
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Star_Hero) {
		return false
	}
	isUpdate := false
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
		heroStarReq := subCfg[1].GetFCValue2()
		if uint32(heroId) != heroStarReq {
			continue

		}
		isUpdate = true
		pa.Value = int64(starLevel)
		//整理客户端显示state
		for i, c := range subCfg {
			if uint32(starLevel) >= c.GetFCValue1() && pa.State[i-1] == MA_ST_INIT {
				pa.State[i-1] = MA_ST_ACT
			}
		}
		logs.Debug("MarketActivitys OnHeroStar %d %d %v", heroId, starLevel, pa)
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true
}

func (ma *PlayerMarketActivitys) UpdateOnHeroStar(acid string, now_t int64) {
	typ := uint32(gamedata.ActHeroStar)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Star_Hero) {
		return
	}
	isUpdate := false
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
		isUpdate = true
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		if subCfg == nil {
			continue
		}
		//整理客户端显示state
		for i, c := range subCfg {
			if uint32(pa.Value) >= c.GetFCValue1() && pa.State[i-1] == MA_ST_INIT {
				pa.State[i-1] = MA_ST_ACT
			}
		}
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
}
