// Code generated by protoc-gen-go.
// source: ProtobufGen_petposition.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type PETPOSITION struct {
	// * 主将ID
	HeroID *uint32 `protobuf:"varint,1,req,def=0" json:"HeroID,omitempty"`
	// * 宠物位置X
	HeroPositionX *float32 `protobuf:"fixed32,2,opt,def=0" json:"HeroPositionX,omitempty"`
	// * 宠物位置Y
	HeroPositionY    *float32 `protobuf:"fixed32,3,opt,def=0" json:"HeroPositionY,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *PETPOSITION) Reset()         { *m = PETPOSITION{} }
func (m *PETPOSITION) String() string { return proto.CompactTextString(m) }
func (*PETPOSITION) ProtoMessage()    {}

const Default_PETPOSITION_HeroID uint32 = 0
const Default_PETPOSITION_HeroPositionX float32 = 0
const Default_PETPOSITION_HeroPositionY float32 = 0

func (m *PETPOSITION) GetHeroID() uint32 {
	if m != nil && m.HeroID != nil {
		return *m.HeroID
	}
	return Default_PETPOSITION_HeroID
}

func (m *PETPOSITION) GetHeroPositionX() float32 {
	if m != nil && m.HeroPositionX != nil {
		return *m.HeroPositionX
	}
	return Default_PETPOSITION_HeroPositionX
}

func (m *PETPOSITION) GetHeroPositionY() float32 {
	if m != nil && m.HeroPositionY != nil {
		return *m.HeroPositionY
	}
	return Default_PETPOSITION_HeroPositionY
}

type PETPOSITION_ARRAY struct {
	Items            []*PETPOSITION `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *PETPOSITION_ARRAY) Reset()         { *m = PETPOSITION_ARRAY{} }
func (m *PETPOSITION_ARRAY) String() string { return proto.CompactTextString(m) }
func (*PETPOSITION_ARRAY) ProtoMessage()    {}

func (m *PETPOSITION_ARRAY) GetItems() []*PETPOSITION {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}