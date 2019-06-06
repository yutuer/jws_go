package market_activity

import (
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func (ma *PlayerMarketActivitys) OnOnlyPay(acid string, iapId uint32, now_t int64) bool {
	ma.UpdateMarketActivity(acid, now_t)
	typ := uint32(gamedata.ActOnlyPay)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return false
	}
	//account, _ := db.ParseAccount(acid)
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
		if len(pa.TmpValue) == 0 {
			pa.TmpValue = make([]int64, len(subCfg)*2)
		}

		pa._OnlypayTmpValueCheck(len(subCfg))

		for i, sub := range subCfg {
			otherIapId, _ := strconv.ParseInt(sub.GetSFCValue2(), 10, 64)
			if (iapId == sub.GetFCValue1() || iapId == sub.GetFCValue2() || iapId == uint32(otherIapId)) && iapId > 0 {
				maxIapCount, _ := strconv.ParseInt(sub.GetSFCValue1(), 10, 64)
				if pa.TmpValue[i*2-2] < maxIapCount {
					pa.TmpValue[i*2-2] += 1
				}
			}
		}
		// 整理客户端显示state
		for i, subCfg := range subCfg {
			ipaCount := pa.TmpValue[i*2-2]    //充值次数
			rewardCount := pa.TmpValue[i*2-1] //领奖次数
			maxIapCount, _ := strconv.ParseFloat(subCfg.GetSFCValue1(), 32)

			if ipaCount > rewardCount && rewardCount < int64(maxIapCount) {
				pa.State[i-1] = MA_ST_ACT
			} else if rewardCount >= int64(maxIapCount) {
				pa.State[i-1] = MA_ST_GOT
			} else {
				pa.State[i-1] = MA_ST_INIT
			}
		}
	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true
}

func (ma *PlayerMarketActivitys) UpdateOnlyPay(acid string, now_t int64) bool {
	typ := uint32(gamedata.ActOnlyPay)
	_pa := ma.getActByTyp(typ)
	if _pa == nil || len(_pa) <= 0 {
		return false
	}
	isUpdate := false
	for _, pa := range _pa {
		//if pa.Balanced {
		//	continue
		//}

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
		pa._OnlypayTmpValueCheck(len(subCfg))
		// 整理客户端显示state
		for i, subCfg := range subCfg {
			ipaCount := pa.TmpValue[i*2-2]
			rewardCount := pa.TmpValue[i*2-1]
			maxIapCount, _ := strconv.ParseFloat(subCfg.GetSFCValue1(), 32)

			if ipaCount > rewardCount && rewardCount < int64(maxIapCount) {
				pa.State[i-1] = MA_ST_ACT
			} else if rewardCount >= int64(maxIapCount) {
				pa.State[i-1] = MA_ST_GOT
			} else {
				pa.State[i-1] = MA_ST_INIT
			}
		}

	}
	if isUpdate {
		ma.SyncObj.SetNeedSync()
	}
	return true

}
