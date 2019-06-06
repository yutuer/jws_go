package gamedata

import (
	"errors"
	"strconv"
	"strings"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type DroidAccountData struct {
	AcID          string
	DroidID       int
	AvatarId      int
	CorpLv        uint32
	Name          string
	AvatarSkills  [helper.AVATAR_SKILL_MAX]uint32
	AvatarLockeds [helper.AVATAR_NUM_CURR]int

	Arousals       [helper.AVATAR_NUM_MAX]uint32
	SkillPractices [helper.CORP_SKILLPRACTICE_MAX]uint32

	EquipsItemID         [helper.PartEquipCount]string
	EquipsItemTrickGroup [helper.PartEquipCount][]string
	EquipUpgrade         [helper.PartEquipCount]uint32
	EquipStar            [helper.PartEquipCount]uint32
	EquipMatEnhLv        [helper.PartEquipCount]uint32
	EquipMatEhn          [helper.PartEquipCount][helper.EQUIP_MAT_ENHANCE_MAT]bool
	Fashions             [helper.AVATAR_NUM_MAX][helper.FashionPart_Count]string

	Generals        []string
	GeneralsStar    []uint32
	GeneralRels     []string
	GeneralRelsStar []uint32

	EquipJade       [helper.EQUIP_SLOT_CURR * helper.JADE_SLOT_MAX]string
	DestGeneralJade []string

	DestinyGeneralID        int
	DestinyGeneralLv        int
	CurrDestinyGeneralSkill [helper.DestinyGeneralSkillMax]int

	Attr   AvatarAttr
	HeroGs int
	CorpGs int
}

func IsAccountIsAnDroid(id string) bool {
	return id == "0:0:DroidID"
}

func (d *DroidAccountData) FromData(data *ProtobufGen.BSCPVPBOT) {
	d.DroidID = int(data.GetBotID())
	d.AcID = fmt.Sprintf("0:0:DroidID")
	d.AvatarId = int(data.GetAID())
	d.CorpLv = data.GetCLv()
	d.Name = data.GetName()
	d.AvatarLockeds = [helper.AVATAR_NUM_CURR]int{0, 1, 2}

	currSkillStr := data.GetCurr_Skills()
	currSkillStrs := strings.Split(currSkillStr, ",")
	for idx, skillStr := range currSkillStrs {
		skillLv, err := strconv.Atoi(skillStr)
		if err != nil {
			panic(err)
		}
		d.AvatarSkills[idx] = uint32(skillLv)
	}

	for _, avatarData := range data.GetRoleInformation_Template() {
		avatarID := int(avatarData.GetRoleID())
		d.Arousals[avatarID] = avatarData.GetCurr_Arousals()
		d.Fashions[avatarID][helper.FashionPart_Armor] = avatarData.GetFashionArmor()
		d.Fashions[avatarID][helper.FashionPart_Weapon] = avatarData.GetFashionWeapon()
	}

	d.DestinyGeneralID = int(data.GetDG())
	d.DestinyGeneralLv = int(data.GetDGLv())

	for _, equipData := range data.GetEquipDetail_Template() {
		partNum := helper.GetEquipSlot(equipData.GetEquipPart())
		if partNum < 0 {
			panic(errors.New("GetEquipSlot Err By " + equipData.GetEquipPart()))
		}
		d.EquipsItemID[partNum] = equipData.GetCurr_Equips()
		d.EquipsItemTrickGroup[partNum] = []string{
			equipData.GetEquipAttr1(),
			equipData.GetEquipAttr2(),
			equipData.GetEquipAttr3(),
			equipData.GetEquipAttr4(),
			equipData.GetEquipAttr5(),
		}
		d.EquipUpgrade[partNum] = equipData.GetCurr_Upgrades()
		d.EquipStar[partNum] = equipData.GetCurr_Stars()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+0] = equipData.GetJade1()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+1] = equipData.GetJade2()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+2] = equipData.GetJade3()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+3] = equipData.GetJade4()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+4] = equipData.GetJade5()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+5] = equipData.GetJade6()
	}

	dgLen := len(data.GetDGJade_Template())
	d.DestGeneralJade = make([]string, dgLen*helper.JADE_SLOT_MAX, dgLen*helper.JADE_SLOT_MAX)
	for _, dgData := range data.GetDGJade_Template() {
		dgID := int(dgData.GetDGID())
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+0] = dgData.GetJade1()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+1] = dgData.GetJade2()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+2] = dgData.GetJade3()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+3] = dgData.GetJade4()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+4] = dgData.GetJade5()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+5] = dgData.GetJade6()
	}

	d.Generals = make([]string, 0, 16)
	d.GeneralsStar = make([]uint32, 0, 16)
	d.GeneralRels = make([]string, 0, 16)
	d.GeneralRelsStar = make([]uint32, 0, 16)
	for _, generalData := range data.GetCurr_General_Template() {
		d.Generals = append(d.Generals, generalData.GetCurr_Generals())
		d.GeneralsStar = append(d.GeneralsStar, generalData.GetCurr_Genstars())
	}

	for _, generalRelsData := range data.GetCurr_General_Relation_Template() {
		d.GeneralRels = append(d.Generals, generalRelsData.GetGeneral_Relation())
		d.GeneralRelsStar = append(d.GeneralsStar, generalRelsData.GetGeneral_Relation_Star())
	}

	for idx, dgskills := range data.GetDestiny_Generals_Skill_Template() {
		d.CurrDestinyGeneralSkill[idx] = int(dgskills.GetCurr_Destiny_Generals_For_Skill())
	}

	d.Attr = AvatarAttr{}
	d.Attr.ATK = data.GetATK()
	d.Attr.DEF = data.GetDEF()
	d.Attr.HP = data.GetHP()
	d.Attr.CritRate = data.GetCritRate()
	d.Attr.CritValue = data.GetCritValue()
	d.Attr.ResilienceRate = data.GetResilienceRate()
	d.Attr.ResilienceValue = data.GetResilienceValue()
	d.Attr.HitRate = data.GetHitRate()
	d.Attr.DodgeRate = data.GetDodgeRate()

	d.HeroGs = d.Attr.GS_Int_NoLog()
	d.CorpGs = d.HeroGs * 3
}

