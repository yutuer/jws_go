// Code generated by protoc-gen-go.
// source: ProtobufGen_tbossvipcontrol.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type TBOSSVIPCONTROL struct {
	// * 主Key
	ID *uint32 `protobuf:"varint,1,req,def=0" json:"ID,omitempty"`
	// * 对应VIP区间
	VIPLower *uint32 `protobuf:"varint,2,opt,def=0" json:"VIPLower,omitempty"`
	// * 对应VIP区间
	VIPUpper *uint32 `protobuf:"varint,3,opt,def=0" json:"VIPUpper,omitempty"`
	// * x次内未得金色及以上品质，则第x+1次走特殊组掉落
	GoodBoxControl *uint32 `protobuf:"varint,4,opt,def=0" json:"GoodBoxControl,omitempty"`
	// * 分子N值
	SepcialN *uint32 `protobuf:"varint,5,opt,def=0" json:"SepcialN,omitempty"`
	// * 分母M值
	SepcialM         *uint32 `protobuf:"varint,6,opt,def=0" json:"SepcialM,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TBOSSVIPCONTROL) Reset()         { *m = TBOSSVIPCONTROL{} }
func (m *TBOSSVIPCONTROL) String() string { return proto.CompactTextString(m) }
func (*TBOSSVIPCONTROL) ProtoMessage()    {}

const Default_TBOSSVIPCONTROL_ID uint32 = 0
const Default_TBOSSVIPCONTROL_VIPLower uint32 = 0
const Default_TBOSSVIPCONTROL_VIPUpper uint32 = 0
const Default_TBOSSVIPCONTROL_GoodBoxControl uint32 = 0
const Default_TBOSSVIPCONTROL_SepcialN uint32 = 0
const Default_TBOSSVIPCONTROL_SepcialM uint32 = 0

func (m *TBOSSVIPCONTROL) GetID() uint32 {
	if m != nil && m.ID != nil {
		return *m.ID
	}
	return Default_TBOSSVIPCONTROL_ID
}

func (m *TBOSSVIPCONTROL) GetVIPLower() uint32 {
	if m != nil && m.VIPLower != nil {
		return *m.VIPLower
	}
	return Default_TBOSSVIPCONTROL_VIPLower
}

func (m *TBOSSVIPCONTROL) GetVIPUpper() uint32 {
	if m != nil && m.VIPUpper != nil {
		return *m.VIPUpper
	}
	return Default_TBOSSVIPCONTROL_VIPUpper
}

func (m *TBOSSVIPCONTROL) GetGoodBoxControl() uint32 {
	if m != nil && m.GoodBoxControl != nil {
		return *m.GoodBoxControl
	}
	return Default_TBOSSVIPCONTROL_GoodBoxControl
}

func (m *TBOSSVIPCONTROL) GetSepcialN() uint32 {
	if m != nil && m.SepcialN != nil {
		return *m.SepcialN
	}
	return Default_TBOSSVIPCONTROL_SepcialN
}

func (m *TBOSSVIPCONTROL) GetSepcialM() uint32 {
	if m != nil && m.SepcialM != nil {
		return *m.SepcialM
	}
	return Default_TBOSSVIPCONTROL_SepcialM
}

type TBOSSVIPCONTROL_ARRAY struct {
	Items            []*TBOSSVIPCONTROL `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *TBOSSVIPCONTROL_ARRAY) Reset()         { *m = TBOSSVIPCONTROL_ARRAY{} }
func (m *TBOSSVIPCONTROL_ARRAY) String() string { return proto.CompactTextString(m) }
func (*TBOSSVIPCONTROL_ARRAY) ProtoMessage()    {}

func (m *TBOSSVIPCONTROL_ARRAY) GetItems() []*TBOSSVIPCONTROL {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
