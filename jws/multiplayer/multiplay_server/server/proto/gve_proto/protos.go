package gve_proto

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
)

const (
	MsgIDPushIDNull = iota
	MsgIDStatePush
)

func mkBool(isOK bool) int32 {
	if isOK {
		return 1
	} else {
		return 0
	}
}

func GenErrorRspPacket(req msgprocessor.IPacket, code int) []byte {
	builder := GetNewBuilder()
	PacketStart(builder)
	PacketAddTyp(builder, int32(msgprocessor.MsgTypRpc))
	PacketAddNumber(builder, req.Number())
	PacketAddCode(builder, int32(code))
	builder.Finish(PacketEnd(builder))
	return builder.Bytes[builder.Head():]
}

func GenPacketRsp(builder *flatbuffers.Builder, typ int32, number int64, code int32, dataTyp byte, datas flatbuffers.UOffsetT) []byte {
	PacketStart(builder)
	PacketAddTyp(builder, (typ))
	PacketAddNumber(builder, (number))
	PacketAddDataType(builder, dataTyp)
	PacketAddData(builder, datas)
	PacketAddCode(builder, (code))
	builder.Finish(PacketEnd(builder))
	return builder.Bytes[builder.Head():]
}

func GenEnterMultiplayGameRsp(req msgprocessor.IPacket, stat *GVEGameState, datas *GVEGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasEnterMultiplayGameRsp,
		MkEnterMultiplayGameRsp(builder, stat, datas))
}

func GenGetGameDataRsp(req msgprocessor.IPacket, stat *GVEGameState, datas *GVEGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasGetGameDatasRsp,
		MkGetGameDataRsp(builder, stat, datas))
}

func GenGetGameStatRsp(req msgprocessor.IPacket, stat *GVEGameState, datas *GVEGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasGetGameStateRsp,
		MkGetGameStatRsp(builder, stat))
}

func GenStatePush(stat *GVEGameState) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		msgprocessor.MsgTypPush,
		0,
		0,
		DatasStatePush,
		MkStatePush(builder, stat))
}

func GenGetGameRwardRsp(req msgprocessor.IPacket, idx int, datas *GVEGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasGetGameRewardsRsp,
		MkGetGameRewardRsp(builder, idx, datas))
}

func MkEnterMultiplayGameRsp(builder *flatbuffers.Builder, stat *GVEGameState, datas *GVEGameDatas) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

	BossAcDatas := make([]flatbuffers.UOffsetT, 0, len(datas.BossAcDatas))
	for i := 0; i < len(datas.BossAcDatas); i++ {
		BossAcDatas = append(BossAcDatas, GenGVEAcData(builder, i, datas.BossModel[i], datas.BossAcDatas[i]))
	}
	AcDatasVector := h.CreateUOffsetTArray(EnterMultiplayGameRspStartAcDatasVector, BossAcDatas)

	PlayerDatas := make([]flatbuffers.UOffsetT, 0, len(datas.PlayerDatas))
	for i := 0; i < len(datas.PlayerDatas); i++ {
		PlayerDatas = append(PlayerDatas, GenAccountInfoData(builder, i, &(datas.PlayerDatas[i])))
	}
	AccDatasVector := h.CreateUOffsetTArray(EnterMultiplayGameRspStartAccDatasVector, PlayerDatas)

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < len(stat.Player); i++ {
		Player = append(Player, GenPlayerState(builder, stat.Player[i]))
	}
	PlayerStatVector := h.CreateUOffsetTArray(EnterMultiplayGameRspStartPlayerStatVector, Player)

	Boss := make([]flatbuffers.UOffsetT, 0, len(stat.Boss))
	for i := 0; i < len(stat.Boss); i++ {
		Boss = append(Boss, GenBossState(builder, stat.Boss[i]))
	}
	BossStatVector := h.CreateUOffsetTArray(EnterMultiplayGameRspStartBossStatVector, Boss)

	EnterMultiplayGameRspStart(builder)
	EnterMultiplayGameRspAddStat(builder, int32(stat.State))
	EnterMultiplayGameRspAddStartTime(builder, stat.StartTime)
	EnterMultiplayGameRspAddEndTime(builder, stat.EndTime)
	EnterMultiplayGameRspAddGameClass(builder, int32(stat.GameClass))
	EnterMultiplayGameRspAddGameScene(builder, h.Get(GameScene))
	EnterMultiplayGameRspAddAccDatas(builder, AccDatasVector)
	EnterMultiplayGameRspAddAcDatas(builder, AcDatasVector)
	EnterMultiplayGameRspAddPlayerStat(builder, PlayerStatVector)
	EnterMultiplayGameRspAddBossStat(builder, BossStatVector)

	return EnterMultiplayGameRspEnd(builder)
}

