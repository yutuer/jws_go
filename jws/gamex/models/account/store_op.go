package account

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/models/store"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerStores struct {
	store.PlayerStores
}

//NewPlayerStores ..
func NewPlayerStores(account db.Account) PlayerStores {
	AccountID := account

	mystore := PlayerStores{
		PlayerStores: *store.NewPlayerStores(AccountID),
	}

	return mystore
}

//Buy ..
func (p *PlayerStores) Buy(storeID, blankID, countP uint32, a *Account, sync interfaces.ISyncRspWithRewards) (
	res bool, warnCode uint32, chg map[uint32]bool) {
	count := countP
	chg = map[uint32]bool{}

	corp := a.Profile.GetCorp()
	lv := corp.GetLvlInfo()
	acid := a.AccountID.String()

	if count <= 0 {
		logs.SentryLogicCritical(acid, "store count <= 0 by id %d.", storeID)
		return false, 0, chg
	}

	store := p.GetStore(storeID)
	if store == nil {
		logs.SentryLogicCritical(acid, "store nil by id %d.", storeID)
		return false, 0, chg
	}

	blank := store.GetBlank(blankID)
	if blank == nil {
		logs.SentryLogicCritical(acid, "blank nil by %d:%d",
			storeID, blankID)
		return false, 0, chg

	}

	goodCfg := gamedata.GetStoreGoodCfg(blank.GoodIndex)
	if goodCfg.GoodCount <= blank.Count {
		logs.Warn("blank HasSell by %d:%d %d <= %d",
			storeID, blankID, goodCfg.GoodCount, blank.Count)
		return false, errCode.ClickTooQuickly, chg
	}

	if blank.Count+count > goodCfg.GoodCount {
		logs.Warn("count > blank.Count Err by %d:%d %d + %d > %d",
			storeID, blankID, blank.Count, count, goodCfg.GoodCount)
		// return false, errCode.ClickTooQuickly, chg
		count = goodCfg.GoodCount - blank.Count
	}

	giveData := gamedata.CostData{}
	if nil != blank.EquipData {
		giveData.AddItemWithData(goodCfg.GoodID, *blank.EquipData, count)
	} else {
		giveData.AddItem(goodCfg.GoodID, count)
	}
	if giveData.HasEquip {
		// 检查装备物品数量
		if a.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
			logs.SentryLogicCritical(acid, "store buy bag full for equip %s %d.", a.AccountID.String(), storeID)
			return false, 0, chg
		}
	}
	if giveData.HasJade {
		if a.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
			logs.SentryLogicCritical(acid, "store buy bag full for jade %s %d.", a.AccountID.String(), storeID)
			return false, 0, chg
		}
	}

	g := &GiveGroup{}
	g.AddCostData(&giveData)
	if !g.IsCanAddItem(a) {
		return false, errCode.AddItemFail_MaxCount, chg
	}

	costData := gamedata.CostData{}
	costData.AddItem(goodCfg.PriceTyp, goodCfg.PriceCount*count)
	cost := &CostGroup{}
	if !cost.AddCostData(a, &costData) {
		logs.SentryLogicCritical(acid, "store CostPrice Add Err by %d - %d %v.",
			storeID, blankID, goodCfg)
		return false, 0, chg
	}

	if !cost.CostBySync(a, sync, fmt.Sprintf("BuyIn%s", helper.StoreString(storeID))) {
		logs.SentryLogicCritical(acid, "store CostPrice Err by %d - %d %v.",
			storeID, blankID, *blank)
		return false, 0, chg
	}

	blank.Count += count

	if !g.GiveBySyncAuto(a, sync, fmt.Sprintf("%sBuy", helper.StoreString(storeID))) {
		logs.SentryLogicCritical(acid, "store Give Err by %d - %d %v.",
			storeID, blankID, *blank)
		return false, 0, chg
	}
	// logiclog
	logiclog.LogStoreBuy(acid, a.Profile.GetCurrAvatar(), a.Profile.GetCorp().GetLvlInfo(),
		a.Profile.ChannelId, helper.StoreString(storeID),
		goodCfg.GoodID, count, goodCfg.PriceTyp, goodCfg.PriceCount,
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")

	// 后刷新 在刷新之前点击的购买界面，在刷新之后点击购买，买到的是刷新之前对应格子的物品
	chg = p.Update(acid, a.Profile.GetProfileNowTime(), lv, a.rander)

	return true, 0, chg
}

//Refresh ..
func (p *PlayerStores) Refresh(storeID uint32, a *Account, sync helper.ISyncRsp) (bool, map[uint32]bool) {
	lv := a.Profile.GetCorp().GetLvlInfo()
	now := a.Profile.GetProfileNowTime()
	acid := a.AccountID.String()
	chg := p.Update(acid, now, lv, a.rander)

	store := p.GetStore(storeID)
	if store == nil {
		logs.SentryLogicCritical(acid, "store nil by id %d.", storeID)
		return false, chg
	}

	vip := a.Profile.GetVip().V
	refreshLimit := gamedata.GetStoreManualRefreshLimit(storeID, vip)

	if store.ManualRefreshCount >= refreshLimit {
		logs.Warn("RefreshStore Num limit by id %d, %d >= %d.", storeID, store.ManualRefreshCount, refreshLimit)
		return false, chg
	}

	c := &CostGroup{}
	costTyp, costNum := gamedata.GetStoreManualRefreshCost(storeID, store.ManualRefreshCount)
	if "" == costTyp || costNum < 0 || !c.AddItem(a, costTyp, costNum) {
		logs.SentryLogicCritical(acid, "RefreshStore Cost AddHc err by id %d, %s:%d.", storeID, costTyp, costNum)
		return false, chg
	}

	if !c.CostBySync(a, sync, fmt.Sprintf("%sRefresh", helper.StoreString(storeID))) {
		logs.SentryLogicCritical(acid, "RefreshStore Cost err by id %d.", storeID)
		return false, chg
	}

	store.ManualRefresh(acid, now, lv, a.GetRand())
	chg[storeID] = true

	return true, chg
}
