package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (a *Account) iOSPayRequest(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Receipt string `codec:"receipt"`
	}{}

	resp := &struct {
		SyncRespWithRewards
	}{}

	acID := a.AccountID.String()

	initReqRsp(
		"PlayerAttr/IOSPayRsp",
		r.RawBytes,
		req, resp, a)

	logs.Info("[%s]iOSPayRequest %s", acID, req.Receipt)

	//client := appstore.New()
	//iapReq := appstore.IAPRequest{
	//	ReceiptData: req.Receipt,
	//}
	//_, err := client.Verify(iapReq)

	var err error
	if err != nil {
		logs.Error("err %s", err.Error())
		return rpcError(resp, 1)
	}

	receiptData, err := pay.ParseReceiptData(req.Receipt)

	if err != nil {
		logs.SentryLogicCritical(acID, "ParseReceiptData Err By %s", err.Error())
		return rpcError(resp, 2)
	}

	logs.Info("[%s]iOSPayRequest pay For %s - %s - %s",
		acID,
		receiptData.PurchaseInfo.TransactionId,
		receiptData.PurchaseInfo.BID,
		receiptData.PurchaseInfo.ProductID)

	if receiptData.PurchaseInfo.BID != "com.taiyouxi.ifsg" {
		logs.SentryLogicCritical(acID, "receiptData.PurchaseInfo.BID Err By %s",
			receiptData.PurchaseInfo.BID)
		return rpcError(resp, 4)
	}

	iapInfo := a.Profile.GetIAPInfo()
	if iapInfo.IsHasTransID(receiptData.PurchaseInfo.TransactionId) {
		logs.Warn("[%s]receiptData.PurchaseInfo.TransactionId Muit Err By %s",
			acID, receiptData.PurchaseInfo.TransactionId)
		resp.mkInfo(a)
		return rpcSuccess(resp)
	}

	iapInfo.AddIAPOrder(
		receiptData.PurchaseInfo.TransactionId,
		receiptData.PurchaseInfo.ProductID)

	// 校验重复 全局
	repeat, err := pay.IsIOSOrderRepeat(receiptData.PurchaseInfo.TransactionId)
	if err != nil {
		logs.SentryLogicCritical(acID, "iOSPayRequest order %s IsIOSOrderRepeat err %v",
			receiptData.PurchaseInfo.TransactionId, err)
		return rpcError(resp, 5)
	}
	if repeat {
		logs.Warn("[%s]receiptData.PurchaseInfo.TransactionId Globe repeat By %s",
			acID, receiptData.PurchaseInfo.TransactionId)
		resp.mkInfo(a)
		return rpcSuccess(resp)
	}

	// 加hc
	costData := &gamedata.CostData{}
	costData.AddIAPGoodByID(receiptData.PurchaseInfo.ProductID, receiptData.PurchaseInfo.TransactionId,
		gamedata.GetIAPIOSPrice(receiptData.PurchaseInfo.ProductID),
		receiptData.PurchaseInfo.PurchaseDateMs[:len(receiptData.PurchaseInfo.PurchaseDateMs)-3])

	if !account.GiveBySync(a.Account, costData, resp, "RmbPay") {
		logs.SentryLogicCritical(acID, "Give IAP Err By %s",
			receiptData.PurchaseInfo.ProductID)
		return rpcError(resp, 3)
	}

	logs.Info("[%s]iOSPayRequest Success  %s - %s - %s", acID,
		receiptData.PurchaseInfo.TransactionId,
		receiptData.PurchaseInfo.BID,
		receiptData.PurchaseInfo.ProductID)

	a.Profile.PayTime = a.Profile.GetProfileNowTime()
	resp.mkInfo(a)
	return rpcSuccess(resp)
}

func (a *Account) iapCardRewardRequest(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ActivityId int `codec:"actId"`
	}{}

	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/IApCardRewardRsp",
		r.RawBytes,
		req, resp, a)

	const (
		_ = iota
		Err_Param
		Err_Month_Card_UnValid
		Err_Life_Card_UnValid
		Err_Week_Card_UnValid
		Err_Give
	)

	iap := a.Profile.GetIAPGoodInfo()
	now_time := a.Profile.GetProfileNowTime()
	data := &gamedata.CostData{}
	var reason string
	switch req.ActivityId {
	case gamedata.IAP_Monthly:
		if iap.MonthlyCardEndTime <= now_time || iap.MonthlyValidTime > now_time {
			logs.Warn("iapCardRewardRequest Err_Month_Card_UnValid")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		for _, loot := range gamedata.IAPMonth.GetLoots() {
			data.AddItem(loot.GetCoinItemID(), loot.GetCount())
		}
		iap.MonthlyValidTime = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_time), now_time)
		// 月卡最后一天领取后,就立刻可以购买,所以这里设置0
		if iap.MonthlyValidTime == iap.MonthlyCardEndTime {
			iap.MonthlyValidTime = 0
			iap.MonthlyCardEndTime = 0
		}
		reason = "MonthCard"
	case gamedata.IAP_Life:
		if !iap.IsLifeCard || iap.LifeCardValidTime > now_time {
			logs.Warn("iapCardRewardRequest Err_Life_Card_UnValid")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		for _, loot := range gamedata.IAPLife.GetLoots() {
			data.AddItem(loot.GetCoinItemID(), loot.GetCount())
		}
		iap.LifeCardValidTime = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_time), now_time)
		reason = "LifeCard"
	case gamedata.IAP_Week:
		if iap.WeekRewardEndTime <= now_time || iap.WeekRewardValidTime > now_time {
			return rpcErrorWithMsg(resp, Err_Week_Card_UnValid, "Err_Week_Card_UnValid")
		}
		for _, loot := range gamedata.IAPWeek.GetLoots() {
			data.AddItem(loot.GetCoinItemID(), loot.GetCount())
		}
		iap.WeekRewardValidTime = pay.WeekStartCommonSecBaseMonday(now_time) + int64(util.WeekSec)
		reason = "WeekCard"
	default:
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}

	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(a.Account, resp, reason) {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}
	resp.OnChangeIAPGoodInfo()
	resp.mkInfo(a)
	return rpcSuccess(resp)
}

