package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 初始化存档
func (p *Account) initAvatars() {
	// 普通数值
	data := gamedata.GetAvatarInitData()
	acid := p.AccountID.String()
	if len(data) <= 0 {
		logs.SentryLogicCritical(acid, "[init]GetAvatarInitData Err")
		return
	}

	give := &gamedata.CostData{}

	give.AddItem(gamedata.VI_Sc0, data[0].GetInitSC())
	give.AddItem(gamedata.VI_Sc1, data[0].GetInitFineIron())
	give.AddItem(gamedata.VI_Hc_Buy, data[0].GetInitHC_Buy())
	give.AddItem(gamedata.VI_Hc_Compensate, data[0].GetInitHC_Compensate())
	give.AddItem(gamedata.VI_Hc_Give, data[0].GetInitHC_Give())

	if !GiveBySync(p, give, nil, "init") {
		logs.SentryLogicCritical(acid, "[init]GiveBySync Err")
		return
	}

	eqs := data[0]
	giveAndThenEquip(p, eqs.GetInitWeapon(), gamedata.PartID_Weapon)
	giveAndThenEquip(p, eqs.GetInitBelt(), gamedata.PartID_Belt)
	giveAndThenEquip(p, eqs.GetInitBracers(), gamedata.PartID_Bracers)
	giveAndThenEquip(p, eqs.GetInitChest(), gamedata.PartID_Chest)
	giveAndThenEquip(p, eqs.GetInitLeggings(), gamedata.PartID_Leggings)
	giveAndThenEquip(p, eqs.GetInitNecklace(), gamedata.PartID_Necklace)
	giveAndThenEquip(p, eqs.GetInitRing(), gamedata.PartID_Ring)

	// 技能解锁
	unlockSkills := gamedata.GetLevelUnlockSkills(1)

	if unlockSkills != nil && len(unlockSkills) > 0 {
		playerSkill := p.Profile.GetAvatarSkill()
		for _, skill := range unlockSkills {
			logs.Trace("UnlockSkill %d %d", skill.AvatarId, skill.SkillId)
			playerSkill.UnlockSkill(skill.AvatarId, skill.SkillId)
		}
		//logs.Trace("Skills %v", playerSkill.Skills)
	}

	// 加背包物品
	for _, item := range gamedata.GetAvatarInitBagData() {
		giveItemToBag(p, item.GetItemID(), item.GetItemCount())
	}

	// 角色装备
	for i := 0; i < helper.AVATAR_NUM_CURR; i++ {
		af := gamedata.GetAvatarInitFashionData(i)
		avatarGiveAndThenEquip(p, i, af.GetInitFWeapon(), gamedata.FashionPart_Weapon)
		avatarGiveAndThenEquip(p, i, af.GetInitFAmor(), gamedata.FashionPart_Armor)
	}
}

func GiveAndThenEquip(p *Account, item_id string, slot int) {
	giveAndThenEquip(p, item_id, slot)
}

func giveAndThenEquip(p *Account, item_id string, slot int) {
	acid := p.AccountID.String()
	if item_id == "" {
		return
	}
	errCode, _, bagId2OldCount := p.BagProfile.AddToBag(p, gamedata.BagItemData{}, item_id, 1)
	if errCode != helper.RES_AddToBag_Success {
		logs.SentryLogicCritical(acid, "[init]AddToBag Err %d", errCode)
		return
	}
	for bagId, _ := range bagId2OldCount {
		p.Profile.GetEquips().EquipImp(slot, bagId)
		break
	}
	return
}
func AvatarGiveAndThenEquip(p *Account, avatar_id int, item_id string, slot int) {
	avatarGiveAndThenEquip(p, avatar_id, item_id, slot)
}
func avatarGiveAndThenEquip(p *Account, avatar_id int, item_id string, slot int) {
	acid := p.AccountID.String()
	if item_id == "" {
		return
	}
	ok, cfg := gamedata.IsFashion(item_id)
	if ok {
		errCode, _, bagId2OldCount := p.Profile.GetFashionBag().AddFashionByTableId(item_id,
			cfg, p.Profile.GetProfileNowTime())
		if errCode != helper.RES_AddToBag_Success {
			logs.SentryLogicCritical(acid, "[init]AddToBag Err %d", errCode)
			return
		}
		for bagId, _ := range bagId2OldCount {
			p.Profile.GetAvatarEquips().EquipImp(avatar_id, slot, bagId)
			break
		}
	}
	return
}

func giveItemToBag(p *Account, item_id string, count uint32) {
	acid := p.AccountID.String()
	if item_id == "" || count <= 0 {
		return
	}
	data := &gamedata.CostData{}
	data.AddItem(item_id, count)

	give := GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p, nil, "init") {
		logs.SentryLogicCritical(acid, "[init]AddItem Fail ")
		return
	}
}

// 是否角色已经解锁
func (p *Account) IsAvatarUnblock(avatarID int) bool {
	/* TODO By FanYang 需要确认当有新角色时, 老玩家是否要自动解锁
	data := gamedata.GetAvatarOpenCondData(avatarID)
	logs.Trace("IsAvatarUnblock %d %v", avatarID, data)
	if data == nil {
		return false
	}

	if data.Typ == 0 {
		return true
	}

	switch data.Typ {
	case 0:
		return true
	case gamedata.FteConditionRoleOpenTypCorpLv:
		corpLv, _ := p.Profile.GetCorp().GetXpInfo()
		//logs.Warn("FteConditionRoleOpenTypCorpLv %d %d %v", avatarID, corpLv, data)
		return corpLv >= uint32(data.Pint)
	case gamedata.FteConditionRoleOpenTypStage:
		//logs.Warn("FteConditionRoleOpenTypStage %d %v", avatarID, data)
		return p.Profile.GetStage().GetStar(data.Pstr) > 0
	}
	logs.Error("IsAvatarUnblock Unknown Typ %v", *data)
	*/
	return p.Profile.GetCorp().IsAvatarHasUnlock(avatarID)
}

//
//
//
//
//
//
