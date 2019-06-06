// Code generated by protoc-gen-go.
// source: ProtobufGen_battlehero.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type BATTLEHERO struct {
	// * 战役ID，1=巴蜀之战 2=官渡之战 3=赤壁之战 4=讨伐董卓
	BattleID *uint32 `protobuf:"varint,1,req,def=0" json:"BattleID,omitempty"`
	// * 所上主将国籍ID 1=蜀 2=魏 3=吴 4=群
	HeroID           *uint32 `protobuf:"varint,2,opt,def=0" json:"HeroID,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *BATTLEHERO) Reset()         { *m = BATTLEHERO{} }
func (m *BATTLEHERO) String() string { return proto.CompactTextString(m) }
func (*BATTLEHERO) ProtoMessage()    {}

const Default_BATTLEHERO_BattleID uint32 = 0
const Default_BATTLEHERO_HeroID uint32 = 0

func (m *BATTLEHERO) GetBattleID() uint32 {
	if m != nil && m.BattleID != nil {
		return *m.BattleID
	}
	return Default_BATTLEHERO_BattleID
}

func (m *BATTLEHERO) GetHeroID() uint32 {
	if m != nil && m.HeroID != nil {
		return *m.HeroID
	}
	return Default_BATTLEHERO_HeroID
}

type BATTLEHERO_ARRAY struct {
	Items            []*BATTLEHERO `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *BATTLEHERO_ARRAY) Reset()         { *m = BATTLEHERO_ARRAY{} }
func (m *BATTLEHERO_ARRAY) String() string { return proto.CompactTextString(m) }
func (*BATTLEHERO_ARRAY) ProtoMessage()    {}

func (m *BATTLEHERO_ARRAY) GetItems() []*BATTLEHERO {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
