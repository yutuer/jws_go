// automatically generated by the FlatBuffers compiler, do not modify

package fenghuomsg

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Attr struct {
	_tab flatbuffers.Table
}

func GetRootAsAttr(buf []byte, offset flatbuffers.UOffsetT) *Attr {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Attr{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Attr) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Attr) Atk() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateAtk(n float32) bool {
	return rcv._tab.MutateFloat32Slot(4, n)
}

func (rcv *Attr) Def() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateDef(n float32) bool {
	return rcv._tab.MutateFloat32Slot(6, n)
}

func (rcv *Attr) Hp() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateHp(n float32) bool {
	return rcv._tab.MutateFloat32Slot(8, n)
}

func (rcv *Attr) CritRate() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateCritRate(n float32) bool {
	return rcv._tab.MutateFloat32Slot(10, n)
}

func (rcv *Attr) ResilienceRate() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateResilienceRate(n float32) bool {
	return rcv._tab.MutateFloat32Slot(12, n)
}

func (rcv *Attr) CritValue() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateCritValue(n float32) bool {
	return rcv._tab.MutateFloat32Slot(14, n)
}

func (rcv *Attr) ResilienceValue() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateResilienceValue(n float32) bool {
	return rcv._tab.MutateFloat32Slot(16, n)
}

func (rcv *Attr) IceDamage() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateIceDamage(n int32) bool {
	return rcv._tab.MutateInt32Slot(18, n)
}

func (rcv *Attr) IceDefense() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateIceDefense(n int32) bool {
	return rcv._tab.MutateInt32Slot(20, n)
}

func (rcv *Attr) IceBonus() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateIceBonus(n float32) bool {
	return rcv._tab.MutateFloat32Slot(22, n)
}

func (rcv *Attr) IceResist() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateIceResist(n float32) bool {
	return rcv._tab.MutateFloat32Slot(24, n)
}

func (rcv *Attr) FireDamage() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(26))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateFireDamage(n int32) bool {
	return rcv._tab.MutateInt32Slot(26, n)
}

func (rcv *Attr) FireDefense() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(28))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateFireDefense(n int32) bool {
	return rcv._tab.MutateInt32Slot(28, n)
}

func (rcv *Attr) FireBonus() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(30))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateFireBonus(n float32) bool {
	return rcv._tab.MutateFloat32Slot(30, n)
}

func (rcv *Attr) FireResist() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(32))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateFireResist(n float32) bool {
	return rcv._tab.MutateFloat32Slot(32, n)
}

func (rcv *Attr) LightingDamage() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(34))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateLightingDamage(n int32) bool {
	return rcv._tab.MutateInt32Slot(34, n)
}

func (rcv *Attr) LightingDefense() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(36))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutateLightingDefense(n int32) bool {
	return rcv._tab.MutateInt32Slot(36, n)
}

func (rcv *Attr) LightingBonus() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(38))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateLightingBonus(n float32) bool {
	return rcv._tab.MutateFloat32Slot(38, n)
}

func (rcv *Attr) LightingResist() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(40))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutateLightingResist(n float32) bool {
	return rcv._tab.MutateFloat32Slot(40, n)
}

func (rcv *Attr) PoisonDamage() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(42))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutatePoisonDamage(n int32) bool {
	return rcv._tab.MutateInt32Slot(42, n)
}

func (rcv *Attr) PoisonDefense() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(44))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Attr) MutatePoisonDefense(n int32) bool {
	return rcv._tab.MutateInt32Slot(44, n)
}

func (rcv *Attr) PoisonBonus() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(46))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutatePoisonBonus(n float32) bool {
	return rcv._tab.MutateFloat32Slot(46, n)
}

