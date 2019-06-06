package account

import (
	"errors"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/general"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func FromAccount(a *helper.Avatar2Client, acc *Account, avatarID int) error {
	eqs, _, lv_evo, lv_star, lv_mat_enh, mat_enh := acc.Profile.GetEquips().CurrInfo()
	bag := acc.BagProfile
	a.Equips = make([][]byte, len(eqs), len(eqs))
	for idx, eq := range eqs {
		if eq != 0 {
			item := bag.GetItem(eq)
			if item == nil {
				logs.Error("FromAccount equips item no find %s %d ",
					acc.AccountID.String(), eq)
				//return errors.New("NoItem")
				continue
			}
			i2c := item.ToClient()
			a.Equips[idx] = encode(i2c)
			a.AppendEquip(i2c)
		}
	}

	a.AvatarEquips = acc.Profile.GetAvatarEquips().CurrByAvatar(avatarID)
	fashions := acc.Profile.GetFashionBag().GetFashionAll()
	a.AllFashions = make([][]byte, 0, len(fashions))
	for _, item := range fashions {
		a.AllFashions = append(a.AllFashions, encode(item))
		a.AppendAllFashions(item)
	}

	a.Title = acc.Profile.GetTitle().GetTitlesForOther(acc)
	a.TitleOn = acc.Profile.GetTitle().GetTitleOnShowForOther(acc)

	a.EquipUpgrade = lv_evo
	a.EquipStar = lv_star
	a.EquipMatEnhLv = lv_mat_enh
	a.EquipMatEnhMax = helper.EQUIP_MAT_ENHANCE_MAT
	a.EquipMatEnh = make([]bool, 0, len(mat_enh)*a.EquipMatEnhMax)
	for _, m := range mat_enh {
		a.EquipMatEnh = append(a.EquipMatEnh, m[:helper.EQUIP_MAT_ENHANCE_MAT]...)
	}
	logs.Debug("Get Account Equip Info: %d, %v, %v", a.EquipMatEnhMax, a.EquipMatEnh, a.EquipMatEnhLv)
	a.SetAcId(acc.AccountID.String())

	a.Arousals = acc.Profile.GetAvatarExp().GetAvatarArousalLv()
	a.AvatarSkills = acc.Profile.GetAvatarSkill().GetByAvatar(avatarID)
	a.SkillPractices = acc.Profile.GetAvatarSkill().GetPracticeLevel()
	a.AvatarId = avatarID
	a.HeroStarLv = acc.Profile.GetHero().HeroStarLevel[:]
	a.HeroLv = acc.Profile.GetHero().HeroLevel[:]
	a.HeroSoulLv = acc.Profile.GetHeroSoul().HeroSoulLevel

	a.Name = acc.Profile.Name
	a.VipLv = acc.Profile.GetVipLevel()
	a.CorpLv, a.CorpXP = acc.Profile.GetCorp().GetXpInfoNoUpdate()

	generals := acc.GeneralProfile.GetAllGeneral()
	a.Generals = make([]string, 0, len(generals))
	a.GeneralStars = make([]uint32, 0, len(generals))
	for _, gen := range generals {
		a.Generals = append(a.Generals, gen.Id)
		a.GeneralStars = append(a.GeneralStars, gen.StarLv)
	}

	generalRels := acc.GeneralProfile.GetAllGeneralRel()
	a.GeneralRels = make([]string, 0, len(generalRels))
	a.GeneralRelLevels = make([]uint32, 0, len(generalRels))
	for _, rel := range generalRels {
		a.GeneralRels = append(a.GeneralRels, rel.Id)
		a.GeneralRelLevels = append(a.GeneralRelLevels, rel.Level)
	}

	a.EquipJade = make([]string, 0, len(acc.Profile.GetEquipJades().Jades))
	for _, j := range acc.Profile.GetEquipJades().Jades {
		if j > 0 {
			item := acc.Profile.GetJadeBag().GetJade(j)
			if item != nil {
				a.EquipJade = append(a.EquipJade, item.TableID)
			}
		}
	}
	a.DestGeneralJade = make([]string, 0, len(acc.Profile.GetDestGeneralJades().DestinyGeneralJade))
	for _, j := range acc.Profile.GetDestGeneralJades().DestinyGeneralJade {
		if j > 0 {
			item := acc.Profile.GetJadeBag().GetJade(j)
			if item != nil {
				a.DestGeneralJade = append(a.DestGeneralJade, item.TableID)
			}
		}
	}

	destinyGenerals := acc.Profile.GetDestinyGeneral()
	a.CurrDestinyGeneralSkill = destinyGenerals.SkillGenerals[:]
	a.DestinyGeneralsID = make([]int, len(destinyGenerals.Generals))
	a.DestinyGeneralsLv = make([]int, len(destinyGenerals.Generals))
	for i, v := range destinyGenerals.Generals {
		a.DestinyGeneralsID[i] = v.Id
		a.DestinyGeneralsLv[i] = v.LevelIndex
	}
	currDestiny := destinyGenerals.GetGeneral(destinyGenerals.CurrGeneralIdx)
	if currDestiny != nil {
		a.DestinyGeneralID = currDestiny.Id
		a.DestinyGeneralLv = currDestiny.LevelIndex
	}

	guildUUID := acc.GuildProfile.GetCurrGuildUUID()
	if guildUUID != "" {
		a.GuildUUID = guildUUID
		a.GuildPos = acc.GuildProfile.GuildPosition
		a.GuildName = acc.GuildProfile.GuildName
	}

	a.AvatarLockeds = acc.Profile.GetCorp().GetUnlockedAvatar()

	_, heroAttrs, heroBaseGs, heroGss, bestHero, cur_gs, _ := gs.GetCurrAttr(NewAccountGsCalculateAdapter(acc))

	a.CorpGs = cur_gs
	a.Gs = heroGss[avatarID]
	a.Attr = encode(heroAttrs[avatarID])

	gsHeroGs, gsHeroBaseGs := gs.GetBestHeroInfo(bestHero, heroBaseGs, heroGss)
	a.GsHeroIds = bestHero
	a.GsHeroGs = gsHeroGs
	a.GsHeroBaseGs = gsHeroBaseGs

	a.PassiveSkillId = acc.Profile.GetHero().HeroSkills[avatarID].PassiveSkill[:]
	a.CounterSkillId = acc.Profile.GetHero().HeroSkills[avatarID].CounterSkill[:]
	a.TriggerSkillId = acc.Profile.GetHero().HeroSkills[avatarID].TriggerSkill[:]
	// TODO by ljz 可能需要增加不显示选项
	a.HeroSwing = acc.Profile.GetHero().HeroSwings[avatarID].CurSwing
	a.MagicPetfigure = acc.Profile.GetHero().GetMagicPetFigure(avatarID)

	return nil
}

