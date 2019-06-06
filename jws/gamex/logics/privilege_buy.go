package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestPrivilegeBuy struct {
	Req
	PrivilegeId int32 `codec:"pvlgId"`
}

type ResponsePrivilegeBuy struct {
	SyncRespWithRewards
	PrivilegeId int32 `codec:"pvlgId"`
}

func (p *Account) PrivilegeBuy(r servers.Request) *servers.Response {
	req := &RequestPrivilegeBuy{}
	resp := &ResponsePrivilegeBuy{}

	initReqRsp(
		"PlayerAttr/PrivilegeBuyResponse",
		r.RawBytes,
		req, resp, p)

	const (
		CODE_MIN = 20
	)

	ok, errcode, warncode := p.Profile.GetPrivilegeBuy().BuyByPrivilege(p.Account, req.PrivilegeId, resp)
	if !ok {
		if warncode > 0 {
			return rpcWarn(resp, warncode)
		}
		return rpcError(resp, errcode+CODE_MIN)
	}
	resp.OnChangeHC()
	resp.OnChangePrivilegeBuy()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
