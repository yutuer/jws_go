package gvg

import (
	"time"

	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/teamboss_proto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MsgResCodeNoGameCurr = iota + 900
	MsgResCodePlayerNoInGame
	MsgResCodeReqPacketErr
)

func (p *GVGPlayer) Init() {
	//客户端,链接服务器后,第1个协议
	p.OnRpc(multiplayMsg.DatasEnterMultiplayGameReq, p.OnEnterMultiplayGameRpc)

	p.OnNotify(multiplayMsg.DatasHPNotify, p.OnHPNotify)
	p.OnNotify(multiplayMsg.DatasReadyMultiplayGameNotify, p.OnReadyMultiplayGameNotify)
	p.OnNotify(multiplayMsg.DatasLeaveMultiplayGameNotify, p.OnLeaveMultiplayGameNotify)
	p.OnNotify(multiplayMsg.DatasEnemyHP, p.OnEnemyHPNotify)
	p.OnNotify(multiplayMsg.DatasPing, p.OnPingNotify)
	p.OnNotify(multiplayMsg.DatasChangeAvatar, p.OnChangeAvatarNotify)

	p.OnForward(p.Onforward)
}

// OnEnterMultiplayGameRpc 玩家请求信息:准备进入多人游戏
func (p *GVGPlayer) OnEnterMultiplayGameRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Info("EnterMultiplayGameRpc: %v", p.AcID)
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		return teamboss_proto.GenErrorRspPacket(reqPacket, MsgResCodeReqPacketErr)
	}
	req := new(multiplayMsg.EnterMultiplayGameReq)
	req.Init(unionTable.Bytes, unionTable.Pos)

	p.game = GVGGamesMgr.GVGGetGame(string(req.RoomID()), string(req.Secret()))
	if p.game == nil {
		logs.Error("OnEnterMultiplayGameRpc player %s game %s not found ",
			p.AcID, req.RoomID())
		return teamboss_proto.GenErrorRspPacket(reqPacket, MsgResCodeNoGameCurr)
	}
	p.AcID = string(req.AccountId())
	err := p.game.EnterPlayer(p)
	if err != nil {
		return teamboss_proto.GenErrorRspPacket(reqPacket, MsgResCodePlayerNoInGame)
	}
	return p.game.SendReqToGame(reqPacket)
}

func (p *GVGPlayer) OnGetGameStateRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Info("OnGetGameStateRpc: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return nil
	}
	return p.game.SendReqToGame(reqPacket)
}

func (p *GVGPlayer) OnGetGameDatasRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Info("OnGetGameDatasRpc: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return nil
	}
	return p.game.SendReqToGame(reqPacket)
}

func (p *GVGPlayer) OnHPNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnHPNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.SendNotifyToGame(reqPacket)
}

func (p *GVGPlayer) OnLeaveMultiplayGameNotify(reqPacket msgprocessor.IPacket) {
	logs.Info("OnLeaveMultiplayGameNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.SendNotifyToGame(reqPacket)
}

func (p *GVGPlayer) Onforward(msg []byte) {
	logs.Trace("OnForward: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.Forward(msg)
}

func (p *GVGPlayer) OnReadyMultiplayGameNotify(reqPacket msgprocessor.IPacket) {
	logs.Info("OnReadyMultiplayGameNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.SendNotifyToGame(reqPacket)
}

func (p *GVGPlayer) OnGetGameRewardsRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Info("OnGetGameRewardsRpc: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return nil
	}
	return p.game.SendReqToGame(reqPacket)
}

func (p *GVGPlayer) OnEnemyHPNotify(reqPacket msgprocessor.IPacket) {
	logs.Debug("OnEnemyHPNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.SendNotifyToGame(reqPacket)
}

func (p *GVGPlayer) OnPingNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnPingNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	pIdx := p.game.Stat.GetPlayerIdx(p.AcID)
	if pIdx != -1 {
		p.game.Stat.Player[pIdx].LastHPDeltaTime = time.Now().Unix()
	}
	p.game.SendNotifyToGame(reqPacket)
}

func (p *GVGPlayer) OnChangeAvatarNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnChangeAvatarNotify: %v", p.AcID)
	if p.game == nil {
		logs.Warn("p.game is nil")
		return
	}
	p.game.SendNotifyToGame(reqPacket)
}
