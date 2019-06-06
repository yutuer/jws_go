// Code generated by protoc-gen-go.
// source: ProtobufGen_gvgconfig.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GVGCONFIG struct {
	// * 胜利基础积分
	BasicGVGPoint *uint32 `protobuf:"varint,8,opt,def=0" json:"BasicGVGPoint,omitempty"`
	// * 机器人得分
	RobotGVGPoint *uint32 `protobuf:"varint,13,opt,def=0" json:"RobotGVGPoint,omitempty"`
	// * 失败得分
	FailGVGPoint *uint32 `protobuf:"varint,14,opt,def=0" json:"FailGVGPoint,omitempty"`
	// * 匹配时间上限（秒）
	MatchLimitTime *uint32 `protobuf:"varint,11,opt,def=0" json:"MatchLimitTime,omitempty"`
	// * 想玩要达到的军团等级
	NeedGuildLevel *uint32 `protobuf:"varint,16,opt,def=0" json:"NeedGuildLevel,omitempty"`
	// * 占领城池上限
	CityHoldMax *uint32 `protobuf:"varint,21,opt,def=0" json:"CityHoldMax,omitempty"`
	// * 黑影模型最低数量1
	ShadowNum1 *uint32 `protobuf:"varint,17,opt,def=0" json:"ShadowNum1,omitempty"`
	// * 黑影模型中等数量2
	ShadowNum2 *uint32 `protobuf:"varint,18,opt,def=0" json:"ShadowNum2,omitempty"`
	// * 黑影模型较多数量3
	ShadowNum3 *uint32 `protobuf:"varint,19,opt,def=0" json:"ShadowNum3,omitempty"`
	// * 黑影模型最高数量4
	ShadowNum4 *uint32 `protobuf:"varint,20,opt,def=0" json:"ShadowNum4,omitempty"`
	// * 排行刷新时间（秒）
	RankFreshTime *uint32 `protobuf:"varint,12,opt,def=0" json:"RankFreshTime,omitempty"`
	// * 匹配扩大搜索时间（秒）
	GSMateTime *uint32 `protobuf:"varint,9,opt,def=0" json:"GSMateTime,omitempty"`
	// * 战力百分比
	GearScoreRange *float32 `protobuf:"fixed32,10,opt,def=0" json:"GearScoreRange,omitempty"`
	// * 开服X小时内不开放（根据StartTime）
	SuspensionHour *uint32 `protobuf:"varint,15,opt,def=0" json:"SuspensionHour,omitempty"`
	// * 周几预告
	ReportWeek *uint32 `protobuf:"varint,1,opt,def=0" json:"ReportWeek,omitempty"`
	// * 预告时间
	AnnounceTime *string `protobuf:"bytes,2,opt,def=" json:"AnnounceTime,omitempty"`
	// * 周几活动重置
	RestartAndStartWeek *uint32 `protobuf:"varint,3,opt,def=0" json:"RestartAndStartWeek,omitempty"`
	// * 重置城池时间
	RestartCityTime *string `protobuf:"bytes,4,opt,def=" json:"RestartCityTime,omitempty"`
	// * 开启活动时间
	StartTime *string `protobuf:"bytes,5,opt,def=" json:"StartTime,omitempty"`
	// * 持续时间(分钟）
	GVGOpeningTime *uint32 `protobuf:"varint,6,opt,def=0" json:"GVGOpeningTime,omitempty"`
	// * 结算持续时间（分钟）
	GVGStatementTime *uint32 `protobuf:"varint,7,opt,def=0" json:"GVGStatementTime,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GVGCONFIG) Reset()         { *m = GVGCONFIG{} }
func (m *GVGCONFIG) String() string { return proto.CompactTextString(m) }
func (*GVGCONFIG) ProtoMessage()    {}

const Default_GVGCONFIG_BasicGVGPoint uint32 = 0
const Default_GVGCONFIG_RobotGVGPoint uint32 = 0
const Default_GVGCONFIG_FailGVGPoint uint32 = 0
const Default_GVGCONFIG_MatchLimitTime uint32 = 0
const Default_GVGCONFIG_NeedGuildLevel uint32 = 0
const Default_GVGCONFIG_CityHoldMax uint32 = 0
const Default_GVGCONFIG_ShadowNum1 uint32 = 0
const Default_GVGCONFIG_ShadowNum2 uint32 = 0
const Default_GVGCONFIG_ShadowNum3 uint32 = 0
const Default_GVGCONFIG_ShadowNum4 uint32 = 0
const Default_GVGCONFIG_RankFreshTime uint32 = 0
const Default_GVGCONFIG_GSMateTime uint32 = 0
const Default_GVGCONFIG_GearScoreRange float32 = 0
const Default_GVGCONFIG_SuspensionHour uint32 = 0
const Default_GVGCONFIG_ReportWeek uint32 = 0
const Default_GVGCONFIG_RestartAndStartWeek uint32 = 0
const Default_GVGCONFIG_GVGOpeningTime uint32 = 0
const Default_GVGCONFIG_GVGStatementTime uint32 = 0

func (m *GVGCONFIG) GetBasicGVGPoint() uint32 {
	if m != nil && m.BasicGVGPoint != nil {
		return *m.BasicGVGPoint
	}
	return Default_GVGCONFIG_BasicGVGPoint
}

func (m *GVGCONFIG) GetRobotGVGPoint() uint32 {
	if m != nil && m.RobotGVGPoint != nil {
		return *m.RobotGVGPoint
	}
	return Default_GVGCONFIG_RobotGVGPoint
}

func (m *GVGCONFIG) GetFailGVGPoint() uint32 {
	if m != nil && m.FailGVGPoint != nil {
		return *m.FailGVGPoint
	}
	return Default_GVGCONFIG_FailGVGPoint
}

func (m *GVGCONFIG) GetMatchLimitTime() uint32 {
	if m != nil && m.MatchLimitTime != nil {
		return *m.MatchLimitTime
	}
	return Default_GVGCONFIG_MatchLimitTime
}

func (m *GVGCONFIG) GetNeedGuildLevel() uint32 {
	if m != nil && m.NeedGuildLevel != nil {
		return *m.NeedGuildLevel
	}
	return Default_GVGCONFIG_NeedGuildLevel
}

func (m *GVGCONFIG) GetCityHoldMax() uint32 {
	if m != nil && m.CityHoldMax != nil {
		return *m.CityHoldMax
	}
	return Default_GVGCONFIG_CityHoldMax
}

func (m *GVGCONFIG) GetShadowNum1() uint32 {
	if m != nil && m.ShadowNum1 != nil {
		return *m.ShadowNum1
	}
	return Default_GVGCONFIG_ShadowNum1
}

func (m *GVGCONFIG) GetShadowNum2() uint32 {
	if m != nil && m.ShadowNum2 != nil {
		return *m.ShadowNum2
	}
	return Default_GVGCONFIG_ShadowNum2
}

func (m *GVGCONFIG) GetShadowNum3() uint32 {
	if m != nil && m.ShadowNum3 != nil {
		return *m.ShadowNum3
	}
	return Default_GVGCONFIG_ShadowNum3
}

func (m *GVGCONFIG) GetShadowNum4() uint32 {
	if m != nil && m.ShadowNum4 != nil {
		return *m.ShadowNum4
	}
	return Default_GVGCONFIG_ShadowNum4
}

func (m *GVGCONFIG) GetRankFreshTime() uint32 {
	if m != nil && m.RankFreshTime != nil {
		return *m.RankFreshTime
	}
	return Default_GVGCONFIG_RankFreshTime
}

func (m *GVGCONFIG) GetGSMateTime() uint32 {
	if m != nil && m.GSMateTime != nil {
		return *m.GSMateTime
	}
	return Default_GVGCONFIG_GSMateTime
}

func (m *GVGCONFIG) GetGearScoreRange() float32 {
	if m != nil && m.GearScoreRange != nil {
		return *m.GearScoreRange
	}
	return Default_GVGCONFIG_GearScoreRange
}

func (m *GVGCONFIG) GetSuspensionHour() uint32 {
	if m != nil && m.SuspensionHour != nil {
		return *m.SuspensionHour
	}
	return Default_GVGCONFIG_SuspensionHour
}

func (m *GVGCONFIG) GetReportWeek() uint32 {
	if m != nil && m.ReportWeek != nil {
		return *m.ReportWeek
	}
	return Default_GVGCONFIG_ReportWeek
}

func (m *GVGCONFIG) GetAnnounceTime() string {
	if m != nil && m.AnnounceTime != nil {
		return *m.AnnounceTime
	}
	return ""
}

func (m *GVGCONFIG) GetRestartAndStartWeek() uint32 {
	if m != nil && m.RestartAndStartWeek != nil {
		return *m.RestartAndStartWeek
	}
	return Default_GVGCONFIG_RestartAndStartWeek
}

func (m *GVGCONFIG) GetRestartCityTime() string {
	if m != nil && m.RestartCityTime != nil {
		return *m.RestartCityTime
	}
	return ""
}

func (m *GVGCONFIG) GetStartTime() string {
	if m != nil && m.StartTime != nil {
		return *m.StartTime
	}
	return ""
}

func (m *GVGCONFIG) GetGVGOpeningTime() uint32 {
	if m != nil && m.GVGOpeningTime != nil {
		return *m.GVGOpeningTime
	}
	return Default_GVGCONFIG_GVGOpeningTime
}

func (m *GVGCONFIG) GetGVGStatementTime() uint32 {
	if m != nil && m.GVGStatementTime != nil {
		return *m.GVGStatementTime
	}
	return Default_GVGCONFIG_GVGStatementTime
}

type GVGCONFIG_ARRAY struct {
	Items            []*GVGCONFIG `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *GVGCONFIG_ARRAY) Reset()         { *m = GVGCONFIG_ARRAY{} }
func (m *GVGCONFIG_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GVGCONFIG_ARRAY) ProtoMessage()    {}

func (m *GVGCONFIG_ARRAY) GetItems() []*GVGCONFIG {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}