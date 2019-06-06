// Code generated by protoc-gen-go.
// source: ProtobufGen_severgroup.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type SEVERGROUP struct {
	// * 服务器ID
	SID *uint32 `protobuf:"varint,1,req,def=0" json:"SID,omitempty"`
	// * 跨服同组（从1开始）
	GroupID *uint32 `protobuf:"varint,2,opt,def=0" json:"GroupID,omitempty"`
	// * 无双竞技场服务器分组
	WspvpGroupID *uint32 `protobuf:"varint,5,opt,def=0" json:"WspvpGroupID,omitempty"`
	// * 无双竞技场机器人配置
	WspvpBot *uint32 `protobuf:"varint,6,opt,def=0" json:"WspvpBot,omitempty"`
	// * 劫营夺粮服务器分组
	RobCropsGroupID *uint32 `protobuf:"varint,7,opt,def=0" json:"RobCropsGroupID,omitempty"`
	// * 服务器付费批次
	Sbatch *uint32 `protobuf:"varint,3,opt,def=0" json:"Sbatch,omitempty"`
	// * 批次生效时间
	EffectiveTime *string `protobuf:"bytes,4,opt,def=" json:"EffectiveTime,omitempty"`
	// * 世界boss服务器分组
	WorldBossGroupID *uint32 `protobuf:"varint,8,opt,def=0" json:"WorldBossGroupID,omitempty"`
	// * 单服限时神将配置
	HGRHotID *uint32 `protobuf:"varint,9,opt,def=0" json:"HGRHotID,omitempty"`
	// * 组队boss服务器分组
	TeamBossGroupID  *uint32 `protobuf:"varint,10,opt,def=0" json:"TeamBossGroupID,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *SEVERGROUP) Reset()         { *m = SEVERGROUP{} }
func (m *SEVERGROUP) String() string { return proto.CompactTextString(m) }
func (*SEVERGROUP) ProtoMessage()    {}

const Default_SEVERGROUP_SID uint32 = 0
const Default_SEVERGROUP_GroupID uint32 = 0
const Default_SEVERGROUP_WspvpGroupID uint32 = 0
const Default_SEVERGROUP_WspvpBot uint32 = 0
const Default_SEVERGROUP_RobCropsGroupID uint32 = 0
const Default_SEVERGROUP_Sbatch uint32 = 0
const Default_SEVERGROUP_WorldBossGroupID uint32 = 0
const Default_SEVERGROUP_HGRHotID uint32 = 0
const Default_SEVERGROUP_TeamBossGroupID uint32 = 0

func (m *SEVERGROUP) GetSID() uint32 {
	if m != nil && m.SID != nil {
		return *m.SID
	}
	return Default_SEVERGROUP_SID
}

func (m *SEVERGROUP) GetGroupID() uint32 {
	if m != nil && m.GroupID != nil {
		return *m.GroupID
	}
	return Default_SEVERGROUP_GroupID
}

func (m *SEVERGROUP) GetWspvpGroupID() uint32 {
	if m != nil && m.WspvpGroupID != nil {
		return *m.WspvpGroupID
	}
	return Default_SEVERGROUP_WspvpGroupID
}

func (m *SEVERGROUP) GetWspvpBot() uint32 {
	if m != nil && m.WspvpBot != nil {
		return *m.WspvpBot
	}
	return Default_SEVERGROUP_WspvpBot
}

func (m *SEVERGROUP) GetRobCropsGroupID() uint32 {
	if m != nil && m.RobCropsGroupID != nil {
		return *m.RobCropsGroupID
	}
	return Default_SEVERGROUP_RobCropsGroupID
}

func (m *SEVERGROUP) GetSbatch() uint32 {
	if m != nil && m.Sbatch != nil {
		return *m.Sbatch
	}
	return Default_SEVERGROUP_Sbatch
}

func (m *SEVERGROUP) GetEffectiveTime() string {
	if m != nil && m.EffectiveTime != nil {
		return *m.EffectiveTime
	}
	return ""
}

func (m *SEVERGROUP) GetWorldBossGroupID() uint32 {
	if m != nil && m.WorldBossGroupID != nil {
		return *m.WorldBossGroupID
	}
	return Default_SEVERGROUP_WorldBossGroupID
}

func (m *SEVERGROUP) GetHGRHotID() uint32 {
	if m != nil && m.HGRHotID != nil {
		return *m.HGRHotID
	}
	return Default_SEVERGROUP_HGRHotID
}

func (m *SEVERGROUP) GetTeamBossGroupID() uint32 {
	if m != nil && m.TeamBossGroupID != nil {
		return *m.TeamBossGroupID
	}
	return Default_SEVERGROUP_TeamBossGroupID
}

type SEVERGROUP_ARRAY struct {
	Items            []*SEVERGROUP `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *SEVERGROUP_ARRAY) Reset()         { *m = SEVERGROUP_ARRAY{} }
func (m *SEVERGROUP_ARRAY) String() string { return proto.CompactTextString(m) }
func (*SEVERGROUP_ARRAY) ProtoMessage()    {}

func (m *SEVERGROUP_ARRAY) GetItems() []*SEVERGROUP {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