func FromAccount2Json(a *helper.Avatar2ClientByJson, acc *Account, avatarID int) error {
	eqs, _, lv_evo, lv_star, lv_mat_enh, mat_enh := acc.Profile.GetEquips().CurrInfo()
	bag := acc.BagProfile
	a.Equips = make([]helper.BagItemToClient, len(eqs), len(eqs))
	for idx, eq := range eqs {
		if eq != 0 {
			item := bag.GetItem(eq)
			if item == nil {
				logs.Error("equips item %d no find", eq)
				return errors.New("NoItem")
			}
			a.Equips[idx] = item.ToClient()
		}
	}

	a.AcID = acc.AccountID.String()
	a.AvatarEquips = acc.Profile.GetAvatarEquips().CurrByAvatar(avatarID)
	fashions := acc.Profile.GetFashionBag().GetFashionAll()
	a.AllFashions = make([]helper.FashionItem, 0, len(fashions))
	for _, item := range fashions {
		a.AllFashions = append(a.AllFashions, item)
	}
	a.Title = acc.Profile.GetTitle().GetTitlesForOther(acc)
	a.TitleOn = acc.Profile.GetTitle().GetTitleOnShowForOther(acc)

	a.HeroStarLv = acc.Profile.GetHero().HeroStarLevel[:]
	a.HeroLv = acc.Profile.GetHero().HeroLevel[:]

	a.EquipUpgrade = lv_evo
	a.EquipStar = lv_star
	a.EquipMatEnhLv = lv_mat_enh
	a.EquipMatEnh = mat_enh
	a.Arousals = acc.Profile.GetAvatarExp().GetAvatarArousalLv()
	a.AvatarSkills = acc.Profile.GetAvatarSkill().GetByAvatar(avatarID)
	a.SkillPractices = acc.Profile.GetAvatarSkill().GetPracticeLevel()
	a.AvatarId = avatarID

	a.Name = acc.Profile.Name
	a.VipLv = acc.Profile.GetVipLevel()
	a.CorpLv, a.CorpXP = acc.Profile.GetCorp().GetXpInfo()

	generals := acc.GeneralProfile.GetAllGeneral()
	a.Generals = make([]string, 0, len(generals))
	a.GeneralStars = make([]uint32, 0, len(generals))
	for _, gen := range generals {
		a.Generals = append(a.Generals, gen.Id)
		a.GeneralStars = append(a.GeneralStars, gen.StarLv)
	}

	generalRels := acc.GeneralProfile.GetAllGeneralRel()
	a.GeneralRels = make([]string, 0, len(generalRels))
	a.GeneralRelLevels = make([]uint32, 0, len(generalRels))
	for _, rel := range generalRels {
		a.GeneralRels = append(a.GeneralRels, rel.Id)
		a.GeneralRelLevels = append(a.GeneralRelLevels, rel.Level)
	}

	a.EquipJade = make([]string, 0, len(acc.Profile.GetEquipJades().Jades))
	for _, j := range acc.Profile.GetEquipJades().Jades {
		if j > 0 {
			item := acc.Profile.GetJadeBag().GetJade(j)
			if item != nil {
				a.EquipJade = append(a.EquipJade, item.TableID)
			}
		}
	}
	a.DestGeneralJade = make([]string, 0, len(acc.Profile.GetDestGeneralJades().DestinyGeneralJade))
	for _, j := range acc.Profile.GetDestGeneralJades().DestinyGeneralJade {
		if j > 0 {
			item := acc.Profile.GetJadeBag().GetJade(j)
			if item != nil {
				a.DestGeneralJade = append(a.DestGeneralJade, item.TableID)
			}
		}
	}

	destinyGenerals := acc.Profile.GetDestinyGeneral()
	a.CurrDestinyGeneralSkill = destinyGenerals.SkillGenerals[:]
	currDestiny := destinyGenerals.GetGeneral(destinyGenerals.CurrGeneralIdx)
	if currDestiny != nil {
		a.DestinyGeneralID = currDestiny.Id
		a.DestinyGeneralLv = currDestiny.LevelIndex
	}

	guildUUID := acc.GuildProfile.GetCurrGuildUUID()
	if guildUUID != "" {
		a.GuildUUID = guildUUID
		a.GuildPos = acc.GuildProfile.GuildPosition
		a.GuildName = acc.GuildProfile.GuildName
	}

	a.PassiveSkillId = acc.Profile.GetHero().HeroSkills[avatarID].PassiveSkill[:]
	a.CounterSkillId = acc.Profile.GetHero().HeroSkills[avatarID].CounterSkill[:]
	a.TriggerSkillId = acc.Profile.GetHero().HeroSkills[avatarID].TriggerSkill[:]
	// TODO by ljz 可能需要增加不显示选项
	a.HeroSwing = acc.Profile.GetHero().HeroSwings[avatarID].CurSwing
	a.MagicPetfigure = acc.Profile.GetHero().GetMagicPetFigure(avatarID)

	a.AvatarLockeds = acc.Profile.GetCorp().GetUnlockedAvatar()

	_, heroAttrs, _, heroGss, _, corpgs, _ := gs.GetCurrAttr(NewAccountGsCalculateAdapter(acc))
	a.Attr = heroAttrs[a.AvatarId].AvatarAttr_
	a.HP = a.Attr.HP
	a.CorpGs = corpgs
	a.Gs = heroGss[a.AvatarId]
	return nil
}

