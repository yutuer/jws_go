package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *PlayerMarketActivitys) ClearExchangePropInfo(acid string, now_t int64, activityID uint32) {
	_pa := ma._preCheck(acid, gamedata.ActExchangeShop, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_value_ExchangeShop) {
		return
	}
	for _, pa := range _pa {
		if pa.ActivityId == activityID {
			pa.TmpValue = make([]int64, len(gamedata.GetHotDatas().HotExchangeShopData.GetExchangePropShowData(activityID)))
		}
	}
}

func (ma *PlayerMarketActivitys) UpdateOnExchangeShopProp(acid string, now_t int64, activityID uint32, index uint32) {
	_pa := ma._preCheck(acid, gamedata.ActExchangeShop, now_t)
	if _pa == nil || len(_pa) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_value_ExchangeShop) {
		return
	}
	for _, pa := range _pa {
		if pa.ActivityId == activityID {
			pa.updateExchangeValue(index, 1)
		}
	}
}

func (pma *PlayerMarketActivity) updateExchangeValue(index uint32, value int) {
	if int(index) > len(pma.TmpValue) {
		logs.Error("exchange shop index error")
		return
	}
	pma.TmpValue[index-1]++
}

func (ma *PlayerMarketActivitys) GetExchangeValue(index uint32, acid string, now_t int64, activityID uint32) int64 {
	_pa := ma._preCheck(acid, gamedata.ActExchangeShop, now_t)
	if _pa == nil || len(_pa) <= 0 {
		logs.Error("exchange activity not act")
		return 0
	}
	account, _ := db.ParseAccount(acid)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_value_ExchangeShop) {
		logs.Error("exchange activity not act")
		return 0
	}
	for _, pa := range _pa {
		if pa.ActivityId == activityID {
			return pa.getExchangeValue(index)
		}
	}
	logs.Error("fatal error for exchange activity")
	return 0
}

func (pma *PlayerMarketActivity) getExchangeValue(index uint32) int64 {
	if int(index) > len(pma.TmpValue) {
		logs.Error("exchange shop index error")
		return 0
	}
	return pma.TmpValue[index-1]
}
