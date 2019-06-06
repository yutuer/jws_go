package logics

import (
	"time"

	"strconv"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/world_boss"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/worldboss"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

// GetWBInfo : 获取世界boss信息
//
func (p *Account) GetWBInfoHandler(req *reqMsgGetWBInfo, resp *rspMsgGetWBInfo) uint32 {
	status, code, err := worldboss.GetInfo(p.AccountID.ShardId, p.AccountID.String())
	if code != crossservice.ErrOK {
		logs.Error("get worldboss info err, code: %d, err: %v", code, err)
		return errCode.CommonInner
	}
	logs.Debug("get wbInfo ret: %v", *status)
	p.ResetWBGotInfo(p.Profile.GetProfileNowTime(), &resp.SyncResp)
	resp.BossID = status.Boss.BossID
	resp.BossScene = status.Boss.SceneID
	resp.BossLevel = int64(status.Boss.Level)
	resp.BossDamaged = int64(status.Boss.HPMax) - int64(status.Boss.HPCurr)
	resp.SelfDamage = int64(status.MyDamage)
	pos := status.MyPos
	if pos > gamedata.GetMaxRankLimit() {
		pos = 0
	}
	resp.SelfRank = int64(pos)
	wbd := p.Profile.GetWorldBossData()
	resp.BuffLevel = int64(wbd.BuffLevel)
	resp.LeftTimes = int64(wbd.BattleTimes) // warn: 已修改为已挑战次数
	resp.GotRewards = wbd.GotRewards
	return 0
}

// BeginWB : 开始挑战世界boss
//
func (p *Account) BeginWBHandler(req *reqMsgBeginWB, resp *rspMsgBeginWB) uint32 {
	wbd := p.Profile.GetWorldBossData()
	// check condition
	nowT := p.Profile.GetProfileNowTime()
	unixT := time.Unix(nowT, 0)
	// time
	if nowT < gamedata.GetTodayWBStartTime(unixT) || nowT >= gamedata.GetTodayWBEndTime(unixT) {
		return errCode.CommonNotInTime
	}
	// corp
	team := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_WORLD_BOSS)
	if len(team) < 1 {
		return errCode.HeroTeamWarn
	}
	for _, item := range team {
		if !p.isLegalCorp(item, unixT) {
			return errCode.HeroTeamWarn
		}
	}
	v, _ := p.Profile.GetVip().GetVIP()

	// times
	if int64(gamedata.GetVIPCfg(int(v)).WorldBosstimes)-int64(wbd.BattleTimes) <= 0 {
		return errCode.CommonCountLimit
	}

	heroGS := 0
	GS := p.Profile.GetData().HeroGs
	for _, item := range team {
		heroGS += GS[item]
	}
	playerInfo := &worldboss.PlayerInfo{
		Acid:      p.AccountID.String(),
		Sid:       uint32(p.AccountID.ShardId),
		Vip:       v,
		Name:      etcd.ParseDisplayShardName2(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), p.AccountID.ShardId) + "-" + p.Profile.Name,
		Level:     p.Profile.GetCorp().Level,
		Gs:        int64(heroGS),
		GuildName: p.GuildProfile.GuildName,
	}
	status, code, err := worldboss.Join(p.AccountID.ShardId, p.AccountID.String(), playerInfo)
	if code != crossservice.ErrOK {
		logs.Error("begin worldboss battle err, code: %d, err: %v", code, err)
		return errCode.CommonInner
	}
	logs.Debug("begin battle wbInfo ret: %v", *status)

	resp.BossID = status.Boss.BossID
	resp.BossHP = int64(status.Boss.HPCurr)
	resp.BossLevel = int64(status.Boss.Level)
	resp.BossDamaged = int64(status.TotalDamage)
	resp.SelfScore = int64(status.MyDamage)
	resp.SelfRank = int64(status.MyPos)
	resp.Seq = int64(status.Boss.Seq)
	resp.WBFewRankInfos = make([][]byte, 0)
	for _, item := range status.Rank {
		resp.WBFewRankInfos = append(resp.WBFewRankInfos, encode(
			WBFewRankInfo{
				Acid:  item.Acid,
				Rank:  int64(item.Pos),
				Name:  item.Name,
				Score: int64(item.Damage),
			}))
	}
	wbd.Damage = status.MyDamage
	wbd.SetHadCostTimes(false)
	wbd.State = world_boss.Battle
	wbd.CurDamage = 0
	wbd.StartBattleTime = p.Profile.GetProfileNowTime()
	_, heroAttrs, _, _, _, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(p.Account))
	maxAttr := float32(-1.0)
	for _, item := range team {
		if heroAttrs[item].ATK > maxAttr {
			maxAttr = heroAttrs[item].ATK
		}
	}
	wbd.MaxHeroATK = maxAttr
	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.BeginWorldBoss{
		AccountID: p.AccountID.String(),
		GS:        p.Profile.GetData().CorpCurrGS,
		AvatarID:  team,
		BuffLevel: wbd.BuffLevel,
		VIP:       int(p.Profile.GetVip().V),
	}, logiclog.LogicTag_BeginWorldBoss, "")
	return 0
}