func FromDroidAccount2Json(a *helper.Avatar2ClientByJson, acc *gamedata.DroidAccountData, curAvatarId int) error {
	accgs := &gs.DroidAccountGsCalculateAdapter{
		D: acc,
	}

	a.AcID = accgs.GetAcID()
	if curAvatarId < 0 {
		a.AvatarId = acc.AvatarId
	} else {
		a.AvatarId = curAvatarId
	}
	a.CorpLv = acc.CorpLv
	a.CorpXP = 0
	a.Arousals = acc.Arousals[:]
	a.AvatarSkills = acc.AvatarSkills[:]
	a.SkillPractices = acc.SkillPractices[:]
	a.AvatarLockeds = acc.AvatarLockeds[:]
	a.HeroStarLv = []uint32{1, 1, 1}
	a.HeroLv = []uint32{1, 1, 1}

	var mat_enh [][]bool
	a.AvatarEquips, a.EquipUpgrade, a.EquipStar, a.EquipMatEnhLv, mat_enh = accgs.CurrEquipInfo()
	a.EquipMatEnh = mat_enh

	a.Equips = make([]helper.BagItemToClient, len(acc.EquipsItemID), len(acc.EquipsItemID))
	for idx, eqID := range a.AvatarEquips {
		bagitem := accgs.GetItem(eqID)
		if bagitem != nil {
			b2c := bagitem.ToClient()
			a.Equips[idx] = b2c
		}
	}
	// a.AvatarEquips 是时装用的
	a.AvatarEquips = []uint32{uint32(10000 * (a.AvatarId + 1)), uint32(10000*(a.AvatarId+1) + 1)}
	a.AllFashions = make([]helper.FashionItem, 0, len(acc.Fashions[a.AvatarId]))
	for i := 0; i < len(acc.Fashions); i++ {
		for j := 0; j < len(acc.Fashions[i]); j++ {
			fi := helper.FashionItem{
				ID:              uint32(10000*(i+1) + j),
				TableID:         acc.Fashions[i][j],
				ExpireTimeStamp: 99999,
			}
			a.AllFashions = append(a.AllFashions, fi)
		}
	}

	logs.Trace("AvatarEquips %v %v", a.AvatarEquips, a.AllFashions)

	a.Generals = acc.Generals
	a.GeneralStars = acc.GeneralsStar
	a.GeneralRels = acc.GeneralRels
	a.GeneralRelLevels = acc.GeneralRelsStar

	a.EquipJade = acc.EquipJade[:]
	a.DestGeneralJade = acc.DestGeneralJade[:]

	a.DestinyGeneralID = acc.DestinyGeneralID
	a.DestinyGeneralLv = acc.DestinyGeneralLv
	a.CurrDestinyGeneralSkill = acc.CurrDestinyGeneralSkill[:]

	if acc.Name == "" {
		randName := gamedata.RandRobotNames(1)
		a.Name = randName[0]
	} else {
		a.Name = acc.Name
	}

	a.CorpGs = acc.CorpGs
	a.Gs = acc.HeroGs
	a.Attr = acc.Attr.AvatarAttr_

	logs.Trace("FromAccountByDroid %d %d", accgs.GetAcID(), a.Gs)

	return nil
}

