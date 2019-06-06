package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// IsEnterTown : 进入城镇
// 进入城镇

// reqMsgIsEnterTown 进入城镇请求消息定义
type reqMsgIsEnterTown struct {
	Req
}

// rspMsgIsEnterTown 进入城镇回复消息定义
type rspMsgIsEnterTown struct {
	SyncResp
}

// IsEnterTown 进入城镇: 进入城镇
func (p *Account) IsEnterTown(r servers.Request) *servers.Response {
	req := new(reqMsgIsEnterTown)
	rsp := new(rspMsgIsEnterTown)

	initReqRsp(
		"Attr/IsEnterTownRsp",
		r.RawBytes,
		req, rsp, p)

	p.Profile.GetWSPVPInfo().IsWSPVPMarquee(p.AccountID.GameId, p.AccountID.ShardId, p.AccountID.String(), p.Profile.Name)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
