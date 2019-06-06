// automatically generated by the FlatBuffers compiler, do not modify

package fenghuomsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ReviveResp struct {
	_tab flatbuffers.Table
}

func GetRootAsReviveResp(buf []byte, offset flatbuffers.UOffsetT) *ReviveResp {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ReviveResp{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ReviveResp) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

///玩家账号ID
func (rcv *ReviveResp) AccountId() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

///玩家账号ID
///房间密码
func (rcv *ReviveResp) Hp() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

///房间密码
func (rcv *ReviveResp) MutateHp(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func ReviveRespStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func ReviveRespAddAccountId(builder *flatbuffers.Builder, accountId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(accountId), 0)
}
func ReviveRespAddHp(builder *flatbuffers.Builder, hp int32) {
	builder.PrependInt32Slot(1, hp, 0)
}
func ReviveRespEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}