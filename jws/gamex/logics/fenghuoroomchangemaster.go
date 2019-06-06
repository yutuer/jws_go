package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FenghuoRoomChangeMaster : 更换房主
// 将自己的房主身份移交给其他人

// reqMsgFenghuoRoomChangeMaster 更换房主请求消息定义
type reqMsgFenghuoRoomChangeMaster struct {
	Req
	RoomType  int64  `codec:"_p1_"` // Room类型
	AccountID string `codec:"_p2_"` // 新房主的AcID
}

// rspMsgFenghuoRoomChangeMaster 更换房主回复消息定义
type rspMsgFenghuoRoomChangeMaster struct {
	SyncResp
}

// FenghuoRoomChangeMaster 更换房主: 将自己的房主身份移交给其他人
func (p *Account) FenghuoRoomChangeMaster(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomChangeMaster)
	rsp := new(rspMsgFenghuoRoomChangeMaster)

	initReqRsp(
		"Attr/FenghuoRoomChangeMasterRsp",
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
	mr.ChangeRoomMaster2Other(ctx,
		p.AccountID.String(),
		req.AccountID,
		p.Tmp.CurrRoomNum)
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
