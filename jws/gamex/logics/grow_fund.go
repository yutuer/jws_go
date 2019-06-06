package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (a *Account) activateGrowFund(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}

	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/ActivateGrowFundRsp",
		r.RawBytes,
		req, resp, a)

	const (
		_ = iota
		Err_Cond
		Err_Had_Activate
		Err_Vip
		Err_Cost_Hc
	)

	if !account.CondCheck(gamedata.Mod_GrowFund, a.Account) {
		return rpcErrorWithMsg(resp, Err_Cond, "Err_Cond")
	}

	fund := a.Profile.GetGrowFund()
	if fund.IsActivate {
		return rpcErrorWithMsg(resp, Err_Had_Activate, "Err_Had_Activate")
	}

	// vip
	vipCfg := gamedata.GetVIPCfg(int(a.Profile.GetVipLevel()))
	if !vipCfg.GrowFund {
		return rpcErrorWithMsg(resp, Err_Vip, "Err_Vip")
	}
	// hc
	cost := &account.CostGroup{}
	if !cost.AddHc(a.Account, int64(gamedata.GetCommonCfg().GetGrowFundCost())) ||
		!cost.CostBySync(a.Account, resp, "ActivateGrowFund") {
		return rpcErrorWithMsg(resp, Err_Cost_Hc, "Err_Cost_Hc")
	}
	fund.IsActivate = true

	resp.OnChangeGrowFund()
	resp.mkInfo(a)
	return rpcSuccess(resp)
}

func (a *Account) awardGrowFund(r servers.Request) *servers.Response {
	req := &struct {
		Req
		LvlId uint32 `codec:"lvlId"`
	}{}

	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/AwardGrowFundRsp",
		r.RawBytes,
		req, resp, a)

	const (
		_ = iota
		Err_Cond
		Err_NotAct
		Err_Cfg
		Err_Repeat
		Err_give
	)

	if !account.CondCheck(gamedata.Mod_GrowFund, a.Account) {
		return rpcErrorWithMsg(resp, Err_Cond, "Err_Cond")
	}

	fund := a.Profile.GetGrowFund()
	if !fund.IsActivate {
		return rpcErrorWithMsg(resp, Err_NotAct, "Err_NotAct")
	}

	cfg := gamedata.GetGrowFund(req.LvlId)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_Cfg, "Err_Cfg")
	}

	if !fund.IfNotBuyThenBuy(req.LvlId) {
		logs.Warn("awardGrowFund Err_Repeat")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	data := &gamedata.CostData{}
	data.AddItem(cfg.GetItemID(), cfg.GetCount())
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(a.Account, resp, "AwardGrowFund") {
		return rpcErrorWithMsg(resp, Err_give, "Err_give")
	}

	resp.OnChangeGrowFund()
	resp.mkInfo(a)
	return rpcSuccess(resp)
}
