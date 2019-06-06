// Code generated by protoc-gen-go.
// source: ProtobufGen_guildworshipcrit.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GUILDWORSHIPCRIT struct {
	// * 祈福的次数
	WorshipDrawID *uint32 `protobuf:"varint,1,req,def=0" json:"WorshipDrawID,omitempty"`
	// * 该次祈福的暴击率%
	WorshipDrawCrit  *uint32 `protobuf:"varint,2,req,def=0" json:"WorshipDrawCrit,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GUILDWORSHIPCRIT) Reset()         { *m = GUILDWORSHIPCRIT{} }
func (m *GUILDWORSHIPCRIT) String() string { return proto.CompactTextString(m) }
func (*GUILDWORSHIPCRIT) ProtoMessage()    {}

const Default_GUILDWORSHIPCRIT_WorshipDrawID uint32 = 0
const Default_GUILDWORSHIPCRIT_WorshipDrawCrit uint32 = 0

func (m *GUILDWORSHIPCRIT) GetWorshipDrawID() uint32 {
	if m != nil && m.WorshipDrawID != nil {
		return *m.WorshipDrawID
	}
	return Default_GUILDWORSHIPCRIT_WorshipDrawID
}

func (m *GUILDWORSHIPCRIT) GetWorshipDrawCrit() uint32 {
	if m != nil && m.WorshipDrawCrit != nil {
		return *m.WorshipDrawCrit
	}
	return Default_GUILDWORSHIPCRIT_WorshipDrawCrit
}

type GUILDWORSHIPCRIT_ARRAY struct {
	Items            []*GUILDWORSHIPCRIT `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

func (m *GUILDWORSHIPCRIT_ARRAY) Reset()         { *m = GUILDWORSHIPCRIT_ARRAY{} }
func (m *GUILDWORSHIPCRIT_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GUILDWORSHIPCRIT_ARRAY) ProtoMessage()    {}

func (m *GUILDWORSHIPCRIT_ARRAY) GetItems() []*GUILDWORSHIPCRIT {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
