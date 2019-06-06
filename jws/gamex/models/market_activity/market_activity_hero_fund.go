package market_activity

import (
	"strconv"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	updateHeroFund_iap = iota
	updateHeroFund_gs
)

const (
	heroFund_ok = iota
	heroFund_none
	heroFund_gm_close
)

func (ma *PlayerMarketActivitys) OnHeroFundByPay(acid string, iapId int, gs int, now_t int64) bool {
	if _, ok := gamedata.GetHotDatas().HeroFoundConfig.ActivityIap[int(iapId)]; !ok {
		iapId -= uutil.IAPID_ONESTORE_2_GOOGLE
		if _, ok := gamedata.GetHotDatas().HeroFoundConfig.ActivityIap[int(iapId)]; !ok {
			return false
		}
	}
	return ma.onHeroFundByPay(acid, int(iapId), gs, now_t)
}

// 充值调用
func (ma *PlayerMarketActivitys) onHeroFundByPay(acid string, iapId int, gs int, now_t int64) bool {
	ma.UpdateMarketActivity(acid, now_t)

	activities, retCode := ma.getAllHeroFund(acid, now_t)
	if retCode > 0 {
		logs.Debug("not find any hero fund, %d", retCode)
		return false
	}
	isUpdate := false
	for _, pa := range activities {
		if ok := ma.isHeroFundAvailable(acid, now_t, pa); !ok {
			continue
		}
		if ok := ma.setHeroFundParam(pa, updateHeroFund_iap, iapId); !ok {
			continue
		}
		if ok := ma.setHeroFundParam(pa, updateHeroFund_gs, gs); !ok {
			continue
		}
		isUpdate = true
		ma.updateSubHeroFund(pa)
		logs.Debug("MarketActivitys OnHeroFundByPay %d %d %v", iapId, gs, pa)
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true
}

func (ma *PlayerMarketActivitys) OnHeroFundByGs(acid string, gs int, now_t int64) bool {
	ma.UpdateMarketActivity(acid, now_t)
	activities, retCode := ma.getAllHeroFund(acid, now_t)
	if retCode > 0 {
		logs.Debug("not find any hero fund, %d", retCode)
		return false
	}
	isUpdate := false
	for _, pa := range activities {
		if ok := pa.isActAvailable(acid, now_t); !ok {
			continue
		}
		if pa.Value == 0 {
			continue // 没有充值, 不用更新战力
		}
		if ok := ma.setHeroFundParam(pa, updateHeroFund_gs, gs); !ok {
			continue
		}
		isUpdate = true
		ma.updateSubHeroFund(pa)
		logs.Debug("MarketActivitys OnHeroFundByGs %d %v", gs, pa)
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true
}

func (ma *PlayerMarketActivitys) UpdateOnHeroFund(acid string, now_t int64) bool {
	activities, retCode := ma.getAllHeroFund(acid, now_t)
	if retCode > 0 {
		logs.Debug("not find any hero fund, %d", retCode)
		return false
	}
	isUpdate := false
	for _, pa := range activities {
		if ok := ma.isHeroFundAvailable(acid, now_t, pa); !ok {
			continue
		}
		isUpdate = true
		ma.updateSubHeroFund(pa)
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true
}

func (ma *PlayerMarketActivitys) getAllHeroFund(acid string, now_t int64) ([]*PlayerMarketActivity, int) {
	_pa := ma.getActByTypeRange(gamedata.ActHeroFund_Begin, gamedata.ActHeroFund_End)
	if _pa == nil || len(_pa) <= 0 {
		return nil, heroFund_none
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_HERO_FUND) {
		return nil, heroFund_gm_close
	}
	return _pa, heroFund_ok
}

func (ma *PlayerMarketActivitys) isHeroFundAvailable(acid string, now_t int64, pa *PlayerMarketActivity) bool {
	if pa.Balanced {
		return false
	}
	activityCfg := gamedata.GetHotDatas().Activity
	simpleCfg := activityCfg.GetActivitySimpleInfoById(pa.ActivityId)
	if simpleCfg == nil {
		return false
	}
	if now_t >= simpleCfg.EndTime || now_t < simpleCfg.StartTime {
		return false
	}
	return true
}

func (ma *PlayerMarketActivitys) setHeroFundParam(pa *PlayerMarketActivity, paramType, paramValue int) bool {
	if paramType == updateHeroFund_iap {
		activityCfg := gamedata.GetHotDatas().Activity
		subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)
		flagCfg, ok := subCfg[1]
		if !ok {
			logs.Error("hero fund: bad config, no sub cfg 1, %d", pa.ActivityId)
			return false
		}
		if ma.isIapMatch(paramValue, flagCfg.GetSFCValue1(), flagCfg.GetSFCValue2()) {
			logs.Debug("hero fund: iap not match, activityId=%d,iapId=%d", pa.ActivityId, paramValue)
			return false
		}
		pa.Value = int64(paramValue)
	} else if paramType == updateHeroFund_gs {
		if len(pa.TmpValue) == 0 {
			pa.TmpValue = make([]int64, 1)
		}
		if int64(paramValue) > pa.TmpValue[0] {
			pa.TmpValue[0] = int64(paramValue)
		}
	}
	return true
}

func (ma *PlayerMarketActivitys) isIapMatch(iapId int, cfg1, cfg2 string) bool {
	if cfg1 != "" {
		if cfgInt, err := strconv.ParseInt(cfg1, 10, 32); err != nil && iapId == int(cfgInt) {
			return true
		}
	}
	if cfg2 != "" {
		tmp := strings.Split(cfg2, ",")
		for _, temp_iap := range tmp {
			if cfgInt, err := strconv.ParseInt(temp_iap, 10, 32); err != nil && iapId == int(cfgInt) {
				return true
			}
		}
	}
	return false
}

func (ma *PlayerMarketActivitys) updateSubHeroFund(pa *PlayerMarketActivity) {
	if pa.Value == 0 {
		return
	}
	if len(pa.TmpValue) == 0 {
		return
	}
	gs := pa.TmpValue[0]
	activityCfg := gamedata.GetHotDatas().Activity
	subCfg := activityCfg.GetMarketActivitySubConfig(pa.ActivityId)

	for i, c := range subCfg {
		if uint32(gs) >= c.GetFCValue1() && pa.State[i-1] == MA_ST_INIT {
			pa.State[i-1] = MA_ST_ACT
		}
	}
}
