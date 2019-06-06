// Code generated by protoc-gen-go.
// source: ProtobufGen_refreshprice.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type REFRESHPRICE struct {
	// * 商店类型
	StoreID *uint32 `protobuf:"varint,1,req,def=0" json:"StoreID,omitempty"`
	// * 刷新次数
	RefreshNum *uint32 `protobuf:"varint,2,req,def=0" json:"RefreshNum,omitempty"`
	// * 刷新货币
	RefreshCoin *string `protobuf:"bytes,3,req,def=" json:"RefreshCoin,omitempty"`
	// * 刷新价格
	RefreshPrice     *uint32 `protobuf:"varint,4,req,def=0" json:"RefreshPrice,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *REFRESHPRICE) Reset()         { *m = REFRESHPRICE{} }
func (m *REFRESHPRICE) String() string { return proto.CompactTextString(m) }
func (*REFRESHPRICE) ProtoMessage()    {}

const Default_REFRESHPRICE_StoreID uint32 = 0
const Default_REFRESHPRICE_RefreshNum uint32 = 0
const Default_REFRESHPRICE_RefreshPrice uint32 = 0

func (m *REFRESHPRICE) GetStoreID() uint32 {
	if m != nil && m.StoreID != nil {
		return *m.StoreID
	}
	return Default_REFRESHPRICE_StoreID
}

func (m *REFRESHPRICE) GetRefreshNum() uint32 {
	if m != nil && m.RefreshNum != nil {
		return *m.RefreshNum
	}
	return Default_REFRESHPRICE_RefreshNum
}

func (m *REFRESHPRICE) GetRefreshCoin() string {
	if m != nil && m.RefreshCoin != nil {
		return *m.RefreshCoin
	}
	return ""
}

func (m *REFRESHPRICE) GetRefreshPrice() uint32 {
	if m != nil && m.RefreshPrice != nil {
		return *m.RefreshPrice
	}
	return Default_REFRESHPRICE_RefreshPrice
}

type REFRESHPRICE_ARRAY struct {
	Items            []*REFRESHPRICE `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *REFRESHPRICE_ARRAY) Reset()         { *m = REFRESHPRICE_ARRAY{} }
func (m *REFRESHPRICE_ARRAY) String() string { return proto.CompactTextString(m) }
func (*REFRESHPRICE_ARRAY) ProtoMessage()    {}

func (m *REFRESHPRICE_ARRAY) GetItems() []*REFRESHPRICE {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
