package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/jws/gamex/modules/room/info"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FenghuoRoomEnter : 进入某一个房间
// 进入房间

// reqMsgFenghuoRoomEnter 进入某一个房间请求消息定义
type reqMsgFenghuoRoomEnter struct {
	Req
	RoomType int64  `codec:"_p1_"` // Room类型
	RoomNum  int64  `codec:"_p2_"` // RoomNum
	RoomID   string `codec:"_p3_"` // 房间ID
}

// rspMsgFenghuoRoomEnter 进入某一个房间回复消息定义
type rspMsgFenghuoRoomEnter struct {
	SyncResp
	RoomID  string `codec:"_p1_"` // 房间ID
	RoomNum int64  `codec:"_p2_"` // 房间号
	Room    []byte `codec:"_p3_"` // 新建房间具体信息
}

// FenghuoRoomEnter 进入某一个房间: 进入房间
func (p *Account) FenghuoRoomEnter(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomEnter)
	rsp := new(rspMsgFenghuoRoomEnter)

	initReqRsp(
		"Attr/FenghuoRoomEnterRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel()

	avatar, lv := p.GetFenghuoAvatarLv()
	mr := room.Get(p.AccountID.ShardId)

	curRoom, errcode := mr.EnterRoom(
		ctx,
		p.AccountID.String(),
		room.FenghuoProfile{
			Name:     p.Profile.Name,
			AcID:     p.AccountID.String(),
			AvatarID: avatar,
			CorpLv:   lv,
			Gs:       p.Profile.GetData().CorpCurrGS,
		},
		int(req.RoomNum), req.RoomID)

	switch errcode {
	case info.RoomWarnCode_ENTER_ROOM_FULL:
		fallthrough
	case info.RoomWarnCode_ENTER_ROOM_NUM_WRONG:
		//可能想加入的时候房间已经不存在了。因为拿到房间的列表可能是不存在的
		fallthrough
	case info.RoomWarnCode_ENTER_ROOM_ALREADYINROOM:
		fallthrough
	case info.RoomWarnCode_ENTER_ROOM_NOTJOINABLE:
		return rpcWarn(rsp, errcode.Code())
	case info.RoomErrCode_UNKNOWN, info.RoomErrCode_CTX_TIMEOUT:
		return rpcError(rsp, errcode.Code())
	}

	if curRoom.ID != "" {
		rsp.RoomID = curRoom.ID
		rsp.RoomNum = req.RoomNum
		rsp.Room = encode(curRoom)
		p.Tmp.CurrRoomNum = int(req.RoomNum)
	} else {
		return rpcWarn(rsp, info.RoomErrCode_UNKNOWN.Code())
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