func (d *DroidAccountData) FromDataExpedition(data *ProtobufGen.EXPEDITIONBOT) {
	d.DroidID = int(data.GetBotID())
	d.AcID = fmt.Sprintf("0:0:DroidID")
	d.AvatarId = int(data.GetAID())
	d.CorpLv = data.GetCLv()
	d.Name = data.GetName()
	d.AvatarLockeds = [helper.AVATAR_NUM_CURR]int{0, 1, 2}

	currSkillStr := data.GetCurr_Skills()
	currSkillStrs := strings.Split(currSkillStr, ",")
	for idx, skillStr := range currSkillStrs {
		skillLv, err := strconv.Atoi(skillStr)
		if err != nil {
			panic(err)
		}
		d.AvatarSkills[idx] = uint32(skillLv)
	}

	for _, avatarData := range data.GetRoleInformation_Template() {
		avatarID := int(avatarData.GetRoleID())
		d.Arousals[avatarID] = avatarData.GetCurr_Arousals()
		d.Fashions[avatarID][helper.FashionPart_Armor] = avatarData.GetFashionArmor()
		d.Fashions[avatarID][helper.FashionPart_Weapon] = avatarData.GetFashionWeapon()
	}

	d.DestinyGeneralID = int(data.GetDG())
	d.DestinyGeneralLv = int(data.GetDGLv())

	for _, equipData := range data.GetEquipDetail_Template() {
		partNum := helper.GetEquipSlot(equipData.GetEquipPart())
		if partNum < 0 {
			panic(errors.New("GetEquipSlot Err By " + equipData.GetEquipPart()))
		}
		d.EquipsItemID[partNum] = equipData.GetCurr_Equips()
		d.EquipsItemTrickGroup[partNum] = []string{
			equipData.GetEquipAttr1(),
			equipData.GetEquipAttr2(),
			equipData.GetEquipAttr3(),
			equipData.GetEquipAttr4(),
			equipData.GetEquipAttr5(),
		}
		d.EquipUpgrade[partNum] = equipData.GetCurr_Upgrades()
		d.EquipStar[partNum] = equipData.GetCurr_Stars()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+0] = equipData.GetJade1()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+1] = equipData.GetJade2()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+2] = equipData.GetJade3()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+3] = equipData.GetJade4()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+4] = equipData.GetJade5()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+5] = equipData.GetJade6()
	}

	dgLen := len(data.GetDGJade_Template())
	d.DestGeneralJade = make([]string, dgLen*helper.JADE_SLOT_MAX, dgLen*helper.JADE_SLOT_MAX)
	for _, dgData := range data.GetDGJade_Template() {
		dgID := int(dgData.GetDGID())
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+0] = dgData.GetJade1()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+1] = dgData.GetJade2()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+2] = dgData.GetJade3()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+3] = dgData.GetJade4()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+4] = dgData.GetJade5()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+5] = dgData.GetJade6()
	}

	d.Generals = make([]string, 0, 16)
	d.GeneralsStar = make([]uint32, 0, 16)
	d.GeneralRels = make([]string, 0, 16)
	d.GeneralRelsStar = make([]uint32, 0, 16)
	for _, generalData := range data.GetCurr_General_Template() {
		d.Generals = append(d.Generals, generalData.GetCurr_Generals())
		d.GeneralsStar = append(d.GeneralsStar, generalData.GetCurr_Genstars())
	}

	for _, generalRelsData := range data.GetCurr_General_Relation_Template() {
		d.GeneralRels = append(d.Generals, generalRelsData.GetGeneral_Relation())
		d.GeneralRelsStar = append(d.GeneralsStar, generalRelsData.GetGeneral_Relation_Star())
	}

	for idx, dgskills := range data.GetDestiny_Generals_Skill_Template() {
		d.CurrDestinyGeneralSkill[idx] = int(dgskills.GetCurr_Destiny_Generals_For_Skill())
	}

	d.Attr = AvatarAttr{}
	d.Attr.ATK = data.GetATK()
	d.Attr.DEF = data.GetDEF()
	d.Attr.HP = data.GetHP()
	d.Attr.CritRate = data.GetCritRate()
	d.Attr.CritValue = data.GetCritValue()
	d.Attr.ResilienceRate = data.GetResilienceRate()
	d.Attr.ResilienceValue = data.GetResilienceValue()
	d.Attr.HitRate = data.GetHitRate()
	d.Attr.DodgeRate = data.GetDodgeRate()

	d.HeroGs = d.Attr.GS_Int_NoLog()
	d.CorpGs = d.HeroGs * 3
}

