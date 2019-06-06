package gamedata

import (
	"fmt"

	"math"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MP_Max = 100
)

//1攻击
//2防御
//3生命
//4暴击率
//5免暴率
//6暴击伤害
//7暴伤减免
//8伤害减免
//9伤害加深
//10闪避
//11命中
const (
	ATK = iota + 1
	DEF
	HP
	CRI_RATE
	DE_CRI_RATE
	CRI_DAMAGE
	DE_CRI_DAMAGE
	DAMAGE_LESS
	DAMAGE_MORE
	DODGE_RATE
	HIT_RATE
)

type AvatarAttr struct {
	helper.AvatarAttr_
}

type avatarAttrAddon AvatarAttr

var baseAttr AvatarAttr // 基础属性值
var baseAttrGs float32  // 基础属性值造成的Gs,计算Gs时需要减去

func (a *avatarAttrAddon) AddTrickAddon(trick_data *ProtobufGen.EUIPTRICKDETAIL) {
	// 2）随机属性
	//   区别随机属性类型，
	//   0为常规类型，在计算总GS时候不考虑他们，因为总属性已经体现。
	//   1为特殊类型，在通过总属性计算GS之后需要加上身上所有的此类型随机属性的GS

	switch trick_data.GetProperty() {
	/*
		case "BAOJI":
			a.CritRate += info.GetValue()
			break
		case "BAOSH":
			a.CritValue += info.GetValue()
			break
		case "MBAO":
			a.ResilienceRate += info.GetValue()
			break
		case "MBAOSH":
			a.ResilienceValue += info.GetValue()
			break
	*/
	case "ATK":
		a.ATK += float32(trick_data.GetValue())
		break
	case "DEF":
		a.DEF += float32(trick_data.GetValue())
		break
	case "HP":
		a.HP += float32(trick_data.GetValue())
		break
	default:
		//   0为常规类型，在计算总GS时候不考虑他们，因为总属性已经体现。
		//   1为特殊类型，在通过总属性计算GS之后需要加上身上所有的此类型随机属性的GS
		if trick_data.GetTrickType() == 1 {
			a.AddGsAddon_(trick_data.GetTrickGS())
		} else {
			//logs.Warn("AddTrickAddon No Eff By %v", info)
		}
	}

}

func (a *avatarAttrAddon) AddEquipAddon(attack, defense, hp float32) {
	a.ATK += (attack)
	a.DEF += (defense)
	a.HP += (hp)
}

func (a *AvatarAttr) AddAttr(attrType uint32, attValue float32) {
	switch attrType {
	case ATK:
		a.ATK += attValue
	case DEF:
		a.DEF += attValue
	case HP:
		a.HP += attValue
	case CRI_RATE:
		a.CritRate += attValue
	case DE_CRI_RATE:
		a.ResilienceRate += attValue
	case CRI_DAMAGE:
		a.CritValue += attValue
	case DE_CRI_DAMAGE:
		a.ResilienceValue += attValue
	case DAMAGE_LESS:
	case DAMAGE_MORE:
		// TODO
	case DODGE_RATE:
		a.DodgeRate += attValue
	case HIT_RATE:
		a.HitRate += attValue
	}
}

func (a *AvatarAttr) SubAttr(attrType uint32, attValue float32) {
	switch attrType {
	case ATK:
		a.ATK -= attValue
	case DEF:
		a.DEF -= attValue
	case HP:
		a.HP -= attValue
	case CRI_RATE:
		a.CritRate -= attValue
	case DE_CRI_RATE:
		a.ResilienceRate -= attValue
	case CRI_DAMAGE:
		a.CritValue -= attValue
	case DE_CRI_DAMAGE:
		a.ResilienceValue -= attValue
	case DAMAGE_LESS:
	case DAMAGE_MORE:
		// TODO
	case DODGE_RATE:
		a.DodgeRate -= attValue
	case HIT_RATE:
		a.HitRate -= attValue
	}
}

