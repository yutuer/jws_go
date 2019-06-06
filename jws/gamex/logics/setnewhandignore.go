package logics

import "vcs.taiyouxi.net/platform/planx/servers"

//SetNewHandIgnore 设置标记玩家已跳过所有新手引导

type reqMsgSetNewHandIgnore struct {
	Req
}

type rspMsgSetNewHandIgnore struct {
	Resp
}

func (a *Account) SetNewHandIgnore(r servers.Request) *servers.Response {
	req := new(reqMsgSetNewHandIgnore)
	rsp := new(rspMsgSetNewHandIgnore)

	initReqRsp(
		"Attr/SetNewHandIgnoreRsp",
		r.RawBytes,
		req, rsp, a)

	a.Profile.SetNewHandIgnore(true)

	return rpcSuccess(rsp)
}
