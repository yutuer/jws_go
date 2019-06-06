// automatically generated by the FlatBuffers compiler, do not modify

package multiplayMsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

///[RPC]获取奖励 resp
type GetGameRewardsRsp struct {
	_tab flatbuffers.Table
}

func GetRootAsGetGameRewardsRsp(buf []byte, offset flatbuffers.UOffsetT) *GetGameRewardsRsp {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &GetGameRewardsRsp{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *GetGameRewardsRsp) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

///是否双倍(0-否 1-是)
func (rcv *GetGameRewardsRsp) IsDouble() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

///是否双倍(0-否 1-是)
func (rcv *GetGameRewardsRsp) MutateIsDouble(n int32) bool {
	return rcv._tab.MutateInt32Slot(4, n)
}

///是否使用HC双倍(0-否 1-是)
func (rcv *GetGameRewardsRsp) IsUseHc() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

///是否使用HC双倍(0-否 1-是)
func (rcv *GetGameRewardsRsp) MutateIsUseHc(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

///奖励ID列表
func (rcv *GetGameRewardsRsp) Rewards(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *GetGameRewardsRsp) RewardsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

///奖励ID列表
///奖励数量列表(已做双倍四倍处理)
func (rcv *GetGameRewardsRsp) Counts(j int) uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetUint32(a + flatbuffers.UOffsetT(j*4))
	}
	return 0
}

func (rcv *GetGameRewardsRsp) CountsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

///奖励数量列表(已做双倍四倍处理)
func GetGameRewardsRspStart(builder *flatbuffers.Builder) {
	builder.StartObject(4)
}
func GetGameRewardsRspAddIsDouble(builder *flatbuffers.Builder, IsDouble int32) {
	builder.PrependInt32Slot(0, IsDouble, 0)
}
func GetGameRewardsRspAddIsUseHc(builder *flatbuffers.Builder, IsUseHc int32) {
	builder.PrependInt32Slot(1, IsUseHc, 0)
}
func GetGameRewardsRspAddRewards(builder *flatbuffers.Builder, Rewards flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(Rewards), 0)
}
func GetGameRewardsRspStartRewardsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func GetGameRewardsRspAddCounts(builder *flatbuffers.Builder, Counts flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(Counts), 0)
}
func GetGameRewardsRspStartCountsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func GetGameRewardsRspEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
