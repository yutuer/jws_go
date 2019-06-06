// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ProtoWrap struct {
	_tab flatbuffers.Table
}

func GetRootAsProtoWrap(buf []byte, offset flatbuffers.UOffsetT) *ProtoWrap {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ProtoWrap{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ProtoWrap) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ProtoWrap) Id() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *ProtoWrap) SubProto(j int) int8 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt8(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *ProtoWrap) SubProtoLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func ProtoWrapStart(builder *flatbuffers.Builder) { builder.StartObject(2) }
func ProtoWrapAddId(builder *flatbuffers.Builder, id flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(id), 0)
}
func ProtoWrapAddSubProto(builder *flatbuffers.Builder, subProto flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(subProto), 0)
}
func ProtoWrapStartSubProtoVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func ProtoWrapEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
