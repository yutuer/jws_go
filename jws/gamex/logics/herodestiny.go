package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ActivateHeroDestiny : 激活指定的宿命
// 激活指定的宿命

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgActivateHeroDestiny 激活指定的宿命请求消息定义
type reqMsgActivateHeroDestiny struct {
	Req
	DestinyId int64 `codec:"dty_id"` // 宿命ID
}

// rspMsgActivateHeroDestiny 激活指定的宿命回复消息定义
type rspMsgActivateHeroDestiny struct {
	SyncRespWithRewards
}

// ActivateHeroDestiny 激活指定的宿命: 激活指定的宿命
func (p *Account) ActivateHeroDestiny(r servers.Request) *servers.Response {
	req := new(reqMsgActivateHeroDestiny)
	rsp := new(rspMsgActivateHeroDestiny)

	initReqRsp(
		"Attr/ActivateHeroDestinyRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ActivateHeroDestinyHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