//AddAttrs ..
func (a *AvatarAttr) AddAttrs(attrs map[uint32]float32) {
	for p, v := range attrs {
		a.AddAttr(p, v)
	}
}

//SubAttrs ..
func (a *AvatarAttr) SubAttrs(attrs map[uint32]float32) {
	for p, v := range attrs {
		a.SubAttr(p, v)
	}
}

func (a *AvatarAttr) AddEquipStarAddon(rate float32) {
	a.ATK += (a.ATK) * rate
	a.DEF += (a.DEF) * rate
	a.HP += (a.HP) * rate
	//logs.Trace("[GS]AddEquipStarAddon to %v,%v,%v", a.ATK, a.DEF, a.HP)
}

func (a *AvatarAttr) Add(addon *avatarAttrAddon) {
	a.ATK += addon.ATK
	a.DEF += addon.DEF
	a.HP += addon.HP
	a.CritRate += addon.CritRate
	a.ResilienceRate += addon.ResilienceRate
	a.CritValue += addon.CritValue
	a.ResilienceValue += addon.ResilienceValue
	a.IceDamage += addon.IceDamage
	a.IceDefense += addon.IceDefense
	a.IceBonus += addon.IceBonus
	a.IceResist += addon.IceResist
	a.FireDamage += addon.FireDamage
	a.FireDefense += addon.FireDefense
	a.FireBonus += addon.FireBonus
	a.FireResist += addon.FireResist
	a.LightingDamage += addon.LightingDamage
	a.LightingDefense += addon.LightingDefense
	a.LightingBonus += addon.LightingBonus
	a.LightingResist += addon.LightingResist
	a.PoisonDamage += addon.PoisonDamage
	a.PoisonDefense += addon.PoisonDefense
	a.PoisonBonus += addon.PoisonBonus
	a.PoisonResist += addon.PoisonResist
	a.HitRate += addon.HitRate
	a.DodgeRate += addon.DodgeRate
	a.AddGSAddon(addon.GetGsAddon_())
}

func (a *AvatarAttr) AddGSAddon(addon uint32) {
	a.AddGsAddon_(addon)
}

func (a *AvatarAttr) AddGSFloatAddon(addon float32) {
	a.AddGsFloatAddon_(addon)
}

func (a *AvatarAttr) AddEquipUpgrade(slot int, lv uint32) {
	info := GetEquipUpgrade(int(lv))
	if info == nil {
		return
	} else {
		/*
			"Weapon":   0,
			"Chest":    1,
			"Necklace": 2,
			"Belt":     3,

			"Ring":     4,
			"Leggings": 5,
			"Bracers":  6,
		*/
		switch slot {
		case 0:
			a.AddBase(info.GetWeapon_ATT(), info.GetWeapon_DEF(), info.GetWeapon_HP())
			break
		case 1:
			a.AddBase(info.GetChest_ATT(), info.GetChest_DEF(), info.GetChest_HP())
			break
		case 2:
			a.AddBase(info.GetNecklace_ATT(), info.GetNecklace_DEF(), info.GetNecklace_HP())
			break
		case 3:
			a.AddBase(info.GetBelt_ATT(), info.GetBelt_DEF(), info.GetBelt_HP())
			break
		case 4:
			a.AddBase(info.GetRing_ATT(), info.GetRing_DEF(), info.GetRing_HP())
			break
		case 5:
			a.AddBase(info.GetLeggings_ATT(), info.GetLeggings_DEF(), info.GetLeggings_HP())
			break
		case 6:
			a.AddBase(info.GetBracers_ATT(), info.GetBracers_DEF(), info.GetBracers_HP())
			break
		}
	}
}

func (a *AvatarAttr) AddBase(atk, def, hp float32) {
	a.ATK += (atk)
	a.DEF += (def)
	a.HP += (hp)
	//logs.Trace("[GS]AddBase %v,%v,%v to %v,%v,%v", atk, def, hp, a.ATK, a.DEF, a.HP)
}

