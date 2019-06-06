package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FenghuoRoomListenStart : 开始订阅房间请求
// 客户端获取当前房间信息,并开始订阅房间变化,之后服务器会向玩家推送房间变化信息

// reqMsgFenghuoRoomListenStart 开始订阅房间请求请求消息定义
type reqMsgFenghuoRoomListenStart struct {
	Req
	RoomType int64 `codec:"rt_"` // 需要订阅的Room类型
}

// rspMsgFenghuoRoomListenStart 开始订阅房间请求回复消息定义
type rspMsgFenghuoRoomListenStart struct {
	SyncResp
}

// FenghuoRoomListenStart 开始订阅房间请求: 客户端获取当前房间信息,并开始订阅房间变化,之后服务器会向玩家推送房间变化信息
func (p *Account) FenghuoRoomListenStart(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomListenStart)
	rsp := new(rspMsgFenghuoRoomListenStart)

	initReqRsp(
		"Attr/FenghuoRoomListenStartRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel()

	rooms, err := room.Get(p.AccountID.ShardId).AttachObserve(
		ctx,
		p.AccountID.String(),
		p.GetMsgNotifyChan())

	if err != nil {
		logs.Error("Fenghuo FenghuoRoomListenStart err: %d, %s", err.Code(), err.Error())
		return rpcError(rsp, err.Code())
	}
	syncMsg := player_msg.RoomsSyncInfo{}
	syncMsg.RoomNew = make([][]byte, 0, len(rooms))
	syncMsg.RoomDel = []int{}

	for _, room := range rooms {
		syncMsg.RoomNew = append(syncMsg.RoomNew, room.ToData())
	}

	rsp.SyncRoom = encode(syncMsg)

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
