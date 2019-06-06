package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PrivilegeBuy struct {
	Id       int32  `json:"id"`
	BuyTimes uint32 `json:"ts"`
}

type PrivilegeBuyInfo struct {
	BuyTimes []PrivilegeBuy `json:"bt"`
}

type privilegeBuyToClient struct {
	Id       int32  `codec:"id"`
	BuyTimes uint32 `codec:"ts"`
}

const (
	_ = iota
	PRIVILEGE_ID_NOT_FOUND
	PRIVILEGE_VIP_NOT_FOUND
	VIP_NOT_ENOUGH
	BUY_TIMES_FULL
	HC_NOT_RNOUGH
	NO_CFG_AWARD
	GIVE_FAIL
)

func (p *PrivilegeBuyInfo) BuyByPrivilege(a *Account, privilegeId int32, sync interfaces.ISyncRspWithRewards) (
	bool, uint32, uint32) {
	cfgTimes, hcCost, found := gamedata.GetPrivilegeInfo(privilegeId)
	if !found {
		return false, PRIVILEGE_ID_NOT_FOUND, 0
	}
	// vip里找特权id，TODO, 目前特权只在vip这有，有了新需求要重构
	vip, found := gamedata.Privilege2Vip(privilegeId)
	if !found {
		return false, PRIVILEGE_VIP_NOT_FOUND, 0
	}
	// 检查vip是否够
	curVip, _ := a.Profile.GetVip().GetVIP()
	if vip > curVip {
		return false, VIP_NOT_ENOUGH, 0
	}
	// 检查购买次数是否够
	pvInfo := &p.BuyTimes[privilegeId-1]
	if pvInfo.Id > 0 && pvInfo.BuyTimes >= cfgTimes {
		return false, 0, errCode.ClickTooQuickly
	}
	// 花钱
	if !a.Profile.GetHC().UseHcGiveFirst(a.AccountID.String(), int64(hcCost), a.Profile.GetProfileNowTime(), "PrivilegeBuy") {
		return false, HC_NOT_RNOUGH, 0
	}
	// 给物品
	awards := gamedata.GetPrivilegeAward(privilegeId)
	if len(awards) <= 0 {
		return false, NO_CFG_AWARD, 0
	}
	// 加次数
	if pvInfo.Id > 0 {
		pvInfo.BuyTimes = pvInfo.BuyTimes + 1
	} else {
		pvInfo.Id = privilegeId
		pvInfo.BuyTimes = 1
	}
	g := GiveGroup{}
	c := gamedata.CostData{}
	for _, award := range awards {
		item_data := gamedata.MakeItemData(a.AccountID.String(), a.GetRand(), award.ItemId)
		if item_data != nil {
			c.AddItemWithData(award.ItemId, *item_data, award.Count)
		} else {
			c.AddItemWithData(award.ItemId, gamedata.BagItemData{}, award.Count)
		}
	}
	g.AddCostData(&c)
	if !g.GiveBySyncAuto(a, sync, "privilegebuy") {
		return false, GIVE_FAIL, 0
	}
	return true, 0, 0
}

func (pvgb *PrivilegeBuyInfo) GetAllSyncInfo() []privilegeBuyToClient {
	res := make([]privilegeBuyToClient, 0, len(pvgb.BuyTimes))
	for _, v := range pvgb.BuyTimes {
		if v.Id > 0 {
			res = append(res, privilegeBuyToClient{
				Id:       v.Id,
				BuyTimes: v.BuyTimes,
			})
		}
	}
	return res
}

func (pvgb *PrivilegeBuyInfo) InitPrivilegeInfo() {
	logs.Debug("old bugtimes %d, new buytimes %d", len(pvgb.BuyTimes), gamedata.GetPrivilegeCfgCount())
	if len(pvgb.BuyTimes) <= 0 {
		pvgb.BuyTimes = make([]PrivilegeBuy, gamedata.GetPrivilegeCfgCount())
	} else if len(pvgb.BuyTimes) < gamedata.GetPrivilegeCfgCount() {
		newBuyTimes := make([]PrivilegeBuy, gamedata.GetPrivilegeCfgCount())
		for i, x := range pvgb.BuyTimes {
			newBuyTimes[i] = x
		}
		pvgb.BuyTimes = make([]PrivilegeBuy, gamedata.GetPrivilegeCfgCount())
		pvgb.BuyTimes = newBuyTimes
	}
}