func (a *AvatarAttr) AddOther(o *AvatarAttr) {
	a.ATK += o.ATK
	a.DEF += o.DEF
	a.HP += o.HP
	a.CritRate += o.CritRate
	a.ResilienceRate += o.ResilienceRate
	a.CritValue += o.CritValue
	a.ResilienceValue += o.ResilienceValue
	a.IceDamage += o.IceDamage
	a.IceDefense += o.IceDefense
	a.IceBonus += o.IceBonus
	a.IceResist += o.IceResist
	a.FireDamage += o.FireDamage
	a.FireDefense += o.FireDefense
	a.FireBonus += o.FireBonus
	a.FireResist += o.FireResist
	a.LightingDamage += o.LightingDamage
	a.LightingDefense += o.LightingDefense
	a.LightingBonus += o.LightingBonus
	a.LightingResist += o.LightingResist
	a.PoisonDamage += o.PoisonDamage
	a.PoisonDefense += o.PoisonDefense
	a.PoisonBonus += o.PoisonBonus
	a.PoisonResist += o.PoisonResist
	a.HitRate += o.HitRate
	a.DodgeRate += o.DodgeRate

	a.AddGsAddon_(o.GetGsAddon_())
	a.AddGsFloatAddon_(o.GetGsFloatAddon_())
	//logs.Trace("[GS]AddOther %v,%v,%v to %v,%v,%v", o.ATK, o.DEF, o.HP, a.ATK, a.DEF, a.HP)
}

func (a *AvatarAttr) Subtract(b *AvatarAttr, c *AvatarAttr) {
	a.ATK += b.ATK - c.ATK
	a.DEF += b.DEF - c.DEF
	a.HP += b.HP - c.HP
	a.CritRate += b.CritRate - c.CritRate
	a.ResilienceRate += b.ResilienceRate - c.ResilienceRate
	a.CritValue += b.CritValue - c.CritValue
	a.ResilienceValue += b.ResilienceValue - c.ResilienceValue
	a.IceDamage += b.IceDamage - c.IceDamage
	a.IceDefense += b.IceDefense - c.IceDefense
	a.IceBonus += b.IceBonus - c.IceBonus
	a.IceResist += b.IceResist - c.IceResist
	a.FireDamage += b.FireDamage - c.FireDamage
	a.FireDefense += b.FireDefense - c.FireDefense
	a.FireBonus += b.FireBonus - c.FireBonus
	a.FireResist += b.FireResist - c.FireResist
	a.LightingDamage += b.LightingDamage - c.LightingDamage
	a.LightingDefense += b.LightingDefense - c.LightingDefense
	a.LightingBonus += b.LightingBonus - c.LightingBonus
	a.LightingResist += b.LightingResist - c.LightingResist
	a.PoisonDamage += b.PoisonDamage - c.PoisonDamage
	a.PoisonDefense += b.PoisonDefense - c.PoisonDefense
	a.PoisonBonus += b.PoisonBonus - c.PoisonBonus
	a.PoisonResist += b.PoisonResist - c.PoisonResist
	a.HitRate += b.HitRate - c.HitRate
	a.DodgeRate += b.DodgeRate - c.DodgeRate
	a.AddGsAddon_(b.GetGsAddon_() - c.GetGsAddon_())
	a.AddGsFloatAddon_(b.GetGsFloatAddon_() - c.GetGsFloatAddon_())
}

func (a *AvatarAttr) AddOnlyNotImportant(addon *AvatarAttr) {
	a.IceDamage += addon.IceDamage
	a.IceDefense += addon.IceDefense
	a.IceBonus += addon.IceBonus
	a.IceResist += addon.IceResist
	a.FireDamage += addon.FireDamage
	a.FireDefense += addon.FireDefense
	a.FireBonus += addon.FireBonus
	a.FireResist += addon.FireResist
	a.LightingDamage += addon.LightingDamage
	a.LightingDefense += addon.LightingDefense
	a.LightingBonus += addon.LightingBonus
	a.LightingResist += addon.LightingResist
	a.PoisonDamage += addon.PoisonDamage
	a.PoisonDefense += addon.PoisonDefense
	a.PoisonBonus += addon.PoisonBonus
	a.PoisonResist += addon.PoisonResist
	a.AddGsAddon_(addon.GetGsAddon_())
	a.AddGsFloatAddon_(addon.GetGsFloatAddon_())
}

