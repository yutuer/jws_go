package fenghuo

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/common"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuomsg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuoproto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *FenghuoPlayer) Init() {
	//客户端,链接服务器后,第1个协议
	p.OnRpc(DatasEnterGameReq, p.EnterGameRpc)
	p.OnRpc(DatasReviveReq, p.ReviveMeRpc)

	p.OnNotify(DatasHPNotify, p.OnHPNotify)
	p.OnNotify(DatasStartFightNotify, p.OnStartFightNotify)

}

// EnterGameRpc 玩家请求信息:准备进入多人游戏
// FIXME by YZH  如何让这个协议有问题的时候主动断开链接, 次协议类似Handshake, 如果链接建立有指定时间内没有收到握手事件, 则主动关闭链接
func (p *FenghuoPlayer) EnterGameRpc(reqPacket msgprocessor.IPacket) []byte {
	logs.Trace("FenghuoPlayer EnterGameRpc")
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		return fenghuoproto.GenErrorRspPacket(reqPacket, common.MsgResCodeReqPacketErr)
	}

	req := new(EnterGameReq)
	req.Init(unionTable.Bytes, unionTable.Pos)

	p.AcID = string(req.AccountId())
	p.PlayerStatus = PlayerStatusWaiting

	logs.Trace("FenghuoPlayer EnterGameRpc %s %s %v", string(req.AccountId()), string(req.RoomID()), req.Secret())
	p.game = FHGamesMgr.GetGame(string(req.RoomID()), string(req.Secret()))
	if p.game == nil {
		return fenghuoproto.GenErrorRspPacket(reqPacket, common.MsgResCodeNoGameCurr)
	}

	idx, err := p.game.EnterPlayer(p)
	if err != nil {
		logs.Error("FenghuoPlayer EnterGameRpc EnterPlayer error %s", err.Error())
		return fenghuoproto.GenErrorRspPacket(reqPacket, common.MsgResCodePlayerNoInGame)
	}

	p.IDX = idx

	logs.Debug("FenghuoPlayer EnterGameRpc idx: %d", idx)

	var ehp, hp flatbuffers.UOffsetT
	builder := GetNewBuilder()
	if p.game.HpSyncable() {
		ehps := p.game.GetEnemyHPs()
		ehp = fenghuoproto.GenIntArray(builder, EnterGameRespStartEnemiesHpVector, ehps[:])

		hps := p.game.GetPlayerHPs()
		hp = fenghuoproto.GenIntArray(builder, EnterGameRespStartHpsVector, hps[:])
	}

	avatars := p.game.GenFenghuoAvatarsFlatbuffer(builder)
	avVector := fenghuoproto.GenUOffsetTArray(builder, EnterGameRespStartAvatarsVector, avatars[:])
	EnterGameRespStart(builder)
	EnterGameRespAddMyidx(builder, int32(p.IDX))
	EnterGameRespAddGamestatus(builder, int8(p.game.GetGameStatus()))
	if p.game.HpSyncable() {
		EnterGameRespAddEnemiesHp(builder, ehp)
		EnterGameRespAddHps(builder, hp)
	}
	EnterGameRespAddAvatars(builder, avVector)
	resp := EnterGameRespEnd(builder)
	return fenghuoproto.GenPacketRsp(
		builder, reqPacket,
		DatasEnterGameResp, resp)
}

func (p *FenghuoPlayer) ReviveMeRpc(reqPacket msgprocessor.IPacket) []byte {
	//FIXME by YZH FenghuoPlayer 战斗中复活协议的处理
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		return fenghuoproto.GenErrorRspPacket(reqPacket, common.MsgResCodeReqPacketErr)
	}

	if p.game == nil {
		return fenghuoproto.GenErrorRspPacket(reqPacket, 1000)
	}

	if !p.game.IsPlayerOnline(p.IDX) {
		return fenghuoproto.GenErrorRspPacket(reqPacket, 1000)
	}

	if p.PlayerStatus != PlayerStatusDead {
		return fenghuoproto.GenErrorRspPacket(reqPacket, 1001)
	}

	req := new(ReviveReq)
	req.Init(unionTable.Bytes, unionTable.Pos)

	if int(req.Idx()) != p.IDX {
		logs.Warn("FenghuoPlayer OnHPNotify got different idx")
		return fenghuoproto.GenErrorRspPacket(reqPacket, 1002)
	}

	acID := p.AcID
	p.game.SetPlayerHP(int(req.Idx()), int(req.Hp()))
	p.PlayerStatus = PlayerStatusHPSync

	builder := GetNewBuilder()
	ReviveRespStart(builder)
	ReviveRespAddAccountId(builder, builder.CreateString(acID))
	ReviveRespAddHp(builder, req.Hp())
	resp := ReviveRespEnd(builder)
	return fenghuoproto.GenPacketRsp(
		builder, reqPacket,
		DatasReviveResp, resp)
}

func (p *FenghuoPlayer) OnHPNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("FenghuoPlayer OnHPNotify")
	if p.PlayerStatus != PlayerStatusHPSync {
		logs.Warn("FenghuoPlayer OnHPNotify is not PlayerStatusHPSync.")
		return
	}

	if p.game == nil {
		return
	}

	if !p.game.IsPlayerOnline(p.IDX) {
		return
	}

	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		return
	}

	req := new(HPNotify)
	req.Init(unionTable.Bytes, unionTable.Pos)

	if int(req.Myidx()) != p.IDX {
		logs.Warn("FenghuoPlayer OnHPNotify got different idx, req:%d, p:%d", req.Myidx(), p.IDX)
		return
	}

	dead := p.game.OnHPNotify(p.IDX, req)
	if dead {
		p.PlayerStatus = PlayerStatusDead
		logs.Trace("FenghuoPlayer EnterGameRpc Player Dead idx: %d", p.IDX)
	}

	//b := fenghuoproto.ForwardHPNotifyToClient(req)
	//p.game.Channel.Broadcast(p.session)
}

func (p *FenghuoPlayer) OnStartFightNotify(reqPacket msgprocessor.IPacket) {
	logs.Trace("OnStartFightNotify %d", p.IDX)
	if p.game == nil {
		logs.Warn("OnStartFightNotify p.game is not ok")
		return
	}

	if !p.game.IsPlayerOnline(p.IDX) {
		logs.Warn("OnStartFightNotify IsPlayerOnline false")
		return
	}

	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		logs.Warn("OnStartFightNotify Data parse")
		return
	}

	req := new(StartFightNotify)
	req.Init(unionTable.Bytes, unionTable.Pos)

	p.PlayerStatus = PlayerStatusHPSync
	p.game.OnStartFightNotify(p.IDX, req)
}