// EndWB : 结束挑战世界boss
//
func (p *Account) EndWBHandler(req *reqMsgEndWB, resp *rspMsgEndWB) uint32 {
	wbd := p.Profile.GetWorldBossData()
	oldBuffLevel := wbd.BuffLevel
	if wbd.IsHadCostTimes() {
		wbd.BuffLevel = 0
	}
	resp.BuffLevel = int64(wbd.BuffLevel)
	v, _ := p.Profile.GetVip().GetVIP()
	resp.LeftTimes = int64(gamedata.GetVIPCfg(int(v)).WorldBosstimes) - int64(wbd.BattleTimes)
	resp.RemindBuyBuff = !wbd.DisableRemindBuyBuff
	resp.RoundDamage = wbd.CurDamage
	// corp
	if wbd.State != world_boss.Battle {
		return 0
	}
	team := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_WORLD_BOSS)
	if len(team) < 1 {
		return 0
	}
	maxAttr := float32(-1.0)
	heroData := p.Profile.GetData()
	for _, item := range team {
		if heroData.HeroBaseAttrs[item].ATK > maxAttr {
			maxAttr = heroData.HeroBaseAttrs[item].ATK
		}
	}
	isCheat := false
	if wbd.MaxHeroATK != -1 {
		maxDamage := world_boss.AntiCheatAllDamage(wbd.CurDamage, wbd.MaxHeroATK, oldBuffLevel,
			p.Profile.GetProfileNowTime()-wbd.StartBattleTime)
		if wbd.CurDamage > maxDamage {
			logs.Warn("player cheat, round damage: %v, maDamage: %v", wbd.CurDamage, maxDamage)
		}
		logs.Debug("cheat info: %v, %v, %v, %v", wbd.CurDamage, wbd.MaxHeroATK, oldBuffLevel,
			p.Profile.GetProfileNowTime()-wbd.StartBattleTime)
	}
	_, code, err := worldboss.Leave(p.AccountID.ShardId, p.AccountID.String(), p.GenTeamInfoDetail(team, oldBuffLevel), isCheat)
	if code != crossservice.ErrOK {
		logs.Error("end worldboss battle err, code: %d, err: %v", code, err)
		return errCode.CommonInner
	}
	wbd.State = world_boss.Idle
	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.EndWorldBoss{
		AccountID: p.AccountID.String(),
		GS:        p.Profile.GetData().CorpCurrGS,
		AvatarID:  team,
		VIP:       int(p.Profile.GetVip().V),
	}, logiclog.LogicTag_EndWorldBoss, "")
	if isCheat {
		return errCode.YouCheat
	}
	return 0
}

