package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ExchangeProp : 兑换商品道具
//
func (p *Account) ExchangePropHandler(req *reqMsgExchangeProp, resp *rspMsgExchangeProp) uint32 {
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_value_ExchangeShop) {
		return errCode.ActivityTimeOut
	}
	propData := gamedata.GetHotDatas().HotExchangeShopData.GetExchangePropData(uint32(req.ExchangeID),
		uint32(req.ActivityID))
	costData := &gamedata.CostData{}
	for _, item := range propData.GetNeed_Table() {
		costData.AddItem(item.GetItemID(), item.GetItemNum())
	}
	if !account.CostBySync(p.Account, costData, resp, "Exchange Shop Prop") {
		return errCode.RewardFail
	}
	giveData := &gamedata.CostData{}
	for _, item := range propData.GetLoot_Table() {
		item.GetLootTime()
		for t := uint32(0); t < item.GetLootTime(); t++ {
			reward := gamedata.GetRandWeightProp(item.GetLootGroupID())
			giveData.AddItem(reward.GetItemID(), reward.GetItemCount())
		}
	}
	logs.Debug("reward data: %v", giveData)
	if !account.GiveBySync(p.Account, giveData, resp, "Exchange Shop Prop") {
		return errCode.RewardFail
	}
	alreadyExchangeTimes := p.Profile.GetMarketActivitys().GetExchangeValue(uint32(req.ExchangeID), p.AccountID.String(),
		p.Profile.GetProfileNowTime(), uint32(req.ActivityID))
	if alreadyExchangeTimes > int64(propData.GetLimitTime()) {
		return errCode.CommonCountLimit
	}
	p.Profile.GetMarketActivitys().UpdateOnExchangeShopProp(p.AccountID.String(),
		p.Profile.GetProfileNowTime(), uint32(req.ActivityID), uint32(req.ExchangeID))
	resp.AlreadyExchangeTimes = p.Profile.GetMarketActivitys().GetExchangeValue(uint32(req.ExchangeID), p.AccountID.String(),
		p.Profile.GetProfileNowTime(), uint32(req.ActivityID))
	return 0
}

// GetExchangeShopInfo : 获取兑换商店信息
//
func (p *Account) GetExchangeShopInfoHandler(req *reqMsgGetExchangeShopInfo, resp *rspMsgGetExchangeShopInfo) uint32 {
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_value_ExchangeShop) {
		return errCode.ActivityTimeOut
	}
	propData := gamedata.GetHotDatas().HotExchangeShopData.GetExchangePropShowData(uint32(req.ActivityID))
	resp.ExchangePropInfo = make([][]byte, 0)
	for _, item := range propData {
		resp.ExchangePropInfo = append(resp.ExchangePropInfo, encode(p.genExchangePropInfo(item)))
	}
	if len(gamedata.GetHotDatas().HotExchangeShopData.GetCanAutoExchangePropData(uint32(req.ActivityID))) > 0 {
		resp.HasAutoProp = 1
	} else {
		resp.HasAutoProp = 0
	}
	return 0
}

func (p *Account) genExchangePropInfo(data *ProtobufGen.HOTSHOP) ExchangePropInfo {
	info := ExchangePropInfo{}
	info.ExchangeCount = make([]int64, 0)
	info.ExchangeID = make([]string, 0)
	for _, item := range data.GetNeed_Table() {
		info.ExchangeID = append(info.ExchangeID, item.GetItemID())
		info.ExchangeCount = append(info.ExchangeCount, int64(item.GetItemNum()))
	}
	info.ShowPropID = data.GetShowItemID()
	info.ShowPropCount = int64(data.GetShowItemNum())
	info.ExchangeLimitTimes = int64(data.GetLimitTime())
	info.ExchangePropIndex = int64(data.GetIndex())
	info.AlreadyExchangeTimes = p.Profile.GetMarketActivitys().GetExchangeValue(data.GetIndex(),
		p.AccountID.String(), p.Profile.GetProfileNowTime(), data.GetActivityID())
	return info
}
