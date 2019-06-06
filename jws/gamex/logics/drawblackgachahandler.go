package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

// DrawBlackGacha : 黑盒宝箱抽奖
// 黑盒宝箱抽奖
func (p *Account) DrawBlackGachaHandler(req *reqMsgDrawBlackGacha, resp *rspMsgDrawBlackGacha) uint32 {
	actId := uint32(req.BlackGachaId)
	subId := uint32(req.BlackGachaSubId)

	p.Profile.BlackGachaInfo.TryDailyReset(p.Account.GetProfileNowTime())

	subAct := p.Profile.BlackGachaInfo.GetSubActivity(actId, subId)
	if subAct == nil {
		return errCode.ActivityTimeOut
	}
	settingCfg := gamedata.GetHotDatas().HotBlackGachaData.GetBlackGachaSettingsCfg(actId, subId)
	if settingCfg == nil {
		return errCode.CommonInvalidParam
	}
	// 次数判断
	count := 0
	if req.IsTen {
		count = 10
	} else {
		count = 1
	}
	if subAct.GachaCount+count > int(settingCfg.GetDailyLimit()) {
		return errCode.CommonCountLimit
	}

	// 随机奖励
	randomCfgs := gamedata.GetHotDatas().Activity.GetActivityGachaNormal(settingCfg.GetItemGroupID())
	cfgArray := randomCfgs.RandomConfigByCount(count)

	// 消耗
	reason := fmt.Sprintf("draw black gacha %d %d %d", actId, subId, req.IsTen)
	if errCode := p.costDrawBlackGacha(count, settingCfg, resp, reason, subAct); errCode != 0 {
		return errCode
	}

	subAct.GachaCount += count

	// 给奖励
	rewardData := &gamedata.CostData{}
	for _, cfg := range cfgArray {
		rewardData.AddItem(cfg.GetItemID(), cfg.GetItemCount())
		trySendRollNotice(p, cfg.GetItemID(), cfg.GetItemCount(), settingCfg.GetMsgQuality(), settingCfg.GetMsgIDS())
	}
	if settingCfg.GetGachaItem() != "" {
		rewardData.AddItem(settingCfg.GetGachaItem(), uint32(count))
	}
	if !account.GiveBySync(p.Account, rewardData, resp, reason) {
		logs.Error("black gacha GiveBySync Err")
		return errCode.ClickTooQuickly
	}

	// 回客户端
	// 抽奖的奖励不能合并， 需要特殊处理
	resp.RewardID = make([]string, 0)
	resp.RewardCount = make([]uint32, 0)
	resp.RewardData = make([]string, 0)
	for _, cfg := range cfgArray {
		resp.RewardID = append(resp.RewardID, cfg.GetItemID())
		resp.RewardCount = append(resp.RewardCount, cfg.GetItemCount())
		resp.RewardData = append(resp.RewardData, "")
	}
	if settingCfg.GetGachaItem() != "" {
		resp.GiveRewardId = settingCfg.GetGachaItem()
		resp.GiveRewardCount = int64(count)
	}
	resp.BlackGachaId = req.BlackGachaId
	resp.GachaInfo = encode(convertGachaInfo2Client(subAct))
	return 0
}

func (p *Account) costDrawBlackGacha(count int, settingCfg *ProtobufGen.BOXSETTINGS,
	resp *rspMsgDrawBlackGacha, reason string, subAct *account.BlackGachaSubInfo) uint32 {
	if count == 10 {
		//十连抽，不免费
		costData := &gamedata.CostData{}
		if p.Profile.GetSC().HasSC(helper.SCId(settingCfg.GetGachaTicket()), int64(settingCfg.GetTTenPrice())) {
			//有抽十次的抽奖券，可以用抽奖券
			costData.AddItem(settingCfg.GetGachaTicket(), settingCfg.GetTTenPrice())
		} else {
			//没有抽十次的抽奖券，用货币抽
			costData.AddItem(settingCfg.GetGachaCoin1(), settingCfg.GetTenPrice1())
		}
		if !account.CostBySync(p.Account, costData, resp, reason) {
			logs.Warn("black gacha CostBySync Err")
			return errCode.ClickTooQuickly
		}
	} else {
		//单抽
		if subAct.TodayFreeUsedCount < int(settingCfg.GetFreeTime()) {
			//免费抽
			subAct.TodayFreeUsedCount++
		} else {
			//不免费抽
			costData := &gamedata.CostData{}
			if p.Profile.GetSC().HasSC(helper.SCId(settingCfg.GetGachaTicket()), int64(settingCfg.GetTAPrice())) {
				//有抽一次的抽奖券，可以用抽奖券
				costData.AddItem(settingCfg.GetGachaTicket(), settingCfg.GetTAPrice())
			} else {
				//没有抽一次的抽奖券，用货币抽
				costData.AddItem(settingCfg.GetGachaCoin1(), settingCfg.GetAPrice1())
			}
			if !account.CostBySync(p.Account, costData, resp, reason) {
				logs.Warn("black gacha CostBySync Err")
				return errCode.ClickTooQuickly
			}
		}
	}
	return 0
}

