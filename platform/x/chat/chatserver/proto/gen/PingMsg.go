// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type PingMsg struct {
	_tab flatbuffers.Table
}

func GetRootAsPingMsg(buf []byte, offset flatbuffers.UOffsetT) *PingMsg {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &PingMsg{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *PingMsg) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *PingMsg) Ping() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func PingMsgStart(builder *flatbuffers.Builder) { builder.StartObject(1) }
func PingMsgAddPing(builder *flatbuffers.Builder, ping flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(ping), 0)
}
func PingMsgEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
