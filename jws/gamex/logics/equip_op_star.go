package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// EquipStarLevelUp 一次升星逻辑 返回 code 暴击类型
func (p *Account) EquipStarLevelUp(
	slot int,
	count int,
	isUseHC bool,
	sync helper.ISyncRsp) (uint32, []int) {
	logs.Trace("[%s]EquipStarLevelUp:%d,%v,%v",
		p.AccountID, slot, isUseHC, count)

	const (
		_                      = iota
		CODE_No_StarLevel_Info // 失败:升星信息缺失
		CODE_No_Enough_Cost    // 失败:没有足够的物品
		CODE_No_Equip          // 失败:升级精炼位置错误，没有装备
		CODE_No_Equip_Info     // 失败:装备信息缺失
		CODE_VIP_Err           // 失败:VIP等级不够
	)

	equips := p.Profile.GetEquips()
	//nowT := p.Profile.GetProfileNowTime()
	resBonus := make([]int, 0, count)
	costSCTyp, costCount, xp, little, big, hcRatio :=
		gamedata.GetEquipStarUpSettings()

	var sumSc, sumMoney, sumHc uint32
	oldStar := equips.GetStarLv(slot)
	for i := 0; i < count; i++ {
		//if isUseHC {
		//	hasUseCount := equips.StarHcUpCount.Get(nowT)
		//	vipCfg := p.Profile.GetMyVipCfg()
		//	if hasUseCount+1 > vipCfg.StarHcUpLimitDaily {
		//		logs.Warn("EquipStarLevelUp CODE_VIP_Err")
		//		return mkCode(CODE_WARN, errCode.ClickTooQuickly), resBonus[:]
		//	} else {
		//		equips.StarHcUpCount.Add(nowT, 1)
		//	}
		//}

		nextLv := equips.GetStarLv(slot) + 1
		xpNeed, bp := gamedata.GetEquipStarLvUpData(nextLv)
		dataStarUp := gamedata.GetEquipStarData(nextLv)
		if xpNeed <= 0 || bp == nil || dataStarUp == nil {
			return 0, resBonus
		}

		// 消耗折算
		costs := new(gamedata.CostData)
		scUse := costCount

		if isUseHC {
			// HC升星
			//starHcUpCount := equips.StarHcUpCount.Count
			//hc := gamedata.GetStarUpHcCount(starHcUpCount - 1)
			hc := uint32(float32(costCount) * hcRatio)
			costs.AddItem(gamedata.VI_Hc, hc)
			sumHc += hc
		} else {
			costs.AddItem(costSCTyp, scUse)
			sumSc += scUse
		}

		costs.AddItem(gamedata.VI_Sc0, dataStarUp.GetSCCost())
		sumMoney += dataStarUp.GetSCCost()

		logs.Trace("costs %v", costs)

		if !account.CostBySync(p.Account, costs, sync, "EquipStarLevelUp") {
			return mkCode(CODE_WARN, errCode.ClickTooQuickly), resBonus
		}

		//
		var bonus uint32 = 1
		bonusTyp := bp.Rand(p.GetRand())
		switch bonusTyp {
		case gamedata.EquipStarLevelUpBonus_Big:
			bonus = big
		case gamedata.EquipStarLevelUpBonus_Little:
			bonus = little
		}

		resBonus = append(resBonus, bonusTyp)

		equips.AddStarXP(slot, xp*bonus)
		for xpNeed > 0 && equips.GetStarXP(slot) >= xpNeed {
			equips.AddStarXP(slot, -xpNeed)
			equips.StarLvUp(slot)
			xpNeed, _ = gamedata.GetEquipStarLvUpData(equips.GetStarLv(slot) + 1)
			if xpNeed <= 0 {
				// 升满级了
				equips.SetStarXP(slot, 0)
			}
		}

		p.updateCondition(account.COND_TYP_EquipStarUp,
			1, 0, "", "", sync)

		p.updateCondition(account.COND_TYP_EquipStarPartCount,
			0, 0, "", "", sync)
	}

	// logiclog
	typ := "one"
	if count > 1 {
		typ = "ten"
	}

	// 装备星级排行榜更新
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByEquipStarLv.Add(&info)
	logiclog.LogEquipStarUp(p.AccountID.String(), p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, typ, slot, oldStar, equips.GetStarLv(slot), sumSc, sumMoney, sumHc,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	// sysnotice
	for _, cfg := range gamedata.EquipStarSysNotice() {
		cd := cfg.Cond
		c := 0
		_, _, _, avatarStars, _, _, _, _ := p.Profile.GetEquips().Curr()
		for i := 0; i < len(avatarStars); i++ {
			if int64(avatarStars[i]) == cd.Param1 {
				c++
			}
		}
		if int64(c) == cd.Param2 {
			alrdy := false
			for _, v := range equips.StarSysNotice {
				if v == cd {
					alrdy = true
					break
				}
			}
			if !alrdy {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfg.Cfg.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", cfg.Star)).Send()
				equips.StarSysNotice = append(equips.StarSysNotice, cd)
			}
			break
		}
	}
	return 0, resBonus
}
