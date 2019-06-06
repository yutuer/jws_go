// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ErrMsg struct {
	_tab flatbuffers.Table
}

func GetRootAsErrMsg(buf []byte, offset flatbuffers.UOffsetT) *ErrMsg {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ErrMsg{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ErrMsg) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ErrMsg) Err() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func ErrMsgStart(builder *flatbuffers.Builder) { builder.StartObject(1) }
func ErrMsgAddErr(builder *flatbuffers.Builder, err flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(err), 0)
}
func ErrMsgEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