func FromAccountByDroid(a *helper.Avatar2Client, acc *gamedata.DroidAccountData, curAvatarId int) error {
	accgs := &gs.DroidAccountGsCalculateAdapter{
		D: acc,
	}

	a.SetAcId(accgs.GetAcID())
	if curAvatarId < 0 {
		a.AvatarId = acc.AvatarId
	} else {
		a.AvatarId = curAvatarId
	}
	a.CorpLv = acc.CorpLv
	a.CorpXP = 0
	a.Arousals = acc.Arousals[:]
	a.AvatarSkills = acc.AvatarSkills[:]
	a.SkillPractices = acc.SkillPractices[:]
	a.AvatarLockeds = acc.AvatarLockeds[:]
	a.HeroStarLv = []uint32{1, 1, 1}
	a.HeroLv = []uint32{1, 1, 1}

	var mat_enh [][]bool
	a.AvatarEquips, a.EquipUpgrade, a.EquipStar, a.EquipMatEnhLv, mat_enh = accgs.CurrEquipInfo()
	a.EquipMatEnhMax = helper.EQUIP_MAT_ENHANCE_MAT
	a.EquipMatEnh = make([]bool, 0, len(mat_enh)*a.EquipMatEnhMax)
	for _, m := range mat_enh {
		a.EquipMatEnh = append(a.EquipMatEnh, m...)
		logs.Debug("Equipmatenh %v", a.EquipMatEnh)
	}
	a.Equips = make([][]byte, len(acc.EquipsItemID), len(acc.EquipsItemID))
	for idx, eqID := range a.AvatarEquips {
		bagitem := accgs.GetItem(eqID)
		if bagitem != nil {
			b2c := bagitem.ToClient()
			a.Equips[idx] = encode(b2c)
			a.AppendEquip(b2c)
		}
	}
	// a.AvatarEquips 是时装用的
	a.AvatarEquips = []uint32{uint32(10000 * (a.AvatarId + 1)), uint32(10000*(a.AvatarId+1) + 1)}
	a.AllFashions = make([][]byte, 0, len(acc.Fashions[a.AvatarId]))
	for i := 0; i < len(acc.Fashions); i++ {
		for j := 0; j < len(acc.Fashions[i]); j++ {
			fi := helper.FashionItem{
				ID:              uint32(10000*(i+1) + j),
				TableID:         acc.Fashions[i][j],
				ExpireTimeStamp: 99999,
			}
			a.AllFashions = append(a.AllFashions, encode(fi))
			a.AppendAllFashions(fi)
		}
	}

	logs.Trace("AvatarEquips %v %v", a.AvatarEquips, a.AllFashions)

	a.Generals = acc.Generals
	a.GeneralStars = acc.GeneralsStar
	a.GeneralRels = acc.GeneralRels
	a.GeneralRelLevels = acc.GeneralRelsStar

	a.EquipJade = acc.EquipJade[:]
	a.DestGeneralJade = acc.DestGeneralJade[:]

	a.DestinyGeneralID = acc.DestinyGeneralID
	a.DestinyGeneralLv = acc.DestinyGeneralLv
	a.CurrDestinyGeneralSkill = acc.CurrDestinyGeneralSkill[:]

	if acc.Name == "" {
		randName := gamedata.RandRobotNames(1)
		a.Name = randName[0]
	} else {
		a.Name = acc.Name
	}
	logs.Debug("DroidData Name is %s", acc.Name)
	logs.Debug("Droid Name is: %s", a.Name)

	a.CorpGs = acc.CorpGs
	a.Gs = acc.HeroGs
	a.Attr = encode(&acc.Attr)

	logs.Trace("FromAccountByDroid %d %d", accgs.GetAcID(), a.Gs)

	return nil
}

