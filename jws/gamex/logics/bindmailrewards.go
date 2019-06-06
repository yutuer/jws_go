package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// BindMailRewards : 客户端绑定邮箱可以发奖
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBindMailRewards 客户端绑定邮箱可以发奖请求消息定义
type reqMsgBindMailRewards struct {
	Req
	ActivityRewardID int64 `codec:"ac_rw_id"` // 活动id
}

// rspMsgBindMailRewards 客户端绑定邮箱可以发奖回复消息定义
type rspMsgBindMailRewards struct {
	SyncRespWithRewards
}

// BindMailRewards 客户端绑定邮箱可以发奖:
func (p *Account) BindMailRewards(r servers.Request) *servers.Response {
	req := new(reqMsgBindMailRewards)
	rsp := new(rspMsgBindMailRewards)

	initReqRsp(
		"Attr/BindMailRewardsRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BindMailRewardsHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
