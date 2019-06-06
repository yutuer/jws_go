package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// Just Debug
type RequestSCDebugOp struct {
	Req
	OpType int   `codec:"op"`
	SCType int   `codec:"typ"`
	Value  int64 `codec:"value"`
}

type ResponseSCDebugOp struct {
	Resp
	SC []int64 `codec:"sc"`
}

func (p *Account) SCDebugOp(r servers.Request) *servers.Response {
	req := &RequestSCDebugOp{}
	resp := &ResponseSCDebugOp{}
	initReqRsp(
		"Debug/SCOpResponse",
		r.RawBytes,
		req, resp, p)

	if req.OpType == 1 {
		p.Profile.GetSC().AddSC(req.SCType, req.Value, "Debug")
	} else if req.OpType == 2 {
		p.Profile.GetSC().UseSC(req.SCType, req.Value, "Debug")
	}

	resp.SC = p.Profile.GetSC().GetAll()
	logs.Trace("sc %v", p.Profile.GetSC().GetAll())
	logs.Trace("resp %v", resp)

	return rpcSuccess(resp)
}