type Avatar2ClientJson struct {
	Acid           string
	AvatarId       int
	CorpLv         uint32
	AvatarArousals map[string]int
	avatarArousals map[int]int
	AvatarSkills   []uint32
	Name           string

	SimplePvpScore int
	SimplePvpRank  int

	Equips        []bag.BagItem
	EquipUpgrade  []uint32
	EquipStar     []uint32
	EquipMatEnhLv []uint32
	EquipMatEnh   [][]bool
	Fashions      []helper.FashionItem

	Generals    []general.General
	GeneralRels []general.Relation

	AvatarJade      []string
	DestGeneralJade []string

	GuildPost     string
	GuildPostTime int64
}

func (a *Avatar2ClientJson) FromAccount(acc *Account, avatarID int) error {
	eqs, _, lv_evo, lv_star, lv_mat_enh, mat_enh := acc.Profile.GetEquips().CurrInfo()
	a.Equips = make([]bag.BagItem, len(eqs), len(eqs))
	for idx, eq := range eqs {
		if eq != 0 {
			item := acc.BagProfile.GetItem(eq)
			if item == nil {
				logs.Error("item %d no find", eq)
				return errors.New("NoItem")
			}
			a.Equips[idx] = *item
		}
	}

	a.EquipUpgrade = lv_evo
	a.EquipStar = lv_star
	a.EquipMatEnhLv = lv_mat_enh
	a.EquipMatEnh = mat_enh
	a.Acid = acc.AccountID.String()

	aEqs, _ := acc.Profile.GetAvatarEquips().Curr()
	a.Fashions = make([]helper.FashionItem, len(aEqs), len(aEqs))
	for idx, eq := range aEqs {
		if eq != 0 {
			ok, item := acc.Profile.GetFashionBag().GetFashionInfo(eq)
			if !ok {
				logs.Error("fashion %d no find", eq)
				return errors.New("NoFashion")
			}
			a.Fashions[idx] = item
		}
	}

	a.AvatarArousals = make(map[string]int, gamedata.AVATAR_NUM_CURR)
	a.avatarArousals = make(map[int]int, gamedata.AVATAR_NUM_CURR)
	for _, avatar_id := range acc.Profile.GetCorp().GetUnlockedAvatar() {
		l := int(acc.Profile.GetAvatarExp().GetArousalLv(avatar_id))
		a.AvatarArousals[fmt.Sprintf("%d", avatar_id)] = l
		a.avatarArousals[avatar_id] = l
	}
	a.AvatarSkills = acc.Profile.GetAvatarSkill().GetByAvatar(avatarID)
	a.AvatarId = avatarID

	a.Name = acc.Profile.Name
	a.CorpLv, _ = acc.Profile.GetCorp().GetXpInfoNoUpdate()

	a.Generals = acc.GeneralProfile.GetAllGeneral()
	a.GeneralRels = acc.GeneralProfile.GetAllGeneralRel()

	a.GuildPost, a.GuildPostTime = acc.GuildProfile.GetPost()
	nowT := acc.Profile.GetProfileNowTime()
	if nowT > a.GuildPostTime {
		a.GuildPost = ""
	}

	equip_jade := acc.Profile.GetEquipJades().Jades
	if equip_jade != nil {
		a.AvatarJade = make([]string, 0, len(equip_jade))
		for _, j := range equip_jade {
			if j > 0 {
				item := acc.Profile.GetJadeBag().GetJade(j)
				if item != nil {
					a.AvatarJade = append(a.AvatarJade, item.TableID)
				} else {
					logs.Error("Avatar2ClientJson FromAccount AvatarJades can't find item %d", j)
				}
			}
		}
	}

	dg_equip_jade := acc.Profile.GetDestGeneralJades().DestinyGeneralJade
	if dg_equip_jade != nil {
		a.DestGeneralJade = make([]string, 0, len(dg_equip_jade))
		for _, j := range dg_equip_jade {
			if j > 0 {
				item := acc.Profile.GetJadeBag().GetJade(j)
				if item != nil {
					a.DestGeneralJade = append(a.DestGeneralJade, item.TableID)
				} else {
					logs.Error("Avatar2ClientJson FromAccount DestGeneralJades can't find item %d", j)
				}
			}
		}
	}
	return nil
}

