// Code generated by protoc-gen-go.
// source: ProtobufGen_starmapconfig.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type STARMAPCONFIG struct {
	// * 批量分解的品质上限
	QuickResolveLimit *uint32 `protobuf:"varint,1,req,def=0" json:"QuickResolveLimit,omitempty"`
	// * 一键占星数量上限
	OneKeyAugurNum *uint32 `protobuf:"varint,2,req,def=0" json:"OneKeyAugurNum,omitempty"`
	// * 发送跑马灯的品质下限
	HorseLampLimit   *uint32 `protobuf:"varint,3,req,def=0" json:"HorseLampLimit,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *STARMAPCONFIG) Reset()         { *m = STARMAPCONFIG{} }
func (m *STARMAPCONFIG) String() string { return proto.CompactTextString(m) }
func (*STARMAPCONFIG) ProtoMessage()    {}

const Default_STARMAPCONFIG_QuickResolveLimit uint32 = 0
const Default_STARMAPCONFIG_OneKeyAugurNum uint32 = 0
const Default_STARMAPCONFIG_HorseLampLimit uint32 = 0

func (m *STARMAPCONFIG) GetQuickResolveLimit() uint32 {
	if m != nil && m.QuickResolveLimit != nil {
		return *m.QuickResolveLimit
	}
	return Default_STARMAPCONFIG_QuickResolveLimit
}

func (m *STARMAPCONFIG) GetOneKeyAugurNum() uint32 {
	if m != nil && m.OneKeyAugurNum != nil {
		return *m.OneKeyAugurNum
	}
	return Default_STARMAPCONFIG_OneKeyAugurNum
}

func (m *STARMAPCONFIG) GetHorseLampLimit() uint32 {
	if m != nil && m.HorseLampLimit != nil {
		return *m.HorseLampLimit
	}
	return Default_STARMAPCONFIG_HorseLampLimit
}

type STARMAPCONFIG_ARRAY struct {
	Items            []*STARMAPCONFIG `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *STARMAPCONFIG_ARRAY) Reset()         { *m = STARMAPCONFIG_ARRAY{} }
func (m *STARMAPCONFIG_ARRAY) String() string { return proto.CompactTextString(m) }
func (*STARMAPCONFIG_ARRAY) ProtoMessage()    {}

func (m *STARMAPCONFIG_ARRAY) GetItems() []*STARMAPCONFIG {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
