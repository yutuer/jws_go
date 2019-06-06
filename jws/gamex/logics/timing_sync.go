package logics

import "vcs.taiyouxi.net/platform/planx/servers"

type RequestTimingSync struct {
	Req
}

type ResponseTimingSync struct {
	SyncResp
}

// 此协议前端在城镇的情况下，会半分钟发一次，用来更新每次sync必带的信息, 和红点信息
func (p *Account) TimingSync(r servers.Request) *servers.Response {
	req := &RequestTimingSync{}
	resp := &ResponseTimingSync{}

	initReqRsp(
		"PlayerAttr/TimingSyncgResp",
		r.RawBytes,
		req, resp, p)

	resp.mkInfo(p)
	return rpcSuccess(resp)
}
