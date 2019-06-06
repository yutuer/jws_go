// Code generated by protoc-gen-go.
// source: ProtobufGen_bscpvpswtcost.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type BSCPVPSWTCOST struct {
	// * 第几次切人
	SwitchNum *uint32 `protobuf:"varint,1,req,def=0" json:"SwitchNum,omitempty"`
	// * 货币类型
	CostType *string `protobuf:"bytes,2,opt,def=" json:"CostType,omitempty"`
	// * 货币数量
	CostValue        *uint32 `protobuf:"varint,3,opt,def=0" json:"CostValue,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *BSCPVPSWTCOST) Reset()         { *m = BSCPVPSWTCOST{} }
func (m *BSCPVPSWTCOST) String() string { return proto.CompactTextString(m) }
func (*BSCPVPSWTCOST) ProtoMessage()    {}

const Default_BSCPVPSWTCOST_SwitchNum uint32 = 0
const Default_BSCPVPSWTCOST_CostValue uint32 = 0

func (m *BSCPVPSWTCOST) GetSwitchNum() uint32 {
	if m != nil && m.SwitchNum != nil {
		return *m.SwitchNum
	}
	return Default_BSCPVPSWTCOST_SwitchNum
}

func (m *BSCPVPSWTCOST) GetCostType() string {
	if m != nil && m.CostType != nil {
		return *m.CostType
	}
	return ""
}

func (m *BSCPVPSWTCOST) GetCostValue() uint32 {
	if m != nil && m.CostValue != nil {
		return *m.CostValue
	}
	return Default_BSCPVPSWTCOST_CostValue
}

type BSCPVPSWTCOST_ARRAY struct {
	Items            []*BSCPVPSWTCOST `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *BSCPVPSWTCOST_ARRAY) Reset()         { *m = BSCPVPSWTCOST_ARRAY{} }
func (m *BSCPVPSWTCOST_ARRAY) String() string { return proto.CompactTextString(m) }
func (*BSCPVPSWTCOST_ARRAY) ProtoMessage()    {}

func (m *BSCPVPSWTCOST_ARRAY) GetItems() []*BSCPVPSWTCOST {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
