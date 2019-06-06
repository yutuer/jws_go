package logics

import (
	"time"

	"fmt"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/city_broadcast"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/jws/gamex/modules/room/info"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	FENGHUO_CREATE_ALREADY_EXIST = 201
)

// FenghuoRoomCreate : 创建一个房间
// 创建一个房间,会返回房间的具体信息

// reqMsgFenghuoRoomCreate 创建一个房间请求消息定义
type reqMsgFenghuoRoomCreate struct {
	Req
	RoomType    int64  `codec:"_p1_"` // Room类型
	Password    string `codec:"_p2_"` // 进入密码
	Degree      int64  `codec:"_p3_"` // 难度IDx
	RewardPower int64  `codec:"_p4_"` // 产出倍率
}

// rspMsgFenghuoRoomCreate 创建一个房间回复消息定义
type rspMsgFenghuoRoomCreate struct {
	SyncResp
	RoomID  string `codec:"_p1_"` // 房间ID
	RoomNum int64  `codec:"_p2_"` // 房间号
	Room    []byte `codec:"_p3_"` // 新建房间具体信息
}

func (p *Account) GetFenghuoAvatarLv() (avatar int, lvl uint32) {
	avatars := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_FENGHUO)
	if avatars == nil || len(avatars) <= 0 {
		avatar = p.Profile.GetCurrAvatar()
	} else {
		avatar = avatars[0]
	}
	lv, _ := p.Profile.GetCorp().GetXpInfo()
	lvl = lv
	return
}

// FenghuoRoomCreate 创建一个房间: 创建一个房间,会返回房间的具体信息
func (p *Account) FenghuoRoomCreate(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomCreate)
	rsp := new(rspMsgFenghuoRoomCreate)

	initReqRsp(
		"Attr/FenghuoRoomCreateRsp",
		r.RawBytes,
		req, rsp, p)

	if p.Tmp.CurrRoomNum > 0 {
		//you already in a room
		//do nothing
	} else {

		sc := p.Profile.GetSC().GetSC(helper.SC_Money)
		hc := p.Profile.GetHC().GetHC()
		moneyEnough := gamedata.FenghuoHasEnoughCurrency(sc, hc, uint32(req.RewardPower), uint32(req.Degree), true)
		if !moneyEnough {
			// 没有钻石则提示充值，没有金币则提示购买金币
			return rpcWarn(rsp, info.RoomWarnCode_NOMONEY.Code())
		}

		// logic imp begin
		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second)
		defer cancel()
		avatar, lv := p.GetFenghuoAvatarLv()

		rooms, err := room.Get(p.AccountID.ShardId).UpdateRoom(
			ctx,
			p.AccountID.String(),
			room.FenghuoProfile{
				Name:     p.Profile.Name,
				AcID:     p.AccountID.String(),
				AvatarID: avatar,
				CorpLv:   lv,
				Gs:       p.Profile.GetData().CorpCurrGS,
			},
			info.Room{
				Type:        int(req.RoomType),
				Password:    req.Password,
				Degree:      int(req.Degree),
				RewardPower: int(req.RewardPower), //房间创建时,初始化当前联机是否消耗双方钻石获取更多奖励和战斗计次
			})
		//logs.Trace("room %v", rooms)
		if err != nil {
			logs.Error("Fenghuo FenghuoRoomCreate %d, %s", err.Code(), err.Error())
			return rpcError(rsp, err.Code())
		}

		if rooms.Num <= 0 {
			return rpcError(rsp, info.RoomErrCode_UNKNOWN.Code())
		}

		rsp.RoomID = rooms.ID
		rsp.RoomNum = int64(rooms.Num)
		rsp.Room = encode(rooms)
		p.Tmp.CurrRoomNum = rooms.Num
	}
	// logic imp end

	// 系统频道广播
	city_broadcast.Pool.UseRes2Send(
		city_broadcast.CBC_Typ_FengHuo,
		p.AccountID.ServerString(), fmt.Sprintf("%d,%s", rsp.RoomNum, rsp.RoomID), nil)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

/////////////////////////////

// FenghuoRoomUpdate : 尝试更新房间内部信息，例如玩家的主将阵容变更
// 更新房间内部信息

// reqMsgFenghuoRoomUpdate 尝试更新房间内部信息，例如玩家的主将阵容变更请求消息定义
type reqMsgFenghuoRoomUpdate struct {
	Req
	RoomType int64 `codec:"_p1_"` // Room类型
	RoomNum  int64 `codec:"_p2_"` // RoomNum
}

// rspMsgFenghuoRoomUpdate 尝试更新房间内部信息，例如玩家的主将阵容变更回复消息定义
type rspMsgFenghuoRoomUpdate struct {
	SyncResp
}

// FenghuoRoomUpdate 尝试更新房间内部信息，例如玩家的主将阵容变更: 更新房间内部信息
func (p *Account) FenghuoRoomUpdate(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoRoomUpdate)
	rsp := new(rspMsgFenghuoRoomUpdate)

	initReqRsp(
		"Attr/FenghuoRoomUpdateRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if p.Tmp.CurrRoomNum > 0 {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second)
		defer cancel()

		avatar, lv := p.GetFenghuoAvatarLv()

		rooms, err := room.Get(p.AccountID.ShardId).UpdateRoom(
			ctx,
			p.AccountID.String(),
			room.FenghuoProfile{
				Name:     p.Profile.Name,
				AcID:     p.AccountID.String(),
				AvatarID: avatar,
				CorpLv:   lv,
				Gs:       p.Profile.GetData().CorpCurrGS,
			},
			info.Room{
				Num: p.Tmp.CurrRoomNum,
			})
		//logs.Trace("room %v", rooms)
		if err != nil {
			logs.Error("Fenghuo FenghuoRoomCreate %d, %s", err.Code(), err.Error())
			return rpcError(rsp, err.Code())
		}

		if rooms.Num <= 0 {
			return rpcError(rsp, info.RoomErrCode_UNKNOWN.Code())
		}

		//rsp.GlobalRoomID = rooms.ID
		//rsp.RoomNum = int64(rooms.Num)
		//rsp.Room = encode(rooms)
		//p.Tmp.CurrRoomNum = rooms.Num
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
