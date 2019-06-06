package logics

import (
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"

	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//SyncResp GetModule
func (s *SyncRespNotify) mkGuildInfo(p *Account) {
	guildID := p.GuildProfile.GuildUUID

	if guildID != "" {
		if p.GuildProfile.HasApplyCanApprove {
			s.OnChangeRedPoint(notify.RedPointTyp_Guild)
		}
	}

	if s.SyncPlayerGuild {
		// player guild info
		playerGuild := guildPlayerInfoToClient{
			GuildUUID:     guildID,
			NextEnterTime: p.GuildProfile.NextEnterGuildTime,
		}
		s.SyncPlayerGuildInfo = encode(playerGuild)
	}

	if s.SyncPlayerGuildApply {
		bs := time.Now().UnixNano()
		aplys := guild.GetModule(p.AccountID.ShardId).GetPlayerApplyInfo(p.AccountID.String())
		metric_send(p.AccountID, "PlayerGuildGetApply", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
		if aplys != nil {
			s.SyncPlayerApplyGuild_UUID = make([]string, len(aplys))
			s.SyncPlayerApplyGuild_Name = make([]string, len(aplys))
			s.SyncPlayerApplyGuild_Lvl = make([]uint32, len(aplys))
			s.SyncPlayerApplyGuild_Notice = make([]string, len(aplys))
			s.SyncPlayerApplyGuild_Time = make([]int64, len(aplys))
			for i, aply := range aplys {
				s.SyncPlayerApplyGuild_UUID[i] = aply.GuildUuid
				s.SyncPlayerApplyGuild_Name[i] = aply.GuildName
				s.SyncPlayerApplyGuild_Lvl[i] = aply.GuildLvl
				s.SyncPlayerApplyGuild_Notice[i] = aply.GuildNotice
				s.SyncPlayerApplyGuild_Time[i] = aply.ApplyTime
			}
		}
	}
	if guildID != "" && p.GuildProfile.WorshipInfo.CheckDailyReset(p.GetProfileNowTime()) {
		s.SyncGuildWorshipNeed = true
	}
	// 保证只去guild读一次数据
	var guildInfo *guild.GuildInfo
	nowTime := time.Now().Unix()
	isNeedSync := s.SyncGuildInfoNeed || s.SyncGuildMemsNeed ||
		s.SyncApplyGuildMemsNeed || s.SyncGuildInventory || s.SyncGuildScienceNeed ||
		s.SyncGuildWorshipNeed || s.SyncGuildRedPacketNeed

	if isNeedSync && guildID != "" {
		bs := time.Now().UnixNano()
		res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
		metric_send(p.AccountID, "GuildGetInfo", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
		if code.HasError() {
			logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
				code, guildID)
		} else {
			guildInfo = res
		}
	}

	if s.SyncGuildInfoNeed && guildID != "" {
		vipLv := int(p.Profile.GetVipLevel())
		nowT := p.Profile.GetProfileNowTime()
		// guild info
		if guildInfo != nil {
			gettedReward, allReward := guildInfo.GatesEnemyData.GetRewardCount()
			bs := time.Now().UnixNano()
			pos, _ := rank.GetModule(p.AccountID.ShardId).RankGuildGs.GetPos(guildID)
			metric_send(p.AccountID, "RankGuildGsGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
			data := guildBasicInfoToClient{
				GuildUUID:                  guildInfo.Base.GuildUUID,
				GuildID:                    guild.GuildItoa(guildInfo.Base.GuildID),
				Name:                       guildInfo.Base.Name,
				Level:                      guildInfo.Base.Level,
				Icon:                       guildInfo.Base.Icon,
				Rank:                       pos,
				GuildGSSum:                 guildInfo.Base.GuildGSSum,
				MemNum:                     uint32(guildInfo.Base.MemNum),
				MaxMem:                     uint32(guildInfo.Base.MaxMemNum),
				ApplyGsLimit:               guildInfo.Base.ApplyGsLimit,
				ApplyAuto:                  guildInfo.Base.ApplyAuto,
				Notice:                     guildInfo.Base.Notice,
				Exp:                        guildInfo.Base.XpCurr,
				NextExp:                    gamedata.GetGuildXpNeedNext(guildInfo.Base.Level),
				GatesEnemyCount:            guildInfo.Base.GetGateEnemyCount(),
				GatesEnemyMemCount:         allReward,
				GatesEnemyMemRewardedCount: gettedReward,
				SignCount:                  p.GuildProfile.GetGuildSignCount(vipLv, nowT),
				RenameTimes:                guildInfo.Base.RenameTimes,
				GuildTmpVer:                guildInfo.Base.GuildTmpVer,
			}
			s.SyncGuildInfo = encode(data)
		}
	}

	if guildID != "" && s.IsNeedSyncGuildActBoss() {
		s.makeGuildBossData(p, nowTime, guildInfo, guildID)
	}

	if s.SyncGuildMemsNeed && guildID != "" && guildInfo != nil {
		s.SyncGuildMems = make([][]byte, 0, len(guildInfo.Members))
		for i := 0; i < int(guildInfo.Base.MemNum); i++ {
			member := guildInfo.Members[i]
			guildBossDamage := int(guildInfo.ActBoss.LastDayDamages.GetSorce(member.AccountID))
			m := guildMemToClient{
				Acid:                    member.AccountID,
				Name:                    member.Name,
				Level:                   member.CorpLv,
				Gs:                      member.CurrCorpGs,
				Vip:                     member.Vip,
				CurrAvatar:              member.CurrAvatar,
				Position:                member.GuildPosition,
				LastLoginTime:           member.LastLoginTime,
				IsGettedGateEnemyReward: guildInfo.GatesEnemyData.GetGetRewardTime(member.AccountID) > 0,
				GuildContribution:       member.Contribution[0],
				GuildSp:                 member.GuildSp,
				GuildSignLastTs:         member.Contribution[1],
				Online:                  member.GetOnline(),
				GuildBossDamage:         guildBossDamage,
				GVGScore:                member.GVGScore,
			}
			if !gamedata.IsSameDayCommon(p.Profile.GetProfileNowTime(), member.Contribution[1]) {
				// 如果数据已经过期一天的话
				m.GuildContribution = 0
			}
			s.SyncGuildMems = append(s.SyncGuildMems, encode(m))
		}
	}

	if s.SyncApplyGuildMemsNeed && guildID != "" &&
		guildInfo != nil && gamedata.CheckApprovePosition(p.GuildProfile.GuildPosition) {
		bs := time.Now().UnixNano()
		aplys := guild.GetModule(p.AccountID.ShardId).GetGuildApplyInfo(guildID, p.AccountID.String())
		metric_send(p.AccountID, "GetGuildApplyInfo", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
		if aplys != nil {
			nowTime := time.Now().Unix()
			s.SyncGuildApplyList = make([][]byte, 0, len(aplys))
			for _, aply := range aplys {
				member := aply.PlayerInfo
				m := guildMemToClient{
					Acid:              member.AccountID,
					Name:              member.Name,
					Level:             member.CorpLv,
					Gs:                member.CurrCorpGs,
					Position:          member.GuildPosition,
					LastLoginTime:     member.LastLoginTime,
					GuildContribution: member.Contribution[0],
				}
				if !gamedata.IsSameDayCommon(nowTime, member.Contribution[1]) {
					// 如果数据已经过期一天的话
					m.GuildContribution = 0
				}
				s.SyncGuildApplyList = append(s.SyncGuildApplyList, encode(m))
			}
		}
		if len(s.SyncGuildApplyList) > 0 {
			s.OnChangeRedPoint(notify.RedPointTyp_Guild)
		} else {
			p.GuildProfile.HasApplyCanApprove = false
		}
	}

	if s.IsChangeGatesEnemyData() {
		gatesEnemyData := resetGatesEnemyDataIfNil(p, guildInfo, guildID)
		s.makeGatesEnemyInfoData(p, gatesEnemyData)
	}

	if s.IsChangeGatesEnemyPushData() {
		gatesEnemyData := resetGatesEnemyDataIfNil(p, guildInfo, guildID)
		s.makeGatesEnemyPushData(gatesEnemyData)
	}

	// 兵临城下开启结束时间，每个协议都带上
	s.SyncGateEnemyStartTime, s.SyncGateEnemyEndTime = p.Profile.GetGatesEnemy().GetActTime(nowTime)

	if s.SyncClientTagNeed {
		tags := p.Profile.GetClientTagInfo().TagState
		s.SyncClientTag = make([]int, len(tags))
		for i, t := range tags {
			s.SyncClientTag[i] = t
		}
	}

	// 切磋就一个红点，先放这吧
	if p.Profile.GetGank().IsRedPoint() {
		s.OnChangeRedPoint(notify.RedPointTyp_Gank)
	}

	// 公会仓库
	if s.SyncGuildInventory && guildID != "" && guildInfo != nil {
		invy := guildInfo.Inventory
		s.SyncGuildInventoryNextRefTime = invy.NextRefTime
		s.SyncGuildInventoryNextResetTime = invy.NextResetTime
		s.SyncGuildInventoryBossCoin = p.Profile.GetSC().GetSC(gamedata.SC_GB)
		s.SyncGuildInventoryLoots = make([][]byte, len(invy.Loots))
		s.SyncGuildInventoryLootMemCount = make([]int, len(invy.Loots))
		s.SyncGuildInventoryLootMemAssign = make([][]byte, 0)
		s.SyncGuildInventorySelfApplyCount = make([]int, len(invy.Loots))
		s.SyncGuildInventoryPreLoots = make([][]byte, 0, len(invy.PrepareLoots))
		for i, l := range invy.Loots {
			//if l.Loot.Count <= 0 { // 客户端需要数量为0的,方便ui操作
			//	continue
			//}
			s.SyncGuildInventoryLoots[i] = encode(l.Loot)
			s.SyncGuildInventoryLootMemCount[i] = len(l.AssignMems)
			selfC := 0
			for index, m := range l.AssignMems {
				if m.Acid == p.AccountID.String() {
					selfC++
				}
				s.SyncGuildInventoryLootMemAssign = append(s.SyncGuildInventoryLootMemAssign, encode(l.AssignMems[index]))
			}
			s.SyncGuildInventorySelfApplyCount[i] = selfC
		}
		for _, prel := range invy.PrepareLoots {
			s.SyncGuildInventoryPreLoots = append(s.SyncGuildInventoryPreLoots, encode(prel))

		}

		s.SyncGuildLostInventoryLoots = make([][]byte, len(guildInfo.LostInventory.Loots))
		s.SyncGuildLostInventoryLootMemCount = make([]int, len(guildInfo.LostInventory.Loots))
		s.SyncGuildLostInventoryLootMemAssign = make([][]byte, 0)
		s.SyncGuildLostInventorySelfApplyCount = make([]int, len(guildInfo.LostInventory.Loots))
		for i, l := range guildInfo.LostInventory.Loots {
			s.SyncGuildLostInventoryLoots[i] = encode(l.Loot)
			s.SyncGuildLostInventoryLootMemCount[i] = len(l.ApplyAcids)
			selfC := 0
			for _, m := range l.ApplyAcids {
				if m.Acid == p.AccountID.String() {
					selfC++
				}
			}
			s.SyncGuildLostInventorySelfApplyCount[i] = selfC
		}
		s.SyncGuildLostInventoryActive = guildInfo.LostInventory.LostActive
	}

	// 工会科技
	if s.SyncGuildScienceNeed && guildID != "" && guildInfo != nil {
		s.SyncGuildScience = make([][]byte, 0, len(guildInfo.Sciences))
		for _, v := range guildInfo.Sciences {
			s.SyncGuildScience = append(s.SyncGuildScience, encode(v))
		}
	}
	if p.GuildProfile.RedPacketInfo.CheckDailyReset(p.GetProfileNowTime()) {
		s.SyncGuildRedPacketNeed = true
	}

	if s.SyncGuildRedPacketNeed || p.GuildProfile.RedPacketInfo.Sync.IsNeedSync() {
		// 公会红包 红包不在公会界面， 所以可能会单独请求
		if guildID != "" && guildInfo == nil {
			res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
			if !code.HasError() {
				guildInfo = res
			}
		}
		s.mkGuildRedPacket(p, guildInfo)
	}

	// 军团膜拜
	if s.SyncGuildWorshipNeed && guildID != "" && guildInfo != nil {
		gg := &p.GuildProfile.WorshipInfo
		s.PersionTakeNum = gg.PersionTakeNum
		s.GuildTakeNum = guildInfo.GuildInfoBase.GuildWorship.WorshipIndex

		s.TakeSign = gg.TakeSign
		s.Reward = gg.Reward
		s.OneTime = gg.OneTime
		s.DoubleTime = gg.DoubleTime
		s.HasReward = gg.HasRewards
		s.IsOpen = guildInfo.GuildInfoBase.GuildWorship.IsOpen
		for _, x := range guildInfo.GuildWorship.WorshipMember {
			if gg.WorshipAccoundID == x.MemberAccountId {
				s.TakeId = gg.TakeId
				break
			} else {
				s.TakeId = -1
			}
		}
	}
}

func (s *SyncRespNotify) makeGatesEnemyInfoData(p *Account, gatesEnemyData *player_msg.PlayerMsgGatesEnemyData) {
	s.SyncGatesEnemyNeed = true
	s.SyncGatesEnemyBossState = p.Profile.GetGatesEnemy().CurrBoss
	s.SyncGatesEnemyRankNames = gatesEnemyData.Names
	s.SyncGatesEnemyRankPoints = gatesEnemyData.Points
	s.SyncGatesEnemyRankFashion = make([]string, 0, len(gatesEnemyData.Fashion)*2+1)
	for i := 0; i < len(gatesEnemyData.Fashion); i++ {
		s.SyncGatesEnemyRankFashion = append(s.SyncGatesEnemyRankFashion, gatesEnemyData.Fashion[i][0])
		s.SyncGatesEnemyRankFashion = append(s.SyncGatesEnemyRankFashion, gatesEnemyData.Fashion[i][1])
	}
	s.SyncGatesEnemyRankWeaponStartLvl = gatesEnemyData.WeaponStartLvl
	s.SyncGatesEnemyRankEqStartLvl = gatesEnemyData.EqStartLvl
	s.SyncGatesEnemyRankStates = gatesEnemyData.MemStats
	s.SyncGatesEnemyRankAvatarIDs = gatesEnemyData.AvatarID
	s.SyncGatesEnemyRankTitleOn = gatesEnemyData.TitleOn
	s.SyncGatesEnemyPointAll = gatesEnemyData.Point
	s.SyncGatesEnemySwing = gatesEnemyData.Swing
	s.SyncGatesEnemyMagicPet = gatesEnemyData.MagicPetfigure
}

func (s *SyncRespNotify) makeGatesEnemyPushData(gatesEnemyData *player_msg.PlayerMsgGatesEnemyData) {
	s.SyncGatesEnemyEnemyInfo = gatesEnemyData.EnemyInfo
	s.SyncGatesEnemyState = gatesEnemyData.State
	s.SyncGatesEnemyStateOverTime = gatesEnemyData.StateOverTime
	s.SyncGatesEnemyKillPoint = gatesEnemyData.KillPoint
	s.SyncGatesEnemyBossMax = gatesEnemyData.BossMax
	s.SyncGatesEnemyBuffMemName = gatesEnemyData.BuffMemName[:]
	s.SyncGatesEnemyBuffCurLv = gatesEnemyData.BuffCurLv
}

func resetGatesEnemyDataIfNil(p *Account, guildInfo *guild.GuildInfo, guildID string) *player_msg.PlayerMsgGatesEnemyData {
	gatesEnemyData := p.Profile.GetGatesEnemy().GetPushData()
	// gatesEnemy不存库， 重登后为空, 这里从工会里面重新复制
	if gatesEnemyData.IsNil() {
		if guildInfo == nil && guildID != "" {
			bs := time.Now().UnixNano()
			res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
			metric_send(p.AccountID, "GetGuildInfo", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
			if code.HasError() {
				logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
					code, guildID)
			} else {
				guildInfo = res
			}
		}
		if guildInfo != nil {
			gatesEnemyData = guildInfo.GatesEnemyData.ToClient()
			p.Profile.GetGatesEnemy().SetPushData(*gatesEnemyData) // 下面的逻辑不会修改p里面的值， 但是下次请求就可以了
		}
	}
	// 保护代码, 删档后应该删除
	if gatesEnemyData.TitleOn == nil {
		gatesEnemyData.TitleOn = make([]string, len(gatesEnemyData.Names))
		for i := 0; i < len(gatesEnemyData.TitleOn); i++ {
			gatesEnemyData.TitleOn[i] = ""
		}
	}
	return gatesEnemyData
}

func (s *SyncRespNotify) makeGuildBossData(p *Account, nowTime int64, guildInfo *guild.GuildInfo, guildID string) {
	playerGuildBossInfo := p.Profile.GetGuildBossInfo()
	if nowTime-playerGuildBossInfo.RefreshTime > 5 || playerGuildBossInfo.Info == nil {
		playerGuildBossInfo.RefreshTime = nowTime
		if guildInfo == nil {
			res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
			if code.HasError() {
				logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
					code, guildID)
				return
			} else {
				guildInfo = res
			}
		}
		if guildInfo != nil {
			s.OnChangeGuildActBoss(guildInfo.ActBoss.ToClient(nowTime))
		}
	} else {
		s.OnChangeGuildActBoss(p.Profile.GetGuildBossInfo().Info)
	}
}
