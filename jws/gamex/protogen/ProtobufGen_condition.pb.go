// Code generated by protoc-gen-go.
// source: ProtobufGen_condition.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type CONDITION struct {
	// * 功能开放的顺序
	ConditionOrder *uint32 `protobuf:"varint,1,req,def=0" json:"ConditionOrder,omitempty"`
	// * 功能名称的枚举
	ConditionID *uint32 `protobuf:"varint,2,req,def=0" json:"ConditionID,omitempty"`
	// * 开放的方式
	ConditionType *uint32 `protobuf:"varint,3,opt,def=0" json:"ConditionType,omitempty"`
	// * 开放的条件
	ConditionValue *string `protobuf:"bytes,4,opt,def=" json:"ConditionValue,omitempty"`
	// * 是否显示提示，0=不显示，1=显示
	ConditionDisplay *uint32 `protobuf:"varint,5,opt,def=0" json:"ConditionDisplay,omitempty"`
	// * 解锁功能名称描述
	SystemName *string `protobuf:"bytes,6,opt,def=" json:"SystemName,omitempty"`
	// * 解锁功能的内容描述
	SystemDescribe *string `protobuf:"bytes,7,opt,def=" json:"SystemDescribe,omitempty"`
	// * 功能解锁条件的描述
	SystemCondition *string `protobuf:"bytes,8,opt,def=" json:"SystemCondition,omitempty"`
	// * 解锁功能的图标
	SystemIcon *string `protobuf:"bytes,9,opt,def=" json:"SystemIcon,omitempty"`
	// * 解锁功能的图标类型
	IconType *uint32 `protobuf:"varint,11,opt,def=0" json:"IconType,omitempty"`
	// * 解锁的关卡
	ConditionPass    *string `protobuf:"bytes,10,opt,def=" json:"ConditionPass,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CONDITION) Reset()         { *m = CONDITION{} }
func (m *CONDITION) String() string { return proto.CompactTextString(m) }
func (*CONDITION) ProtoMessage()    {}

const Default_CONDITION_ConditionOrder uint32 = 0
const Default_CONDITION_ConditionID uint32 = 0
const Default_CONDITION_ConditionType uint32 = 0
const Default_CONDITION_ConditionDisplay uint32 = 0
const Default_CONDITION_IconType uint32 = 0

func (m *CONDITION) GetConditionOrder() uint32 {
	if m != nil && m.ConditionOrder != nil {
		return *m.ConditionOrder
	}
	return Default_CONDITION_ConditionOrder
}

func (m *CONDITION) GetConditionID() uint32 {
	if m != nil && m.ConditionID != nil {
		return *m.ConditionID
	}
	return Default_CONDITION_ConditionID
}

func (m *CONDITION) GetConditionType() uint32 {
	if m != nil && m.ConditionType != nil {
		return *m.ConditionType
	}
	return Default_CONDITION_ConditionType
}

func (m *CONDITION) GetConditionValue() string {
	if m != nil && m.ConditionValue != nil {
		return *m.ConditionValue
	}
	return ""
}

func (m *CONDITION) GetConditionDisplay() uint32 {
	if m != nil && m.ConditionDisplay != nil {
		return *m.ConditionDisplay
	}
	return Default_CONDITION_ConditionDisplay
}

func (m *CONDITION) GetSystemName() string {
	if m != nil && m.SystemName != nil {
		return *m.SystemName
	}
	return ""
}

func (m *CONDITION) GetSystemDescribe() string {
	if m != nil && m.SystemDescribe != nil {
		return *m.SystemDescribe
	}
	return ""
}

func (m *CONDITION) GetSystemCondition() string {
	if m != nil && m.SystemCondition != nil {
		return *m.SystemCondition
	}
	return ""
}

func (m *CONDITION) GetSystemIcon() string {
	if m != nil && m.SystemIcon != nil {
		return *m.SystemIcon
	}
	return ""
}

func (m *CONDITION) GetIconType() uint32 {
	if m != nil && m.IconType != nil {
		return *m.IconType
	}
	return Default_CONDITION_IconType
}

func (m *CONDITION) GetConditionPass() string {
	if m != nil && m.ConditionPass != nil {
		return *m.ConditionPass
	}
	return ""
}

type CONDITION_ARRAY struct {
	Items            []*CONDITION `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *CONDITION_ARRAY) Reset()         { *m = CONDITION_ARRAY{} }
func (m *CONDITION_ARRAY) String() string { return proto.CompactTextString(m) }
func (*CONDITION_ARRAY) ProtoMessage()    {}

func (m *CONDITION_ARRAY) GetItems() []*CONDITION {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
