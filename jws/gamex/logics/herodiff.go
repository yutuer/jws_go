package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// StartHeroDiffFight : 开始出奇制胜战斗
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgStartHeroDiffFight 开始出奇制胜战斗请求消息定义
type reqMsgStartHeroDiffFight struct {
	Req
}

// rspMsgStartHeroDiffFight 开始出奇制胜战斗回复消息定义
type rspMsgStartHeroDiffFight struct {
	SyncResp
}

// StartHeroDiffFight 开始出奇制胜战斗:
func (p *Account) StartHeroDiffFight(r servers.Request) *servers.Response {
	req := new(reqMsgStartHeroDiffFight)
	rsp := new(rspMsgStartHeroDiffFight)

	initReqRsp(
		"Attr/StartHeroDiffFightRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.StartHeroDiffFightHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OverHeroDiffFight : 出奇制胜战斗结束
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgOverHeroDiffFight 出奇制胜战斗结束请求消息定义
type reqMsgOverHeroDiffFight struct {
	ReqWithAnticheat
	Score int64 `codec:"score"` // 分数
}

// rspMsgOverHeroDiffFight 出奇制胜战斗结束回复消息定义
type rspMsgOverHeroDiffFight struct {
	SyncRespWithRewardsAnticheat
}

// OverHeroDiffFight 出奇制胜战斗结束:
func (p *Account) OverHeroDiffFight(r servers.Request) *servers.Response {
	req := new(reqMsgOverHeroDiffFight)
	rsp := new(rspMsgOverHeroDiffFight)

	initReqRsp(
		"Attr/OverHeroDiffFightRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.OverHeroDiffFightHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
