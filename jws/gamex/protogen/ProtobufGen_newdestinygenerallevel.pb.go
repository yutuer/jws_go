// Code generated by protoc-gen-go.
// source: ProtobufGen_newdestinygenerallevel.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type NEWDESTINYGENERALLEVEL struct {
	// * 神将ID
	DestinyGeneralID *uint32 `protobuf:"varint,1,req,def=0" json:"DestinyGeneralID,omitempty"`
	// * 神将等级ID
	DestinyGeneralLevelID *uint32 `protobuf:"varint,2,opt,def=0" json:"DestinyGeneralLevelID,omitempty"`
	// * 升到本级的经验
	DestinyGeneralExp *uint32 `protobuf:"varint,3,opt,def=0" json:"DestinyGeneralExp,omitempty"`
	// * 小暴击概率
	DGLittleBonusRate *float32 `protobuf:"fixed32,4,opt,def=0" json:"DGLittleBonusRate,omitempty"`
	// * 大暴击概率
	DGBigBonusRate *float32 `protobuf:"fixed32,5,opt,def=0" json:"DGBigBonusRate,omitempty"`
	// * 累计攻击力
	AttackIncrease *uint32 `protobuf:"varint,6,opt,def=0" json:"AttackIncrease,omitempty"`
	// * 累计防御力
	DefenseIncrease *uint32 `protobuf:"varint,7,opt,def=0" json:"DefenseIncrease,omitempty"`
	// * 累计生命值
	HPIncrease       *uint32 `protobuf:"varint,8,opt,def=0" json:"HPIncrease,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NEWDESTINYGENERALLEVEL) Reset()         { *m = NEWDESTINYGENERALLEVEL{} }
func (m *NEWDESTINYGENERALLEVEL) String() string { return proto.CompactTextString(m) }
func (*NEWDESTINYGENERALLEVEL) ProtoMessage()    {}

const Default_NEWDESTINYGENERALLEVEL_DestinyGeneralID uint32 = 0
const Default_NEWDESTINYGENERALLEVEL_DestinyGeneralLevelID uint32 = 0
const Default_NEWDESTINYGENERALLEVEL_DestinyGeneralExp uint32 = 0
const Default_NEWDESTINYGENERALLEVEL_DGLittleBonusRate float32 = 0
const Default_NEWDESTINYGENERALLEVEL_DGBigBonusRate float32 = 0
const Default_NEWDESTINYGENERALLEVEL_AttackIncrease uint32 = 0
const Default_NEWDESTINYGENERALLEVEL_DefenseIncrease uint32 = 0
const Default_NEWDESTINYGENERALLEVEL_HPIncrease uint32 = 0

func (m *NEWDESTINYGENERALLEVEL) GetDestinyGeneralID() uint32 {
	if m != nil && m.DestinyGeneralID != nil {
		return *m.DestinyGeneralID
	}
	return Default_NEWDESTINYGENERALLEVEL_DestinyGeneralID
}

func (m *NEWDESTINYGENERALLEVEL) GetDestinyGeneralLevelID() uint32 {
	if m != nil && m.DestinyGeneralLevelID != nil {
		return *m.DestinyGeneralLevelID
	}
	return Default_NEWDESTINYGENERALLEVEL_DestinyGeneralLevelID
}

func (m *NEWDESTINYGENERALLEVEL) GetDestinyGeneralExp() uint32 {
	if m != nil && m.DestinyGeneralExp != nil {
		return *m.DestinyGeneralExp
	}
	return Default_NEWDESTINYGENERALLEVEL_DestinyGeneralExp
}

func (m *NEWDESTINYGENERALLEVEL) GetDGLittleBonusRate() float32 {
	if m != nil && m.DGLittleBonusRate != nil {
		return *m.DGLittleBonusRate
	}
	return Default_NEWDESTINYGENERALLEVEL_DGLittleBonusRate
}

func (m *NEWDESTINYGENERALLEVEL) GetDGBigBonusRate() float32 {
	if m != nil && m.DGBigBonusRate != nil {
		return *m.DGBigBonusRate
	}
	return Default_NEWDESTINYGENERALLEVEL_DGBigBonusRate
}

func (m *NEWDESTINYGENERALLEVEL) GetAttackIncrease() uint32 {
	if m != nil && m.AttackIncrease != nil {
		return *m.AttackIncrease
	}
	return Default_NEWDESTINYGENERALLEVEL_AttackIncrease
}

func (m *NEWDESTINYGENERALLEVEL) GetDefenseIncrease() uint32 {
	if m != nil && m.DefenseIncrease != nil {
		return *m.DefenseIncrease
	}
	return Default_NEWDESTINYGENERALLEVEL_DefenseIncrease
}

func (m *NEWDESTINYGENERALLEVEL) GetHPIncrease() uint32 {
	if m != nil && m.HPIncrease != nil {
		return *m.HPIncrease
	}
	return Default_NEWDESTINYGENERALLEVEL_HPIncrease
}

type NEWDESTINYGENERALLEVEL_ARRAY struct {
	Items            []*NEWDESTINYGENERALLEVEL `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte                    `json:"-"`
}

func (m *NEWDESTINYGENERALLEVEL_ARRAY) Reset()         { *m = NEWDESTINYGENERALLEVEL_ARRAY{} }
func (m *NEWDESTINYGENERALLEVEL_ARRAY) String() string { return proto.CompactTextString(m) }
func (*NEWDESTINYGENERALLEVEL_ARRAY) ProtoMessage()    {}

func (m *NEWDESTINYGENERALLEVEL_ARRAY) GetItems() []*NEWDESTINYGENERALLEVEL {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}