// GetWBRankInfo : 获取世界boss伤害排行榜信息
//
func (p *Account) GetWBRankInfoHandler(req *reqMsgGetWBRankInfo, resp *rspMsgGetWBRankInfo) uint32 {
	const (
		global    = 1
		formation = 2
	)
	if req.Typ == global {
		status, code, err := worldboss.GetRank(p.AccountID.ShardId, p.AccountID.String())
		if code != crossservice.ErrOK {
			logs.Error("get worldboss rank info err, code: %d, err: %v", code, err)
			return errCode.CommonInner
		}
		logs.Debug("get rank wbInfo ret: %v", *status)
		// self
		ids := make([]int64, 0)
		star := make([]int64, 0)
		for _, hero := range status.MyRank.Team {
			ids = append(ids, int64(hero.Idx))
			star = append(star, int64(hero.StarLevel))
		}
		pos := status.MyRank.Pos
		if pos > gamedata.GetMaxRankLimit() {
			pos = 0
		}
		resp.SelfWBRankInfo = encode(WBRankInfo{
			Acid:     p.AccountID.String(),
			Rank:     int64(pos),
			Name:     p.Profile.Name,
			Score:    int64(status.MyRank.Damage),
			HeroID:   ids,
			HeroStar: star,
		})
		// rank
		resp.WBRankInfos = make([][]byte, 0)
		for _, item := range status.Top {
			ids := make([]int64, 0)
			star := make([]int64, 0)
			for _, hero := range item.Team {
				ids = append(ids, int64(hero.Idx))
				star = append(star, int64(hero.StarLevel))
			}
			resp.WBRankInfos = append(resp.WBRankInfos, encode(
				WBRankInfo{
					Acid:     item.Acid,
					Rank:     int64(item.Pos),
					Name:     item.Name,
					Score:    int64(item.Damage),
					HeroID:   ids,
					HeroStar: star,
				}))
		}
	} else if req.Typ == formation {
		status, code, err := worldboss.GetFormationRank(p.AccountID.ShardId, p.AccountID.String())
		if code != crossservice.ErrOK {
			logs.Error("get worldboss formation rank info err, code: %d, err: %v", code, err)
			return errCode.CommonInner
		}
		logs.Debug("get formation rank wbInfo ret: %v", *status)
		// self
		ids := make([]int64, 0)
		star := make([]int64, 0)
		for _, hero := range status.MyRank.Team {
			ids = append(ids, int64(hero.Idx))
			star = append(star, int64(hero.StarLevel))
		}
		pos := status.MyRank.Pos
		if pos > gamedata.GetMaxRankLimit() {
			pos = 0
		}
		resp.SelfWBRankInfo = encode(WBRankInfo{
			Acid:      p.AccountID.String(),
			Rank:      int64(pos),
			Name:      p.Profile.Name,
			Score:     int64(status.MyRank.Damage),
			HeroID:    ids,
			HeroStar:  star,
			BuffLevel: int64(status.MyRank.BuffLevel),
		})
		// rank
		resp.WBRankInfos = make([][]byte, 0)
		for _, item := range status.Top {
			ids := make([]int64, 0)
			star := make([]int64, 0)
			for _, hero := range item.Team {
				ids = append(ids, int64(hero.Idx))
				star = append(star, int64(hero.StarLevel))
			}
			resp.WBRankInfos = append(resp.WBRankInfos, encode(
				WBRankInfo{
					Acid:      item.Acid,
					Rank:      int64(item.Pos),
					Name:      item.Name,
					Score:     int64(item.Damage),
					HeroID:    ids,
					HeroStar:  star,
					BuffLevel: int64(item.BuffLevel),
				}))
		}
	}

	return 0
}

// UpdateBattleInfo : 战斗中更新自己的伤害量、boss血量、伤害排行榜
//
func (p *Account) UpdateBattleInfoHandler(req *reqMsgUpdateBattleInfo, resp *rspMsgUpdateBattleInfo) uint32 {
	gs := p.Profile.GetData().HeroGs
	anticheatGsCfg := game.Cfg.AntiCheat[account.CheckerHeroGS]
	antiCheatGsThreshold := anticheatGsCfg.ParamInt
	logs.Debug("anticheat gs threshold: %v", antiCheatGsThreshold)
	if len(req.CheatParam)%2 == 0 {
		for i := 0; i < len(req.CheatParam); i += 2 {
			if int(req.CheatParam[i]) > len(gs) {
				logs.Error("wb cheat param err, param: %v", req.CheatParam)
				break
			}
			if req.CheatParam[i+1]-int64(gs[req.CheatParam[i]]) > antiCheatGsThreshold {
				logs.Debug("cheat client gs: %v, server gs: %v", req.CheatParam[i+1], gs[req.CheatParam[i]])
				resp.IsCheat = true
				return 0
			}
		}
	}
	wbd := p.Profile.GetWorldBossData()
	if wbd.State != world_boss.Battle {
		resp.BossID = ""
		return 0
	}
	damage := req.SelfDamage
	if p.Profile.GetProfileNowTime()-wbd.StartBattleTime >= int64(gamedata.GetWBBattleValidTime()) {
		damage = 0
	}

	if wbd.MaxHeroATK != -1 {
		oldDamage := damage
		damage = world_boss.AntiCheatSingleDamage(oldDamage, wbd.MaxHeroATK, wbd.BuffLevel)
		if oldDamage > damage {
			logs.Debug("cheat info: %v, %v, %v", oldDamage, wbd.MaxHeroATK, wbd.BuffLevel)
			logs.Warn("player damage is too high for %v, %v", damage, oldDamage)
		}
	}
	attackInfo := &worldboss.AttackInfo{
		Damage: uint64(damage),
		Level:  uint32(req.BossLevel),
	}
	status, code, err := worldboss.Attack(p.AccountID.ShardId, p.AccountID.String(), attackInfo)
	if code != crossservice.ErrOK {
		logs.Error("update worldboss battle info err, code: %d, err: %v", code, err)
		resp.BossID = ""
		return 0
	}
	logs.Debug("update battle wbInfo ret: %v", *status)

	resp.BossID = status.Boss.BossID
	resp.BossHP = int64(status.Boss.HPCurr)
	resp.BossLevel = int64(status.Boss.Level)
	resp.BossDamaged = int64(status.TotalDamage)
	pos := status.MyPos
	if pos > gamedata.GetMaxRankLimit() {
		pos = 0
	}
	resp.SelfScore = int64(status.MyDamage)
	resp.SelfRank = int64(pos)
	resp.Seq = int64(status.Boss.Seq)
	resp.WBFewRankInfos = make([][]byte, 0)
	for _, item := range status.Rank {
		resp.WBFewRankInfos = append(resp.WBFewRankInfos, encode(
			WBFewRankInfo{
				Acid:  item.Acid,
				Rank:  int64(item.Pos),
				Name:  item.Name,
				Score: int64(item.Damage),
			}))
	}
	wbd.SetDamageInDay(status.MyDamage, time.Unix(p.Profile.GetProfileNowTime(), 0))
	wbd.CurDamage = int64(status.DamageRound)
	if req.SelfDamage > 0 {
		if !wbd.IsHadCostTimes() {
			wbd.SetHadCostTimes(true)
			wbd.BattleTimes++
			p.updateCondition(account.COND_TYP_WorldBoss, 1, 0, "", "", resp)
		}
	}
	return 0
}

