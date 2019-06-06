package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FenghuoRoomListenEnd : 停止订阅房间请求
// 客户端停止订阅房间变化

// reqMsgFenghuoRoomListenEnd 停止订阅房间请求请求消息定义
type reqMsgFenghuoRoomListenEnd struct {
	Req
	RoomType int64 `codec:"rt_"` // 订阅的Room类型
}

// rspMsgFenghuoRoomListenEnd 停止订阅房间请求回复消息定义
type rspMsgFenghuoRoomListenEnd struct {
	SyncResp
}

// FenghuoRoomListenEnd 停止订阅房间请求: 客户端停止订阅房间变化
func (p *Account) FenghuoRoomListenEnd(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomListenEnd)
	rsp := new(rspMsgFenghuoRoomListenEnd)

	initReqRsp(
		"Attr/FenghuoRoomListenEndRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel()

	room.Get(p.AccountID.ShardId).DetachObserve(
		ctx,
		p.AccountID.String())
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
