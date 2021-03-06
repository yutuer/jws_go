// Code generated by protoc-gen-go.
// source: ProtobufGen_level_condition.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type LEVEL_CONDITION struct {
	// * 关卡ID
	LevelID          *string                      `protobuf:"bytes,1,req,name=levelID,def=" json:"levelID,omitempty"`
	ConditionData    []*LEVEL_CONDITION_Condition `protobuf:"bytes,2,rep" json:"ConditionData,omitempty"`
	XXX_unrecognized []byte                       `json:"-"`
}

func (m *LEVEL_CONDITION) Reset()         { *m = LEVEL_CONDITION{} }
func (m *LEVEL_CONDITION) String() string { return proto.CompactTextString(m) }
func (*LEVEL_CONDITION) ProtoMessage()    {}

func (m *LEVEL_CONDITION) GetLevelID() string {
	if m != nil && m.LevelID != nil {
		return *m.LevelID
	}
	return ""
}

func (m *LEVEL_CONDITION) GetConditionData() []*LEVEL_CONDITION_Condition {
	if m != nil {
		return m.ConditionData
	}
	return nil
}

type LEVEL_CONDITION_Condition struct {
	// * 条件类型
	ConditionType *int32 `protobuf:"varint,1,opt,name=conditionType,def=0" json:"conditionType,omitempty"`
	// * 条件参数1
	ConditionValue1 *string `protobuf:"bytes,2,opt,name=conditionValue1,def=" json:"conditionValue1,omitempty"`
	// * 条件参数2
	ConditionValue2 *float32 `protobuf:"fixed32,3,opt,name=conditionValue2,def=0" json:"conditionValue2,omitempty"`
	// * 条件说明
	ConditionIDS     *string `protobuf:"bytes,4,opt,name=conditionIDS,def=" json:"conditionIDS,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LEVEL_CONDITION_Condition) Reset()         { *m = LEVEL_CONDITION_Condition{} }
func (m *LEVEL_CONDITION_Condition) String() string { return proto.CompactTextString(m) }
func (*LEVEL_CONDITION_Condition) ProtoMessage()    {}

const Default_LEVEL_CONDITION_Condition_ConditionType int32 = 0
const Default_LEVEL_CONDITION_Condition_ConditionValue2 float32 = 0

func (m *LEVEL_CONDITION_Condition) GetConditionType() int32 {
	if m != nil && m.ConditionType != nil {
		return *m.ConditionType
	}
	return Default_LEVEL_CONDITION_Condition_ConditionType
}

func (m *LEVEL_CONDITION_Condition) GetConditionValue1() string {
	if m != nil && m.ConditionValue1 != nil {
		return *m.ConditionValue1
	}
	return ""
}

func (m *LEVEL_CONDITION_Condition) GetConditionValue2() float32 {
	if m != nil && m.ConditionValue2 != nil {
		return *m.ConditionValue2
	}
	return Default_LEVEL_CONDITION_Condition_ConditionValue2
}

func (m *LEVEL_CONDITION_Condition) GetConditionIDS() string {
	if m != nil && m.ConditionIDS != nil {
		return *m.ConditionIDS
	}
	return ""
}

type LEVEL_CONDITION_ARRAY struct {
	Items            []*LEVEL_CONDITION `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *LEVEL_CONDITION_ARRAY) Reset()         { *m = LEVEL_CONDITION_ARRAY{} }
func (m *LEVEL_CONDITION_ARRAY) String() string { return proto.CompactTextString(m) }
func (*LEVEL_CONDITION_ARRAY) ProtoMessage()    {}

func (m *LEVEL_CONDITION_ARRAY) GetItems() []*LEVEL_CONDITION {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