func (a *Avatar2ClientJson) ToGs() float32 {
	subModuleGs := make([]int, helper.Gs_Module_Count)
	attr := gamedata.GetBaseAvatarAttr(a.Acid, a.CorpLv, subModuleGs)

	a.GsEquip(&attr)
	a.GsGeneral(&attr)
	a.GsSkill(&attr)

	gs := attr.GS()
	logs.Trace("attr %v gs %v", attr, gs)
	return gs
}

func (a *Avatar2ClientJson) GsEquip(attr *gamedata.AvatarAttr) {
	if attr == nil {
		return
	}

	for idx, equip := range a.Equips {
		item := &equip
		a.GsEquipItem(item, a.EquipStar[idx], attr)
		attr.AddEquipEvolution(idx, a.EquipUpgrade[idx])
		attr.AddEquipMatEnhance(idx, a.EquipMatEnhLv[idx], a.EquipMatEnh[idx])
	}

	logs.Trace("GsEquip %v", attr)
	return
}

func (a *Avatar2ClientJson) GsEquipItem(item *bag.BagItem, star uint32, attr *gamedata.AvatarAttr) {
	item_data, data_ok := gamedata.GetProtoItem(item.TableID)
	if !data_ok {
		logs.Warn("GsEquipItem Data No Found by %v", item.ID)
		return
	}

	rankLimit := item_data.GetRankLimit()
	playerRank := a.GuildPost
	if rankLimit != "" && rankLimit != playerRank {
		// 官阶不符 跳过
		return
	}

	attrAdd := &gamedata.AvatarAttr{}

	attrAdd.AddBase(item_data.GetAttack(), item_data.GetDefense(), item_data.GetHP())

	data := item.GetItemData()

	if data == nil {
		logs.Warn("GetItemData Data No Found by %v", item)
		return
	}

	for _, tr := range data.TrickGroup {
		if tr != "" {
			a := gamedata.GetTrickDetailAttrAddon(tr)
			if a != nil {
				attrAdd.Add(a)
			}
		}
	}

	var addRate float32 = 0
	starInfo := gamedata.GetEquipStarData(star)
	if starInfo == nil {
		addRate = 0.0
	} else {
		addRate = starInfo.GetAddition()
	}

	attrAdd.AddEquipStarAddon(addRate)
	logs.Trace("attrAdd by %v", attrAdd)

	attr.AddOther(attrAdd)
	return
}

