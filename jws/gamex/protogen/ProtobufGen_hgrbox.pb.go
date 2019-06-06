// Code generated by protoc-gen-go.
// source: ProtobufGen_hgrbox.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type HGRBOX struct {
	// * 运营类型ID
	ActivityID *uint32 `protobuf:"varint,4,opt,def=0" json:"ActivityID,omitempty"`
	// * ID
	BoxID *uint32 `protobuf:"varint,1,opt,def=0" json:"BoxID,omitempty"`
	// * 需要达到的积分
	NeedPoint        *uint32            `protobuf:"varint,2,opt,def=0" json:"NeedPoint,omitempty"`
	Loot_Table       []*HGRBOX_LootRule `protobuf:"bytes,3,rep" json:"Loot_Table,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *HGRBOX) Reset()         { *m = HGRBOX{} }
func (m *HGRBOX) String() string { return proto.CompactTextString(m) }
func (*HGRBOX) ProtoMessage()    {}

const Default_HGRBOX_ActivityID uint32 = 0
const Default_HGRBOX_BoxID uint32 = 0
const Default_HGRBOX_NeedPoint uint32 = 0

func (m *HGRBOX) GetActivityID() uint32 {
	if m != nil && m.ActivityID != nil {
		return *m.ActivityID
	}
	return Default_HGRBOX_ActivityID
}

func (m *HGRBOX) GetBoxID() uint32 {
	if m != nil && m.BoxID != nil {
		return *m.BoxID
	}
	return Default_HGRBOX_BoxID
}

func (m *HGRBOX) GetNeedPoint() uint32 {
	if m != nil && m.NeedPoint != nil {
		return *m.NeedPoint
	}
	return Default_HGRBOX_NeedPoint
}

func (m *HGRBOX) GetLoot_Table() []*HGRBOX_LootRule {
	if m != nil {
		return m.Loot_Table
	}
	return nil
}

type HGRBOX_LootRule struct {
	// * 掉落1
	ItemID *string `protobuf:"bytes,1,opt,def=" json:"ItemID,omitempty"`
	// * 掉落数量
	ItemNum          *uint32 `protobuf:"varint,2,opt,def=0" json:"ItemNum,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *HGRBOX_LootRule) Reset()         { *m = HGRBOX_LootRule{} }
func (m *HGRBOX_LootRule) String() string { return proto.CompactTextString(m) }
func (*HGRBOX_LootRule) ProtoMessage()    {}

const Default_HGRBOX_LootRule_ItemNum uint32 = 0

func (m *HGRBOX_LootRule) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *HGRBOX_LootRule) GetItemNum() uint32 {
	if m != nil && m.ItemNum != nil {
		return *m.ItemNum
	}
	return Default_HGRBOX_LootRule_ItemNum
}

type HGRBOX_ARRAY struct {
	Items            []*HGRBOX `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *HGRBOX_ARRAY) Reset()         { *m = HGRBOX_ARRAY{} }
func (m *HGRBOX_ARRAY) String() string { return proto.CompactTextString(m) }
func (*HGRBOX_ARRAY) ProtoMessage()    {}

func (m *HGRBOX_ARRAY) GetItems() []*HGRBOX {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
