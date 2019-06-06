// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type BroadCastMsg struct {
	_tab flatbuffers.Table
}

func GetRootAsBroadCastMsg(buf []byte, offset flatbuffers.UOffsetT) *BroadCastMsg {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &BroadCastMsg{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *BroadCastMsg) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *BroadCastMsg) Typ() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *BroadCastMsg) Msg() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func BroadCastMsgStart(builder *flatbuffers.Builder) { builder.StartObject(2) }
func BroadCastMsgAddTyp(builder *flatbuffers.Builder, typ flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(typ), 0)
}
func BroadCastMsgAddMsg(builder *flatbuffers.Builder, msg flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(msg), 0)
}
func BroadCastMsgEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