func (a *Avatar2ClientJson) GsGeneral(attr *gamedata.AvatarAttr) {
	for _, g := range a.Generals {
		atk, def, hp := gamedata.GetGeneralStarAttr(g.Id, g.StarLv)
		attr.AddBase(atk, def, hp)
	}

	for _, r := range a.GeneralRels {
		atk, def, hp := gamedata.GetGeneralRelLvlAttr(r.Id, r.Level)
		attr.AddBase(atk, def, hp)
	}
	return
}

func (a *Avatar2ClientJson) GsSkill(attr *gamedata.AvatarAttr) {
	for idx, lv := range a.AvatarSkills {
		skill_cfg := gamedata.GetSkillLevelConfig(a.AvatarId, idx)
		if skill_cfg == nil {
			logs.Warn("Skill Cfg Nil by %d %d", a.AvatarId, idx)
			continue
		}
		if int(lv) < len(skill_cfg.SkillGS) && int(lv) >= 0 {
			attr.AddGSAddon(skill_cfg.SkillGS[int(lv)])
		}
	}
	return
}

func (a *Avatar2ClientJson) GsJade(attr *gamedata.AvatarAttr) {
	for _, j := range a.AvatarJade {
		_, cfg := gamedata.IsJade(j)
		attr.AddBase(cfg.GetAttack(), cfg.GetDefense(), cfg.GetHP())
	}
	for _, j := range a.DestGeneralJade {
		_, cfg := gamedata.IsJade(j)
		attr.AddBase(cfg.GetAttack(), cfg.GetDefense(), cfg.GetHP())
	}
}