func (d *DroidAccountData) FromDataGVG(data *ProtobufGen.GVGGUARD) {
	d.DroidID = int(data.GetBotID())
	d.AcID = fmt.Sprintf("0:0:GVGGUARD")
	d.AvatarId = int(data.GetAID())
	d.CorpLv = data.GetCLv()
	d.Name = data.GetName()
	d.AvatarLockeds = [helper.AVATAR_NUM_CURR]int{0, 1, 2}

	currSkillStr := data.GetCurr_Skills()
	currSkillStrs := strings.Split(currSkillStr, ",")
	for idx, skillStr := range currSkillStrs {
		skillLv, err := strconv.Atoi(skillStr)
		if err != nil {
			panic(err)
		}
		d.AvatarSkills[idx] = uint32(skillLv)
	}

	for _, avatarData := range data.GetRoleInformation_Template() {
		avatarID := int(avatarData.GetRoleID())
		d.Arousals[avatarID] = avatarData.GetCurr_Arousals()
		d.Fashions[avatarID][helper.FashionPart_Armor] = avatarData.GetFashionArmor()
		d.Fashions[avatarID][helper.FashionPart_Weapon] = avatarData.GetFashionWeapon()
	}

	d.DestinyGeneralID = int(data.GetDG())
	d.DestinyGeneralLv = int(data.GetDGLv())

	for _, equipData := range data.GetEquipDetail_Template() {
		partNum := helper.GetEquipSlot(equipData.GetEquipPart())
		if partNum < 0 {
			panic(errors.New("GetEquipSlot Err By " + equipData.GetEquipPart()))
		}
		d.EquipsItemID[partNum] = equipData.GetCurr_Equips()
		d.EquipsItemTrickGroup[partNum] = []string{
			equipData.GetEquipAttr1(),
			equipData.GetEquipAttr2(),
			equipData.GetEquipAttr3(),
			equipData.GetEquipAttr4(),
			equipData.GetEquipAttr5(),
		}
		d.EquipUpgrade[partNum] = equipData.GetCurr_Upgrades()
		d.EquipStar[partNum] = equipData.GetCurr_Stars()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+0] = equipData.GetJade1()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+1] = equipData.GetJade2()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+2] = equipData.GetJade3()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+3] = equipData.GetJade4()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+4] = equipData.GetJade5()
		d.EquipJade[partNum*helper.JADE_SLOT_MAX+5] = equipData.GetJade6()
	}

	dgLen := len(data.GetDGJade_Template())
	d.DestGeneralJade = make([]string, dgLen*helper.JADE_SLOT_MAX, dgLen*helper.JADE_SLOT_MAX)
	for _, dgData := range data.GetDGJade_Template() {
		dgID := int(dgData.GetDGID())
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+0] = dgData.GetJade1()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+1] = dgData.GetJade2()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+2] = dgData.GetJade3()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+3] = dgData.GetJade4()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+4] = dgData.GetJade5()
		d.DestGeneralJade[dgID*helper.JADE_SLOT_MAX+5] = dgData.GetJade6()
	}

	d.Generals = make([]string, 0, 16)
	d.GeneralsStar = make([]uint32, 0, 16)
	d.GeneralRels = make([]string, 0, 16)
	d.GeneralRelsStar = make([]uint32, 0, 16)
	for _, generalData := range data.GetCurr_General_Template() {
		d.Generals = append(d.Generals, generalData.GetCurr_Generals())
		d.GeneralsStar = append(d.GeneralsStar, generalData.GetCurr_Genstars())
	}

	for _, generalRelsData := range data.GetCurr_General_Relation_Template() {
		d.GeneralRels = append(d.Generals, generalRelsData.GetGeneral_Relation())
		d.GeneralRelsStar = append(d.GeneralsStar, generalRelsData.GetGeneral_Relation_Star())
	}

	for idx, dgskills := range data.GetDestiny_Generals_Skill_Template() {
		d.CurrDestinyGeneralSkill[idx] = int(dgskills.GetCurr_Destiny_Generals_For_Skill())
	}

	d.Attr = AvatarAttr{}
	d.Attr.ATK = data.GetATK()
	d.Attr.DEF = data.GetDEF()
	d.Attr.HP = data.GetHP()
	d.Attr.CritRate = data.GetCritRate()
	d.Attr.CritValue = data.GetCritValue()
	d.Attr.ResilienceRate = data.GetResilienceRate()
	d.Attr.ResilienceValue = data.GetResilienceValue()
	d.Attr.HitRate = data.GetHitRate()
	d.Attr.DodgeRate = data.GetDodgeRate()

	d.HeroGs = d.Attr.GS_Int_NoLog()
	d.CorpGs = d.HeroGs * 3
}
