// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ReassignNotify struct {
	_tab flatbuffers.Table
}

func GetRootAsReassignNotify(buf []byte, offset flatbuffers.UOffsetT) *ReassignNotify {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ReassignNotify{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ReassignNotify) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func ReassignNotifyStart(builder *flatbuffers.Builder)                    { builder.StartObject(0) }
func ReassignNotifyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