func MkGetGameStatRsp(builder *flatbuffers.Builder, stat *GVEGameState) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < len(stat.Player); i++ {
		Player = append(Player, GenPlayerState(builder, stat.Player[i]))
	}
	PlayerStatVector := h.CreateUOffsetTArray(GetGameStateRspStartPlayerStatVector, Player)

	Boss := make([]flatbuffers.UOffsetT, 0, len(stat.Boss))
	for i := 0; i < len(stat.Boss); i++ {
		Boss = append(Boss, GenBossState(builder, stat.Boss[i]))
	}
	BossStatVector := h.CreateUOffsetTArray(GetGameStateRspStartBossStatVector, Boss)

	GetGameStateRspStart(builder)
	GetGameStateRspAddStat(builder, int32(stat.State))
	GetGameStateRspAddStartTime(builder, stat.StartTime)
	GetGameStateRspAddEndTime(builder, stat.EndTime)
	GetGameStateRspAddGameClass(builder, int32(stat.GameClass))
	GetGameStateRspAddGameScene(builder, h.Get(GameScene))
	GetGameStateRspAddPlayerStat(builder, PlayerStatVector)
	GetGameStateRspAddBossStat(builder, BossStatVector)

	return GetGameStateRspEnd(builder)
}

func MkGetGameDataRsp(builder *flatbuffers.Builder, stat *GVEGameState, datas *GVEGameDatas) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

	BossAcDatas := make([]flatbuffers.UOffsetT, 0, len(datas.BossAcDatas))
	for i := 0; i < len(datas.BossAcDatas); i++ {
		BossAcDatas = append(BossAcDatas, GenGVEAcData(builder, i, datas.BossModel[i], datas.BossAcDatas[i]))
	}
	AcDatasVector := h.CreateUOffsetTArray(GetGameDatasRspStartAcDatasVector, BossAcDatas)

	PlayerDatas := make([]flatbuffers.UOffsetT, 0, len(datas.PlayerDatas))
	for i := 0; i < len(datas.PlayerDatas); i++ {
		PlayerDatas = append(PlayerDatas, GenAccountInfoData(builder, i, &(datas.PlayerDatas[i])))
	}
	AccDatasVector := h.CreateUOffsetTArray(GetGameDatasRspStartAccDatasVector, PlayerDatas)

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < len(stat.Player); i++ {
		Player = append(Player, GenPlayerState(builder, stat.Player[i]))
	}
	PlayerStatVector := h.CreateUOffsetTArray(GetGameDatasRspStartPlayerStatVector, Player)

	Boss := make([]flatbuffers.UOffsetT, 0, len(stat.Boss))
	for i := 0; i < len(stat.Boss); i++ {
		Boss = append(Boss, GenBossState(builder, stat.Boss[i]))
	}
	BossStatVector := h.CreateUOffsetTArray(GetGameDatasRspStartBossStatVector, Boss)

	GetGameDatasRspStart(builder)
	GetGameDatasRspAddStat(builder, int32(stat.State))
	GetGameDatasRspAddStartTime(builder, stat.StartTime)
	GetGameDatasRspAddEndTime(builder, stat.EndTime)
	GetGameDatasRspAddGameClass(builder, int32(stat.GameClass))
	GetGameDatasRspAddGameScene(builder, h.Get(GameScene))
	GetGameDatasRspAddAccDatas(builder, AccDatasVector)
	GetGameDatasRspAddAcDatas(builder, AcDatasVector)
	GetGameDatasRspAddPlayerStat(builder, PlayerStatVector)
	GetGameDatasRspAddBossStat(builder, BossStatVector)

	return GetGameDatasRspEnd(builder)
}

func MkStatePush(builder *flatbuffers.Builder, stat *GVEGameState) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < len(stat.Player); i++ {
		Player = append(Player, GenPlayerState(builder, stat.Player[i]))
	}
	PlayerStatVector := h.CreateUOffsetTArray(StatePushStartPlayerStatVector, Player)

	Boss := make([]flatbuffers.UOffsetT, 0, len(stat.Boss))
	for i := 0; i < len(stat.Boss); i++ {
		Boss = append(Boss, GenBossState(builder, stat.Boss[i]))
	}
	BossStatVector := h.CreateUOffsetTArray(StatePushStartBossStatVector, Boss)

	StatePushStart(builder)
	StatePushAddStat(builder, int32(stat.State))
	StatePushAddStartTime(builder, stat.StartTime)
	StatePushAddEndTime(builder, stat.EndTime)
	StatePushAddGameClass(builder, int32(stat.GameClass))
	StatePushAddGameScene(builder, h.Get(GameScene))
	StatePushAddPlayerStat(builder, PlayerStatVector)
	StatePushAddBossStat(builder, BossStatVector)

	return StatePushEnd(builder)
}

func MkGetGameRewardRsp(builder *flatbuffers.Builder, idx int, datas *GVEGameDatas) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	Rewards := h.CreateStringArray(GetGameRewardsRspStartRewardsVector, datas.PlayerDatas[idx].Reward)
	Counts := h.CreateUInt32Array(GetGameRewardsRspStartCountsVector, datas.PlayerDatas[idx].Count)

	GetGameRewardsRspStart(builder)
	GetGameRewardsRspAddIsDouble(builder, mkBool(datas.PlayerDatas[idx].IsDouble))
	GetGameRewardsRspAddIsUseHc(builder, mkBool(datas.PlayerDatas[idx].IsUseHc))
	GetGameRewardsRspAddRewards(builder, Rewards)
	GetGameRewardsRspAddCounts(builder, Counts)

	return GetGameRewardsRspEnd(builder)
}
