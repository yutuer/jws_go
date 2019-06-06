// Code generated by protoc-gen-go.
// source: ProtobufGen_giftactivitylist.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GIFTACTIVITYLIST struct {
	// * 活动ID
	ActivityID *uint32 `protobuf:"varint,1,req,def=0" json:"ActivityID,omitempty"`
	// * 活动时间类型
	TimeType *uint32 `protobuf:"varint,2,opt,def=0" json:"TimeType,omitempty"`
	// * 领取奖励类型
	GiftAcceptType *uint32 `protobuf:"varint,3,opt,def=0" json:"GiftAcceptType,omitempty"`
	// * 开始时间
	StartTime *string `protobuf:"bytes,4,opt,def=" json:"StartTime,omitempty"`
	// * 结束时间
	EndTime *string `protobuf:"bytes,5,opt,def=" json:"EndTime,omitempty"`
	// * 活动标题
	ActivityTitle    *string `protobuf:"bytes,6,opt,def=" json:"ActivityTitle,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GIFTACTIVITYLIST) Reset()         { *m = GIFTACTIVITYLIST{} }
func (m *GIFTACTIVITYLIST) String() string { return proto.CompactTextString(m) }
func (*GIFTACTIVITYLIST) ProtoMessage()    {}

const Default_GIFTACTIVITYLIST_ActivityID uint32 = 0
const Default_GIFTACTIVITYLIST_TimeType uint32 = 0
const Default_GIFTACTIVITYLIST_GiftAcceptType uint32 = 0

func (m *GIFTACTIVITYLIST) GetActivityID() uint32 {
	if m != nil && m.ActivityID != nil {
		return *m.ActivityID
	}
	return Default_GIFTACTIVITYLIST_ActivityID
}

func (m *GIFTACTIVITYLIST) GetTimeType() uint32 {
	if m != nil && m.TimeType != nil {
		return *m.TimeType
	}
	return Default_GIFTACTIVITYLIST_TimeType
}

func (m *GIFTACTIVITYLIST) GetGiftAcceptType() uint32 {
	if m != nil && m.GiftAcceptType != nil {
		return *m.GiftAcceptType
	}
	return Default_GIFTACTIVITYLIST_GiftAcceptType
}

func (m *GIFTACTIVITYLIST) GetStartTime() string {
	if m != nil && m.StartTime != nil {
		return *m.StartTime
	}
	return ""
}

func (m *GIFTACTIVITYLIST) GetEndTime() string {
	if m != nil && m.EndTime != nil {
		return *m.EndTime
	}
	return ""
}

func (m *GIFTACTIVITYLIST) GetActivityTitle() string {
	if m != nil && m.ActivityTitle != nil {
		return *m.ActivityTitle
	}
	return ""
}

type GIFTACTIVITYLIST_ARRAY struct {
	Items            []*GIFTACTIVITYLIST `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

func (m *GIFTACTIVITYLIST_ARRAY) Reset()         { *m = GIFTACTIVITYLIST_ARRAY{} }
func (m *GIFTACTIVITYLIST_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GIFTACTIVITYLIST_ARRAY) ProtoMessage()    {}

func (m *GIFTACTIVITYLIST_ARRAY) GetItems() []*GIFTACTIVITYLIST {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
