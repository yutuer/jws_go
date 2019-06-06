package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// TwitterShare : Twitter分享
// 

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTwitterShare Twitter分享请求消息定义
type reqMsgTwitterShare struct {
	Req
}

// rspMsgTwitterShare Twitter分享回复消息定义
type rspMsgTwitterShare struct {
	SyncRespWithRewards
}

// TwitterShare Twitter分享: 
func (p *Account) TwitterShare(r servers.Request) *servers.Response {
	req := new(reqMsgTwitterShare)
	rsp := new(rspMsgTwitterShare)

	initReqRsp(
		"Attr/TwitterShareRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TwitterShareHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// LineShare : Line分享
// 

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgLineShare Line分享请求消息定义
type reqMsgLineShare struct {
	Req
}

// rspMsgLineShare Line分享回复消息定义
type rspMsgLineShare struct {
	SyncRespWithRewards
}

// LineShare Line分享:
func (p *Account) LineShare(r servers.Request) *servers.Response {
	req := new(reqMsgLineShare)
	rsp := new(rspMsgLineShare)

	initReqRsp(
		"Attr/LineShareRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.LineShareHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


