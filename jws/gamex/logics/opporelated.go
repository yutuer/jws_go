package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// OppoSign : oppo签到
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgOppoSign oppo签到请求消息定义
type reqMsgOppoSign struct {
	Req
}

// rspMsgOppoSign oppo签到回复消息定义
type rspMsgOppoSign struct {
	SyncRespWithRewards
}

// OppoSign oppo签到:
func (p *Account) OppoSign(r servers.Request) *servers.Response {
	req := new(reqMsgOppoSign)
	rsp := new(rspMsgOppoSign)

	initReqRsp(
		"Attr/OppoSignRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.OppoSignHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OppoDailyQuest : oppo每日任务
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgOppoDailyQuest oppo每日任务请求消息定义
type reqMsgOppoDailyQuest struct {
	Req
}

// rspMsgOppoDailyQuest oppo每日任务回复消息定义
type rspMsgOppoDailyQuest struct {
	SyncRespWithRewards
}

// OppoDailyQuest oppo每日任务:
func (p *Account) OppoDailyQuest(r servers.Request) *servers.Response {
	req := new(reqMsgOppoDailyQuest)
	rsp := new(rspMsgOppoDailyQuest)

	initReqRsp(
		"Attr/OppoDailyQuestRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.OppoDailyQuestHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OppoLogin : oppo每日任务
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgOppoLogin oppo每日任务请求消息定义
type reqMsgOppoLogin struct {
	Req
	IsOppo bool `codec:"is_oppo"` // 是否是oppo游戏中心启动
}

// rspMsgOppoLogin oppo每日任务回复消息定义
type rspMsgOppoLogin struct {
	SyncResp
}

// OppoLogin oppo每日任务:
func (p *Account) OppoLogin(r servers.Request) *servers.Response {
	req := new(reqMsgOppoLogin)
	rsp := new(rspMsgOppoLogin)

	initReqRsp(
		"Attr/OppoLoginRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.OppoLoginHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
