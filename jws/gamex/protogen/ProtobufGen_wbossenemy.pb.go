// Code generated by protoc-gen-go.
// source: ProtobufGen_wbossenemy.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type WBOSSENEMY struct {
	// * 敌兵的ID
	BossID *string `protobuf:"bytes,1,req,def=" json:"BossID,omitempty"`
	// * 在界面展示的立绘
	PanelShowPic *string `protobuf:"bytes,29,opt,def=" json:"PanelShowPic,omitempty"`
	// * 敌兵的模型
	CharacterID *string `protobuf:"bytes,3,opt,def=" json:"CharacterID,omitempty"`
	// * 敌兵的类型
	Type *string `protobuf:"bytes,4,opt,def=" json:"Type,omitempty"`
	// * 克制ID
	Idid *string `protobuf:"bytes,5,opt,def=" json:"Idid,omitempty"`
	// * 敌兵的名字
	NameIDs *string `protobuf:"bytes,6,opt,def=" json:"NameIDs,omitempty"`
	// *
	StageIDs *string `protobuf:"bytes,7,opt,def=" json:"StageIDs,omitempty"`
	// *
	Faction *string `protobuf:"bytes,8,opt,def=" json:"Faction,omitempty"`
	// *
	IsPlayer *uint32 `protobuf:"varint,9,opt,def=0" json:"IsPlayer,omitempty"`
	// *
	Speed *uint32 `protobuf:"varint,10,opt,def=0" json:"Speed,omitempty"`
	// *
	AngleSpeed *uint32 `protobuf:"varint,11,opt,def=0" json:"AngleSpeed,omitempty"`
	// * 生命系数（前端用了但后端并没有用，填1不要变！）
	HitPointCoefficient *float32 `protobuf:"fixed32,12,opt,def=0" json:"HitPointCoefficient,omitempty"`
	// * 血条数
	HPSectionNum *uint32 `protobuf:"varint,13,opt,def=1" json:"HPSectionNum,omitempty"`
	// * 护甲下限
	ThresholdMin *uint32 `protobuf:"varint,14,opt,def=0" json:"ThresholdMin,omitempty"`
	// * 护甲上限
	ThresholdMax *uint32 `protobuf:"varint,15,opt,def=0" json:"ThresholdMax,omitempty"`
	// * 护盾系数
	ThresholdRatio *uint32 `protobuf:"varint,16,opt,def=0" json:"ThresholdRatio,omitempty"`
	// *
	Guard *uint32 `protobuf:"varint,17,opt,def=0" json:"Guard,omitempty"`
	// * 护盾吸收率
	ShieldAbsorbRate *float32 `protobuf:"fixed32,18,opt,def=0.8" json:"ShieldAbsorbRate,omitempty"`
	// * 攻击系数
	PhysicalDamageCoefficient *float32 `protobuf:"fixed32,19,opt,def=0" json:"PhysicalDamageCoefficient,omitempty"`
	// * 防御系数
	PhysicalResistCoefficient *float32 `protobuf:"fixed32,20,opt,def=0" json:"PhysicalResistCoefficient,omitempty"`
	// *
	CritRate *float32 `protobuf:"fixed32,21,opt,def=0" json:"CritRate,omitempty"`
	// *
	CritDamage *float32 `protobuf:"fixed32,22,opt,def=0" json:"CritDamage,omitempty"`
	// * 闪避率
	DodgeRate *float32 `protobuf:"fixed32,30,opt,def=0" json:"DodgeRate,omitempty"`
	// * 命中率
	HitRate *float32 `protobuf:"fixed32,31,opt,def=0" json:"HitRate,omitempty"`
	// * 无法被黑洞牵引
	CantbeBlackHole *uint32 `protobuf:"varint,23,opt,def=0" json:"CantbeBlackHole,omitempty"`
	// * 无法被击飞、挑起等特殊效果作用
	CantbeSpecialHit *uint32 `protobuf:"varint,24,opt,def=0" json:"CantbeSpecialHit,omitempty"`
	// * 无法被冰冻效果作用
	CantbeFrozen *uint32                `protobuf:"varint,32,opt,def=0" json:"CantbeFrozen,omitempty"`
	Loots        []*WBOSSENEMY_LootRule `protobuf:"bytes,25,rep" json:"Loots,omitempty"`
	// *
	Equip1 *string `protobuf:"bytes,26,opt,def=" json:"Equip1,omitempty"`
	// *
	Equip2 *string `protobuf:"bytes,27,opt,def=" json:"Equip2,omitempty"`
	// * 光环BuffIDs
	Aura             *string `protobuf:"bytes,28,opt,def=" json:"Aura,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *WBOSSENEMY) Reset()         { *m = WBOSSENEMY{} }
func (m *WBOSSENEMY) String() string { return proto.CompactTextString(m) }
func (*WBOSSENEMY) ProtoMessage()    {}

const Default_WBOSSENEMY_IsPlayer uint32 = 0
const Default_WBOSSENEMY_Speed uint32 = 0
const Default_WBOSSENEMY_AngleSpeed uint32 = 0
const Default_WBOSSENEMY_HitPointCoefficient float32 = 0
const Default_WBOSSENEMY_HPSectionNum uint32 = 1
const Default_WBOSSENEMY_ThresholdMin uint32 = 0
const Default_WBOSSENEMY_ThresholdMax uint32 = 0
const Default_WBOSSENEMY_ThresholdRatio uint32 = 0
const Default_WBOSSENEMY_Guard uint32 = 0
const Default_WBOSSENEMY_ShieldAbsorbRate float32 = 0.8
const Default_WBOSSENEMY_PhysicalDamageCoefficient float32 = 0
const Default_WBOSSENEMY_PhysicalResistCoefficient float32 = 0
const Default_WBOSSENEMY_CritRate float32 = 0
const Default_WBOSSENEMY_CritDamage float32 = 0
const Default_WBOSSENEMY_DodgeRate float32 = 0
const Default_WBOSSENEMY_HitRate float32 = 0
const Default_WBOSSENEMY_CantbeBlackHole uint32 = 0
const Default_WBOSSENEMY_CantbeSpecialHit uint32 = 0
const Default_WBOSSENEMY_CantbeFrozen uint32 = 0

func (m *WBOSSENEMY) GetBossID() string {
	if m != nil && m.BossID != nil {
		return *m.BossID
	}
	return ""
}

func (m *WBOSSENEMY) GetPanelShowPic() string {
	if m != nil && m.PanelShowPic != nil {
		return *m.PanelShowPic
	}
	return ""
}

func (m *WBOSSENEMY) GetCharacterID() string {
	if m != nil && m.CharacterID != nil {
		return *m.CharacterID
	}
	return ""
}

func (m *WBOSSENEMY) GetType() string {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ""
}

func (m *WBOSSENEMY) GetIdid() string {
	if m != nil && m.Idid != nil {
		return *m.Idid
	}
	return ""
}

func (m *WBOSSENEMY) GetNameIDs() string {
	if m != nil && m.NameIDs != nil {
		return *m.NameIDs
	}
	return ""
}

func (m *WBOSSENEMY) GetStageIDs() string {
	if m != nil && m.StageIDs != nil {
		return *m.StageIDs
	}
	return ""
}

func (m *WBOSSENEMY) GetFaction() string {
	if m != nil && m.Faction != nil {
		return *m.Faction
	}
	return ""
}

func (m *WBOSSENEMY) GetIsPlayer() uint32 {
	if m != nil && m.IsPlayer != nil {
		return *m.IsPlayer
	}
	return Default_WBOSSENEMY_IsPlayer
}

func (m *WBOSSENEMY) GetSpeed() uint32 {
	if m != nil && m.Speed != nil {
		return *m.Speed
	}
	return Default_WBOSSENEMY_Speed
}

func (m *WBOSSENEMY) GetAngleSpeed() uint32 {
	if m != nil && m.AngleSpeed != nil {
		return *m.AngleSpeed
	}
	return Default_WBOSSENEMY_AngleSpeed
}

func (m *WBOSSENEMY) GetHitPointCoefficient() float32 {
	if m != nil && m.HitPointCoefficient != nil {
		return *m.HitPointCoefficient
	}
	return Default_WBOSSENEMY_HitPointCoefficient
}

func (m *WBOSSENEMY) GetHPSectionNum() uint32 {
	if m != nil && m.HPSectionNum != nil {
		return *m.HPSectionNum
	}
	return Default_WBOSSENEMY_HPSectionNum
}

func (m *WBOSSENEMY) GetThresholdMin() uint32 {
	if m != nil && m.ThresholdMin != nil {
		return *m.ThresholdMin
	}
	return Default_WBOSSENEMY_ThresholdMin
}

func (m *WBOSSENEMY) GetThresholdMax() uint32 {
	if m != nil && m.ThresholdMax != nil {
		return *m.ThresholdMax
	}
	return Default_WBOSSENEMY_ThresholdMax
}

func (m *WBOSSENEMY) GetThresholdRatio() uint32 {
	if m != nil && m.ThresholdRatio != nil {
		return *m.ThresholdRatio
	}
	return Default_WBOSSENEMY_ThresholdRatio
}

func (m *WBOSSENEMY) GetGuard() uint32 {
	if m != nil && m.Guard != nil {
		return *m.Guard
	}
	return Default_WBOSSENEMY_Guard
}

func (m *WBOSSENEMY) GetShieldAbsorbRate() float32 {
	if m != nil && m.ShieldAbsorbRate != nil {
		return *m.ShieldAbsorbRate
	}
	return Default_WBOSSENEMY_ShieldAbsorbRate
}

func (m *WBOSSENEMY) GetPhysicalDamageCoefficient() float32 {
	if m != nil && m.PhysicalDamageCoefficient != nil {
		return *m.PhysicalDamageCoefficient
	}
	return Default_WBOSSENEMY_PhysicalDamageCoefficient
}

func (m *WBOSSENEMY) GetPhysicalResistCoefficient() float32 {
	if m != nil && m.PhysicalResistCoefficient != nil {
		return *m.PhysicalResistCoefficient
	}
	return Default_WBOSSENEMY_PhysicalResistCoefficient
}

func (m *WBOSSENEMY) GetCritRate() float32 {
	if m != nil && m.CritRate != nil {
		return *m.CritRate
	}
	return Default_WBOSSENEMY_CritRate
}

func (m *WBOSSENEMY) GetCritDamage() float32 {
	if m != nil && m.CritDamage != nil {
		return *m.CritDamage
	}
	return Default_WBOSSENEMY_CritDamage
}

func (m *WBOSSENEMY) GetDodgeRate() float32 {
	if m != nil && m.DodgeRate != nil {
		return *m.DodgeRate
	}
	return Default_WBOSSENEMY_DodgeRate
}

func (m *WBOSSENEMY) GetHitRate() float32 {
	if m != nil && m.HitRate != nil {
		return *m.HitRate
	}
	return Default_WBOSSENEMY_HitRate
}

func (m *WBOSSENEMY) GetCantbeBlackHole() uint32 {
	if m != nil && m.CantbeBlackHole != nil {
		return *m.CantbeBlackHole
	}
	return Default_WBOSSENEMY_CantbeBlackHole
}

func (m *WBOSSENEMY) GetCantbeSpecialHit() uint32 {
	if m != nil && m.CantbeSpecialHit != nil {
		return *m.CantbeSpecialHit
	}
	return Default_WBOSSENEMY_CantbeSpecialHit
}

func (m *WBOSSENEMY) GetCantbeFrozen() uint32 {
	if m != nil && m.CantbeFrozen != nil {
		return *m.CantbeFrozen
	}
	return Default_WBOSSENEMY_CantbeFrozen
}

func (m *WBOSSENEMY) GetLoots() []*WBOSSENEMY_LootRule {
	if m != nil {
		return m.Loots
	}
	return nil
}

func (m *WBOSSENEMY) GetEquip1() string {
	if m != nil && m.Equip1 != nil {
		return *m.Equip1
	}
	return ""
}

func (m *WBOSSENEMY) GetEquip2() string {
	if m != nil && m.Equip2 != nil {
		return *m.Equip2
	}
	return ""
}

func (m *WBOSSENEMY) GetAura() string {
	if m != nil && m.Aura != nil {
		return *m.Aura
	}
	return ""
}

type WBOSSENEMY_LootRule struct {
	// *
	LootTemplate *string `protobuf:"bytes,1,opt,def=" json:"LootTemplate,omitempty"`
	// *
	LootTimes        *uint32 `protobuf:"varint,2,opt,def=0" json:"LootTimes,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *WBOSSENEMY_LootRule) Reset()         { *m = WBOSSENEMY_LootRule{} }
func (m *WBOSSENEMY_LootRule) String() string { return proto.CompactTextString(m) }
func (*WBOSSENEMY_LootRule) ProtoMessage()    {}

const Default_WBOSSENEMY_LootRule_LootTimes uint32 = 0

func (m *WBOSSENEMY_LootRule) GetLootTemplate() string {
	if m != nil && m.LootTemplate != nil {
		return *m.LootTemplate
	}
	return ""
}

func (m *WBOSSENEMY_LootRule) GetLootTimes() uint32 {
	if m != nil && m.LootTimes != nil {
		return *m.LootTimes
	}
	return Default_WBOSSENEMY_LootRule_LootTimes
}

type WBOSSENEMY_ARRAY struct {
	Items            []*WBOSSENEMY `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *WBOSSENEMY_ARRAY) Reset()         { *m = WBOSSENEMY_ARRAY{} }
func (m *WBOSSENEMY_ARRAY) String() string { return proto.CompactTextString(m) }
func (*WBOSSENEMY_ARRAY) ProtoMessage()    {}

func (m *WBOSSENEMY_ARRAY) GetItems() []*WBOSSENEMY {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
