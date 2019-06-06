// Code generated by protoc-gen-go.
// source: ProtobufGen_generalstart.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GENERALSTART struct {
	// * 副将ID
	GeneralID *string `protobuf:"bytes,1,req,def=" json:"GeneralID,omitempty"`
	// * 星级
	StarLevel *uint32 `protobuf:"varint,2,opt,def=0" json:"StarLevel,omitempty"`
	// * 升至本星级所需碎片个数
	PieceNum *uint32 `protobuf:"varint,3,opt,def=0" json:"PieceNum,omitempty"`
	// * 升至本星级所需SC数量
	PieceSC                  *uint32                         `protobuf:"varint,4,opt,def=0" json:"PieceSC,omitempty"`
	GeneralProperty_Template []*GENERALSTART_GeneralProperty `protobuf:"bytes,5,rep" json:"GeneralProperty_Template,omitempty"`
	XXX_unrecognized         []byte                          `json:"-"`
}

func (m *GENERALSTART) Reset()         { *m = GENERALSTART{} }
func (m *GENERALSTART) String() string { return proto.CompactTextString(m) }
func (*GENERALSTART) ProtoMessage()    {}

const Default_GENERALSTART_StarLevel uint32 = 0
const Default_GENERALSTART_PieceNum uint32 = 0
const Default_GENERALSTART_PieceSC uint32 = 0

func (m *GENERALSTART) GetGeneralID() string {
	if m != nil && m.GeneralID != nil {
		return *m.GeneralID
	}
	return ""
}

func (m *GENERALSTART) GetStarLevel() uint32 {
	if m != nil && m.StarLevel != nil {
		return *m.StarLevel
	}
	return Default_GENERALSTART_StarLevel
}

func (m *GENERALSTART) GetPieceNum() uint32 {
	if m != nil && m.PieceNum != nil {
		return *m.PieceNum
	}
	return Default_GENERALSTART_PieceNum
}

func (m *GENERALSTART) GetPieceSC() uint32 {
	if m != nil && m.PieceSC != nil {
		return *m.PieceSC
	}
	return Default_GENERALSTART_PieceSC
}

func (m *GENERALSTART) GetGeneralProperty_Template() []*GENERALSTART_GeneralProperty {
	if m != nil {
		return m.GeneralProperty_Template
	}
	return nil
}

type GENERALSTART_GeneralProperty struct {
	// * 属性ID
	Property *string `protobuf:"bytes,1,opt,def=" json:"Property,omitempty"`
	// * 属性值
	Value            *float32 `protobuf:"fixed32,2,opt,def=0" json:"Value,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *GENERALSTART_GeneralProperty) Reset()         { *m = GENERALSTART_GeneralProperty{} }
func (m *GENERALSTART_GeneralProperty) String() string { return proto.CompactTextString(m) }
func (*GENERALSTART_GeneralProperty) ProtoMessage()    {}

const Default_GENERALSTART_GeneralProperty_Value float32 = 0

func (m *GENERALSTART_GeneralProperty) GetProperty() string {
	if m != nil && m.Property != nil {
		return *m.Property
	}
	return ""
}

func (m *GENERALSTART_GeneralProperty) GetValue() float32 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return Default_GENERALSTART_GeneralProperty_Value
}

type GENERALSTART_ARRAY struct {
	Items            []*GENERALSTART `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *GENERALSTART_ARRAY) Reset()         { *m = GENERALSTART_ARRAY{} }
func (m *GENERALSTART_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GENERALSTART_ARRAY) ProtoMessage()    {}

func (m *GENERALSTART_ARRAY) GetItems() []*GENERALSTART {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
