package teamboss

import (
	"time"

	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/teamboss_proto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *TBGame) SendNotifyToGame(reqPacket msgprocessor.IPacket) int {
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		logs.Error("TBGame.SendNotifyToGame reqPacket.Data fail %v", reqPacket)
		return MsgResCodeReqPacketErr
	}

	msg := TBGameCommandMsg{}

	switch reqPacket.DataType() {
	case DatasHPNotify:
		req := new(HPNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_hp = req
	case DatasLeaveMultiplayGameNotify:
		req := new(LeaveMultiplayGameNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_leave = req
	case DatasReadyMultiplayGameNotify:
		req := new(ReadyMultiplayGameNotify)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_ready = req
	case DatasEnemyHP:
		req := new(EnemyHP)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_enemyHP = req
	case DatasPing:
		req := new(Ping)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_ping = req
	}

	if r != nil {
		r.PushCommand(&msg)
	}
	return 0
}

func (r *TBGame) Forward(msg []byte) {
	m := TBGameCommandMsg{}
	m.msg = msg
	r.PushCommand(&m)
}

// 2. [Notify]主动离开战斗服务器(状态算退出)
func (r *TBGame) onLeaveRoom(req *LeaveMultiplayGameNotify) int {
	r.Stat.SetPlayerLeave(string(req.AccountId()))
	r.PushGameState()
	return 0
}

// 3. [Notify]准备开始战斗
func (r *TBGame) onReadyToGame(req *ReadyMultiplayGameNotify) int {
	if r.Stat.State == teamboss_proto.GameStateWaitReady || r.Stat.State == teamboss_proto.GameStateWaitOnline {
		r.Stat.SetPlayerStat(string(req.AccountId()), teamboss_proto.PlayerStateReady)
		r.Stat.SetPlayerLatency(string(req.AccountId()), req.Latency())
		if r.lead == "" {
			r.lead = string(req.AccountId())
		} else {
			leadPlayer := r.Stat.Player[r.Stat.GetPlayerIdx(r.lead)]
			if req.Latency() < leadPlayer.Latency {
				r.lead = string(req.AccountId())
			}
		}

		if idx := r.Stat.GetPlayerIdx(string(req.AccountId())); idx >= 0 {
			if req.PlayerHP() > 0 {
				r.Stat.Player[idx].Hp = int(req.PlayerHP())
			}
		}

		r.PushGameState()
	}
	return 0
}

// 4. [Notify]伤害\损失HP通知
func (r *TBGame) onHpDeta(req *HPNotify) int {
	pIdx := r.Stat.GetPlayerIdx(string(req.AccountId()))
	r.Stat.PlayerHpDeta(pIdx, int(req.PlayerHpD()))
	r.Stat.LastDamageType = req.DamageTyp()
	for i := 0; i < req.BossHpDLength(); i++ {
		bossHpD := int(req.BossHpD(i))
		r.Stat.BossHpDeta(i, bossHpD)
		r.Stat.AddHatred(i, pIdx, bossHpD)
	}

	for i := 0; i < req.BossArmorLength(); i++ {
		bossArmorDelta := req.BossArmor(i)
		r.Stat.BossArmorDeta(i, bossArmorDelta)
	}
	for i := 0; i < req.OthersHpDLength(); i++ {
		p := &PlayerState{}
		if req.OthersHpD(p, i) {
			idx := r.Stat.GetPlayerIdx(string(p.AccountID()))
			if idx != -1 {
				r.Stat.PlayerHpDeta(idx, int(p.Hp(0)))
			}
		} else {
			logs.Error("get info err on hp deta")
		}
	}
	r.PushGameState()
	return 0
}

func (r *TBGame) onEnemyHPDeta(req *EnemyHP) {
	if int(req.Waves()) > len(r.Stat.EnemyWaveHP) {
		logs.Error("receive client EnemyHP req wave: %v err", req.Waves())
		return
	}
	waveHP := r.Stat.EnemyWaveHP[req.Waves()]
	for i := range waveHP {
		waveHP[i] += uint32(req.Hp(i))
	}
	r.Stat.EnemyWaveHP[req.Waves()] = waveHP
	// select lead
	data := teamboss_proto.GenEnemyWaveHP(&r.Stat, req.Waves())
	r.broadcastMsg(data)
}

func (r *TBGame) onPing(req *Ping) {
	index := r.Stat.GetPlayerIdx(string(req.Acid()))
	if index != -1 {
		r.Stat.Player[index].LastPingTime = time.Now().Unix()
	}
}
