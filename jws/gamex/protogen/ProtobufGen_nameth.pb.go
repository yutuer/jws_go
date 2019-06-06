// Code generated by protoc-gen-go.
// source: ProtobufGen_nameth.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type NAMETH struct {
	// * 索引，第一列不能为汉字
	Index *uint32 `protobuf:"varint,1,req,def=0" json:"Index,omitempty"`
	// * 家鄉
	HomeTown *string `protobuf:"bytes,2,opt,def=" json:"HomeTown,omitempty"`
	// * 姓
	FamilyName *string `protobuf:"bytes,3,opt,def=" json:"FamilyName,omitempty"`
	// * 名
	FirstName        *string `protobuf:"bytes,4,opt,def=" json:"FirstName,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NAMETH) Reset()         { *m = NAMETH{} }
func (m *NAMETH) String() string { return proto.CompactTextString(m) }
func (*NAMETH) ProtoMessage()    {}

const Default_NAMETH_Index uint32 = 0

func (m *NAMETH) GetIndex() uint32 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return Default_NAMETH_Index
}

func (m *NAMETH) GetHomeTown() string {
	if m != nil && m.HomeTown != nil {
		return *m.HomeTown
	}
	return ""
}

func (m *NAMETH) GetFamilyName() string {
	if m != nil && m.FamilyName != nil {
		return *m.FamilyName
	}
	return ""
}

func (m *NAMETH) GetFirstName() string {
	if m != nil && m.FirstName != nil {
		return *m.FirstName
	}
	return ""
}

type NAMETH_ARRAY struct {
	Items            []*NAMETH `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *NAMETH_ARRAY) Reset()         { *m = NAMETH_ARRAY{} }
func (m *NAMETH_ARRAY) String() string { return proto.CompactTextString(m) }
func (*NAMETH_ARRAY) ProtoMessage()    {}

func (m *NAMETH_ARRAY) GetItems() []*NAMETH {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
