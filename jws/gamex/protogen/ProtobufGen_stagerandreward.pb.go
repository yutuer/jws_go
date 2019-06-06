// Code generated by protoc-gen-go.
// source: ProtobufGen_stagerandreward.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type STAGERANDREWARD struct {
	// * 关卡ID
	StageID           *string                        `protobuf:"bytes,1,req,def=" json:"StageID,omitempty"`
	SRewardRand_Table []*STAGERANDREWARD_SRewardRand `protobuf:"bytes,2,rep" json:"SRewardRand_Table,omitempty"`
	XXX_unrecognized  []byte                         `json:"-"`
}

func (m *STAGERANDREWARD) Reset()         { *m = STAGERANDREWARD{} }
func (m *STAGERANDREWARD) String() string { return proto.CompactTextString(m) }
func (*STAGERANDREWARD) ProtoMessage()    {}

func (m *STAGERANDREWARD) GetStageID() string {
	if m != nil && m.StageID != nil {
		return *m.StageID
	}
	return ""
}

func (m *STAGERANDREWARD) GetSRewardRand_Table() []*STAGERANDREWARD_SRewardRand {
	if m != nil {
		return m.SRewardRand_Table
	}
	return nil
}

type STAGERANDREWARD_SRewardRand struct {
	// * 物品组ID
	ItemGroupID *string `protobuf:"bytes,1,opt,def=" json:"ItemGroupID,omitempty"`
	// * 物品组掉落概率
	SRandRate        *uint32 `protobuf:"varint,2,opt,def=0" json:"SRandRate,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *STAGERANDREWARD_SRewardRand) Reset()         { *m = STAGERANDREWARD_SRewardRand{} }
func (m *STAGERANDREWARD_SRewardRand) String() string { return proto.CompactTextString(m) }
func (*STAGERANDREWARD_SRewardRand) ProtoMessage()    {}

const Default_STAGERANDREWARD_SRewardRand_SRandRate uint32 = 0

func (m *STAGERANDREWARD_SRewardRand) GetItemGroupID() string {
	if m != nil && m.ItemGroupID != nil {
		return *m.ItemGroupID
	}
	return ""
}

func (m *STAGERANDREWARD_SRewardRand) GetSRandRate() uint32 {
	if m != nil && m.SRandRate != nil {
		return *m.SRandRate
	}
	return Default_STAGERANDREWARD_SRewardRand_SRandRate
}

type STAGERANDREWARD_ARRAY struct {
	Items            []*STAGERANDREWARD `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *STAGERANDREWARD_ARRAY) Reset()         { *m = STAGERANDREWARD_ARRAY{} }
func (m *STAGERANDREWARD_ARRAY) String() string { return proto.CompactTextString(m) }
func (*STAGERANDREWARD_ARRAY) ProtoMessage()    {}

func (m *STAGERANDREWARD_ARRAY) GetItems() []*STAGERANDREWARD {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