func (a *AvatarAttr) AddEquipEvolution(slot int, lv uint32) {
	info := GetEquipEvolution(slot, int(lv))
	if info == nil {
		return
	} else {
		a.AddBase(info.GetATK(), info.GetDEF(), info.GetHP())
	}
}

func (a *AvatarAttr) AddEquipMatEnhance(slot int, lv uint32, mat []bool) {
	cfg := GetEquipMatEnhCfg(slot, lv)
	if cfg != nil {
		a.AddBase(cfg.GetTotalATK(), cfg.GetTotalDEF(), cfg.GetTotalHP())
	}
	cfg = GetEquipMatEnhCfg(slot, lv+1)
	if cfg != nil {
		ms := cfg.GetMaterials_Table()
		for i, b := range mat {
			if b {
				a.AddBase(ms[i].GetATK(), ms[i].GetDEF(), ms[i].GetHP())
			}
		}
	}
}

func (a *AvatarAttr) AddEquipStar(slot int, eqGS float32, lv uint32) {
	info := GetEquipStarData(lv)
	if info == nil {
		return
	} else {
		a.AddGSFloatAddon(eqGS * info.GetAddition())
	}
}

func (a *AvatarAttr) GS() float32 {
	radio := GetPlayerGSRadio()

	var gs float32
	gs += float32(a.ATK) * radio.GetATKRadio()
	gs += float32(a.DEF) * radio.GetDEFRadio()
	gs += float32(a.HP) * radio.GetHPRadio()

	gs += a.CritRate * radio.GetCritRateRadio()
	gs += a.ResilienceRate * radio.GetResilienceRateRadio()
	gs += a.CritValue * radio.GetCritValueRadio()
	gs += a.ResilienceValue * radio.GetResilienceValueRadio()

	gs += float32(a.IceDamage) * radio.GetIceDamageRadio()
	gs += float32(a.IceDefense) * radio.GetIceDefenseRadio()

	gs += a.IceBonus * radio.GetIceBonusRadio()
	gs += a.IceResist * radio.GetIceResistRadio()

	gs += float32(a.FireDamage) * radio.GetFireDamageRadio()
	gs += float32(a.FireDefense) * radio.GetFireDefenseRadio()

	gs += a.FireBonus * radio.GetFireBonusRadio()
	gs += a.FireResist * radio.GetFireResistRadio()

	gs += float32(a.LightingDamage) * radio.GetLightingDamageRadio()
	gs += float32(a.LightingDefense) * radio.GetLightingDefenseRadio()

	gs += a.LightingBonus * radio.GetLightingBonusRadio()
	gs += a.LightingResist * radio.GetLightingResistRadio()

	gs += float32(a.PoisonDamage) * radio.GetPoisonDamageRadio()
	gs += float32(a.PoisonDefense) * radio.GetPoisonDefenseRadio()

	gs += a.PoisonBonus * radio.GetPoisonBonusRadio()
	gs += a.PoisonResist * radio.GetPoisonResistRadio()

	gs += a.HitRate * radio.GetHitRateRadio()
	gs += a.DodgeRate * radio.GetDodgeRateRadio()

	if float32(a.GetGsAddon_()) >= 0 {
		gs += float32(a.GetGsAddon_())
	}

	if a.GetGsFloatAddon_() >= 0 {
		gs += a.GetGsFloatAddon_()
	}
	res := gs - baseAttrGs
	return res
}

