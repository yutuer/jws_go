// automatically generated, do not modify

package gen

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type UnRegNotify struct {
	_tab flatbuffers.Table
}

func GetRootAsUnRegNotify(buf []byte, offset flatbuffers.UOffsetT) *UnRegNotify {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &UnRegNotify{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *UnRegNotify) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *UnRegNotify) AccountId() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *UnRegNotify) Name() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func UnRegNotifyStart(builder *flatbuffers.Builder) { builder.StartObject(2) }
func UnRegNotifyAddAccountId(builder *flatbuffers.Builder, accountId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(accountId), 0)
}
func UnRegNotifyAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(name), 0)
}
func UnRegNotifyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
