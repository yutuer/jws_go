// automatically generated by the FlatBuffers compiler, do not modify

package multiplayMsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

/// [Notify]伤害\损失HP通知
type BossData struct {
	_tab flatbuffers.Table
}

func GetRootAsBossData(buf []byte, offset flatbuffers.UOffsetT) *BossData {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &BossData{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *BossData) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *BossData) BossReleaseSkillID() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *BossData) MutateBossReleaseSkillID(n int32) bool {
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *BossData) BossComboCount() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *BossData) MutateBossComboCount(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func (rcv *BossData) BossStartAttackPos(obj *Vector) *Vector {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Vector)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *BossData) BossStartAttackDir(obj *Vector) *Vector {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Vector)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *BossData) BossStartAttackTimeStamp() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *BossData) MutateBossStartAttackTimeStamp(n int64) bool {
	return rcv._tab.MutateInt64Slot(12, n)
}

func (rcv *BossData) SimpleData(obj *BossSimpleData) *BossSimpleData {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(BossSimpleData)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func BossDataStart(builder *flatbuffers.Builder) {
	builder.StartObject(6)
}
func BossDataAddBossReleaseSkillID(builder *flatbuffers.Builder, bossReleaseSkillID int32) {
	builder.PrependInt32Slot(0, bossReleaseSkillID, 0)
}
func BossDataAddBossComboCount(builder *flatbuffers.Builder, bossComboCount int32) {
	builder.PrependInt32Slot(1, bossComboCount, 0)
}
func BossDataAddBossStartAttackPos(builder *flatbuffers.Builder, bossStartAttackPos flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(bossStartAttackPos), 0)
}
func BossDataAddBossStartAttackDir(builder *flatbuffers.Builder, bossStartAttackDir flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(bossStartAttackDir), 0)
}
func BossDataAddBossStartAttackTimeStamp(builder *flatbuffers.Builder, bossStartAttackTimeStamp int64) {
	builder.PrependInt64Slot(4, bossStartAttackTimeStamp, 0)
}
func BossDataAddSimpleData(builder *flatbuffers.Builder, simpleData flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(simpleData), 0)
}
func BossDataEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}