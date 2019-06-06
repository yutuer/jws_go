// Code generated by protoc-gen-go.
// source: ProtobufGen_initializeitem.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type INITIALIZEITEM struct {
	// * 物品ID
	ItemID *string `protobuf:"bytes,1,opt,def=" json:"ItemID,omitempty"`
	// * 物品数量
	ItemCount        *uint32 `protobuf:"varint,2,opt,def=1" json:"ItemCount,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *INITIALIZEITEM) Reset()         { *m = INITIALIZEITEM{} }
func (m *INITIALIZEITEM) String() string { return proto.CompactTextString(m) }
func (*INITIALIZEITEM) ProtoMessage()    {}

const Default_INITIALIZEITEM_ItemCount uint32 = 1

func (m *INITIALIZEITEM) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *INITIALIZEITEM) GetItemCount() uint32 {
	if m != nil && m.ItemCount != nil {
		return *m.ItemCount
	}
	return Default_INITIALIZEITEM_ItemCount
}

type INITIALIZEITEM_ARRAY struct {
	Items            []*INITIALIZEITEM `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *INITIALIZEITEM_ARRAY) Reset()         { *m = INITIALIZEITEM_ARRAY{} }
func (m *INITIALIZEITEM_ARRAY) String() string { return proto.CompactTextString(m) }
func (*INITIALIZEITEM_ARRAY) ProtoMessage()    {}

func (m *INITIALIZEITEM_ARRAY) GetItems() []*INITIALIZEITEM {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