func trySendRollNotice(p *Account, itemId string, itemCount uint32, noticeLimit uint32, noticeId uint32) {
	if cfg, ok := gamedata.GetProtoItem(itemId); ok {
		if uint32(cfg.GetRareLevel()) >= noticeLimit {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(noticeId)).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_ItemId, itemId).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", itemCount)).Send()
		}
	}
}

// GetBlackGachaInfo : 黑盒宝箱信息
// 黑盒宝箱信息
func (p *Account) GetBlackGachaInfoHandler(req *reqMsgGetBlackGachaInfo, resp *rspMsgGetBlackGachaInfo) uint32 {
	p.UpdateMarketActivity()
	p.Profile.BlackGachaInfo.TryDailyReset(p.Account.GetProfileNowTime())
	gachaInfo := &p.Profile.BlackGachaInfo
	resp.HeroActivityId = int64(gachaInfo.BlackGachaHeroInfo.ActivityId)
	if resp.HeroActivityId > 0 {
		resp.HeroInfo = make([][]byte, 0)
		for _, act := range gachaInfo.BlackGachaHeroInfo.SubActivies {
			tempAct := act
			resp.HeroInfo = append(resp.HeroInfo, encode(convertGachaInfo2Client(&tempAct)))
		}
	}
	resp.WeaponActivityId = int64(gachaInfo.BlackGachaWeaponInfo.ActivityId)
	if resp.WeaponActivityId > 0 {
		resp.WeaponInfo = make([][]byte, 0)
		for _, act := range gachaInfo.BlackGachaWeaponInfo.SubActivies {
			tempAct := act
			resp.WeaponInfo = append(resp.WeaponInfo, encode(convertGachaInfo2Client(&tempAct)))
		}
	}
	return 0
}

func (p *Account) UpdateMarketActivity() {
	p.Profile.GetMarketActivitys().UpdateMarketActivity(p.AccountID.String(), p.Profile.GetProfileNowTime())
	p.CheckBlackGachaActivity()
}

func (p *Account) CheckBlackGachaActivity() {
	hasHeroAct, hasWeaponAct := false, false
	acts := gamedata.GetHotDatas().Activity.GetAllActivityInfoValid(p.Profile.ChannelQuickId, p.GetProfileNowTime())
	for _, act := range acts {
		if !hasHeroAct && act.ActivityType == gamedata.ActBlackGachaHero {
			hasHeroAct = true
			p.updateBlackGachaInfo(act.ActivityId, act.ActivityType)
		}
		if !hasWeaponAct && act.ActivityType == gamedata.ActBlackGachaWeapon {
			hasWeaponAct = true
			p.updateBlackGachaInfo(act.ActivityId, act.ActivityType)
		}
	}
	if !hasHeroAct {
		p.updateBlackGachaInfo(0, gamedata.ActBlackGachaHero)
	}
	if !hasWeaponAct {
		p.updateBlackGachaInfo(0, gamedata.ActBlackGachaWeapon)
	}
	tryInitBlackGacha(&p.Profile.BlackGachaInfo.BlackGachaHeroInfo)
	tryInitBlackGacha(&p.Profile.BlackGachaInfo.BlackGachaWeaponInfo)
}

func convertGachaInfo2Client(act *account.BlackGachaSubInfo) BlackGachaActivity {
	return BlackGachaActivity{
		TodayFreeUsedCount: int64(act.TodayFreeUsedCount),
		GachaCount:         int64(act.GachaCount),
		HasClaimedReward:   act.HasClaimedExtraReward,
		SubActivityId:      int64(act.SubId),
	}
}

