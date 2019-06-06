package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestDebugAddGeneralNum struct {
	Req
	GeneralId string `codec:"gid"`
	Num       uint32 `codec:"num"`
}

type ResponseDebugAddGeneralNum struct {
	SyncResp
}

func (p *Account) DebugAddGeneralNum(r servers.Request) *servers.Response {
	req := &RequestDebugAddGeneralNum{}
	resp := &ResponseDebugAddGeneralNum{}

	initReqRsp(
		"Debug/DebugAddGeneralNumResponse",
		r.RawBytes,
		req, resp, p)

	logs.Error("DebugAddGeneralNum %s %d", req.GeneralId, req.Num)
	p.GeneralProfile.AddGeneralNum(req.GeneralId, req.Num, "debug")

	resp.OnChangeGeneralAllChange()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
