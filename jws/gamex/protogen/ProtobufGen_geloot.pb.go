// Code generated by protoc-gen-go.
// source: ProtobufGen_geloot.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GELOOT struct {
	// * 关卡ID
	EGLevelID *string `protobuf:"bytes,1,req,def=" json:"EGLevelID,omitempty"`
	// * 获得的活动积分
	GEPoint *uint32 `protobuf:"varint,2,opt,def=0" json:"GEPoint,omitempty"`
	// * 获得的杀戮值
	KillingValue     *uint32         `protobuf:"varint,3,opt,def=0" json:"KillingValue,omitempty"`
	Fixed_Loot       []*GELOOT_Loot1 `protobuf:"bytes,4,rep" json:"Fixed_Loot,omitempty"`
	Random_Loot      []*GELOOT_Loot2 `protobuf:"bytes,5,rep" json:"Random_Loot,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *GELOOT) Reset()         { *m = GELOOT{} }
func (m *GELOOT) String() string { return proto.CompactTextString(m) }
func (*GELOOT) ProtoMessage()    {}

const Default_GELOOT_GEPoint uint32 = 0
const Default_GELOOT_KillingValue uint32 = 0

func (m *GELOOT) GetEGLevelID() string {
	if m != nil && m.EGLevelID != nil {
		return *m.EGLevelID
	}
	return ""
}

func (m *GELOOT) GetGEPoint() uint32 {
	if m != nil && m.GEPoint != nil {
		return *m.GEPoint
	}
	return Default_GELOOT_GEPoint
}

func (m *GELOOT) GetKillingValue() uint32 {
	if m != nil && m.KillingValue != nil {
		return *m.KillingValue
	}
	return Default_GELOOT_KillingValue
}

func (m *GELOOT) GetFixed_Loot() []*GELOOT_Loot1 {
	if m != nil {
		return m.Fixed_Loot
	}
	return nil
}

func (m *GELOOT) GetRandom_Loot() []*GELOOT_Loot2 {
	if m != nil {
		return m.Random_Loot
	}
	return nil
}

type GELOOT_Loot1 struct {
	// * 固定掉落类型
	FixedLootID *string `protobuf:"bytes,1,opt,def=" json:"FixedLootID,omitempty"`
	// * 固定掉落数量
	FixedLootNumber  *uint32 `protobuf:"varint,2,opt,def=0" json:"FixedLootNumber,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GELOOT_Loot1) Reset()         { *m = GELOOT_Loot1{} }
func (m *GELOOT_Loot1) String() string { return proto.CompactTextString(m) }
func (*GELOOT_Loot1) ProtoMessage()    {}

const Default_GELOOT_Loot1_FixedLootNumber uint32 = 0

func (m *GELOOT_Loot1) GetFixedLootID() string {
	if m != nil && m.FixedLootID != nil {
		return *m.FixedLootID
	}
	return ""
}

func (m *GELOOT_Loot1) GetFixedLootNumber() uint32 {
	if m != nil && m.FixedLootNumber != nil {
		return *m.FixedLootNumber
	}
	return Default_GELOOT_Loot1_FixedLootNumber
}

type GELOOT_Loot2 struct {
	// * 掉落组ID
	LootGroupID *string `protobuf:"bytes,1,opt,def=" json:"LootGroupID,omitempty"`
	// * 掉落概率
	LootProbability  *float32 `protobuf:"fixed32,2,opt,def=0" json:"LootProbability,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *GELOOT_Loot2) Reset()         { *m = GELOOT_Loot2{} }
func (m *GELOOT_Loot2) String() string { return proto.CompactTextString(m) }
func (*GELOOT_Loot2) ProtoMessage()    {}

const Default_GELOOT_Loot2_LootProbability float32 = 0

func (m *GELOOT_Loot2) GetLootGroupID() string {
	if m != nil && m.LootGroupID != nil {
		return *m.LootGroupID
	}
	return ""
}

func (m *GELOOT_Loot2) GetLootProbability() float32 {
	if m != nil && m.LootProbability != nil {
		return *m.LootProbability
	}
	return Default_GELOOT_Loot2_LootProbability
}

type GELOOT_ARRAY struct {
	Items            []*GELOOT `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *GELOOT_ARRAY) Reset()         { *m = GELOOT_ARRAY{} }
func (m *GELOOT_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GELOOT_ARRAY) ProtoMessage()    {}

func (m *GELOOT_ARRAY) GetItems() []*GELOOT {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}