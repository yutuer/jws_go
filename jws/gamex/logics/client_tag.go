package logics

import "vcs.taiyouxi.net/platform/planx/servers"

type RequestSetClientTag struct {
	Req
	Index int `codec:"i"`
	Val   int `codec:"v"`
}

type ResponseSetClientTag struct {
	SyncResp
}

func (p *Account) SetClientTag(r servers.Request) *servers.Response {
	req := &RequestSetClientTag{}
	resp := &ResponseSetClientTag{}

	initReqRsp(
		"PlayerAttr/SetClientTagResp",
		r.RawBytes,
		req, resp, p)

	if err := p.Profile.GetClientTagInfo().SetTag(req.Index, req.Val); err != nil {
		return rpcErrorWithMsg(resp, 1, err.Error())
	}
	resp.OnChangeClientTagInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
