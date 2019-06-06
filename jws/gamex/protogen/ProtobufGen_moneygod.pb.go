// Code generated by protoc-gen-go.
// source: ProtobufGen_moneygod.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type MONEYGOD struct {
	// * 招财次数
	GodLevel *uint32 `protobuf:"varint,1,req,def=0" json:"GodLevel,omitempty"`
	// * 所需HC数额
	CostHC *uint32 `protobuf:"varint,2,opt,def=0" json:"CostHC,omitempty"`
	// * 所需VIP等级
	VIPlevel *uint32 `protobuf:"varint,3,opt,def=0" json:"VIPlevel,omitempty"`
	// * 获得物品ID
	ItemID *string `protobuf:"bytes,4,opt,def=" json:"ItemID,omitempty"`
	// * 获得数量下限
	MinNum *int32 `protobuf:"varint,5,opt,def=0" json:"MinNum,omitempty"`
	// * 获得刷量上限
	MaxNum *int32 `protobuf:"varint,6,opt,def=1" json:"MaxNum,omitempty"`
	// * 是否开启跑马灯
	OpenAdv          *uint32          `protobuf:"varint,7,opt,def=0" json:"OpenAdv,omitempty"`
	Fixed_Num        []*MONEYGOD_Num1 `protobuf:"bytes,8,rep" json:"Fixed_Num,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *MONEYGOD) Reset()         { *m = MONEYGOD{} }
func (m *MONEYGOD) String() string { return proto.CompactTextString(m) }
func (*MONEYGOD) ProtoMessage()    {}

const Default_MONEYGOD_GodLevel uint32 = 0
const Default_MONEYGOD_CostHC uint32 = 0
const Default_MONEYGOD_VIPlevel uint32 = 0
const Default_MONEYGOD_MinNum int32 = 0
const Default_MONEYGOD_MaxNum int32 = 1
const Default_MONEYGOD_OpenAdv uint32 = 0

func (m *MONEYGOD) GetGodLevel() uint32 {
	if m != nil && m.GodLevel != nil {
		return *m.GodLevel
	}
	return Default_MONEYGOD_GodLevel
}

func (m *MONEYGOD) GetCostHC() uint32 {
	if m != nil && m.CostHC != nil {
		return *m.CostHC
	}
	return Default_MONEYGOD_CostHC
}

func (m *MONEYGOD) GetVIPlevel() uint32 {
	if m != nil && m.VIPlevel != nil {
		return *m.VIPlevel
	}
	return Default_MONEYGOD_VIPlevel
}

func (m *MONEYGOD) GetItemID() string {
	if m != nil && m.ItemID != nil {
		return *m.ItemID
	}
	return ""
}

func (m *MONEYGOD) GetMinNum() int32 {
	if m != nil && m.MinNum != nil {
		return *m.MinNum
	}
	return Default_MONEYGOD_MinNum
}

func (m *MONEYGOD) GetMaxNum() int32 {
	if m != nil && m.MaxNum != nil {
		return *m.MaxNum
	}
	return Default_MONEYGOD_MaxNum
}

func (m *MONEYGOD) GetOpenAdv() uint32 {
	if m != nil && m.OpenAdv != nil {
		return *m.OpenAdv
	}
	return Default_MONEYGOD_OpenAdv
}

func (m *MONEYGOD) GetFixed_Num() []*MONEYGOD_Num1 {
	if m != nil {
		return m.Fixed_Num
	}
	return nil
}

type MONEYGOD_Num1 struct {
	// * 获得数量下限
	SMinNum *int32 `protobuf:"varint,4,opt,def=0" json:"SMinNum,omitempty"`
	// * 获得刷量上限
	SMaxNum *int32 `protobuf:"varint,5,opt,def=1" json:"SMaxNum,omitempty"`
	// * 区间的权重
	Weight           *int32 `protobuf:"varint,3,opt,def=0" json:"Weight,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *MONEYGOD_Num1) Reset()         { *m = MONEYGOD_Num1{} }
func (m *MONEYGOD_Num1) String() string { return proto.CompactTextString(m) }
func (*MONEYGOD_Num1) ProtoMessage()    {}

const Default_MONEYGOD_Num1_SMinNum int32 = 0
const Default_MONEYGOD_Num1_SMaxNum int32 = 1
const Default_MONEYGOD_Num1_Weight int32 = 0

func (m *MONEYGOD_Num1) GetSMinNum() int32 {
	if m != nil && m.SMinNum != nil {
		return *m.SMinNum
	}
	return Default_MONEYGOD_Num1_SMinNum
}

func (m *MONEYGOD_Num1) GetSMaxNum() int32 {
	if m != nil && m.SMaxNum != nil {
		return *m.SMaxNum
	}
	return Default_MONEYGOD_Num1_SMaxNum
}

func (m *MONEYGOD_Num1) GetWeight() int32 {
	if m != nil && m.Weight != nil {
		return *m.Weight
	}
	return Default_MONEYGOD_Num1_Weight
}

type MONEYGOD_ARRAY struct {
	Items            []*MONEYGOD `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *MONEYGOD_ARRAY) Reset()         { *m = MONEYGOD_ARRAY{} }
func (m *MONEYGOD_ARRAY) String() string { return proto.CompactTextString(m) }
func (*MONEYGOD_ARRAY) ProtoMessage()    {}

func (m *MONEYGOD_ARRAY) GetItems() []*MONEYGOD {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