func (a *AvatarAttr) GS_2(starCfg *ProtobufGen.HEROSTAR) float32 {
	radio := GetPlayerGSRadio()

	var gs float32
	gs += float32(a.ATK) * radio.GetATKRadio()
	gs += float32(a.DEF) * radio.GetDEFRadio()
	gs += float32(a.HP) * radio.GetHPRadio()

	gs += (a.CritRate - starCfg.GetCritRate()) * radio.GetCritRateRadio()
	gs += (a.ResilienceRate - starCfg.GetResilienceRate()) * radio.GetResilienceRateRadio()
	gs += (a.CritValue - starCfg.GetCritValue()) * radio.GetCritValueRadio()
	gs += (a.ResilienceValue - starCfg.GetResilienceValue()) * radio.GetResilienceValueRadio()

	gs += float32(a.IceDamage) * radio.GetIceDamageRadio()
	gs += float32(a.IceDefense) * radio.GetIceDefenseRadio()

	gs += a.IceBonus * radio.GetIceBonusRadio()
	gs += a.IceResist * radio.GetIceResistRadio()

	gs += float32(a.FireDamage) * radio.GetFireDamageRadio()
	gs += float32(a.FireDefense) * radio.GetFireDefenseRadio()

	gs += a.FireBonus * radio.GetFireBonusRadio()
	gs += a.FireResist * radio.GetFireResistRadio()

	gs += float32(a.LightingDamage) * radio.GetLightingDamageRadio()
	gs += float32(a.LightingDefense) * radio.GetLightingDefenseRadio()

	gs += a.LightingBonus * radio.GetLightingBonusRadio()
	gs += a.LightingResist * radio.GetLightingResistRadio()

	gs += float32(a.PoisonDamage) * radio.GetPoisonDamageRadio()
	gs += float32(a.PoisonDefense) * radio.GetPoisonDefenseRadio()

	gs += a.PoisonBonus * radio.GetPoisonBonusRadio()
	gs += a.PoisonResist * radio.GetPoisonResistRadio()

	gs += a.HitRate * radio.GetHitRateRadio()
	gs += a.DodgeRate * radio.GetDodgeRateRadio()

	if float32(a.GetGsAddon_()) >= 0 {
		gs += float32(a.GetGsAddon_())
	}

	if a.GetGsFloatAddon_() >= 0 {
		gs += a.GetGsFloatAddon_()
	}
	res := gs - baseAttrGs
	return res
}

func (a *AvatarAttr) GS_Int(acid string, avatar_id int, showTyp string) int {
	fGS := a.GS()
	if acid != "" {
		logs.Debug("%s GetCurrAttr %s %d %f", acid, showTyp, avatar_id, fGS)
	}
	return int(fGS + 0.5)
}

// 新版GS计算方法, 暴击率， 暴击伤害，暴击减免，暴击伤害（属性值-对应武将初始属性值）*GS系数
func (a *AvatarAttr) GS_Int_2(starCfg *ProtobufGen.HEROSTAR, acid string, avatar_id int, showTyp string) int {
	fGS := a.GS_2(starCfg)
	if acid != "" {
		logs.Debug("%s GetCurrAttr %s %d %f", acid, showTyp, avatar_id, fGS)
	}
	return int(fGS + 0.5)
}

func (a *AvatarAttr) GS_Int_NoLog() int {
	fGS := a.GS()
	return int(fGS + 0.5)
}

