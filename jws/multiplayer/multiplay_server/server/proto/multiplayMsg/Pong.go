// automatically generated by the FlatBuffers compiler, do not modify

package multiplayMsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Pong struct {
	_tab flatbuffers.Table
}

func GetRootAsPong(buf []byte, offset flatbuffers.UOffsetT) *Pong {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Pong{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Pong) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func PongStart(builder *flatbuffers.Builder) {
	builder.StartObject(0)
}
func PongEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
