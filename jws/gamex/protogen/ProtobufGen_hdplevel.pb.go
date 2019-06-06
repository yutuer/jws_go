// Code generated by protoc-gen-go.
// source: ProtobufGen_hdplevel.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type HDPLEVEL struct {
	// * 奖励分类关卡ID
	HeroDiffLevel *uint32 `protobuf:"varint,1,req,def=0" json:"HeroDiffLevel,omitempty"`
	// * 对应关卡的ID
	LevelInfoID *string `protobuf:"bytes,2,req,def=" json:"LevelInfoID,omitempty"`
	// * buffID
	BuffID *string `protobuf:"bytes,5,opt,def=" json:"BuffID,omitempty"`
	// * buff叠加间隔(s）
	BuffTime         *float32             `protobuf:"fixed32,6,opt,def=0" json:"BuffTime,omitempty"`
	ItemData_Table   []*HDPLEVEL_ItemData `protobuf:"bytes,4,rep" json:"ItemData_Table,omitempty"`
	XXX_unrecognized []byte               `json:"-"`
}

func (m *HDPLEVEL) Reset()         { *m = HDPLEVEL{} }
func (m *HDPLEVEL) String() string { return proto.CompactTextString(m) }
func (*HDPLEVEL) ProtoMessage()    {}

const Default_HDPLEVEL_HeroDiffLevel uint32 = 0
const Default_HDPLEVEL_BuffTime float32 = 0

func (m *HDPLEVEL) GetHeroDiffLevel() uint32 {
	if m != nil && m.HeroDiffLevel != nil {
		return *m.HeroDiffLevel
	}
	return Default_HDPLEVEL_HeroDiffLevel
}

func (m *HDPLEVEL) GetLevelInfoID() string {
	if m != nil && m.LevelInfoID != nil {
		return *m.LevelInfoID
	}
	return ""
}

func (m *HDPLEVEL) GetBuffID() string {
	if m != nil && m.BuffID != nil {
		return *m.BuffID
	}
	return ""
}

func (m *HDPLEVEL) GetBuffTime() float32 {
	if m != nil && m.BuffTime != nil {
		return *m.BuffTime
	}
	return Default_HDPLEVEL_BuffTime
}

func (m *HDPLEVEL) GetItemData_Table() []*HDPLEVEL_ItemData {
	if m != nil {
		return m.ItemData_Table
	}
	return nil
}

type HDPLEVEL_ItemData struct {
	// * 展示道具一
	ItemID *string `protobuf:"bytes,1,req,def=" json:"ItemID,omitempty"`
	// * 道具一数量
	ItemNum          *uint32 `protobuf:"varint,2,opt,def=0" json:"ItemNum,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *HDPLEVEL_ItemData) Reset()         { *m = HDPLEVEL_ItemData{} }
func (m *HDPLEVEL_ItemData) String() string { return proto.CompactTextString(m) }
func (*HDPLEVEL_ItemData) ProtoMessage()    {}

const Default_HDPLEVEL_ItemData_ItemNum uint32 = 0

func (m *HDPLEVEL_ItemData) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *HDPLEVEL_ItemData) GetItemNum() uint32 {
	if m != nil && m.ItemNum != nil {
		return *m.ItemNum
	}
	return Default_HDPLEVEL_ItemData_ItemNum
}

type HDPLEVEL_ARRAY struct {
	Items            []*HDPLEVEL `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *HDPLEVEL_ARRAY) Reset()         { *m = HDPLEVEL_ARRAY{} }
func (m *HDPLEVEL_ARRAY) String() string { return proto.CompactTextString(m) }
func (*HDPLEVEL_ARRAY) ProtoMessage()    {}

func (m *HDPLEVEL_ARRAY) GetItems() []*HDPLEVEL {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
