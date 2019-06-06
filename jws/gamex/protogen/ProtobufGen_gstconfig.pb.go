// Code generated by protoc-gen-go.
// source: ProtobufGen_gstconfig.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GSTCONFIG struct {
	// * 统一捐献点
	StandardSP *uint32 `protobuf:"varint,1,req,def=0" json:"StandardSP,omitempty"`
	// * log每日重置时间
	LogDailyResetTime *string `protobuf:"bytes,5,req,def=" json:"LogDailyResetTime,omitempty"`
	// * log每周重置时间
	LogWeeklyResetTime *string `protobuf:"bytes,6,req,def=" json:"LogWeeklyResetTime,omitempty"`
	// * log每周重置日
	LogWeeklyResetDay *uint32 `protobuf:"varint,7,req,def=0" json:"LogWeeklyResetDay,omitempty"`
	XXX_unrecognized  []byte  `json:"-"`
}

func (m *GSTCONFIG) Reset()         { *m = GSTCONFIG{} }
func (m *GSTCONFIG) String() string { return proto.CompactTextString(m) }
func (*GSTCONFIG) ProtoMessage()    {}

const Default_GSTCONFIG_StandardSP uint32 = 0
const Default_GSTCONFIG_LogWeeklyResetDay uint32 = 0

func (m *GSTCONFIG) GetStandardSP() uint32 {
	if m != nil && m.StandardSP != nil {
		return *m.StandardSP
	}
	return Default_GSTCONFIG_StandardSP
}

func (m *GSTCONFIG) GetLogDailyResetTime() string {
	if m != nil && m.LogDailyResetTime != nil {
		return *m.LogDailyResetTime
	}
	return ""
}

func (m *GSTCONFIG) GetLogWeeklyResetTime() string {
	if m != nil && m.LogWeeklyResetTime != nil {
		return *m.LogWeeklyResetTime
	}
	return ""
}

func (m *GSTCONFIG) GetLogWeeklyResetDay() uint32 {
	if m != nil && m.LogWeeklyResetDay != nil {
		return *m.LogWeeklyResetDay
	}
	return Default_GSTCONFIG_LogWeeklyResetDay
}

type GSTCONFIG_ARRAY struct {
	Items            []*GSTCONFIG `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *GSTCONFIG_ARRAY) Reset()         { *m = GSTCONFIG_ARRAY{} }
func (m *GSTCONFIG_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GSTCONFIG_ARRAY) ProtoMessage()    {}

func (m *GSTCONFIG_ARRAY) GetItems() []*GSTCONFIG {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}