// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type RegResp struct {
	_tab flatbuffers.Table
}

func GetRootAsRegResp(buf []byte, offset flatbuffers.UOffsetT) *RegResp {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &RegResp{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *RegResp) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *RegResp) RoomId() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *RegResp) Roles(obj *RoleInfo, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		if obj == nil {
			obj = new(RoleInfo)
		}
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *RegResp) RolesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func RegRespStart(builder *flatbuffers.Builder) { builder.StartObject(2) }
func RegRespAddRoomId(builder *flatbuffers.Builder, roomId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(roomId), 0)
}
func RegRespAddRoles(builder *flatbuffers.Builder, roles flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(roles), 0)
}
func RegRespStartRolesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func RegRespEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
