package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

// FBShare : FBShare分享
// FBShare分享

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFBShare FBShare分享请求消息定义
type reqMsgFBShare struct {
	Req
}

// rspMsgFBShare FBShare分享回复消息定义
type rspMsgFBShare struct {
	SyncResp
}

// FBShare FBShare分享: FBShare分享
func (p *Account) FBShare(r servers.Request) *servers.Response {
	req := new(reqMsgFBShare)
	rsp := new(rspMsgFBShare)

	initReqRsp(
		"Attr/FBShareRsp",
		r.RawBytes,
		req, rsp, p)

	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_FaceBookShare) {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}
	p.updateCondition(account.COND_TYP_FB_Share,
		1, 0, "", "", rsp)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
