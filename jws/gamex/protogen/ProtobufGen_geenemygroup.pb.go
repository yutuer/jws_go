// Code generated by protoc-gen-go.
// source: ProtobufGen_geenemygroup.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GEENEMYGROUP struct {
	// * 军团的ID
	EnemyGroupID *uint32 `protobuf:"varint,1,req,def=0" json:"EnemyGroupID,omitempty"`
	// * 激活所需杀戮值
	ActiveCondition *uint32 `protobuf:"varint,2,req,def=0" json:"ActiveCondition,omitempty"`
	// * 数量上限
	NumberLimit *uint32 `protobuf:"varint,3,opt,def=0" json:"NumberLimit,omitempty"`
	// * 刷新时间（秒）
	RenovateTime *uint32 `protobuf:"varint,4,opt,def=0" json:"RenovateTime,omitempty"`
	// * 关卡ID
	EGLevelID *string `protobuf:"bytes,5,opt,def=" json:"EGLevelID,omitempty"`
	// * 军团名字
	EnemyName *string `protobuf:"bytes,6,opt,def=" json:"EnemyName,omitempty"`
	// * 军团图片
	EnemyImage *string `protobuf:"bytes,7,opt,def=" json:"EnemyImage,omitempty"`
	// * 军团头像（仅限BOSS）
	EnemyPortrait    *string `protobuf:"bytes,8,opt,def=" json:"EnemyPortrait,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GEENEMYGROUP) Reset()         { *m = GEENEMYGROUP{} }
func (m *GEENEMYGROUP) String() string { return proto.CompactTextString(m) }
func (*GEENEMYGROUP) ProtoMessage()    {}

const Default_GEENEMYGROUP_EnemyGroupID uint32 = 0
const Default_GEENEMYGROUP_ActiveCondition uint32 = 0
const Default_GEENEMYGROUP_NumberLimit uint32 = 0
const Default_GEENEMYGROUP_RenovateTime uint32 = 0

func (m *GEENEMYGROUP) GetEnemyGroupID() uint32 {
	if m != nil && m.EnemyGroupID != nil {
		return *m.EnemyGroupID
	}
	return Default_GEENEMYGROUP_EnemyGroupID
}

func (m *GEENEMYGROUP) GetActiveCondition() uint32 {
	if m != nil && m.ActiveCondition != nil {
		return *m.ActiveCondition
	}
	return Default_GEENEMYGROUP_ActiveCondition
}

func (m *GEENEMYGROUP) GetNumberLimit() uint32 {
	if m != nil && m.NumberLimit != nil {
		return *m.NumberLimit
	}
	return Default_GEENEMYGROUP_NumberLimit
}

func (m *GEENEMYGROUP) GetRenovateTime() uint32 {
	if m != nil && m.RenovateTime != nil {
		return *m.RenovateTime
	}
	return Default_GEENEMYGROUP_RenovateTime
}

func (m *GEENEMYGROUP) GetEGLevelID() string {
	if m != nil && m.EGLevelID != nil {
		return *m.EGLevelID
	}
	return ""
}

func (m *GEENEMYGROUP) GetEnemyName() string {
	if m != nil && m.EnemyName != nil {
		return *m.EnemyName
	}
	return ""
}

func (m *GEENEMYGROUP) GetEnemyImage() string {
	if m != nil && m.EnemyImage != nil {
		return *m.EnemyImage
	}
	return ""
}

func (m *GEENEMYGROUP) GetEnemyPortrait() string {
	if m != nil && m.EnemyPortrait != nil {
		return *m.EnemyPortrait
	}
	return ""
}

type GEENEMYGROUP_ARRAY struct {
	Items            []*GEENEMYGROUP `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *GEENEMYGROUP_ARRAY) Reset()         { *m = GEENEMYGROUP_ARRAY{} }
func (m *GEENEMYGROUP_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GEENEMYGROUP_ARRAY) ProtoMessage()    {}

func (m *GEENEMYGROUP_ARRAY) GetItems() []*GEENEMYGROUP {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}