package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FenghuoRoomDelete : 解散当前房间
// 解散当前自己为房主的房间

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFenghuoRoomDelete 解散当前房间请求消息定义
type reqMsgFenghuoRoomDelete struct {
	Req
	RoomType int64 `codec:"_p1_"` // Room类型
}

// rspMsgFenghuoRoomDelete 解散当前房间回复消息定义
type rspMsgFenghuoRoomDelete struct {
	SyncResp
}

// FenghuoRoomDelete 解散当前房间: 解散当前自己为房主的房间
func (p *Account) FenghuoRoomDelete(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomDelete)
	rsp := new(rspMsgFenghuoRoomDelete)

	initReqRsp(
		"Attr/FenghuoRoomDeleteRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	logs.Error("there is no Imp for FenghuoRoomDelete")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
