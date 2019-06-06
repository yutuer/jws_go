package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"
)

func (p *Account) recoverAward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		RecoverId uint32 `codec:"rid"`
		IsHcAward bool   `codec:"ishc"` // 是否hc领奖，false即领免费的
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}
	initReqRsp(
		"PlayerAttr/RecoverAwardResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Cost_Not_Enough
		Err_Give
	)

	give := &account.GiveGroup{}
	cost := &account.CostGroup{}
	resCode, errMsg := p.Profile.GetRecover().Award(p.Account, req.RecoverId, req.IsHcAward, give, cost, resp)
	if resCode > 0 {
		return rpcErrorWithMsg(resp, resCode+20, errMsg)
	}
	if !cost.CostBySync(p.Account, resp, "RecoverAward") {
		return rpcErrorWithMsg(resp, Err_Cost_Not_Enough, "Err_Cost_Not_Enough")
	}
	if !give.GiveBySyncAuto(p.Account, resp, "RecoverAward") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}
	resp.OnChangeRecover()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
