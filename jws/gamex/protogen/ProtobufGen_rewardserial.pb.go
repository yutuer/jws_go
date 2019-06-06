// Code generated by protoc-gen-go.
// source: ProtobufGen_rewardserial.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type REWARDSERIAL struct {
	// * 小组ID
	SerialID *uint32 `protobuf:"varint,1,req,def=0" json:"SerialID,omitempty"`
	// * 组内序号
	SubID *uint32 `protobuf:"varint,2,opt,def=0" json:"SubID,omitempty"`
	// * 角色限定
	RoleLimit *uint32 `protobuf:"varint,3,opt,def=0" json:"RoleLimit,omitempty"`
	// * 物品ID
	ItemID *string `protobuf:"bytes,4,opt,def=" json:"ItemID,omitempty"`
	// * 物品数量
	Count            *uint32 `protobuf:"varint,5,opt,def=0" json:"Count,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *REWARDSERIAL) Reset()         { *m = REWARDSERIAL{} }
func (m *REWARDSERIAL) String() string { return proto.CompactTextString(m) }
func (*REWARDSERIAL) ProtoMessage()    {}

const Default_REWARDSERIAL_SerialID uint32 = 0
const Default_REWARDSERIAL_SubID uint32 = 0
const Default_REWARDSERIAL_RoleLimit uint32 = 0
const Default_REWARDSERIAL_Count uint32 = 0

func (m *REWARDSERIAL) GetSerialID() uint32 {
	if m != nil && m.SerialID != nil {
		return *m.SerialID
	}
	return Default_REWARDSERIAL_SerialID
}

func (m *REWARDSERIAL) GetSubID() uint32 {
	if m != nil && m.SubID != nil {
		return *m.SubID
	}
	return Default_REWARDSERIAL_SubID
}

func (m *REWARDSERIAL) GetRoleLimit() uint32 {
	if m != nil && m.RoleLimit != nil {
		return *m.RoleLimit
	}
	return Default_REWARDSERIAL_RoleLimit
}

func (m *REWARDSERIAL) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *REWARDSERIAL) GetCount() uint32 {
	if m != nil && m.Count != nil {
		return *m.Count
	}
	return Default_REWARDSERIAL_Count
}

type REWARDSERIAL_ARRAY struct {
	Items            []*REWARDSERIAL `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *REWARDSERIAL_ARRAY) Reset()         { *m = REWARDSERIAL_ARRAY{} }
func (m *REWARDSERIAL_ARRAY) String() string { return proto.CompactTextString(m) }
func (*REWARDSERIAL_ARRAY) ProtoMessage()    {}

func (m *REWARDSERIAL_ARRAY) GetItems() []*REWARDSERIAL {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}