// 获取角色的基本属性值
func GetBaseAvatarAttr(acid string, corp_lv uint32, subModuleGs []int) (res AvatarAttr) {
	base_values := GetPlayerBasicAtt()
	lv_attrs := GetPlayerLevelAttr(corp_lv)

	if lv_attrs == nil {
		logs.Error("GetBaseAvatarAttr Err by corp_lv %d ", corp_lv)
		return
	}

	// 1. 其余基本属性在ATTRIBUTES中, 不受等级影响
	res.CritRate = base_values.GetCritRate()
	res.ResilienceRate = base_values.GetResilienceRate()
	res.CritValue = base_values.GetCritValue()
	res.ResilienceValue = base_values.GetResilienceValue()
	res.IceDamage = base_values.GetIceDamage()
	res.IceDefense = base_values.GetIceDefense()
	res.IceBonus = base_values.GetIceBonus()
	res.IceResist = base_values.GetIceResist()
	res.FireDamage = base_values.GetFireDamage()
	res.FireDefense = base_values.GetFireDefense()
	res.FireBonus = base_values.GetFireBonus()
	res.FireResist = base_values.GetFireResist()
	res.LightingDamage = base_values.GetLightingDamage()
	res.LightingDefense = base_values.GetLightingDefense()
	res.LightingBonus = base_values.GetLightingBonus()
	res.LightingResist = base_values.GetLightingResist()
	res.PoisonDamage = base_values.GetPoisonDamage()
	res.PoisonDefense = base_values.GetPoisonDefense()
	res.PoisonBonus = base_values.GetPoisonBonus()
	res.PoisonResist = base_values.GetPoisonResist()
	res.HitRate = base_values.GetHitRate()
	res.DodgeRate = base_values.GetDodgeRate()

	// 战队的攻防血
	res.ATK += float32(lv_attrs.GetAttack())
	res.DEF += float32(lv_attrs.GetDefense())
	res.HP += float32(lv_attrs.GetHP())

	subModuleGs[helper.Gs_Module_CorpLvl] = int(res.GS())

	logs.Trace("%s Add GS By CorpLvl&Arousal %v %v", acid,
		subModuleGs[helper.Gs_Module_CorpLvl], res)

	return
}

func (a *AvatarAttr) String() string {
	return fmt.Sprintf("%f %f %f %f %f %f %f", a.ATK,
		a.DEF,
		a.HP,
		a.CritRate,
		a.ResilienceRate,
		a.CritValue,
		a.ResilienceValue)
}

func (a *AvatarAttr) GetCompressAttr(diff uint32, ADHGS uint32) AvatarAttr {
	compressAvatar := *a
	compressAvatar.HP = TBCompressAttr(diff, a.HP, ADHGS)
	compressAvatar.ATK = TBCompressAttr(diff, a.ATK, ADHGS)
	compressAvatar.DEF = TBCompressAttr(diff, a.DEF, ADHGS)
	return compressAvatar
}

//压缩属性方法
/*
	玩家该主将属性*（难度限定战力值/该主将战力值）=比值属性（记为M）
	压缩后属性值N=（该主将战力值/难度限定战力值）开四次方*M
	N就是玩家在这场战斗的最终属性。
*/
func TBCompressAttr(diff uint32, attr float32, ADHGS uint32) float32 {
	dataCfg := GetTBossMainDataByDiff(diff)
	if ADHGS <= dataCfg.GetHeroGSLimit() {
		return attr
	}
	M := attr * float32(dataCfg.GetHeroGSLimit()) / float32(ADHGS)
	N := math.Sqrt(math.Sqrt(float64(ADHGS)/float64(dataCfg.GetHeroGSLimit()))) * float64(M)
	return float32(N)
}

/*
	组队boss用的方法：只算攻防血的战力
*/
func (a *AvatarAttr) GS_OnlyATKDEF(starCfg *ProtobufGen.HEROSTAR) float32 {
	radio := GetPlayerGSRadio()
	var gs float32
	gs += float32(a.ATK) * radio.GetATKRadio()
	gs += float32(a.DEF) * radio.GetDEFRadio()
	gs += float32(a.HP) * radio.GetHPRadio()
	if float32(a.GetGsAddon_()) >= 0 {
		gs += float32(a.GetGsAddon_())
	}
	if a.GetGsFloatAddon_() >= 0 {
		gs += a.GetGsFloatAddon_()
	}
	res := gs - baseAttrGs
	return res
}

/*
	根据index选择属性值
*/
func (a *AvatarAttr) ATTR_choice(index uint32) float32 {
	switch index {
	case ATK:
		return a.ATK
	case DEF:
		return a.DEF
	case HP:
		return a.HP
	case CRI_RATE:
		return a.CritRate
	case DE_CRI_RATE:
		return a.ResilienceRate
	case CRI_DAMAGE:
		return a.CritValue
	case DE_CRI_DAMAGE:
		return a.ResilienceValue
	default:
		return 0
	}
}
