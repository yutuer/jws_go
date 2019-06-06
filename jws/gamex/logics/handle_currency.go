package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

const Pay_Result_MonyerCat = "Moneycat Cost"

func addScChgHandle(a *Account) {
	a.AddHandle(events.NewHandler().WithScChg(func(isAdd bool, typ int, oldV, chgV int64, reason string) {
		aid := a.AccountID.String()
		cur_avatar := a.Profile.GetCurrAvatar()
		if isAdd {
			a.GetGiveCurrencyLog(aid, cur_avatar, a.Profile.GetCorp().GetLvlInfo(),
				a.Profile.ChannelId, reason, helper.SCString(typ), oldV, chgV,
				func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
		} else {
			logiclog.LogCostCurrency(aid, cur_avatar, a.Profile.GetCorp().GetLvlInfo(),
				a.Profile.ChannelId, reason, helper.SCString(typ), oldV, chgV, a.Profile.GetVipLevel(),
				func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}))
}

func addHcChgHandle(a *Account) {
	a.AddHandle(events.NewHandler().WithHcChg(func(isAdd bool, typ int, oldV, chgV int64, reason string) {
		aid := a.AccountID.String()
		cur_avatar := a.Profile.GetCurrAvatar()
		if isAdd {
			a.GetGiveCurrencyLog(aid, cur_avatar, a.Profile.GetCorp().GetLvlInfo(),
				a.Profile.ChannelId, reason, helper.HCString(typ), oldV, chgV,
				func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")

			// 48.每日购买钻石, P1次数, P2所在天数(0为所有历史钻石，1为指定天钻石)
			a.updateCondition(account.COND_TYP_7Day_BuyHC_Today,
				0, 0, "", "", nil)
			if helper.HC_From_Buy == typ && reason != Pay_Result_MonyerCat {
				a.Profile.GetMarketActivitys().OnPay(aid, chgV, a.Profile.GetProfileNowTime())
				a.Profile.GetMarketActivitys().SyncObj.SetNeedSync()
			}
		} else {
			logiclog.LogCostCurrency(aid, cur_avatar, a.Profile.GetCorp().GetLvlInfo(),
				a.Profile.ChannelId, reason, helper.HCString(typ), oldV, chgV, a.Profile.GetVipLevel(),
				func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")

			// hc消耗，每种hc都会触发一次event；游戏逻辑一般之关系hc总量变化，所以只关心HC_TYPE_COUNT
			if typ == gamedata.HC_TYPE_COUNT && reason != Pay_Result_MonyerCat {
				// 49.每日消费钻石, P1次数, P2所在天数
				a.updateCondition(account.COND_TYP_7Day_CostHC_Today,
					0, 0, "", "", nil)

				if a.Profile.GetMarketActivitys().OnHcCost(aid, chgV, a.Profile.GetProfileNowTime()) {
					a.Profile.GetMarketActivitys().SyncObj.SetNeedSync()
				}
				a.Profile.GetRedPacket7day().UpdateSaveHc(a.Profile.GetProfileNowTime(), chgV)
			}
		}
	}))
}
