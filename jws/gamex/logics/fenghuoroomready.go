package logics

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FenghuoRoomReady : 房间准备
// 在房间中准备

// reqMsgFenghuoRoomReady 房间准备请求消息定义
type reqMsgFenghuoRoomReady struct {
	Req
	RoomType int64 `codec:"_p1_"` // Room类型
}

// rspMsgFenghuoRoomReady 房间准备回复消息定义
type rspMsgFenghuoRoomReady struct {
	SyncResp
}

// FenghuoRoomReady 房间准备: 在房间中准备
func (p *Account) FenghuoRoomReady(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomReady)
	rsp := new(rspMsgFenghuoRoomReady)

	initReqRsp(
		"Attr/FenghuoRoomReadyRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if p.Tmp.CurrRoomNum <= 0 {
		return rpcSuccess(rsp)
	}

	ctx1, cancel1 := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel1()

	ri, err := room.Get(p.AccountID.ShardId).GetRoomInfo(ctx1,
		p.AccountID.String(),
		p.Tmp.CurrRoomNum)

	if err != nil {
		logs.Error("Fenghuo RoomReady error: %s", err.Error())
		return rpcError(rsp, room.ROOM_ERR_GETROOMINFO_NOFOUND_ROOMID)
	}

	sc := p.Profile.GetSC().GetSC(helper.SC_Money)
	hc := p.Profile.GetHC().GetHC()
	moneyEnough := gamedata.FenghuoHasEnoughCurrency(sc, hc, uint32(ri.GetRewardPower()), uint32(ri.Degree), true)
	if !moneyEnough {
		// 没有钻石则提示充值，没有金币则提示购买金币
		return rpcWarn(rsp, room.ROOM_WARN_CREATEROOM_NOENOUGH_MONEY)
	}

	accData := &helper.Avatar2ClientByJson{}
	// 用客户端设置的阵容里的角色
	curAvatar := p.Profile.CurrAvatar
	heroTm := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_FENGHUO)
	if heroTm != nil && len(heroTm) > 0 {
		curAvatar = heroTm[0]
	}
	account.FromAccount2Json(accData, p.Account, curAvatar)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second)
	defer cancel()

	code := room.Get(p.AccountID.ShardId).Ready(
		ctx,
		p.AccountID.String(),
		p.Tmp.CurrRoomNum, accData)

	switch code {
	case room.ROOM_WARN_MASTER_WAITING_OTHERS_READY:
		return rpcWarn(rsp, uint32(code))
	case room.ROOM_ERR_UNKNOWN:
		return rpcError(rsp, uint32(code))
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)

}
