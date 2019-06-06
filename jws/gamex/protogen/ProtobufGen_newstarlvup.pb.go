// Code generated by protoc-gen-go.
// source: ProtobufGen_newstarlvup.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type NEWSTARLVUP struct {
	// * 升星等级
	StarLV *uint32 `protobuf:"varint,1,req,def=0" json:"StarLV,omitempty"`
	// * 升星属性加成百分比
	Addition *float32 `protobuf:"fixed32,2,req,def=0" json:"Addition,omitempty"`
	// * 升星所需经验值
	StarUpExp *uint32 `protobuf:"varint,3,opt,def=0" json:"StarUpExp,omitempty"`
	// * 小暴击概率
	LittleBonusRate *float32 `protobuf:"fixed32,4,opt,def=0" json:"LittleBonusRate,omitempty"`
	// * 大暴击概率
	BigBonusRate *float32 `protobuf:"fixed32,5,opt,def=0" json:"BigBonusRate,omitempty"`
	// * 单次升星所需的金币
	SCCost *uint32 `protobuf:"varint,6,opt,def=0" json:"SCCost,omitempty"`
	// * 升星突破所用的道具
	StarBreakThroughItem *string `protobuf:"bytes,7,opt,def=" json:"StarBreakThroughItem,omitempty"`
	// * 升星突破所需的数量
	StarBreakThroughCost *uint32 `protobuf:"varint,8,opt,def=1" json:"StarBreakThroughCost,omitempty"`
	XXX_unrecognized     []byte  `json:"-"`
}

func (m *NEWSTARLVUP) Reset()         { *m = NEWSTARLVUP{} }
func (m *NEWSTARLVUP) String() string { return proto.CompactTextString(m) }
func (*NEWSTARLVUP) ProtoMessage()    {}

const Default_NEWSTARLVUP_StarLV uint32 = 0
const Default_NEWSTARLVUP_Addition float32 = 0
const Default_NEWSTARLVUP_StarUpExp uint32 = 0
const Default_NEWSTARLVUP_LittleBonusRate float32 = 0
const Default_NEWSTARLVUP_BigBonusRate float32 = 0
const Default_NEWSTARLVUP_SCCost uint32 = 0
const Default_NEWSTARLVUP_StarBreakThroughCost uint32 = 1

func (m *NEWSTARLVUP) GetStarLV() uint32 {
	if m != nil && m.StarLV != nil {
		return *m.StarLV
	}
	return Default_NEWSTARLVUP_StarLV
}

func (m *NEWSTARLVUP) GetAddition() float32 {
	if m != nil && m.Addition != nil {
		return *m.Addition
	}
	return Default_NEWSTARLVUP_Addition
}

func (m *NEWSTARLVUP) GetStarUpExp() uint32 {
	if m != nil && m.StarUpExp != nil {
		return *m.StarUpExp
	}
	return Default_NEWSTARLVUP_StarUpExp
}

func (m *NEWSTARLVUP) GetLittleBonusRate() float32 {
	if m != nil && m.LittleBonusRate != nil {
		return *m.LittleBonusRate
	}
	return Default_NEWSTARLVUP_LittleBonusRate
}

func (m *NEWSTARLVUP) GetBigBonusRate() float32 {
	if m != nil && m.BigBonusRate != nil {
		return *m.BigBonusRate
	}
	return Default_NEWSTARLVUP_BigBonusRate
}

func (m *NEWSTARLVUP) GetSCCost() uint32 {
	if m != nil && m.SCCost != nil {
		return *m.SCCost
	}
	return Default_NEWSTARLVUP_SCCost
}

func (m *NEWSTARLVUP) GetStarBreakThroughItem() string {
	if m != nil && m.StarBreakThroughItem != nil {
		return *m.StarBreakThroughItem
	}
	return ""
}

func (m *NEWSTARLVUP) GetStarBreakThroughCost() uint32 {
	if m != nil && m.StarBreakThroughCost != nil {
		return *m.StarBreakThroughCost
	}
	return Default_NEWSTARLVUP_StarBreakThroughCost
}

type NEWSTARLVUP_ARRAY struct {
	Items            []*NEWSTARLVUP `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *NEWSTARLVUP_ARRAY) Reset()         { *m = NEWSTARLVUP_ARRAY{} }
func (m *NEWSTARLVUP_ARRAY) String() string { return proto.CompactTextString(m) }
func (*NEWSTARLVUP_ARRAY) ProtoMessage()    {}

func (m *NEWSTARLVUP_ARRAY) GetItems() []*NEWSTARLVUP {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
