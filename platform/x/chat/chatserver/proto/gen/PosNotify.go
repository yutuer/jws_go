// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type PosNotify struct {
	_tab flatbuffers.Table
}

func GetRootAsPosNotify(buf []byte, offset flatbuffers.UOffsetT) *PosNotify {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &PosNotify{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *PosNotify) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *PosNotify) AccountId() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *PosNotify) Name() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *PosNotify) Pos(obj *RolePos) *RolePos {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(RolePos)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func PosNotifyStart(builder *flatbuffers.Builder) { builder.StartObject(3) }
func PosNotifyAddAccountId(builder *flatbuffers.Builder, accountId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(accountId), 0)
}
func PosNotifyAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(name), 0)
}
func PosNotifyAddPos(builder *flatbuffers.Builder, pos flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(pos), 0)
}
func PosNotifyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
