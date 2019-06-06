// automatically generated by the FlatBuffers compiler, do not modify

package multiplayMsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

/// [EnemyHP]伤害\损失HP通知
type EnemyHP struct {
	_tab flatbuffers.Table
}

func GetRootAsEnemyHP(buf []byte, offset flatbuffers.UOffsetT) *EnemyHP {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &EnemyHP{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *EnemyHP) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *EnemyHP) Waves() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *EnemyHP) MutateWaves(n int32) bool {
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *EnemyHP) Hp(j int) int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt64(a + flatbuffers.UOffsetT(j*8))
	}
	return 0
}

func (rcv *EnemyHP) HpLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func EnemyHPStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func EnemyHPAddWaves(builder *flatbuffers.Builder, waves int32) {
	builder.PrependInt32Slot(0, waves, 0)
}
func EnemyHPAddHp(builder *flatbuffers.Builder, hp flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(hp), 0)
}
func EnemyHPStartHpVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(8, numElems, 8)
}
func EnemyHPEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
