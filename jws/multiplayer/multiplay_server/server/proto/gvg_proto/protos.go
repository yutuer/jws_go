package gvg_proto

import (
	"time"

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

func GenEnterMultiplayGameRsp(req msgprocessor.IPacket, stat *GVGGameState, datas *GVGGameDatas, acid string) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasEnterMultiplayGameRsp,
		MkEnterMultiplayGameRsp(builder, stat, datas, acid))
}

func GenGetGameDataRsp(req msgprocessor.IPacket, stat *GVGGameState, datas *GVGGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasGetGameDatasRsp,
		MkGetGameDataRsp(builder, stat, datas))
}

func GenGetGameStatRsp(req msgprocessor.IPacket, stat *GVGGameState, datas *GVGGameDatas) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		req.Typ(),
		req.Number(),
		0,
		DatasGetGameStateRsp,
		MkGetGameStatRsp(builder, stat))
}

func GenStatePush(stat *GVGGameState, lead string) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		msgprocessor.MsgTypPush,
		0,
		0,
		DatasStatePush,
		MkStatePush(builder, stat, lead))
}

func GenGetGameRwardRsp(req msgprocessor.IPacket, idx int, datas *GVGGameDatas) []byte {
	return nil
}

func GenEnemyWaveHP(state *GVGGameState, wave int32) []byte {
	builder := GetNewBuilder()
	return GenPacketRsp(builder,
		msgprocessor.MsgTypPush,
		0,
		0,
		DatasEnemyHP,
		MkEnemyWaveHP(builder, state, wave))
	return nil
}

func MkEnterMultiplayGameRsp(builder *flatbuffers.Builder, stat *GVGGameState, datas *GVGGameDatas, acid string) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

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
	nowT := time.Now()
	EnterMultiplayGameRspStart(builder)
	EnterMultiplayGameRspAddStat(builder, int32(stat.State))
	EnterMultiplayGameRspAddStartTime(builder, stat.StartTime)
	EnterMultiplayGameRspAddEndTime(builder, stat.EndTime)
	EnterMultiplayGameRspAddGameClass(builder, int32(stat.GameClass))
	EnterMultiplayGameRspAddGameScene(builder, h.Get(GameScene))
	EnterMultiplayGameRspAddPlayerStat(builder, PlayerStatVector)
	EnterMultiplayGameRspAddBossStat(builder, BossStatVector)
	EnterMultiplayGameRspAddTimeStampS(builder, nowT.Unix())
	EnterMultiplayGameRspAddTimeStampNS(builder, int32(nowT.Nanosecond()))
	EnterMultiplayGameRspAddPos(builder, int32(datas.PlayerDatas[acid].Pos))
	return EnterMultiplayGameRspEnd(builder)
}

func MkGetGameStatRsp(builder *flatbuffers.Builder, stat *GVGGameState) flatbuffers.UOffsetT {
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

func MkGetGameDataRsp(builder *flatbuffers.Builder, stat *GVGGameState, datas *GVGGameDatas) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

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
	GetGameDatasRspAddPlayerStat(builder, PlayerStatVector)
	GetGameDatasRspAddBossStat(builder, BossStatVector)

	return GetGameDatasRspEnd(builder)
}

func MkStatePush(builder *flatbuffers.Builder, stat *GVGGameState, lead string) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 32)

	GameScene := h.Pre(builder.CreateString(stat.GameScene))

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < len(stat.Player); i++ {
		Player = append(Player, GenPlayerStateWithPos(builder, stat.Player[i], lead))
	}
	PlayerStatVector := h.CreateUOffsetTArray(StatePushStartPlayerStatVector, Player)

	Boss := make([]flatbuffers.UOffsetT, 0, len(stat.Boss))
	for i := 0; i < len(stat.Boss); i++ {
		Boss = append(Boss, GenBossState(builder, stat.Boss[i]))
	}
	BossStatVector := h.CreateUOffsetTArray(StatePushStartBossStatVector, Boss)
	param := make([]string, len(stat.Param))

	Param := make([]flatbuffers.UOffsetT, 0, len(param))
	for k, v := range stat.Param {
		Param = append(Param, GenStateParam(builder, k, v))
	}
	StatePushStartBossArmorVector(builder, len(stat.Boss))
	for _, item := range stat.Boss {
		builder.PrependInt64(item.Armor)
	}
	BossArmor := builder.EndVector(len(stat.Boss))
	StatePushStart(builder)
	StatePushAddStat(builder, int32(stat.State))
	StatePushAddStartTime(builder, stat.StartTime)
	StatePushAddEndTime(builder, stat.EndTime)
	StatePushAddGameClass(builder, int32(stat.GameClass))
	StatePushAddGameScene(builder, h.Get(GameScene))
	StatePushAddPlayerStat(builder, PlayerStatVector)
	StatePushAddBossStat(builder, BossStatVector)
	StatePushAddBossArmor(builder, BossArmor)
	StatePushAddLastDamageTyp(builder, stat.LastDamageType)
	return StatePushEnd(builder)
}

func MkEnemyWaveHP(builder *flatbuffers.Builder, stat *GVGGameState, wave int32) flatbuffers.UOffsetT {
	waves := stat.EnemyWaveHP[wave]
	length := len(waves)
	EnemyHPStartHpVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependInt64(int64(waves[i]))
	}
	EnemyHPVector := builder.EndVector(length)
	EnemyHPStart(builder)
	EnemyHPAddWaves(builder, wave)
	EnemyHPAddHp(builder, EnemyHPVector)
	return EnemyHPEnd(builder)
}
