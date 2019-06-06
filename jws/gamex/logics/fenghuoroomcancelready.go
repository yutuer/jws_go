package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FenghuoRoomCancelReady : 房间取消准备
// 在房间中取消准备

// reqMsgFenghuoRoomCancelReady 房间取消准备请求消息定义
type reqMsgFenghuoRoomCancelReady struct {
	Req
	RoomType int64 `codec:"_p1_"` // Room类型
}

// rspMsgFenghuoRoomCancelReady 房间取消准备回复消息定义
type rspMsgFenghuoRoomCancelReady struct {
	SyncResp
}

// FenghuoRoomCancelReady 房间取消准备: 在房间中取消准备
func (p *Account) FenghuoRoomCancelReady(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomCancelReady)
	rsp := new(rspMsgFenghuoRoomCancelReady)

	initReqRsp(
		"Attr/FenghuoRoomCancelReadyRsp",
		r.RawBytes,
		req, rsp, p)
	// logic imp begin
	if p.Tmp.CurrRoomNum <= 0 {
		return rpcError(rsp, room.ROOM_ERR_UNKNOWN)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel()

	//simpleInfo := p.GetSimpleInfo()
	mr := room.Get(p.AccountID.ShardId)
	mr.CancelReady(ctx,
		p.AccountID.String(),
		p.Tmp.CurrRoomNum)
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