// UseBuff : 使用上古之力
//
func (p *Account) UseBuffHandler(req *reqMsgUseBuff, resp *rspMsgUseBuff) uint32 {
	wbd := p.Profile.GetWorldBossData()
	wbc := gamedata.GetWBConfig()
	if wbd.BuffLevel+int(req.UseCount) > int(wbc.GetBuffNumLimit()) {
		return errCode.CommonMaxLimit
	}
	costData := gamedata.CostData{}
	costData.AddItem(wbc.GetBuffCost(), wbc.GetItemNum()*uint32(req.UseCount))
	if !account.CostBySync(p.Account, &costData, resp, "buy world boss buf") {
		return errCode.CommonLessMoney
	}
	wbd.BuffLevel += int(req.UseCount)
	resp.BuffLevel = int64(wbd.BuffLevel)
	return 0
}

// GetRankRewards : 领取排行奖励
//
func (p *Account) GetWBRankRewardsHandler(req *reqMsgGetWBRankRewards, resp *rspMsgGetWBRankRewards) uint32 {
	wbd := p.Profile.GetWorldBossData()
	if wbd.HadGetRewards(int(req.LevelID)) {
		return errCode.RewardFail
	}
	status, code, err := worldboss.GetInfo(p.AccountID.ShardId, p.AccountID.String())
	if code != crossservice.ErrOK {
		logs.Error("get worldboss info err, code: %d, err: %v", code, err)
		return errCode.CommonInner
	}

	damageReward := gamedata.GetWBDamageRewards(uint32(req.LevelID))
	if status.MyDamage <= uint64(damageReward.GetNeedDemage()) {
		return errCode.CommonConditionFalse
	}
	wbd.SetRewards(int(req.LevelID))
	giveData := gamedata.CostData{}
	for _, item := range damageReward.GetLoot_Table() {
		giveData.AddItem(item.GetItemID(), item.GetItemNum())
	}
	if !account.GiveBySync(p.Account, &giveData, resp, "buy world boss buf") {
		return errCode.CommonLessMoney
	}
	resp.GotRewards = wbd.GotRewards
	return 0
}

