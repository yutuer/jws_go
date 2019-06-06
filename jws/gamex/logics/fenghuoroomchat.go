package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FenghuoRoomChat : 房间中发送聊天信息
// 在房间中发一个聊天信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFenghuoRoomChat 房间中发送聊天信息请求消息定义
type reqMsgFenghuoRoomChat struct {
	Req
	RoomType int64  `codec:"_p1_"` // Room类型
	Message  string `codec:"_p2_"` // 信息内容
}

// rspMsgFenghuoRoomChat 房间中发送聊天信息回复消息定义
type rspMsgFenghuoRoomChat struct {
	SyncResp
}

// FenghuoRoomChat 房间中发送聊天信息: 在房间中发一个聊天信息
func (p *Account) FenghuoRoomChat(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomChat)
	rsp := new(rspMsgFenghuoRoomChat)

	initReqRsp(
		"Attr/FenghuoRoomChatRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	logs.Error("there is no Imp for FenghuoRoomChat")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
