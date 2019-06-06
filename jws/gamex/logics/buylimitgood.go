package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// BuyLimitGood : 购买限时商品
// 请求购买限时商品

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBuyLimitGood 购买限时商品请求消息定义
type reqMsgBuyLimitGood struct {
	Req
	GoodId   int64 `codec:"goodid"` // 商品ID
	BuyCount int64 `codec:"buy_count"`
}

// rspMsgBuyLimitGood 购买限时商品回复消息定义
type rspMsgBuyLimitGood struct {
	SyncRespWithRewards
}

// BuyLimitGood 购买限时商品: 请求购买限时商品
func (p *Account) BuyLimitGood(r servers.Request) *servers.Response {
	req := new(reqMsgBuyLimitGood)
	rsp := new(rspMsgBuyLimitGood)

	initReqRsp(
		"Attr/BuyLimitGoodRsp",
		r.RawBytes,
		req, rsp, p)

	// 作弊检查
	if req.BuyCount < 0 || req.BuyCount > uutil.CHEAT_INT_MAX {
		return rpcErrorWithMsg(rsp, 99, "BuyLimitGood BuyCount cheat")
	}

	errRsp := p.doBuyLimitGood(req.GoodId, req.BuyCount, rsp)
	if errRsp != nil {
		return errRsp
	}

	rsp.onChangeLimitShop()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) doBuyLimitGood(goodId, buyCount int64, rsp *rspMsgBuyLimitGood) *servers.Response {
	p.StoreProfile.LimitShop.CheckConsistence()

	goodConfig, ok := gamedata.GetHotDatas().LimitGoodConfig.GetLimitGoodConfig(goodId)
	if !ok {
		logs.Warn("buyLimitGood: fail to find limit good config, goodId = %d", goodId)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if !p.canBuyMoreLimitGood(goodId, buyCount, goodConfig.Item.GetTimesLimit()) {
		logs.Warn("buyLimitGood: has bought limit good, goodId = %d", goodId)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if p.isLimitGoodTimeOut(goodConfig) {
		logs.Warn("buyLimitGood: limit good is time out, goodId = %d", goodId)
		return rpcWarn(rsp, errCode.LimitGoodTimeOut)
	}
	if p.Profile.GetVipLevel() < goodConfig.Item.GetVIPLevel() {
		logs.Warn("buyLimitGood: vip level limit, goodId = %d, player vip Level =%d", goodId, p.Profile.GetVipLevel())
		return rpcWarn(rsp, errCode.ActivityNotValid)
	}
	p.recordBuyLimitGood(goodId, buyCount)
	if errCode := p.rewardOnBuyGoodLimit(goodConfig, int(buyCount), rsp); errCode != 0 {
		return rpcWarn(rsp, errCode)
	}
	return nil
}

// 购买的达到上限
func (p *Account) canBuyMoreLimitGood(goodId, buyCount int64, limitCount uint32) bool {
	boughtInfo := p.StoreProfile.LimitShop.HasBoughtGoods
	boughtCount := p.StoreProfile.LimitShop.HasBoughtCount
	if boughtInfo == nil {
		return true // 未购买过
	}
	for i, buyId := range boughtInfo {
		if buyId == int(goodId) {
			return boughtCount[i]+int(buyCount) <= int(limitCount)
		}
	}
	return true // 未购买过
}

func (p *Account) isLimitGoodTimeOut(config *gamedata.HotLimitGoodCfg) bool {
	now := p.GetProfileNowTime()
	return config.StartTime > now || now > (config.StartTime+int64(config.Duration))
}

func (p *Account) recordBuyLimitGood(goodId, buyCount int64) {
	if p.StoreProfile.LimitShop.HasBoughtGoods == nil {
		p.StoreProfile.LimitShop.HasBoughtGoods = make([]int, 0)
		p.StoreProfile.LimitShop.HasBoughtCount = make([]int, 0)
	}
	limitShop := &p.StoreProfile.LimitShop
	for i, buyId := range limitShop.HasBoughtGoods {
		if buyId == int(goodId) {
			limitShop.HasBoughtCount[i] += int(buyCount)
			return
		}
	}
	limitShop.HasBoughtGoods = append(limitShop.HasBoughtGoods, int(goodId))
	limitShop.HasBoughtCount = append(limitShop.HasBoughtCount, int(buyCount))
}

func (p *Account) rewardOnBuyGoodLimit(config *gamedata.HotLimitGoodCfg, buyCount int, resp *rspMsgBuyLimitGood) uint32 {
	costData := &gamedata.CostData{}
	costData.AddItem(config.Item.GetCoinItemID(), config.Item.GetCurrentPrice()*uint32(buyCount))

	rewardData := &gamedata.CostData{}
	for _, rewardCfg := range config.Item.GetFixed_Loot() {
		if rewardCfg.GetItemID() != "" {
			rewardData.AddItem(rewardCfg.GetItemID(), rewardCfg.GetGoodsCount()*uint32(buyCount))
		}
	}

	reason := fmt.Sprintf("buy limit good %d", config.Item.GetLimitGoodsID())
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	if ok := account.GiveBySync(p.Account, rewardData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	return 0
}