// IAPPaySuccess : 支付成功通知服务器
// 支付成功通知服务器的协议

// reqMsgIAPPaySuccess 支付成功通知服务器请求消息定义
type reqMsgIAPPaySuccess struct {
	Req
	GoodIdx int64 `codec:"gidx"` // iapmain配置表的id
}

// rspMsgIAPPaySuccess 支付成功通知服务器回复消息定义
type rspMsgIAPPaySuccess struct {
	SyncResp
}

// IAPPaySuccess 支付成功通知服务器: 支付成功通知服务器的协议，但客户端可能不发这个协议
func (p *Account) IAPPaySuccess(r servers.Request) *servers.Response {
	req := new(reqMsgIAPPaySuccess)
	rsp := new(rspMsgIAPPaySuccess)

	initReqRsp(
		"PlayerAttr/IAPPaySuccessRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Cfg
	)

	idx := uint32(req.GoodIdx)
	cfg := gamedata.GetIAPInfo(idx)
	if cfg == nil {
		return rpcErrorWithMsg(rsp, Err_Cfg, "Err_Cfg")
	}
	pg := p.Profile.GetIAPGoodInfo()
	found := false
	for _, iap := range pg.Infos {
		if iap.GoodIdx == idx {
			found = true
		}
	}
	//sbatch := p.Profile.GetIAPGoodInfo().GetCurrenServerSbatch(p.AccountID.ShardId, p.Profile.GetProfileNowTime())
	if !found {
		pg.Infos = append(pg.Infos, pay.PayGoodInfo{
			GoodIdx:         idx,
			FirstGiveSerial: 0,
		})
		rsp.OnChangeIAPGoodInfo()
		rsp.mkInfo(p)
	}

	return rpcSuccess(rsp)
}

// AwardLevelGift : 等级礼包领奖
// 等级礼包领奖的协议

// reqMsgAwardLevelGift 等级礼包领奖请求消息定义
type reqMsgAwardLevelGift struct {
	Req
	LevelGiftId string `codec:"lgid"` // 等级礼包id
}

// rspMsgAwardLevelGift 等级礼包领奖回复消息定义
type rspMsgAwardLevelGift struct {
	SyncRespWithRewards
}

// AwardLevelGift 等级礼包领奖: 等级礼包领奖的协议
func (p *Account) AwardLevelGift(r servers.Request) *servers.Response {
	req := new(reqMsgAwardLevelGift)
	rsp := new(rspMsgAwardLevelGift)

	initReqRsp(
		"PlayerAttr/AwardLevelGiftRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Not_Gift_Can_Award
		Err_Give
	)

	iap := p.Profile.GetIAPGoodInfo()
	h := false
	for i, v := range iap.LevelGiftIdWaitAward {
		if v == req.LevelGiftId {
			tmp := iap.LevelGiftIdWaitAward[:i]
			tmp = append(tmp, iap.LevelGiftIdWaitAward[i+1:]...)
			iap.LevelGiftIdWaitAward = tmp
			h = true
			break
		}
	}
	if !h {
		logs.Warn("AwardLevelGift Not_Gift_Can_Award")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	// 发奖
	data := &gamedata.PriceDatas{}
	cfg := gamedata.GetLevelGiftCfg(req.LevelGiftId)
	for _, v := range cfg.GetLevelGiftAward_Template() {
		if !gamedata.IsFixedIDItemID(v.GetItemID()) {
			d := gamedata.MakeItemData(p.AccountID.String(), p.Account.GetRand(), v.GetItemID())
			data.AddItemWithData(v.GetItemID(), *d, v.GetCount())
		} else {
			data.AddItem(v.GetItemID(), v.GetCount())
		}
	}
	give := &account.GiveGroup{}
	give.AddCostData(data.Gives())
	if !give.GiveBySyncAuto(p.Account, rsp, "AwardLevelGift") {
		return rpcErrorWithMsg(rsp, Err_Give, "Err_Give")
	}

	rsp.OnChangeIAPGoodInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