// 同步运营里面的活动和profile里面的活动， 结算已经删除的或者结束的活动
func (p *Account) updateBlackGachaInfo(newActId uint32, actType uint32) {
	if actType == gamedata.ActBlackGachaHero {
		oldId := p.Profile.BlackGachaInfo.BlackGachaHeroInfo.ActivityId
		if oldId != newActId {
			p.sendBlackGachaMail(oldId, actType, &p.Profile.BlackGachaInfo.BlackGachaHeroInfo)
			p.Profile.BlackGachaInfo.BlackGachaHeroInfo.Reset(newActId)
		}
	} else if actType == gamedata.ActBlackGachaWeapon {
		oldId := p.Profile.BlackGachaInfo.BlackGachaWeaponInfo.ActivityId
		if oldId != newActId {
			p.sendBlackGachaMail(oldId, actType, &p.Profile.BlackGachaInfo.BlackGachaWeaponInfo)
			p.Profile.BlackGachaInfo.BlackGachaWeaponInfo.Reset(newActId)
		}
	}
}

func (p *Account) sendBlackGachaMail(actId, actType uint32, activity *account.BlackGachaActivity) {
	for _, subAct := range activity.SubActivies {
		subCfg := gamedata.GetHotDatas().HotBlackGachaData.GetBlackGachaSettingsCfg(actId, subAct.SubId)
		lowRewards := gamedata.GetHotDatas().HotBlackGachaData.GetAllSubBlackGachaLowest(subAct.SubId)
		for _, reward := range lowRewards {
			if subAct.GachaCount >= int(reward.GetLowestTimes()) && !subAct.ContainsReward(int64(reward.GetLowestTimes())) {
				mailId := 0
				if actType == gamedata.ActBlackGachaHero {
					mailId = mail_sender.IDS_MAIL_BLACK_GACHA_HERO_TITLE
				} else {
					mailId = mail_sender.IDS_MAIL_BLACK_GACHA_GWC_TITLE
				}
				itemMap := make(map[string]uint32)
				for _, loot := range reward.GetFixed_Loot() {
					itemMap[loot.GetItemID()] = loot.GetItemCount()
				}
				mail_sender.BatchSendMail2Account(p.AccountID.String(),
					timail.Mail_send_By_Black_Gacha,
					mailId,
					[]string{subCfg.GetActivitySubName(), fmt.Sprintf("%d", reward.GetLowestTimes())},
					itemMap,
					"BlackGachaMail",
					true,
				)
			}
		}
	}
}

func tryInitBlackGacha(activity *account.BlackGachaActivity) {
	if activity.ActivityId != 0 && activity.SubActivies == nil {
		subCfgs := gamedata.GetHotDatas().HotBlackGachaData.GetAllSubGachaSettingsCfg(activity.ActivityId)
		activity.SubActivies = make([]account.BlackGachaSubInfo, 0)
		for _, cfg := range subCfgs {
			activity.SubActivies = append(activity.SubActivies, account.BlackGachaSubInfo{
				SubId: cfg.GetActivitySubID(),
			})
		}
	}
}

// ClaimBlackGachaExtraReward : 获取黑盒宝箱的额外奖励
// 累计抽奖一定次数后会有额外的奖励
func (p *Account) ClaimBlackGachaExtraRewardHandler(req *reqMsgClaimBlackGachaExtraReward, resp *rspMsgClaimBlackGachaExtraReward) uint32 {
	p.Profile.BlackGachaInfo.TryDailyReset(p.Account.GetProfileNowTime())
	actId := uint32(req.BlackGachaActivityId)
	subId := uint32(req.BlackGachaActivitySubId)
	rewardId := req.RewardId

	if actId == 0 || subId == 0 {
		return errCode.CommonInvalidParam
	}
	subAct := p.Profile.BlackGachaInfo.GetSubActivity(actId, subId)
	if subAct.ContainsReward(rewardId) {
		return errCode.CommonCountLimit
	}
	rewardCfg := gamedata.GetHotDatas().HotBlackGachaData.GetBlackGachaLowest(subId, uint32(rewardId))
	if rewardCfg == nil {
		return errCode.CommonInvalidParam
	}
	if subAct.GachaCount < int(rewardCfg.GetLowestTimes()) {
		return errCode.CommonConditionFalse
	}

	subAct.AddHasClaimedReward(rewardId)

	rewardData := &gamedata.CostData{}
	for _, cfg := range rewardCfg.GetFixed_Loot() {
		rewardData.AddItem(cfg.GetItemID(), cfg.GetItemCount())
	}
	reason := fmt.Sprintf("claim black gacha reward %d %d %d", actId, subId, rewardId)
	if ok := account.GiveBySync(p.Account, rewardData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	resp.BlackGachaActivityId = req.BlackGachaActivityId
	resp.BlackGachaActivitySubId = req.BlackGachaActivitySubId
	resp.GachaInfo = encode(convertGachaInfo2Client(subAct))
	return 0
}
