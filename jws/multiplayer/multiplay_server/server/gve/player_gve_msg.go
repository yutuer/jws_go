package gve

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/common"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Player) Init() {
	//客户端,链接服务器后,第1个协议
	p.OnRpc(multiplayMsg.DatasEnterMultiplayGameReq, p.OnEnterMultiplayGameRpc)
	//客户端,链接服务器后,第2个协议,只是告知客户端奖励有什么,用于显示
	p.OnRpc(multiplayMsg.DatasGetGameRewardsReq, p.OnGetGameRewardsRpc)

	//客户端可能没有使用如下两个协议. By YZH 2016.10.5
	p.OnRpc(multiplayMsg.DatasGetGameDatasReq, p.OnGetGameStateRpc)
	//客户端再断线重连,过程貌似尝试用这个接口,  但是似乎断线重连的处理本身不是必须的,因此这个协议不重要了
	p.OnRpc(multiplayMsg.DatasGetGameStateReq, p.OnGetGameDatasRpc)

	p.OnNotify(multiplayMsg.DatasHPNotify, p.OnHPNotify)

	p.OnNotify(multiplayMsg.DatasReadyMultiplayGameNotify, p.OnReadyMultiplayGameNotify)
	p.OnNotify(multiplayMsg.DatasLeaveMultiplayGameNotify, p.OnLeaveMultiplayGameNotify)

}

// OnEnterMultiplayGameRpc 玩家请求信息:准备进入多人游戏
func (p *Player) OnEnterMultiplayGameRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Trace("EnterMultiplayGameRpc")
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		return gve_proto.GenErrorRspPacket(reqPacket, common.MsgResCodeReqPacketErr)
	}
	req := new(multiplayMsg.EnterMultiplayGameReq)
	req.Init(unionTable.Bytes, unionTable.Pos)

	logs.Trace("EnterMultiplayGameRpc %v %v %v", req.AccountId(), req.RoomID(), req.Secret())
	p.game = GVEGamesMgr.GVEGetGame(string(req.RoomID()), string(req.Secret()))
	if p.game == nil {
		logs.Error("OnEnterMultiplayGameRpc player %s game %s not found ",
			p.AcID, req.RoomID())
		return gve_proto.GenErrorRspPacket(reqPacket, common.MsgResCodeNoGameCurr)
	}
	p.AcID = string(req.AccountId())
	err := p.game.EnterPlayer(p)
	if err != nil {
		return gve_proto.GenErrorRspPacket(reqPacket, common.MsgResCodePlayerNoInGame)
	}
	return p.game.SendReqToGame(reqPacket)
}

func (p *Player) OnGetGameStateRpc(reqPacket msgprocessor.IPacket) []byte {
	return p.game.SendReqToGame(reqPacket)
}

func (p *Player) OnGetGameDatasRpc(reqPacket msgprocessor.IPacket) []byte {
	return p.game.SendReqToGame(reqPacket)
}

func (p *Player) OnHPNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnHPNotify")
	p.game.SendNotifyToGame(p.AcID, reqPacket)
}

func (p *Player) OnLeaveMultiplayGameNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnLeaveMultiplayGameNotify")
	p.game.SendNotifyToGame(p.AcID, reqPacket)
}

func (p *Player) OnReadyMultiplayGameNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnReadyMultiplayGameNotify")
	p.game.SendNotifyToGame(p.AcID, reqPacket)
}

func (p *Player) OnGetGameRewardsRpc(reqPacket msgprocessor.IPacket) []byte {
	return p.game.SendReqToGame(reqPacket)
}
