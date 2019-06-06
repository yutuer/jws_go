// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type RoleEnterNotify struct {
	_tab flatbuffers.Table
}

func GetRootAsRoleEnterNotify(buf []byte, offset flatbuffers.UOffsetT) *RoleEnterNotify {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &RoleEnterNotify{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *RoleEnterNotify) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *RoleEnterNotify) Role(obj *RoleInfo) *RoleInfo {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(RoleInfo)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func RoleEnterNotifyStart(builder *flatbuffers.Builder) { builder.StartObject(1) }
func RoleEnterNotifyAddRole(builder *flatbuffers.Builder, role flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(role), 0)
}
func RoleEnterNotifyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
