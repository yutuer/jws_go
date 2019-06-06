package gve_proto

import (
	"github.com/google/flatbuffers/go"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
)

func MkStateTest(builder *flatbuffers.Builder, stat *GVEGameState) []byte {
	builder.Reset()
	h := NewFlatBufferHelper(builder, 32)

	Player := make([]flatbuffers.UOffsetT, 0, len(stat.Player))
	for i := 0; i < 1; i++ {
		Player = append(Player, GenPlayerState(builder, stat.Player[i]))
	}
	PlayerStatVector := h.CreateUOffsetTArray(TestMultStartPlayerStatVector, Player)

	TestMultStart(builder)
	TestMultAddPlayerStat(builder, PlayerStatVector)

	builder.Finish(TestMultEnd(builder))
	return builder.Bytes[builder.Head():]
}
