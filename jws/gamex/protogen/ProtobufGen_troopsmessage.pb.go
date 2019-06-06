// Code generated by protoc-gen-go.
// source: ProtobufGen_troopsmessage.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type TROOPSMESSAGE struct {
	// * 关卡ID
	LevelID          *string                         `protobuf:"bytes,1,req,name=levelID,def=" json:"levelID,omitempty"`
	Troop_Table      []*TROOPSMESSAGE_TroopCondition `protobuf:"bytes,2,rep" json:"Troop_Table,omitempty"`
	XXX_unrecognized []byte                          `json:"-"`
}

func (m *TROOPSMESSAGE) Reset()         { *m = TROOPSMESSAGE{} }
func (m *TROOPSMESSAGE) String() string { return proto.CompactTextString(m) }
func (*TROOPSMESSAGE) ProtoMessage()    {}

func (m *TROOPSMESSAGE) GetLevelID() string {
	if m != nil && m.LevelID != nil {
		return *m.LevelID
	}
	return ""
}

func (m *TROOPSMESSAGE) GetTroop_Table() []*TROOPSMESSAGE_TroopCondition {
	if m != nil {
		return m.Troop_Table
	}
	return nil
}

type TROOPSMESSAGE_TroopCondition struct {
	// * 第几波兵
	WaveNumber *uint32 `protobuf:"varint,1,opt,def=0" json:"WaveNumber,omitempty"`
	// * 敌兵的ID
	WBossID *string `protobuf:"bytes,2,opt,def=" json:"WBossID,omitempty"`
	// * 该ID兵的数量
	EnemyNumber      *uint32 `protobuf:"varint,3,opt,def=0" json:"EnemyNumber,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TROOPSMESSAGE_TroopCondition) Reset()         { *m = TROOPSMESSAGE_TroopCondition{} }
func (m *TROOPSMESSAGE_TroopCondition) String() string { return proto.CompactTextString(m) }
func (*TROOPSMESSAGE_TroopCondition) ProtoMessage()    {}

const Default_TROOPSMESSAGE_TroopCondition_WaveNumber uint32 = 0
const Default_TROOPSMESSAGE_TroopCondition_EnemyNumber uint32 = 0

func (m *TROOPSMESSAGE_TroopCondition) GetWaveNumber() uint32 {
	if m != nil && m.WaveNumber != nil {
		return *m.WaveNumber
	}
	return Default_TROOPSMESSAGE_TroopCondition_WaveNumber
}

func (m *TROOPSMESSAGE_TroopCondition) GetWBossID() string {
	if m != nil && m.WBossID != nil {
		return *m.WBossID
	}
	return ""
}

func (m *TROOPSMESSAGE_TroopCondition) GetEnemyNumber() uint32 {
	if m != nil && m.EnemyNumber != nil {
		return *m.EnemyNumber
	}
	return Default_TROOPSMESSAGE_TroopCondition_EnemyNumber
}

type TROOPSMESSAGE_ARRAY struct {
	Items            []*TROOPSMESSAGE `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *TROOPSMESSAGE_ARRAY) Reset()         { *m = TROOPSMESSAGE_ARRAY{} }
func (m *TROOPSMESSAGE_ARRAY) String() string { return proto.CompactTextString(m) }
func (*TROOPSMESSAGE_ARRAY) ProtoMessage()    {}

func (m *TROOPSMESSAGE_ARRAY) GetItems() []*TROOPSMESSAGE {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
