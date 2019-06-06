package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FenghuoRoomLeave : 离开当前房间
// 离开当前房间，如果当前不在房间中则无效

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFenghuoRoomLeave 离开当前房间请求消息定义
type reqMsgFenghuoRoomLeave struct {
	Req
	RoomType int64  `codec:"_p1_"` // Room类型
	KickAcID string `codec:"_p2_"` // 不为空时,尝试踢人
}

// rspMsgFenghuoRoomLeave 离开当前房间回复消息定义
type rspMsgFenghuoRoomLeave struct {
	SyncResp
}

// FenghuoRoomLeave 离开当前房间: 离开当前房间，如果当前不在房间中则无效
func (p *Account) FenghuoRoomLeave(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomLeave)
	rsp := new(rspMsgFenghuoRoomLeave)

	initReqRsp(
		"Attr/FenghuoRoomLeaveRsp",
		r.RawBytes,
		req, rsp, p)

	if p.Tmp.CurrRoomNum > 0 {
		// logic imp begin
		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second)
		defer cancel()

		leaveAcid := req.KickAcID
		mr := room.Get(p.AccountID.ShardId)

		mr.LeaveRoom(
			ctx,
			p.AccountID.String(),
			leaveAcid,
			p.Tmp.CurrRoomNum,
		)
	}

	if req.KickAcID == "" {
		//Not a kick command
		p.Tmp.CurrRoomNum = 0
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