func (rcv *Attr) PoisonResist() float32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(48))
	if o != 0 {
		return rcv._tab.GetFloat32(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Attr) MutatePoisonResist(n float32) bool {
	return rcv._tab.MutateFloat32Slot(48, n)
}

func AttrStart(builder *flatbuffers.Builder) {
	builder.StartObject(23)
}
func AttrAddAtk(builder *flatbuffers.Builder, atk float32) {
	builder.PrependFloat32Slot(0, atk, 0.0)
}
func AttrAddDef(builder *flatbuffers.Builder, def float32) {
	builder.PrependFloat32Slot(1, def, 0.0)
}
func AttrAddHp(builder *flatbuffers.Builder, hp float32) {
	builder.PrependFloat32Slot(2, hp, 0.0)
}
func AttrAddCritRate(builder *flatbuffers.Builder, critRate float32) {
	builder.PrependFloat32Slot(3, critRate, 0.0)
}
func AttrAddResilienceRate(builder *flatbuffers.Builder, resilienceRate float32) {
	builder.PrependFloat32Slot(4, resilienceRate, 0.0)
}
func AttrAddCritValue(builder *flatbuffers.Builder, critValue float32) {
	builder.PrependFloat32Slot(5, critValue, 0.0)
}
func AttrAddResilienceValue(builder *flatbuffers.Builder, resilienceValue float32) {
	builder.PrependFloat32Slot(6, resilienceValue, 0.0)
}
func AttrAddIceDamage(builder *flatbuffers.Builder, iceDamage int32) {
	builder.PrependInt32Slot(7, iceDamage, 0)
}
func AttrAddIceDefense(builder *flatbuffers.Builder, iceDefense int32) {
	builder.PrependInt32Slot(8, iceDefense, 0)
}
func AttrAddIceBonus(builder *flatbuffers.Builder, iceBonus float32) {
	builder.PrependFloat32Slot(9, iceBonus, 0.0)
}
func AttrAddIceResist(builder *flatbuffers.Builder, iceResist float32) {
	builder.PrependFloat32Slot(10, iceResist, 0.0)
}
func AttrAddFireDamage(builder *flatbuffers.Builder, fireDamage int32) {
	builder.PrependInt32Slot(11, fireDamage, 0)
}
func AttrAddFireDefense(builder *flatbuffers.Builder, fireDefense int32) {
	builder.PrependInt32Slot(12, fireDefense, 0)
}
func AttrAddFireBonus(builder *flatbuffers.Builder, fireBonus float32) {
	builder.PrependFloat32Slot(13, fireBonus, 0.0)
}
func AttrAddFireResist(builder *flatbuffers.Builder, fireResist float32) {
	builder.PrependFloat32Slot(14, fireResist, 0.0)
}
func AttrAddLightingDamage(builder *flatbuffers.Builder, lightingDamage int32) {
	builder.PrependInt32Slot(15, lightingDamage, 0)
}
func AttrAddLightingDefense(builder *flatbuffers.Builder, lightingDefense int32) {
	builder.PrependInt32Slot(16, lightingDefense, 0)
}
func AttrAddLightingBonus(builder *flatbuffers.Builder, lightingBonus float32) {
	builder.PrependFloat32Slot(17, lightingBonus, 0.0)
}
func AttrAddLightingResist(builder *flatbuffers.Builder, lightingResist float32) {
	builder.PrependFloat32Slot(18, lightingResist, 0.0)
}
func AttrAddPoisonDamage(builder *flatbuffers.Builder, poisonDamage int32) {
	builder.PrependInt32Slot(19, poisonDamage, 0)
}
func AttrAddPoisonDefense(builder *flatbuffers.Builder, poisonDefense int32) {
	builder.PrependInt32Slot(20, poisonDefense, 0)
}
func AttrAddPoisonBonus(builder *flatbuffers.Builder, poisonBonus float32) {
	builder.PrependFloat32Slot(21, poisonBonus, 0.0)
}
func AttrAddPoisonResist(builder *flatbuffers.Builder, poisonResist float32) {
	builder.PrependFloat32Slot(22, poisonResist, 0.0)
}
func AttrEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
