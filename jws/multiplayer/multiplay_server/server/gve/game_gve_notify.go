package gve

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/common"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GVEGame) SendNotifyToGame(acid string, reqPacket msgprocessor.IPacket) int {
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		logs.Error("GVEGame.SendNotifyToGame reqPacket.Data fail %v", reqPacket)
		return common.MsgResCodeReqPacketErr
	}

	msg := GVEGameCommandMsg{}

	switch reqPacket.DataType() {
	case multiplayMsg.DatasHPNotify:
		req := new(HPNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_hp = req
	case multiplayMsg.DatasLeaveMultiplayGameNotify:
		req := new(LeaveMultiplayGameNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_leave = req
	case multiplayMsg.DatasReadyMultiplayGameNotify:
		req := new(ReadyMultiplayGameNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_ready = req
	}


	if r != nil {
		r.PushCommand(&msg)
	} else {
		logs.Warn("SendNotifyToGame r.cmdChannel nil, acid %v, typ %v",
			acid, reqPacket.DataType())
	}

	return 0
}

// 2. [Notify]主动离开战斗服务器(状态算退出)
func (r *GVEGame) onLeaveRoom(req *LeaveMultiplayGameNotify) int {
	r.Stat.SetPlayerLeave(string(req.AccountId()))
	r.PushGameState()
	return 0
}

// 3. [Notify]准备开始战斗
func (r *GVEGame) onReadyToGame(req *ReadyMultiplayGameNotify) int {
	if r.Stat.State == gve_proto.GameStateWaitReady || r.Stat.State == gve_proto.GameStateWaitOnline {
		r.Stat.SetPlayerStat(string(req.AccountId()), gve_proto.PlayerStateReady)
		r.PushGameState()
	}
	return 0
}

// 4. [Notify]伤害\损失HP通知
func (r *GVEGame) onHpDeta(req *HPNotify) int {
	// TBD 暂时不阻拦hp变化通知
	//if r.Stat.State == data.GameStateFighting {
	pIdx := r.Stat.GetPlayerIdx(string(req.AccountId()))
	r.Stat.PlayerHpDeta(pIdx, int(req.PlayerHpD()))

	for i := 0; i < req.BossHpDLength(); i++ {
		bossHpD := int(req.BossHpD(i))
		r.Stat.BossHpDeta(i, bossHpD)
		r.Stat.AddHatred(i, pIdx, bossHpD)
	}
	r.PushGameState()
	//}
	return 0
}