func (p *Account) GetWBPlayerDetailHandler(req *reqMsgGetWBPlayerDetail, resp *rspMsgGetWBPlayerDetail) uint32 {
	ac, err := db.ParseAccount(req.AcID)
	if err != nil {
		return errCode.CommonInvalidParam
	}
	status, code, err := worldboss.PlayerDetail(p.AccountID.ShardId, p.AccountID.String(), req.AcID)
	if code != crossservice.ErrOK {
		logs.Error("get world boss player detail info err, code: %d, err: %v", code, err)
		return errCode.CommonInner
	}
	logs.Debug("get world boss player detail info: %v", status)
	/*********************** danger code, will remove************/
	if len(status.TeamInfoDetail.EquipAttr) == 0 {
		status.TeamInfoDetail.EquipAttr = make([]int64, 3)
	}
	if len(status.TeamInfoDetail.DestinyAttr) == 0 {
		status.TeamInfoDetail.DestinyAttr = make([]int64, 3)
	}
	if len(status.TeamInfoDetail.JadeAttr) == 0 {
		status.TeamInfoDetail.JadeAttr = make([]int64, 3)
	}
	/*************************************************************/
	info := WSPVPRankInfo{
		Sid:               strconv.Itoa(int(ac.ShardId)),
		Name:              status.PlayerInfo.Name,
		VipLevel:          int64(status.PlayerInfo.Vip),
		CorpLevel:         int64(status.PlayerInfo.Level),
		AllGs:             status.PlayerInfo.Gs,
		GuildName:         status.PlayerInfo.GuildName,
		BestHeroIdx:       make([]int64, 0),
		BestHeroLevel:     make([]int64, 0),
		BestHeroStarLevel: make([]int64, 0),
		BestHeroBaseGs:    make([]int64, 0),
		BestHeroExtraGs:   make([]int64, 0),
		EquipAttr:         status.TeamInfoDetail.EquipAttr,
		DestinyAttr:       status.TeamInfoDetail.DestinyAttr,
		JadeAttr:          status.TeamInfoDetail.JadeAttr,
	}
	for _, item := range status.TeamInfoDetail.Team {
		info.BestHeroIdx = append(info.BestHeroIdx, int64(item.Idx))
		info.BestHeroLevel = append(info.BestHeroLevel, int64(item.Level))
		info.BestHeroStarLevel = append(info.BestHeroStarLevel, int64(item.StarLevel))
		info.BestHeroBaseGs = append(info.BestHeroBaseGs, item.BaseGs)
		info.BestHeroExtraGs = append(info.BestHeroExtraGs, item.ExtraGs)
	}
	resp.DetailInfo = encode(info)
	return 0
}

func (p *Account) SetBuyBuffReminderHandler(req *reqMsgSetBuyBuffReminder, resp *rspMsgSetBuyBuffReminder) uint32 {
	wbd := p.Profile.GetWorldBossData()
	wbd.DisableRemindBuyBuff = !req.RemindBuyBuff
	resp.RemindBuyBuff = !wbd.DisableRemindBuyBuff
	return 0
}

func (p *Account) ResetWBGotInfo(nowT int64, rsp *SyncResp) {
	wbd := p.Profile.GetWorldBossData()
	rt := gamedata.GetTodayWBResetTime(time.Unix(nowT, 0))
	if nowT >= rt && rt != wbd.LastResetTime {
		// send reward
		items := make(map[string]uint32, 0)
		rewards := gamedata.GetWBDamageRewardsWithDmg(wbd.Damage)
		for _, reward := range rewards {
			if wbd.HadGetRewards(int(reward.GetID())) {
				continue
			}
			for _, loot := range reward.GetLoot_Table() {
				items[loot.GetItemID()] += loot.GetItemNum()
			}
		}
		if len(items) > 0 {
			// send mail
			mail_sender.BatchSendMail2Account(p.AccountID.String(), timail.Mail_send_By_Common,
				mail_sender.IDS_MAIL_WB_DPSREWARD_TITLE,
				[]string{}, items, "WorldBossDamageRewards", false)
		}
		wbd.Reset()
		p.Profile.GetHeroTeams().ResetHeroTeam(gamedata.LEVEL_TYPE_WORLD_BOSS)
		rsp.OnChangeHeroTeam()
		wbd.LastResetTime = rt
		logs.Debug("reset wb got info")
	}
}

func (p *Account) isLegalCorp(hero int, nowT time.Time) bool {
	//限制阵容（1魏，2蜀，3吴，4群，5男，6女，0不限）
	const (
		ALL = iota
		WEI
		SHU
		WU
		QUN
		MALE
		FEMALE
	)
	info := gamedata.GetHeroData(hero)
	switch gamedata.GetTodayValidHero(nowT) {
	case ALL:
		return true
	case WEI:
		if info.Nationality == 2 {
			return true
		}
	case SHU:
		if info.Nationality == 1 {
			return true
		}
	case WU:
		if info.Nationality == 3 {
			return true
		}
	case QUN:
		if info.Nationality == 4 {
			return true
		}
	case MALE:
		if info.Sex == 1 {
			return true
		}
	case FEMALE:
		if info.Sex == 0 {
			return true
		}
	}
	logs.Error("illegal world boss limit data")
	return false
}
