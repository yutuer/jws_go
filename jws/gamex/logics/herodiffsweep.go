package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// HeroDiffSweep : 出奇制胜扫荡
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgHeroDiffSweep 出奇制胜扫荡请求消息定义
type reqMsgHeroDiffSweep struct {
	Req
}

// rspMsgHeroDiffSweep 出奇制胜扫荡回复消息定义
type rspMsgHeroDiffSweep struct {
	SyncRespWithRewards
}

// HeroDiffSweep 出奇制胜扫荡:
func (p *Account) HeroDiffSweep(r servers.Request) *servers.Response {
	req := new(reqMsgHeroDiffSweep)
	rsp := new(rspMsgHeroDiffSweep)

	initReqRsp(
		"Attr/HeroDiffSweepRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.HeroDiffSweepHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
