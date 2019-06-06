// Code generated by protoc-gen-go.
// source: ProtobufGen_hotgachashow.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type HOTGACHASHOW struct {
	// * 活动ID
	ActivityID       *uint32                       `protobuf:"varint,1,opt,def=0" json:"ActivityID,omitempty"`
	Item_Table       []*HOTGACHASHOW_ItemCondition `protobuf:"bytes,5,rep" json:"Item_Table,omitempty"`
	XXX_unrecognized []byte                        `json:"-"`
}

func (m *HOTGACHASHOW) Reset()         { *m = HOTGACHASHOW{} }
func (m *HOTGACHASHOW) String() string { return proto.CompactTextString(m) }
func (*HOTGACHASHOW) ProtoMessage()    {}

const Default_HOTGACHASHOW_ActivityID uint32 = 0

func (m *HOTGACHASHOW) GetActivityID() uint32 {
	if m != nil && m.ActivityID != nil {
		return *m.ActivityID
	}
	return Default_HOTGACHASHOW_ActivityID
}

func (m *HOTGACHASHOW) GetItem_Table() []*HOTGACHASHOW_ItemCondition {
	if m != nil {
		return m.Item_Table
	}
	return nil
}

type HOTGACHASHOW_ItemCondition struct {
	// * 物品ID
	ItemID *string `protobuf:"bytes,1,opt,def=" json:"ItemID,omitempty"`
	// * 物品数量
	ItemCount *uint32 `protobuf:"varint,2,opt,def=0" json:"ItemCount,omitempty"`
	// * 是否稀有
	Unusual          *uint32 `protobuf:"varint,3,opt,def=0" json:"Unusual,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *HOTGACHASHOW_ItemCondition) Reset()         { *m = HOTGACHASHOW_ItemCondition{} }
func (m *HOTGACHASHOW_ItemCondition) String() string { return proto.CompactTextString(m) }
func (*HOTGACHASHOW_ItemCondition) ProtoMessage()    {}

const Default_HOTGACHASHOW_ItemCondition_ItemCount uint32 = 0
const Default_HOTGACHASHOW_ItemCondition_Unusual uint32 = 0

func (m *HOTGACHASHOW_ItemCondition) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *HOTGACHASHOW_ItemCondition) GetItemCount() uint32 {
	if m != nil && m.ItemCount != nil {
		return *m.ItemCount
	}
	return Default_HOTGACHASHOW_ItemCondition_ItemCount
}

func (m *HOTGACHASHOW_ItemCondition) GetUnusual() uint32 {
	if m != nil && m.Unusual != nil {
		return *m.Unusual
	}
	return Default_HOTGACHASHOW_ItemCondition_Unusual
}

type HOTGACHASHOW_ARRAY struct {
	Items            []*HOTGACHASHOW `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *HOTGACHASHOW_ARRAY) Reset()         { *m = HOTGACHASHOW_ARRAY{} }
func (m *HOTGACHASHOW_ARRAY) String() string { return proto.CompactTextString(m) }
func (*HOTGACHASHOW_ARRAY) ProtoMessage()    {}

func (m *HOTGACHASHOW_ARRAY) GetItems() []*HOTGACHASHOW {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}