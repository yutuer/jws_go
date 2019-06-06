package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) RefreshShop(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ShopId uint32 `codec:"s"`
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerAttr/RefreshShopResponse",
		r.RawBytes,
		req, resp, p)

	if p.StoreProfile.GetShop(req.ShopId).ShopRefresh(p.Profile.GetProfileNowTime()) {
		resp.OnChangeShopChange(req.ShopId)
		resp.mkInfo(p)
	}
	return rpcSuccess(resp)
}

func (p *Account) BuyInShop(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ShopId uint32 `codec:"s"`
		GoodId string `codec:"g"`
		Count  int    `codec:"c"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}
	initReqRsp(
		"PlayerAttr/BuyInShopResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Shop_Or_Good_Nor_Found
		Err_Good_Cfg_Not_Found
		Err_Good_ValidTime_Cfg_Wrong
		Err_Good_Not_ValidTime
		Err_Good_Vip_Not_Satisfy
		Err_Good_Count_Not_Enough
		Err_Bag_Full_For_Equip
		Err_HC_Not_Enough
		Err_Give_Item
		Err_Add_Times
	)

	// 作弊
	if req.Count < 0 || req.Count > uutil.CHEAT_INT_MAX {
		return rpcErrorWithMsg(resp, 99, fmt.Sprintf("BuyInShop count cheat"))
	}

	// 刷新商店
	now_time := p.Profile.GetProfileNowTime()
	shop := p.StoreProfile.GetShop(req.ShopId)
	shop.ShopRefresh(now_time)

	if !gamedata.IsGoodInShop(req.ShopId, req.GoodId) {
		return rpcErrorWithMsg(resp, Err_Shop_Or_Good_Nor_Found,
			fmt.Sprintf("Err_Shop_Or_Good_Nor_Found shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}
	goodCfg := gamedata.GetGoodCfg(req.GoodId)
	if goodCfg == nil {
		return rpcErrorWithMsg(resp, Err_Good_Cfg_Not_Found,
			fmt.Sprintf("Err_Good_Cfg_Not_Found shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}

	// 商品是否有效期内
	ok, err := gamedata.IsGoodTimeValid(goodCfg, now_time)
	if err != nil {
		return rpcErrorWithMsg(resp, Err_Good_ValidTime_Cfg_Wrong,
			fmt.Sprintf("Err_Good_ValidTime_Cfg_Wrong shop %d good %s count %d err %v",
				req.ShopId, req.GoodId, req.Count, err))
	}
	if !ok {
		return rpcErrorWithMsg(resp, Err_Good_Not_ValidTime,
			fmt.Sprintf("Err_Godd_Not_ValidTime shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}

	// vip限制
	if goodCfg.GetVIPLimit() > 0 {
		v, _ := p.Profile.Vip.GetVIP()
		if v < goodCfg.GetVIPLimit() {
			return rpcErrorWithMsg(resp, Err_Good_Vip_Not_Satisfy,
				fmt.Sprintf("Err_Good_Vip_Not_Satisfy shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
		}
	}

	// 剩余数量检查
	if goodCfg.GetCountLimit() > 0 {
		if int(goodCfg.GetCountLimit()) < shop.GetGoodUseTimes(req.GoodId)+req.Count {
			logs.Warn("BuyInShop Err_Good_Count_Not_Enough shop %d good %s count %d", req.ShopId, req.GoodId, req.Count)
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
	}
	// 检查包裹是否满
	giveData := gamedata.CostData{}
	giveData.AddItem(goodCfg.GetItemID(), uint32(req.Count))
	if giveData.HasEquip && p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
		return rpcErrorWithMsg(resp, Err_Bag_Full_For_Equip,
			fmt.Sprintf("Err_Bag_Full_For_Equip shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}
	if giveData.HasJade && p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcErrorWithMsg(resp, Err_Bag_Full_For_Equip,
			fmt.Sprintf("Err_Bag_Full_For_Jade shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}

	// 扣钱
	costData := gamedata.CostData{}
	moneyCost := goodCfg.GetCurrentPrice() * uint32(req.Count)
	costData.AddItem(goodCfg.GetCoinItemID(), moneyCost)
	if !account.CostBySync(p.Account, &costData, resp, fmt.Sprintf("BuyIn%s", helper.ShopString(req.ShopId))) {
		logs.Warn("Err_HC_Not_Enough shop %d good %s count %d", req.ShopId, req.GoodId, req.Count)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 给物品
	if !account.GiveBySync(p.Account, &giveData, resp, fmt.Sprintf("BuyIn%s", helper.ShopString(req.ShopId))) {
		return rpcErrorWithMsg(resp, Err_Give_Item,
			fmt.Sprintf("Err_Give_Item shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}
	if !shop.AddGoodUseTimes(req.GoodId, req.Count) {
		return rpcErrorWithMsg(resp, Err_Add_Times,
			fmt.Sprintf("Err_Add_Times shop %d good %s count %d", req.ShopId, req.GoodId, req.Count))
	}

	// logiclog
	logiclog.LogShopBuy(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, helper.ShopString(req.ShopId),
		goodCfg.GetItemID(), uint32(req.Count), goodCfg.GetCoinItemID(), moneyCost,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	resp.OnChangeShopChange(req.ShopId)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type ShopToClient struct {
	ShopId          uint32 `codec:"s"`
	NextRefreshTime int64  `codec:"nt"`
	GoodNum         int    `codec:"gn"`
}

type GoodToClient struct {
	GoodId   string `codec:"g"`
	UseTimes int    `codec:"ut"`
}

func (p *Account) getAllShopsInfoForClient() ([]ShopToClient, []GoodToClient) {
	resShops := make([]ShopToClient, len(p.StoreProfile.Shops))
	resGoods := make([]GoodToClient, 0, 30)
	for i, shop := range p.StoreProfile.Shops {
		var goodNum int
		for _, good := range shop.Goods {
			if !gamedata.IsGoodInShop(shop.ShopTyp, good.GoodId) {
				continue
			}
			resGoods = append(resGoods, GoodToClient{good.GoodId, good.UseTimes})
			goodNum++
		}
		nextRefTime := util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(shop.LastTime), shop.LastTime)
		resShops[i] = ShopToClient{shop.ShopTyp, nextRefTime, goodNum}
	}
	return resShops, resGoods
}

func (p *Account) getShopsInfoForClient(shopId uint32) ([]ShopToClient, []GoodToClient) {
	resShops := make([]ShopToClient, 1)
	resGoods := make([]GoodToClient, 0, 30)
	if shopId < uint32(len(p.StoreProfile.Shops)) {
		shop := p.StoreProfile.Shops[shopId]
		var goodNum int
		for _, good := range shop.Goods {
			if !gamedata.IsGoodInShop(shop.ShopTyp, good.GoodId) {
				continue
			}
			resGoods = append(resGoods, GoodToClient{good.GoodId, good.UseTimes})
			goodNum++
		}
		nextRefTime := util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(shop.LastTime), shop.LastTime)
		resShops[0] = ShopToClient{shop.ShopTyp, nextRefTime, goodNum}
	}
	return resShops, resGoods
}
