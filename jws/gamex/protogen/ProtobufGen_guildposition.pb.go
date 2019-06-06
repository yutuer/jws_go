// Code generated by protoc-gen-go.
// source: ProtobufGen_guildposition.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type GUILDPOSITION struct {
	// * 公会职位
	Position *uint32 `protobuf:"varint,1,req,def=0" json:"Position,omitempty"`
	// * 数量上限
	PositionNumber *uint32 `protobuf:"varint,2,opt,def=0" json:"PositionNumber,omitempty"`
	// * 职位名称
	PositionName *string `protobuf:"bytes,3,opt,def=" json:"PositionName,omitempty"`
	// * 任命副会长
	AppointVP *uint32 `protobuf:"varint,4,opt,def=0" json:"AppointVP,omitempty"`
	// * 任命精英
	AppointElite *uint32 `protobuf:"varint,5,opt,def=0" json:"AppointElite,omitempty"`
	// * 任命会员
	AppointMember *uint32 `protobuf:"varint,6,opt,def=0" json:"AppointMember,omitempty"`
	// * 踢出成员
	KickMember *uint32 `protobuf:"varint,7,opt,def=0" json:"KickMember,omitempty"`
	// * 同意添加成员
	AddMember *uint32 `protobuf:"varint,8,opt,def=0" json:"AddMember,omitempty"`
	// * 开启兵临城下的权限
	ActiveGE *uint32 `protobuf:"varint,9,opt,def=0" json:"ActiveGE,omitempty"`
	// * 分配仓库奖励权限
	AllotPower *uint32 `protobuf:"varint,10,opt,def=0" json:"AllotPower,omitempty"`
	// * 修改公会名称权限
	ReNamePower *uint32 `protobuf:"varint,11,opt,def=0" json:"ReNamePower,omitempty"`
	// * 修改公会公告权限
	RenNewsPower     *uint32 `protobuf:"varint,12,opt,def=0" json:"RenNewsPower,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *GUILDPOSITION) Reset()         { *m = GUILDPOSITION{} }
func (m *GUILDPOSITION) String() string { return proto.CompactTextString(m) }
func (*GUILDPOSITION) ProtoMessage()    {}

const Default_GUILDPOSITION_Position uint32 = 0
const Default_GUILDPOSITION_PositionNumber uint32 = 0
const Default_GUILDPOSITION_AppointVP uint32 = 0
const Default_GUILDPOSITION_AppointElite uint32 = 0
const Default_GUILDPOSITION_AppointMember uint32 = 0
const Default_GUILDPOSITION_KickMember uint32 = 0
const Default_GUILDPOSITION_AddMember uint32 = 0
const Default_GUILDPOSITION_ActiveGE uint32 = 0
const Default_GUILDPOSITION_AllotPower uint32 = 0
const Default_GUILDPOSITION_ReNamePower uint32 = 0
const Default_GUILDPOSITION_RenNewsPower uint32 = 0

func (m *GUILDPOSITION) GetPosition() uint32 {
	if m != nil && m.Position != nil {
		return *m.Position
	}
	return Default_GUILDPOSITION_Position
}

func (m *GUILDPOSITION) GetPositionNumber() uint32 {
	if m != nil && m.PositionNumber != nil {
		return *m.PositionNumber
	}
	return Default_GUILDPOSITION_PositionNumber
}

func (m *GUILDPOSITION) GetPositionName() string {
	if m != nil && m.PositionName != nil {
		return *m.PositionName
	}
	return ""
}

func (m *GUILDPOSITION) GetAppointVP() uint32 {
	if m != nil && m.AppointVP != nil {
		return *m.AppointVP
	}
	return Default_GUILDPOSITION_AppointVP
}

func (m *GUILDPOSITION) GetAppointElite() uint32 {
	if m != nil && m.AppointElite != nil {
		return *m.AppointElite
	}
	return Default_GUILDPOSITION_AppointElite
}

func (m *GUILDPOSITION) GetAppointMember() uint32 {
	if m != nil && m.AppointMember != nil {
		return *m.AppointMember
	}
	return Default_GUILDPOSITION_AppointMember
}

func (m *GUILDPOSITION) GetKickMember() uint32 {
	if m != nil && m.KickMember != nil {
		return *m.KickMember
	}
	return Default_GUILDPOSITION_KickMember
}

func (m *GUILDPOSITION) GetAddMember() uint32 {
	if m != nil && m.AddMember != nil {
		return *m.AddMember
	}
	return Default_GUILDPOSITION_AddMember
}

func (m *GUILDPOSITION) GetActiveGE() uint32 {
	if m != nil && m.ActiveGE != nil {
		return *m.ActiveGE
	}
	return Default_GUILDPOSITION_ActiveGE
}

func (m *GUILDPOSITION) GetAllotPower() uint32 {
	if m != nil && m.AllotPower != nil {
		return *m.AllotPower
	}
	return Default_GUILDPOSITION_AllotPower
}

func (m *GUILDPOSITION) GetReNamePower() uint32 {
	if m != nil && m.ReNamePower != nil {
		return *m.ReNamePower
	}
	return Default_GUILDPOSITION_ReNamePower
}

func (m *GUILDPOSITION) GetRenNewsPower() uint32 {
	if m != nil && m.RenNewsPower != nil {
		return *m.RenNewsPower
	}
	return Default_GUILDPOSITION_RenNewsPower
}

type GUILDPOSITION_ARRAY struct {
	Items            []*GUILDPOSITION `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *GUILDPOSITION_ARRAY) Reset()         { *m = GUILDPOSITION_ARRAY{} }
func (m *GUILDPOSITION_ARRAY) String() string { return proto.CompactTextString(m) }
func (*GUILDPOSITION_ARRAY) ProtoMessage()    {}

func (m *GUILDPOSITION_ARRAY) GetItems() []*GUILDPOSITION {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}