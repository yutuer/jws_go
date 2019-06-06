package gs

import (
	"sort"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/battlearmy"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/general"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type DataToCalculateGS interface {
	GetAcID() string

	GetCorpLv() uint32
	GetUnlockedAvatar() []int
	GetArousalLv(t int) uint32

	CurrEquipInfo() ([]uint32, []uint32, []uint32, []uint32, [][]bool)
	GetItem(id uint32) *bag.BagItem
	GetPost() (string, int64)
	GetProfileNowTime() int64

	GetAllGeneral() []general.General
	GetAllGeneralRel() []general.Relation

	GetSkillPractices() []uint32

	GetEquipJades() []uint32
	GetDestinyGeneralJades() []uint32
	GetJadeData(id uint32) (*ProtobufGen.Item, bool)

	GetLastGeneralGiveGs() *gamedata.DestinyGeneralLevelData
	GetFashionAll() []helper.FashionItem

	GetHeroStarLv() []uint32
	GetHeroLv() []uint32
	GetHeroTalent(avatarId int) []uint32
	GetHeroTalentPointCost(avatarId int) uint32
	GetHeroSoulLv() uint32

	GetTitle() []string

	GetSwingLv(avatarId int) int

	GetSwingStarLv(avatarId int) int

	GetCompanionActiveData(avatarId int) []*gamedata.CompanionActiveConfig
	GetCompanionEvolveData(avatarId int) []*gamedata.CompanionEvolveConfig

	GetExclusiveWeaponData(avatarId int) (int, []float32)

	GetHeroDestinyData() ([]int, []int)

	GetHeroAstrologySouls(avatarId int) map[uint32]float32

	GetMagicPet(avatarID int) map[uint32]float32

	GetBattleArmyData(avatarID int) *battlearmy.BattleArmy
}

func GetCurrAttr(a DataToCalculateGS) (
	heroBaseAttrs map[int]*gamedata.AvatarAttr,
	heroAttrs map[int]*gamedata.AvatarAttr,
	mheroBaseGs map[int]int, mheroGs map[int]int, bestHero []int, corpGs int, sortGs heroGss) {

	subModuleGs := make([]int, helper.Gs_Module_Count)
	// 一、计算战队属性

	// 计算Avatar当前的GS
	corp_lvl := a.GetCorpLv()

	// 1）常规属性：
	//    攻、防、血、暴击、爆伤、免爆、免爆伤、元素攻、元素防
	//    会为每个属性设置一个战力参数，
	//    由战队各个激活的角色总属性中的各项属与其对应的战力参数相乘，相加得到常规属性部分的战力
	attr := gamedata.GetBaseAvatarAttr(a.GetAcID(), corp_lvl, subModuleGs)

	// 装备攻防血
	// 根据参数求战力

	// 2）随机属性
	//   区别随机属性类型，
	//   0为常规类型，在计算总GS时候不考虑他们，因为总属性已经体现。
	//   1为特殊类型，在通过总属性计算GS之后需要加上身上所有的此类型随机属性的GS
	GsEquip(a, &attr, subModuleGs)

	// 3）副将
	// 带来的出场释放技能 给一个固定的数值 在最后求和
	GsGeneral(a, &attr, subModuleGs)

	// 4）技能战力
	// 每个技能分配一个对应的战斗力，角色全部技能战力求和。
	GsSkill(a, &attr)

	// 5) 龙玉（宝石）
	// 角色和神将身上的宝石的战力求和
	GsJade(a, &attr, subModuleGs)

	// 6) 时装
	// 角色时装的战力求和
	// 时装有过期，目前在sync会先refresh时装，保证gs的正确，其他地方需要单独处理
	//	GsFashion(a, &attr, subModuleGs)

	// 7) 神将
	GsDestinyGenerals(a, &attr, subModuleGs)

	// 8) 称号
	GsTitle(a, &attr, subModuleGs)

	// 9) 武魂
	GsHeroSoul(a, &attr, subModuleGs)

	logs.Debug("%s GetCurrAttr corpattr %v", a.GetAcID(), attr)

	// 二、算各个主将的属性
	hgs := make(heroGss, 0, gamedata.AVATAR_NUM_MAX)

	heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs = getEveryAvatarGS(a, heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs, attr)

	bestHero = make([]int, 0, helper.CorpHeroGsNum)
	for i := 0; i < helper.CorpHeroGsNum && i < len(hgs); i++ {
		corpGs += hgs[i].Gs
		bestHero = append(bestHero, hgs[i].HeroId)
	}
	return heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, bestHero, corpGs, hgs
}
func getEveryAvatarGS(
	a DataToCalculateGS,
	heroBaseAttrs,
	heroAttrs map[int]*gamedata.AvatarAttr,
	mheroBaseGs map[int]int,
	mheroGs map[int]int,
	hgs heroGss,
	attr gamedata.AvatarAttr) (
	map[int]*gamedata.AvatarAttr,
	map[int]*gamedata.AvatarAttr,
	map[int]int,
	map[int]int,
	heroGss) {

	heroBaseAttrs = make(map[int]*gamedata.AvatarAttr, gamedata.AVATAR_NUM_MAX)
	heroAttrs = make(map[int]*gamedata.AvatarAttr, gamedata.AVATAR_NUM_MAX)
	mheroGs = make(map[int]int, gamedata.AVATAR_NUM_MAX)
	mheroBaseGs = make(map[int]int, gamedata.AVATAR_NUM_MAX)
	talentGS := gamedata.GetHeroCommonConfig().GetHSPointGS()

	//每个武将用于计算战阵加成的战力值
	heroBattleArmyGs := make(map[int]*gamedata.AvatarAttr, gamedata.AVATAR_NUM_MAX)
	cost_talent := make(map[int]uint32)
	startCfg := make(map[int]*ProtobufGen.HEROSTAR)
	for _, avatar_id := range a.GetUnlockedAvatar() {
		cost_talent[avatar_id] = a.GetHeroTalentPointCost(avatar_id)
		battleArmyAttr, base, attr, cfg := heroAttr(a, avatar_id, &attr)
		startCfg[avatar_id] = cfg
		heroBattleArmyGs[avatar_id] = battleArmyAttr
		heroBaseAttrs[avatar_id] = base
		heroAttrs[avatar_id] = attr
		//武魂在这里先算好，战阵对武魂没有影响
		mheroBaseGs[avatar_id] = base.GS_Int_2(startCfg[avatar_id], a.GetAcID(), avatar_id, "herobasegs") +
			int(cost_talent[avatar_id]*talentGS)
	}
	//为每个解锁武将添加战阵提供的属性加成
	for _, avatar_id := range a.GetUnlockedAvatar() {
		country := gamedata.GetHeroCountry(avatar_id)
		//战阵加成数值
		battleArmyAttr := make(map[uint32]float32)
		for i, v := range a.GetBattleArmyData(avatar_id).GetBattleArmyLocs() {
			battleArmyLev := gamedata.GetBattleArmyLevel(uint32(i), uint32(v.Lev))
			battleArmy := gamedata.GetBattleArmyByStruct(uint32(country), uint32(i))
			_value := battleArmyLev.GetValue()
			battleArmyAttr[battleArmy.GetArmyType()] = _value * heroBattleArmyGs[avatar_id].ATTR_choice(battleArmy.GetArmyType())
		}
		heroAttrs[avatar_id].AddAttrs(battleArmyAttr)
		mheroGs[avatar_id] = heroAttrs[avatar_id].GS_Int_2(startCfg[avatar_id], a.GetAcID(), avatar_id, "herogs") +
			int(cost_talent[avatar_id]*talentGS)
		hgs = append(hgs, heroGs{
			HeroId: avatar_id,
			Gs:     mheroGs[avatar_id],
		})
	}
	sort.Sort(hgs)
	logs.Trace("[cyt]heroBaseAttrs:%v, heroAttrs:%v, mheroBaseGs:%v, mheroGs:%v, hgs:%v",heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs)
	return heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs
}

func GetBestHeroInfo(bestHero []int, heroBaseGs map[int]int, heroGs map[int]int) (
	gsHeroGs []int, gsHeroBaseGs []int) {
	gsHeroGs = make([]int, 0, 3)
	gsHeroBaseGs = make([]int, 0, 3)
	for _, h := range bestHero {
		gsHeroGs = append(gsHeroGs, heroGs[h])
		gsHeroBaseGs = append(gsHeroBaseGs, heroBaseGs[h])
	}
	return
}

type heroGs struct {
	HeroId int
	Gs     int
}
type heroGss []heroGs

func (hgss heroGss) Len() int { return len(hgss) }

func (hgss heroGss) Less(i, j int) bool {
	return hgss[i].Gs > hgss[j].Gs
}

func (hgss heroGss) Swap(i, j int) {
	hgss[i], hgss[j] = hgss[j], hgss[i]
}

func GsEquip(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) gamedata.AvatarAttr {
	if attr == nil {
		return gamedata.AvatarAttr{}
	}
	gsAttr := gamedata.AvatarAttr{}
	eqs, lv_evo, lv_star, lv_mat_enh, mat_enh := a.CurrEquipInfo()

	var equip, eqTrick, eqStar, eqEvol, eqME gamedata.AvatarAttr
	var eqBT, eqBTS gamedata.AvatarAttr
	for idx, equip_id := range eqs {
		item := a.GetItem(equip_id)
		if item == nil {
			continue
		}
		var eqSelf gamedata.AvatarAttr
		if !gsEquipItem(item, a, &eqSelf) {
			continue
		}
		gsEquipItemTrick(item, &eqSelf)
		eqBT.AddOther(&eqSelf)
		gsEquipItemStar(lv_star[idx], &eqSelf)
		eqBTS.AddOther(&eqSelf)
		// 再算一次，主要因为gs要分模块统计，而且star是比例加成的
		gsEquipItem(item, a, &equip)
		gsEquipItemTrick(item, &eqTrick)

		eqEvol.AddEquipEvolution(idx, lv_evo[idx])
		eqME.AddEquipMatEnhance(idx, lv_mat_enh[idx], mat_enh[idx])
	}
	// 算star的加成
	eqStar.Subtract(&eqBTS, &eqBT)

	subModuleGs[helper.Gs_Module_Equip] = int(equip.GS())
	subModuleGs[helper.Gs_Module_Equip_Trick] = int(eqTrick.GS())
	subModuleGs[helper.Gs_Module_Equip_StarUp] = int(eqStar.GS())
	subModuleGs[helper.Gs_Module_Equip_Evolution] = int(eqEvol.GS())
	subModuleGs[helper.Gs_Module_Equip_Mat_Enhance] = int(eqME.GS())

	attr.AddOther(&eqBTS)
	attr.AddOther(&eqEvol)
	attr.AddOther(&eqME)

	logs.Trace("%s Add Gs By equip %v %v %v %v %v %v %v %v %v", a.GetAcID(),
		subModuleGs[helper.Gs_Module_Equip],
		subModuleGs[helper.Gs_Module_Equip_Trick],
		subModuleGs[helper.Gs_Module_Equip_StarUp],
		subModuleGs[helper.Gs_Module_Equip_Evolution],
		subModuleGs[helper.Gs_Module_Equip_Mat_Enhance],
		eqBTS, eqEvol, eqME, *attr)

	gsAttr.AddOther(&eqBTS)
	gsAttr.AddOther(&eqEvol)
	gsAttr.AddOther(&eqME)
	return gsAttr
}

func gsEquipItem(item *bag.BagItem, a DataToCalculateGS, attr *gamedata.AvatarAttr) bool {
	item_data, data_ok := gamedata.GetProtoItem(item.TableID)
	if !data_ok {
		logs.Warn("GsEquipItem Data No Found by %v", item.ID)
		return false
	}

	if a != nil {
		rankLimit := item_data.GetRankLimit()
		playerRank, rankOverTime := a.GetPost()
		nowT := a.GetProfileNowTime()
		if rankLimit != "" && rankLimit != playerRank {
			// 官阶不符 跳过
			return false
		}

		if rankLimit != "" && rankOverTime < nowT {
			return false
		}
	}

	attr.AddBase(item_data.GetAttack(), item_data.GetDefense(), item_data.GetHP())
	return true
}

func gsEquipItemTrick(item *bag.BagItem, attr *gamedata.AvatarAttr) {
	data := item.GetItemData()

	if data == nil {
		logs.Warn("GetItemData Data No Found by %v", item)
		return
	}

	for _, tr := range data.TrickGroup {
		if tr != "" {
			aa := gamedata.GetTrickDetailAttrAddon(tr)
			if aa != nil {
				attr.Add(aa)
			}
		}
	}
}

func gsEquipItemStar(star uint32, attr *gamedata.AvatarAttr) {
	var addRate float32 = 0
	starInfo := gamedata.GetEquipStarData(star)
	if starInfo == nil {
		addRate = 0.0
	} else {
		addRate = starInfo.GetAddition()
	}

	attr.AddEquipStarAddon(addRate)
}

func GsGeneral(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) {
	var genAttr, relAttr gamedata.AvatarAttr
	for _, gen := range a.GetAllGeneral() {
		atk, def, hp := gamedata.GetGeneralStarAttr(gen.Id, gen.StarLv)
		genAttr.AddBase(atk, def, hp)
	}
	for _, rel := range a.GetAllGeneralRel() {
		atk, def, hp := gamedata.GetGeneralRelLvlAttr(rel.Id, rel.Level)
		relAttr.AddBase(atk, def, hp)
	}
	subModuleGs[helper.Gs_Module_General] = int(genAttr.GS())
	subModuleGs[helper.Gs_Module_General_Rel] = int(relAttr.GS())
	attr.AddOther(&genAttr)
	attr.AddOther(&relAttr)

	logs.Trace("%s Add Gs By General %v %v %v %v %v", a.GetAcID(),
		subModuleGs[helper.Gs_Module_General],
		subModuleGs[helper.Gs_Module_General_Rel], genAttr, relAttr, *attr)

	return
}

func GsSkill(a DataToCalculateGS, attr *gamedata.AvatarAttr) {
	skill_data := a.GetSkillPractices()
	var skillAttr gamedata.AvatarAttr

	for idx, lv := range skill_data {
		skill_cfg := gamedata.GetSkillPracticeLevelInfo(idx)
		if skill_cfg == nil {
			logs.Warn("Skill Cfg Nil by %d %d", idx)
			continue
		}
		if int(lv) < len(skill_cfg.ATK) && int(lv) < len(skill_cfg.DEF) && int(lv) < len(skill_cfg.HP) {
			skillAttr.AddBase(
				float32(skill_cfg.ATK[lv]),
				float32(skill_cfg.DEF[lv]),
				float32(skill_cfg.HP[lv]))
		}
	}
	attr.AddOther(&skillAttr)

	logs.Trace("%s Add Gs By Skill %v %v", a.GetAcID(), skillAttr, *attr)
	return
}

func GsJade(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) gamedata.AvatarAttr {
	var jadeAttr gamedata.AvatarAttr
	for _, jade := range a.GetEquipJades() {
		if jade > 0 {
			itemcfg, _ := a.GetJadeData(jade)
			jadeAttr.AddBase(itemcfg.GetAttack(), itemcfg.GetDefense(), itemcfg.GetHP())
		}
	}
	for _, jade := range a.GetDestinyGeneralJades() {
		if jade > 0 {
			itemcfg, _ := a.GetJadeData(jade)
			jadeAttr.AddBase(itemcfg.GetAttack(), itemcfg.GetDefense(), itemcfg.GetHP())
		}
	}
	subModuleGs[helper.Gs_Module_Jade] = int(jadeAttr.GS())
	attr.AddOther(&jadeAttr)
	logs.Trace("%s Add Gs By Jade %v %v %v", a.GetAcID(), subModuleGs[helper.Gs_Module_Jade],
		jadeAttr, *attr)
	return jadeAttr
}

//func GsFashion(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) {
//	var fashionAttr gamedata.AvatarAttr
//	for _, f := range a.GetFashionAll() {
//		_, cfg := gamedata.IsFashion(f.TableID)
//		fashionAttr.AddBase(cfg.GetAttack(), cfg.GetDefense(), cfg.GetHP())
//	}
//	subModuleGs[helper.Gs_Module_Fashion] = int(fashionAttr.GS())
//	attr.AddOther(&fashionAttr)
//
//	logs.Trace("%s Add Gs By Fashion %v %v %v", a.GetAcID(), subModuleGs[helper.Gs_Module_Fashion],
//		fashionAttr, *attr)
//}

func GsDestinyGenerals(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) gamedata.AvatarAttr {
	var dgAttr gamedata.AvatarAttr
	data := a.GetLastGeneralGiveGs()
	if data != nil {
		dgAttr.AddBase(data.Atk, data.Def, data.Hp)
	}
	subModuleGs[helper.Gs_Module_DestinyGeneral] = int(dgAttr.GS())
	attr.AddOther(&dgAttr)
	logs.Trace("%s Add Gs By DestinyGenerals %v %v %v", a.GetAcID(), subModuleGs[helper.Gs_Module_DestinyGeneral],
		dgAttr, *attr)
	return dgAttr
}

func GsHeroLevel(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) {
	heroAttr := &gamedata.AvatarAttr{}
	heroLv := a.GetHeroStarLv()
	for heroIdx, lv := range heroLv {
		data := gamedata.GetHeroData(heroIdx)
		if lv > 0 && data != nil && len(data.LvData) >= int(lv) {
			lvData := data.LvData[int(lv)]
			heroAttr.AddBase(
				lvData.ATK,
				lvData.DEF,
				lvData.HP)
		}
	}
	subModuleGs[helper.Gs_Module_Hero] = int(heroAttr.GS())
	attr.AddOther(heroAttr)
	logs.Trace("%s Add Gs By Hero %v %v %v", a.GetAcID(), subModuleGs[helper.Gs_Module_Hero], heroAttr, *attr)
}

func GsTitle(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) {
	titleAttr := &gamedata.AvatarAttr{}
	for _, title := range a.GetTitle() {
		cfg := gamedata.GetTitleCfg(title)
		if cfg != nil {
			titleAttr.AddBase(cfg.GetAttack(), cfg.GetDefense(), cfg.GetHP())
		}
	}
	subModuleGs[helper.Gs_Module_Title] = int(titleAttr.GS())
	attr.AddOther(titleAttr)
	logs.Trace("%s Add Gs By Title %v %v %v", a.GetAcID(), subModuleGs[helper.Gs_Module_Title], titleAttr, *attr)
}

func GsHeroSoul(a DataToCalculateGS, attr *gamedata.AvatarAttr, subModuleGs []int) {
	lv := a.GetHeroSoulLv()
	if lv > 0 {
		soulAttr := &gamedata.AvatarAttr{}
		cfg := gamedata.GetHeroSoulLvlConfig(lv)
		if cfg != nil {
			soulAttr.AddBase(cfg.GetATK(), cfg.GetDEF(), cfg.GetHP())
		}
		attr.AddOther(soulAttr)
		logs.Trace("%s Add Gs By HeroSoul %v %v ", a.GetAcID(), soulAttr, attr)
	}
}

func heroAttr(a DataToCalculateGS, avatar_id int, corpAttr *gamedata.AvatarAttr) (heroAttrWithOutMagicPetAndAstrologySouls, heroBaseAttr,
	heroAttr *gamedata.AvatarAttr, c *ProtobufGen.HEROSTAR) {

	heroBaseAttr = &gamedata.AvatarAttr{}
	heroAttr = &gamedata.AvatarAttr{}

	_stars := a.GetHeroStarLv()
	_lvls := a.GetHeroLv()
	_talent := a.GetHeroTalent(avatar_id)
	star := _stars[avatar_id]
	lvl := _lvls[avatar_id]
	acfg := gamedata.GetHeroData(avatar_id)
	scfg := acfg.LvData[int(star)]
	c = scfg.Cfg
	comc := gamedata.GetHeroCommonConfig()
	_l := float32(uint32(comc.GetHeroLevelBase()) + lvl)

	force := float32(c.GetForce()) +
		float32(_talent[gamedata.Talent_Idx_Force])*
			gamedata.GetHeroTalentConfig(gamedata.Talent_Idx_Force).GetHeroTalentPara()
	intellect := float32(c.GetIntellect()) +
		float32(_talent[gamedata.Talent_Idx_Intellect])*
			gamedata.GetHeroTalentConfig(gamedata.Talent_Idx_Intellect).GetHeroTalentPara()
	endurance := float32(c.GetEndurance()) +
		float32(_talent[gamedata.Talent_Idx_Endurance])*
			gamedata.GetHeroTalentConfig(gamedata.Talent_Idx_Endurance).GetHeroTalentPara()

	var swingAtk float32
	var swingHP float32
	var swingDef float32
	// 神翼战力加成
	swingLvInfo := gamedata.GetHeroSwingLvUpInfo(a.GetSwingLv(avatar_id))
	if swingLvInfo != nil {
		swingAtk += swingLvInfo.GetATK()
		swingHP += swingLvInfo.GetHP()
		swingDef += swingLvInfo.GetDEF()
	}
	swingStarInfo := gamedata.GetHeroSwingStarLvUpInfo(a.GetSwingStarLv(avatar_id))
	if swingStarInfo != nil {
		swingAtk += swingStarInfo.GetATK()
		swingHP += swingStarInfo.GetHP()
		swingDef += swingStarInfo.GetDEF()
	}

	var fashionAtk float32
	var fashionHP float32
	var fashionDef float32

	aEquips := a.GetFashionAll()
	for _, idData := range aEquips {
		ok, fashionInfo := gamedata.IsFashion(idData.TableID)
		if ok && fashionInfo.GetRoleOnly() == int32(avatar_id) {
			fashionAtk += fashionInfo.GetAttack()
			fashionHP += fashionInfo.GetHP()
			fashionDef += fashionInfo.GetDefense()
		}
	}

	heroBaseAttr.ATK = _l*c.GetATKGrowth() + swingAtk + fashionAtk
	heroAttr.ATK = corpAttr.ATK * force * comc.GetHeorGrowthRate()

	heroBaseAttr.DEF = _l*c.GetDEFGrowth() + swingDef + fashionDef
	heroAttr.DEF = corpAttr.DEF * intellect * comc.GetHeorGrowthRate()

	heroBaseAttr.HP = _l*c.GetHPGrowth() + swingHP + fashionHP
	heroAttr.HP = corpAttr.HP * endurance * comc.GetHeorGrowthRate()

	heroBaseAttr.CritRate = c.GetCritRate()
	heroBaseAttr.CritValue = c.GetCritValue()
	heroBaseAttr.ResilienceRate = c.GetResilienceRate()
	heroBaseAttr.ResilienceValue = c.GetResilienceValue()
	heroBaseAttr.HitRate = c.GetHitRate()
	heroBaseAttr.DodgeRate = c.GetDodgeRate()

	heroAttr.CritRate = corpAttr.CritRate
	heroAttr.CritValue = corpAttr.CritValue
	heroAttr.ResilienceRate = corpAttr.ResilienceRate
	heroAttr.ResilienceValue = corpAttr.ResilienceValue
	heroAttr.HitRate = corpAttr.HitRate
	heroAttr.DodgeRate = corpAttr.DodgeRate

	heroBaseAttr.Force = force
	heroBaseAttr.Intellect = intellect
	heroBaseAttr.Endurance = endurance

	heroAttr.Force = force
	heroAttr.Intellect = intellect
	heroAttr.Endurance = endurance

	// 情缘系统对战斗属性的影响
	heroCompanionAttr := companionAttr(a, avatar_id)
	heroBaseAttr.AddOther(heroCompanionAttr)

	// 神兵系统对战斗属性的影响
	heroBaseAttr.AddOther(exclusiveWeaponAttr(a, avatar_id))

	// 宿命系统对战斗属性的影响
	fateIds, levelIds := a.GetHeroDestinyData()
	heroBaseAttr.AddOther(heroDestinyAttr(fateIds, levelIds, avatar_id))

	_heroBaseAttrWithOutMagicPetAndAstrologySouls := *heroBaseAttr

	heroAttrWithOutMagicPetAndAstrologySouls = &_heroBaseAttrWithOutMagicPetAndAstrologySouls

	// 武将的星图增加属性
	heroBaseAttr.AddAttrs(a.GetHeroAstrologySouls(avatar_id))

	// 武将的灵宠增加属性
	heroBaseAttr.AddAttrs(a.GetMagicPet(avatar_id))

	heroAttr.AddOnlyNotImportant(corpAttr)

	heroAttr.AddOther(heroBaseAttr)

	logs.Debug("%s calcheroAttr corpAttr %d baseWithOutMagicPetAndAstrologySouls %v ,base %v attr %v %f %f %f",
		a.GetAcID(), avatar_id, heroAttrWithOutMagicPetAndAstrologySouls, heroBaseAttr, heroAttr, force, intellect, endurance)

	return
}

func companionAttr(a DataToCalculateGS, avatar_id int) *gamedata.AvatarAttr {
	// 情缘激活
	activeConfigs := a.GetCompanionActiveData(avatar_id)
	retAttr := &gamedata.AvatarAttr{}
	for _, config := range activeConfigs {
		retAttr.AddAttr(config.Config.GetActiveReward(), config.Config.GetActiveRewardValue())
	}
	logs.Debug("calc companion base atrr, %v", retAttr)
	// 情缘进阶
	evolveConfigs := a.GetCompanionEvolveData(avatar_id)
	for _, config := range evolveConfigs {
		for _, reward := range config.Config.GetReward_Table() {
			logs.Debug("calc companion: %d, %f", reward.GetRewardType(), reward.GetRewardValue())
			retAttr.AddAttr(reward.GetRewardType(), reward.GetRewardValue())
		}
	}
	logs.Debug("calc companion evolve atrr, %v", retAttr)
	return retAttr
}

func exclusiveWeaponAttr(a DataToCalculateGS, avatar_id int) *gamedata.AvatarAttr {
	pAttr := &gamedata.AvatarAttr{}
	quality, attrs := a.GetExclusiveWeaponData(avatar_id)
	if quality == 0 {
		return pAttr
	}
	evolveCfg := gamedata.GetEvolveGloryWeaponCfg(quality)
	for i, v := range attrs {
		attrId := i + 1
		attrCfg := gamedata.GetWeaponAttrById(evolveCfg, attrId)
		if attrCfg.GetIsBonusByQuality() == 1 {
			v = v * (1 + evolveCfg.GetDevelopBonus())
		}
		pAttr.AddAttr(uint32(attrId), v)
	}
	logs.Debug("calc excluseive evolve atrr, %d, %v", avatar_id, pAttr)
	return pAttr
}

func heroDestinyAttr(destinyIds []int, levelIds []int, avatarId int) *gamedata.AvatarAttr {
	pAttr := &gamedata.AvatarAttr{}
	for i, destinyId := range destinyIds {
		cfg := gamedata.GetHeroDestinyById(destinyId)
		if cfg.ContainsAvatarId(avatarId) {
			level := levelIds[i]
			levelCfg := gamedata.GetFateLevelConfig(destinyId, level)
			for _, attr := range levelCfg.GetFateAttr_Table() {
				pAttr.AddAttr(attr.GetAttrType(), attr.GetAttrValue())
			}
		}
	}
	return pAttr
}

// corpGs = 9个最强武将的战力
func GetCurrAttrForWspvp(a DataToCalculateGS) (
	heroAttrs map[int]*gamedata.AvatarAttr,
	mheroGs map[int]int, mheroBaseGs map[int]int, bestHero []int, corpGs int,
	extraAttr [3][3]float32) {

	subModuleGs := make([]int, helper.Gs_Module_Count)
	// 一、计算战队属性

	// 计算Avatar当前的GS
	corp_lvl := a.GetCorpLv()

	// 1）常规属性：
	//    攻、防、血、暴击、爆伤、免爆、免爆伤、元素攻、元素防
	//    会为每个属性设置一个战力参数，
	//    由战队各个激活的角色总属性中的各项属与其对应的战力参数相乘，相加得到常规属性部分的战力
	attr := gamedata.GetBaseAvatarAttr(a.GetAcID(), corp_lvl, subModuleGs)

	// 装备攻防血
	// 根据参数求战力

	// 2）随机属性
	//   区别随机属性类型，
	//   0为常规类型，在计算总GS时候不考虑他们，因为总属性已经体现。
	//   1为特殊类型，在通过总属性计算GS之后需要加上身上所有的此类型随机属性的GS
	equipAttr := GsEquip(a, &attr, subModuleGs)

	// 3）副将
	// 带来的出场释放技能 给一个固定的数值 在最后求和
	GsGeneral(a, &attr, subModuleGs)

	// 4）技能战力
	// 每个技能分配一个对应的战斗力，角色全部技能战力求和。
	GsSkill(a, &attr)

	// 5) 龙玉（宝石）
	// 角色和神将身上的宝石的战力求和
	jadeAttr := GsJade(a, &attr, subModuleGs)

	// 6) 时装
	// 角色时装的战力求和
	// 时装有过期，目前在sync会先refresh时装，保证gs的正确，其他地方需要单独处理
	//	GsFashion(a, &attr, subModuleGs)

	// 7) 神将
	destinyAttr := GsDestinyGenerals(a, &attr, subModuleGs)

	// 8) 称号
	GsTitle(a, &attr, subModuleGs)

	// 9) 武魂
	GsHeroSoul(a, &attr, subModuleGs)

	logs.Debug("%s GetCurrAttr corpattr %v", a.GetAcID(), attr)

	// 二、算各个主将的属性
	var heroBaseAttrs map[int]*gamedata.AvatarAttr
	hgs := make(heroGss, 0, gamedata.AVATAR_NUM_MAX)

	heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs = getEveryAvatarGS(a, heroBaseAttrs, heroAttrs, mheroBaseGs, mheroGs, hgs, attr)

	bestHero = make([]int, 0, helper.WspvpBestHeroCount)
	for i := 0; i < helper.WspvpBestHeroCount && i < len(hgs); i++ {
		corpGs += hgs[i].Gs
		bestHero = append(bestHero, hgs[i].HeroId)
	}
	extraAttrArray := []gamedata.AvatarAttr{equipAttr, destinyAttr, jadeAttr}
	for i := 0; i < 3; i++ {
		extraAttr[i][0] = extraAttrArray[i].ATK
		extraAttr[i][1] = extraAttrArray[i].DEF
		extraAttr[i][2] = extraAttrArray[i].HP
	}
	return heroAttrs, mheroGs, mheroBaseGs, bestHero, corpGs, extraAttr
}

func GetCurrAttrForWB(a DataToCalculateGS) (extraAttr [3][3]float32) {
	subModuleGs := make([]int, helper.Gs_Module_Count)
	corp_lvl := a.GetCorpLv()
	attr := gamedata.GetBaseAvatarAttr(a.GetAcID(), corp_lvl, subModuleGs)
	equipAttr := GsEquip(a, &attr, subModuleGs)
	jadeAttr := GsJade(a, &attr, subModuleGs)
	destinyAttr := GsDestinyGenerals(a, &attr, subModuleGs)
	extraAttrArray := []gamedata.AvatarAttr{equipAttr, destinyAttr, jadeAttr}
	for i := 0; i < 3; i++ {
		extraAttr[i][0] = extraAttrArray[i].ATK
		extraAttr[i][1] = extraAttrArray[i].DEF
		extraAttr[i][2] = extraAttrArray[i].HP
	}
	return extraAttr
}
