package logics

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/buy"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// SetWSPVPDefenseFormation : 布阵无双争霸防守阵容
// 无双争霸防守阵容
func (p *Account) SetWSPVPDefenseFormationHandler(req *reqMsgSetWSPVPDefenseFormation, resp *rspMsgSetWSPVPDefenseFormation) uint32 {
	formation := req.AvatarId
	logs.Debug("set formation from client, %v", formation)
	if !p.isFormationAvailable(formation) {
		return errCode.WSPVPFormationError
	}
	p.Profile.WSPVPPersonalInfo.SetFormation(formation)
	_, heroAttrs, _, heroGs, _, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(p.Account))
	formationInfo := p.makeFormationInfo(heroAttrs, heroGs)
	ws_pvp.SaveFormation2DB(p.GetWSPVPGroupId(), p.AccountID.String(), formationInfo)
	resp.OnChangeWSPVP()
	return 0
}

func (p *Account) makeFormationInfo(heroAttr map[int]*gamedata.AvatarAttr, heroGs map[int]int) ws_pvp.WSPVPFormationInfo {
	formation := p.GetDefenseFormation()
	resInfo := ws_pvp.WSPVPFormationInfo{
		Avatar: make([]ws_pvp.WSPVPHeroInfo, 0),
		Group:  formation,
	}

	for _, idx := range formation {
		idxInt := int(idx)
		if idxInt == -1 {
			continue
		}
		attr := heroAttr[idxInt]

		heroInfo := ws_pvp.WSPVPHeroInfo{
			Idx:            idxInt,
			Attr:           encode(attr),
			StarLevel:      int(p.Profile.GetHero().GetStar(idxInt)),
			Gs:             int64(heroGs[idxInt]),
			AvatarSkills:   p.Profile.GetAvatarSkill().GetByAvatar(idxInt),
			AvatarFashion:  p.getEquipFashionTids(idxInt),
			HeroSwing:      p.Profile.GetHero().GetSwing(idxInt).CurSwing,
			MagicPetfigure: p.Profile.GetHero().GetMagicPetFigure(idxInt),
			PassiveSkillId: p.Profile.GetHero().HeroSkills[idx].PassiveSkill[:],
			CounterSkillId: p.Profile.GetHero().HeroSkills[idx].CounterSkill[:],
			TriggerSkillId: p.Profile.GetHero().HeroSkills[idx].TriggerSkill[:],
		}
		resInfo.Avatar = append(resInfo.Avatar, heroInfo)
	}
	resInfo.DesSkills = make([]int64, len(p.Profile.DestinyGenerals.SkillGenerals))
	for i := 0; i < len(resInfo.DesSkills); i++ {
		resInfo.DesSkills[i] = int64(p.Profile.DestinyGenerals.SkillGenerals[i])
	}
	return resInfo
}

func (p *Account) getEquipFashionTids(avatarId int) []string {
	fashionIds := p.Profile.GetAvatarEquips().CurrByAvatar(avatarId)
	result := make([]string, len(fashionIds))
	for i, id := range fashionIds {
		has, item := p.Profile.GetFashionBag().GetFashionInfo(id)
		if has {
			result[i] = item.TableID
		}
	}
	return result
}

func (p *Account) isFormationAvailable(formation []int64) bool {
	if len(formation) != 9 {
		return false
	}
	// 每3个为一队， 每队至少有一人
	count := [3]int{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if formation[3*i+j] >= 0 {
				count[i]++
			}
		}
	}
	if count[0] <= 0 || count[1] <= 0 || count[2] <= 0 {
		return false
	}
	// 检测是否有重复的
	formationMap := make(map[int64]int)
	for _, idx := range formation {
		if idx != -1 {
			formationMap[idx] = formationMap[idx] + 1
		}
	}
	for _, value := range formationMap {
		if value > 1 {
			return false
		}
	}
	return true
}

// GetMatchOpponent : 获取匹配对手
// 随机4个对手
func (p *Account) GetMatchOpponentHandler(req *reqMsgGetMatchOpponent, resp *rspMsgGetMatchOpponent) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	if p.isInLockTimeRange() {
		return errCode.WSPVPLockingTime
	}
	willRefresh, myNewRank, errCode := p.checkAndCostRefresh(req.ForceRefresh, resp)
	if errCode != 0 {
		return errCode
	}
	if myNewRank != -1 {
		p.Profile.WSPVPPersonalInfo.Rank = myNewRank
		resp.IsMyRankChanged = true
		resp.OnChangeWSPVP()
	}
	rank := p.Profile.WSPVPPersonalInfo.Rank
	var opponentSimples []*WSPVPOpp
	if willRefresh {
		opponentRanks := gamedata.GetWSPVPFourOppopent(rank)
		logs.Debug("got opponent rank, %d", opponentRanks)
		simpleInfos := ws_pvp.GetSimpleOppInfo(p.GetWSPVPGroupId(), opponentRanks[:])
		saveOpponents := make([]account.WSPVPOppSimpleInfo, 0)
		for i, sim := range simpleInfos {
			saveOpponents = append(saveOpponents, account.WSPVPOppSimpleInfo{
				Acid:      sim.Acid,
				Rank:      int64(opponentRanks[i]),
				ServerId:  int64(sim.ServerId),
				Name:      sim.Name,
				GuildName: sim.GuildName,
				TitleId:   sim.TitleId,
				VipLevel:  int64(sim.VipLevel),
			})
		}
		p.Profile.WSPVPPersonalInfo.OpponentSimpleInfo = saveOpponents
	}
	opponentSimples = convertOpponentSimpleInfo2Client(p.Profile.WSPVPPersonalInfo.OpponentSimpleInfo)
	resp.Opponent = make([][]byte, len(opponentSimples))
	for i, simple := range opponentSimples {
		resp.Opponent[i] = encode(simple)
	}
	return 0
}

func convertOpponentSimpleInfo2Client(info []account.WSPVPOppSimpleInfo) []*WSPVPOpp {
	opponentSimples := make([]*WSPVPOpp, 0)
	for _, simp := range info {
		opponentSimples = append(opponentSimples, &WSPVPOpp{
			Acid:      simp.Acid,
			Rank:      simp.Rank,
			ServerId:  gamedata.GetSidName(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), uint(simp.ServerId)),
			Name:      simp.Name,
			GuildName: simp.GuildName,
			TitleId:   simp.TitleId,
			VipLevel:  simp.VipLevel,
		})
	}
	return opponentSimples
}

// 是否刷新，自己的新排名, errCode
func (p *Account) checkAndCostRefresh(forceRefresh bool, resp *rspMsgGetMatchOpponent) (bool, int, uint32) {
	willRefresh := false
	if forceRefresh {
		errC, _ := p.BuyImp(buy.Buy_Typ_WSPVP_Refresh, "", resp)
		if errC != 0 {
			return false, -1, errCode.WSPVPNotTimesLeft
		}
		resp.OnChangeBuy()
		willRefresh = true
	}
	myNewRank := -1
	if !willRefresh {
		willRefresh, myNewRank = p.isRanksChanged()
	}
	return willRefresh, myNewRank, 0
}

// 检测排名是否发生变化, 检测自己的排名和上次对手的排名是否发生变化
// 自己的排名有变化， 更新下
func (p *Account) isRanksChanged() (bool, int) {
	opponents := p.Profile.WSPVPPersonalInfo.OpponentSimpleInfo
	if len(opponents) == 0 {
		return true, -1
	}
	oldRanks := make([]int, 0)
	acids := make([]string, 0)
	checkMyRank := false
	if p.Profile.WSPVPPersonalInfo.Rank > 0 {
		acids = append(acids, p.AccountID.String())
		oldRanks = append(oldRanks, p.Profile.WSPVPPersonalInfo.Rank)
		checkMyRank = true
	}
	for _, opp := range opponents {
		acids = append(acids, opp.Acid)
		oldRanks = append(oldRanks, int(opp.Rank))
	}
	logs.Debug("oldranks %v", oldRanks)
	newRanks := ws_pvp.GetRanks(p.GetWSPVPGroupId(), acids)
	if newRanks == nil || len(newRanks) != len(oldRanks) {
		logs.Debug("new ranks %v", newRanks)
		return true, -1
	}
	for i, rank := range oldRanks {
		if rank != newRanks[i] {
			if i == 0 && checkMyRank {
				return true, newRanks[i]
			} else {
				return true, -1
			}
		}
	}
	return false, -1
}

const LockTimeDelay = 30

func (p *Account) isInLockTimeRange() bool {
	hour, min, _ := util.Clock(time.Unix(p.GetProfileNowTime(), 0))
	return gamedata.IsInWsPvpRange(hour*60 + min)
}

// LockWSPVPBattle : 锁定无双争霸战斗
// 锁定单个武将
func (p *Account) LockWSPVPBattleHandler(req *reqMsgLockWSPVPBattle, resp *rspMsgLockWSPVPBattle) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	if p.isInLockTimeRange() {
		return errCode.WSPVPLockingTime
	}
	if !p.Profile.GetCounts().Has(counter.CounterTypeWspvpChallenge, p) {
		return errCode.WSPVPNotTimesLeft
	}
	// 检查是否锁定正确
	lockAcid := req.Acid
	var simpleInfo *account.WSPVPOppSimpleInfo
	if simpleInfo = p.Profile.WSPVPPersonalInfo.GetOpponentSimpInfo(lockAcid); simpleInfo == nil {
		logs.Debug("lock error 1, %v,", p.Profile.WSPVPPersonalInfo.OpponentSimpleInfo)
		return errCode.WSPVPErrorLockAcid
	}
	// 检查是否正在锁定中，如果没有锁定，直接锁定
	lockTime := time.Now().Unix() + int64(gamedata.WsPvpMainCfg.Config.GetChoiceMaxTime()) + LockTimeDelay
	if !ws_pvp.TryLockOpponent(p.GetWSPVPGroupId(), lockAcid, p.AccountID.String(), lockTime) {
		logs.Debug("lock error 2")
		return errCode.WSPVPHasLocked
	}
	var lockingOpp *ws_pvp.WSPVPInfo
	if ws_pvp.IsRobotId(lockAcid) {
		lockingOpp = buildWspvpRobotInfo(simpleInfo)
		if lockingOpp == nil {
			return errCode.WSPVPCannotGetRobotInfo
		}
	} else {
		// 获取并设置锁定信息
		lockingOpp = ws_pvp.GetRankPlayerAllInfo(p.GetWSPVPGroupId(), lockAcid)
		if lockingOpp == nil {
			return errCode.WSPVPNotInRankYet
		}
	}
	p.Profile.WSPVPPersonalInfo.LockingOppInfo = lockingOpp
	p.Profile.WSPVPPersonalInfo.LockingExpireTime = lockTime
	makeOpponentLockInfo(lockingOpp, resp)
	resp.MyFormation = CheckAndChangeFormation(p.Profile.WSPVPPersonalInfo.GetMyAttackFormation())
	return 0
}

func buildWspvpRobotInfo(simpleInfo *account.WSPVPOppSimpleInfo) *ws_pvp.WSPVPInfo {
	cfg := gamedata.GetWSPVPMatchConfig(int(simpleInfo.Rank))
	if cfg == nil {
		logs.Debug("absence of rank, %d", simpleInfo.Rank)
		return nil
	}
	pvpInfo := new(ws_pvp.WSPVPInfo)
	pvpInfo.Acid = simpleInfo.Acid
	pvpInfo.Name = simpleInfo.Name
	pvpInfo.ServerId = int(simpleInfo.ServerId)
	pvpInfo.GuildName = simpleInfo.GuildName

	robotData := gamedata.GetDroidForWspvp(uint32(cfg.GetRobotID()))
	if robotData == nil {
		logs.Debug("absence of robot, %d", cfg.GetRobotID())
		return nil
	}

	pvpInfo.Formation.Group = gamedata.Random9RobotIds()
	pvpInfo.Formation.Avatar = make([]ws_pvp.WSPVPHeroInfo, 9)
	for i, id := range pvpInfo.Formation.Group {
		pvpInfo.Formation.Avatar[i] = ws_pvp.WSPVPHeroInfo{
			Idx:          int(id),
			Attr:         encode(robotData.Attr),
			StarLevel:    gamedata.RandomRobotStar(),
			Gs:           int64(robotData.HeroGs),
			AvatarSkills: robotData.AvatarSkills[:],
		}
	}
	return pvpInfo
}

func makeOpponentLockInfo(pvpInfo *ws_pvp.WSPVPInfo, resp *rspMsgLockWSPVPBattle) {
	resp.Formation = CheckAndChangeFormation(pvpInfo.Formation.Group)
	resp.HeroStar = make([]int64, 0)
	resp.CorpGs = make([]int64, 0)
	for _, idx := range resp.Formation {
		if idx == -1 {
			resp.HeroStar = append(resp.HeroStar, 0)
			resp.CorpGs = append(resp.CorpGs, 0)
		} else {
			hero := getOpponentByIdx(pvpInfo, int(idx))
			resp.HeroStar = append(resp.HeroStar, int64(hero.StarLevel))
			resp.CorpGs = append(resp.CorpGs, hero.Gs)
		}
	}
}

func getOpponentByIdx(pvpInfo *ws_pvp.WSPVPInfo, idx int) *ws_pvp.WSPVPHeroInfo {
	for _, info := range pvpInfo.Formation.Avatar {
		if info.Idx == idx {
			return &info
		}
	}
	return nil
}

// BeginWSPVPBattle : 开始无双争霸战斗
// 选好己方阵容，开始挑战
func (p *Account) BeginWSPVPBattleHandler(req *reqMsgBeginWSPVPBattle, resp *rspMsgBeginWSPVPBattle) uint32 {
	if !p.isFormationAvailable(req.Formation) {
		return errCode.WSPVPFormationError
	}

	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	nowTime := time.Now().Unix()
	// 锁定过期
	if nowTime > p.Profile.WSPVPPersonalInfo.LockingExpireTime ||
		p.Profile.WSPVPPersonalInfo.LockingOppInfo == nil {
		p.Profile.WSPVPPersonalInfo.CleanLockInfo()
		return errCode.WSPVPOpponentLockTimeOut
	}
	expireTime := nowTime + int64(gamedata.WsPvpMainCfg.Config.GetFightMaxTime()) + LockTimeDelay
	// 重置锁定信息
	if !ws_pvp.TryLockOpponent(p.GetWSPVPGroupId(),
		p.Profile.WSPVPPersonalInfo.LockingOppInfo.Acid,
		p.AccountID.String(),
		expireTime) {
		return errCode.WSPVPOpponentLockTimeOut
	}
	// 次数不足
	if !p.Profile.GetCounts().Use(counter.CounterTypeWspvpChallenge, p) {
		return errCode.WSPVPNotTimesLeft
	}
	p.Profile.WSPVPPersonalInfo.LockingExpireTime = expireTime
	p.Profile.WSPVPPersonalInfo.HasChallengeCount++
	resp.OnChangeGameMode(counter.CounterTypeWspvpChallenge)
	// 保存最新进攻阵容
	p.Profile.WSPVPPersonalInfo.MyAttackFormation = req.Formation
	// 返回对手详细的战斗信息
	resp.OpponentBattleInfo = make([][]byte, len(p.Profile.WSPVPPersonalInfo.LockingOppInfo.Formation.Avatar))
	formationInfo := p.Profile.WSPVPPersonalInfo.LockingOppInfo.Formation
	for i, opp := range formationInfo.Avatar {
		avatarSkills := make([]int64, len(opp.AvatarSkills))
		for i := range opp.AvatarSkills {
			avatarSkills[i] = int64(opp.AvatarSkills[i])
		}
		resp.OpponentBattleInfo[i] = encode(OpponentInfo{
			Idx:            int64(opp.Idx),
			Attr:           opp.Attr,
			Gs:             opp.Gs,
			Skills:         avatarSkills,
			Fashions:       opp.AvatarFashion,
			PassiveSkillId: opp.PassiveSkillId,
			CounterSkillId: opp.CounterSkillId,
			TriggerSkillId: opp.TriggerSkillId,
			Star:           int64(opp.StarLevel),
			HeroWing:       int64(opp.HeroSwing),
		})
	}
	resp.CurrDestinyGeneralSkill = formationInfo.DesSkills
	return 0
}

// EndWSPVPBattle : 结束无双争霸战斗
// 3场战斗结束一起发送, 返回内容走sync
func (p *Account) EndWSPVPBattleHandler(req *reqMsgEndWSPVPBattle, resp *rspMsgEndWSPVPBattle) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	if cheatCode := p.AntiCheatCheckWithCode(&resp.RespWithAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_Wspvp); cheatCode != 0 {
		return cheatCode
	}
	if p.Profile.WSPVPPersonalInfo.LockingOppInfo == nil || p.Profile.WSPVPPersonalInfo.LockingOppInfo.Acid == "" {
		resp.OnChangeWSPVP()
		return 0
	}
	win := req.BattleResult
	acid1 := p.AccountID.String()
	acid2 := p.Profile.WSPVPPersonalInfo.LockingOppInfo.Acid
	opponent := p.Profile.WSPVPPersonalInfo.LockingOppInfo
	newRank1, newRank2, rankChanged := -1, -1, false
	groupId := p.GetWSPVPGroupId()
	// 超时
	if p.Profile.WSPVPPersonalInfo.LockingExpireTime < time.Now().Unix() {
		resp.BattleResult = false
		resp.ServerResult = 1
		p.Profile.WSPVPPersonalInfo.CleanLockInfo()
		resp.OnChangeWSPVP()
		return 0
	}
	if win {
		newRank1, newRank2, rankChanged = ws_pvp.SwapRank(p.GetWSPVPGroupId(), acid1, acid2)
		if newRank1 == newRank2 && newRank1 == -1 {
			return errCode.WSPVPDBError
		}
		p.onSwapRank(groupId, acid2, newRank2)
		if rankChanged && newRank1 > 0 {
			oldRank := p.Profile.WSPVPPersonalInfo.Rank
			p.Profile.WSPVPPersonalInfo.Rank = newRank1
			p.Profile.WSPVPPersonalInfo.OnRankChanged(oldRank)
		}
	} else {
		newRanks := ws_pvp.GetRanks(p.GetWSPVPGroupId(), []string{acid1, acid2})
		if newRanks == nil {
			return errCode.WSPVPDBError
		}
		newRank1 = newRanks[0]
		newRank2 = newRanks[1]
		rankChanged = false
	}
	if newRank1 > 0 {
		wsPvpInfo := p.convertWspvpInfo()
		logs.Debug("update player info to db %v", wsPvpInfo)
		ws_pvp.SavePlayerInfo(groupId, wsPvpInfo)
	}
	ws_pvp.RecordLog(groupId, acid1, true, win, getRankChange(newRank2, newRank1, rankChanged), opponent.Name, opponent.GuildName, newRank1)
	if !ws_pvp.IsRobotId(acid2) {
		ws_pvp.RecordLog(groupId, acid2, false, !win, getRankChange(newRank1, newRank2, rankChanged), p.Profile.Name, p.GuildProfile.GuildName, newRank2)
	}
	BeAttackerInfo := p.Profile.GetWSPVPInfo().LockingOppInfo
	// 重置挑战信息
	p.Profile.WSPVPPersonalInfo.CleanLockInfo()
	// 释放锁定
	ws_pvp.UnlockOpponent(groupId, acid2, p.AccountID.String())
	resp.BattleResult = win
	resp.ServerResult = 0
	resp.OnChangeWSPVP()
	RecordBILog(p, newRank1, newRank2, rankChanged, win, BeAttackerInfo)
	return 0
}

func RecordBILog(p *Account, newRank1, newRank2 int, rankChanged, isWin bool, BeAttackerInfo *ws_pvp.WSPVPInfo) {
	attacker := logiclog.WspvpPlayer{
		AccountID:  p.AccountID.String(),
		CorpLvl:    int(p.Profile.GetCorp().GetLvlInfo()),
		IsWin:      isWin,
		AvatarStar: make([]int, 9),
		AvatarGs:   make([]int, 9),
		AvatarID:   make([]int, 9),
	}
	beAttacker := logiclog.WspvpPlayer{
		AccountID:  BeAttackerInfo.Acid,
		CorpLvl:    int(BeAttackerInfo.CorpLv),
		IsWin:      !isWin,
		AvatarStar: make([]int, 9),
		AvatarGs:   make([]int, 9),
		AvatarID:   make([]int, 9),
	}
	_, _, _, heroGs, _, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(p.Account))
	for i, idx := range p.Profile.GetWSPVPInfo().MyAttackFormation {
		attacker.AvatarID[i] = int(idx)
		if idx != -1 {
			attacker.AvatarStar[i] = int(p.Profile.GetHero().GetStar(int(idx)))
			attacker.AvatarGs[i] = heroGs[int(idx)]
		}
	}
	for i, idx := range BeAttackerInfo.Formation.Group {
		beAttacker.AvatarID[i] = int(idx)
		if idx != -1 {
			hero := getOpponentByIdx(BeAttackerInfo, int(idx))
			if hero != nil {
				beAttacker.AvatarStar[i] = hero.StarLevel
				beAttacker.AvatarGs[i] = int(hero.Gs)
			}
		}
	}
	if rankChanged {
		attacker.BeforeRank = newRank2
		attacker.AfterRank = newRank1
		beAttacker.BeforeRank = newRank1
		beAttacker.AfterRank = newRank2
	} else {
		attacker.BeforeRank = newRank1
		attacker.AfterRank = newRank1
		beAttacker.BeforeRank = newRank2
		beAttacker.AfterRank = newRank2
	}
	logs.Debug("wspvp challenge bi log, %v, %v", attacker, beAttacker)
	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.WspvpChallenge{
		Attacker:   attacker,
		BeAttacker: beAttacker,
	}, logiclog.LogicTag_WspvpChallenge, "")
}

func getRankChange(oldRank, newRank int, rankChanged bool) int {
	if !rankChanged {
		return 0
	}
	if oldRank <= 0 && newRank <= 0 {
		return 0
	}
	if oldRank <= 0 {
		return 10000 - newRank
	}
	if newRank <= 0 {
		return oldRank - 10000
	}
	return oldRank - newRank
}

func (p *Account) onSwapRank(groupId int, acid2 string, rank2 int) {
	if rank2 == 0 {
		// 被打的人已经不在榜单中
		// 删除acid2, 更新acid1
		ws_pvp.DelPlayerFromInfo(groupId, acid2)
	}
}

func (p *Account) convertWspvpInfo() *ws_pvp.WSPVPInfo {
	heroAttrs, heroGs, baseGs, bestHero, corpGs, extraAttrs := gs.GetCurrAttrForWspvp(account.NewAccountGsCalculateAdapter(p.Account))
	info := new(ws_pvp.WSPVPInfo)
	info.Name = p.Profile.Name
	info.Acid = p.AccountID.String()
	info.AvatarId = p.Profile.CurrAvatar
	info.CorpLv = uint32(p.Profile.CorpInf.Level)
	info.GuildName = p.GuildProfile.GuildName
	info.ServerId = int(p.AccountID.ShardId)
	info.VipLevel = int(p.Profile.GetVipLevel())
	info.AllGs = int64(corpGs)
	info.TitleId = p.Profile.GetTitle().GetTitleOnShowForOther(p.Account)
	info.ExtraAttr = ws_pvp.ExtraAttr{
		EquipAttr:   extraAttrs[0],
		DestinyAttr: extraAttrs[1],
		JadeAttr:    extraAttrs[2],
	}
	info.BestHeros.BestHeroInfo = make([]ws_pvp.BestHeroInfo, len(bestHero))
	for i, idx := range bestHero {
		info.BestHeros.BestHeroInfo[i].Idx = int64(idx)
		info.BestHeros.BestHeroInfo[i].Level = int(p.Profile.GetHero().HeroLevel[idx])
		info.BestHeros.BestHeroInfo[i].StarLevel = int(p.Profile.GetHero().GetStar(idx))
		info.BestHeros.BestHeroInfo[i].BaseGs = int64(baseGs[idx])
		info.BestHeros.BestHeroInfo[i].ExtraGs = int64(heroGs[idx] - baseGs[idx])
	}
	info.Formation = p.makeFormationInfo(heroAttrs, heroGs)
	return info
}

// UnlockWSPVPBattle : 取消无双战斗的锁定
// 角色点击返回，需要发送解锁信息
func (p *Account) UnlockWSPVPBattleHandler(req *reqMsgUnlockWSPVPBattle, resp *rspMsgUnlockWSPVPBattle) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	if p.Profile.WSPVPPersonalInfo.LockingOppInfo == nil {
		return errCode.WSPVPErrorLockAcid
	}
	lockingAcid := p.Profile.WSPVPPersonalInfo.LockingOppInfo.Acid
	ws_pvp.UnlockOpponent(p.GetWSPVPGroupId(), lockingAcid, p.AccountID.String())
	// 重置挑战信息
	p.Profile.WSPVPPersonalInfo.CleanLockInfo()
	return 0
}

// GetWSPVPBattleLog : 获取无双争霸的战斗日志
// 无双争霸的战斗日志
func (p *Account) GetWSPVPBattleLogHandler(req *reqMsgGetWSPVPBattleLog, resp *rspMsgGetWSPVPBattleLog) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	logInfo := ws_pvp.GetWSPVPLog(p.GetWSPVPGroupId(), p.AccountID.String())
	resp.WSPVPLog = make([][]byte, len(logInfo))
	for i, log := range logInfo {
		resp.WSPVPLog[i] = encode(log)
	}
	return 0
}

// ClaimWSPVPReward : 领取无双争霸相关奖励
// 领取历史最高排名奖励和每小时累计的奖励
// 如何确保一致性
func (p *Account) ClaimWSPVPRewardHandler(req *reqMsgClaimWSPVPReward, resp *rspMsgClaimWSPVPReward) uint32 {
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	// 0=历史最高排名奖励 1=累计奖励 2=宝箱
	p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
	switch req.RewardType {
	case 0:
		return p.claimBestRankReward(int(req.BestRankRewardId), resp)
	case 1:
		return p.claimTimeReward(resp)
	case 2:
		return p.claimBoxReward(int(req.BestRankRewardId), resp)
	}
	return 0
}

func (p *Account) claimBestRankReward(tableId int, resp *rspMsgClaimWSPVPReward) uint32 {
	if p.Profile.WSPVPPersonalInfo.HasClaimedBestRankReward(tableId) {
		return errCode.WSPVPHasClaimedReward
	}
	cfg := gamedata.GetWsPvpBestRankRewardCfg(tableId)
	if p.Profile.WSPVPPersonalInfo.BestRank > int(cfg.GetStart()) ||
		p.Profile.WSPVPPersonalInfo.BestRank == 0 {
		return errCode.WSPVPFailToClaimReward
	}
	p.Profile.WSPVPPersonalInfo.AddBestRankReward(tableId)
	rewardData := &gamedata.CostData{}
	for _, rewardCfg := range cfg.GetFixed_Loot() {
		if rewardCfg.GetFixedLootID() != "" {
			rewardData.AddItem(rewardCfg.GetFixedLootID(), rewardCfg.GetFixedLootNumber())
		}
	}
	reason := fmt.Sprintf("claim wspvp best rank %d", tableId)
	if ok := account.GiveBySync(p.Account, rewardData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	resp.OnChangeWSPVP()
	return 0
}

const MaxClaimNumByOnce = 999999

// 奖励更新方案
// 1 登录的时候扫描战斗日志， 将离线时候排名变化的时间累计奖励计算出来
// 2 在线发现排名一旦发生变化， 更新累计奖励
// 可能存在的问题是在线发现排名变化的时间会延迟, 玩家可能会多一些奖励
func (p *Account) claimTimeReward(resp *rspMsgClaimWSPVPReward) uint32 {
	nowTime := time.Now().Unix()
	rank := p.Profile.WSPVPPersonalInfo.Rank
	rewardId, rewardCount := p.Profile.WSPVPPersonalInfo.CalcNotClaimedReward(rank, p.Profile.WSPVPPersonalInfo.LastRankChangeTime, nowTime)
	p.Profile.WSPVPPersonalInfo.NotClaimedReward += rewardCount
	if p.Profile.WSPVPPersonalInfo.NotClaimedReward <= 0 {
		return 0
	}
	if p.Profile.WSPVPPersonalInfo.NotClaimedReward > MaxClaimNumByOnce {
		p.Profile.WSPVPPersonalInfo.NotClaimedReward = MaxClaimNumByOnce
	}
	rewardData := &gamedata.CostData{}
	rewardData.AddItem(rewardId, uint32(p.Profile.WSPVPPersonalInfo.NotClaimedReward))

	reason := fmt.Sprintf("claim wspvp time reward")
	if ok := account.GiveBySync(p.Account, rewardData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	p.Profile.WSPVPPersonalInfo.NotClaimedReward = 0
	p.Profile.WSPVPPersonalInfo.LastRankChangeTime = nowTime
	resp.OnChangeWSPVP()
	return 0
}

func (p *Account) claimBoxReward(id int, resp *rspMsgClaimWSPVPReward) uint32 {
	if p.Profile.WSPVPPersonalInfo.HasClaimedBoxReward(id) {
		return errCode.WSPVPHasClaimedReward
	}
	cfg := gamedata.GetWsPvpChallengeReward(id)
	if p.Profile.WSPVPPersonalInfo.HasChallengeCount < int(cfg.GetWinNum()) {
		return errCode.WSPVPFailToClaimReward
	}
	p.Profile.WSPVPPersonalInfo.AddBoxReward(id)
	rewardData := &gamedata.CostData{}
	for _, rewardCfg := range cfg.GetWin_Loot() {
		if rewardCfg.GetRewardID() != "" {
			rewardData.AddItem(rewardCfg.GetRewardID(), rewardCfg.GetRewardNum())
		}
	}
	reason := fmt.Sprintf("claim wspvp best rank %d", id)
	if ok := account.GiveBySync(p.Account, rewardData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	resp.OnChangeWSPVP()
	return 0
}

// GetWSPVPPlayerInfo : 查看无双争霸排行榜上的人的信息
// 简短信息
func (p *Account) GetWSPVPPlayerInfoHandler(req *reqMsgGetWSPVPPlayerInfo, resp *rspMsgGetWSPVPPlayerInfo) uint32 {
	acid := req.AccountId
	if ws_pvp.IsRobotId(acid) {
		return errCode.WSPVPCannotGetRobotInfo
	}
	pvpInfo := ws_pvp.GetRankPlayerAllInfo(p.GetWSPVPGroupId(), acid)
	if pvpInfo == nil {
		return errCode.WSPVPNotInRankYet
	}
	rankInfo := &WSPVPRankInfo{
		Sid:               gamedata.GetSidName(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), uint(pvpInfo.ServerId)),
		Name:              pvpInfo.Name,
		VipLevel:          int64(pvpInfo.VipLevel),
		CorpLevel:         int64(pvpInfo.CorpLv),
		AllGs:             pvpInfo.AllGs,
		GuildName:         pvpInfo.GuildName,
		BestHeroIdx:       make([]int64, 9),
		BestHeroLevel:     make([]int64, 9),
		BestHeroStarLevel: make([]int64, 9),
		BestHeroBaseGs:    make([]int64, 9),
		BestHeroExtraGs:   make([]int64, 9),
		EquipAttr:         make([]int64, 3),
		DestinyAttr:       make([]int64, 3),
		JadeAttr:          make([]int64, 3),
	}
	for i, best := range pvpInfo.BestHeros.BestHeroInfo {
		rankInfo.BestHeroIdx[i] = best.Idx
		rankInfo.BestHeroLevel[i] = int64(best.Level)
		rankInfo.BestHeroStarLevel[i] = int64(best.StarLevel)
		rankInfo.BestHeroBaseGs[i] = best.BaseGs
		rankInfo.BestHeroExtraGs[i] = best.ExtraGs
	}
	for i := 0; i < 3; i++ {
		rankInfo.EquipAttr[i] = int64(pvpInfo.ExtraAttr.EquipAttr[i])
		rankInfo.DestinyAttr[i] = int64(pvpInfo.ExtraAttr.DestinyAttr[i])
		rankInfo.JadeAttr[i] = int64(pvpInfo.ExtraAttr.JadeAttr[i])
	}
	resp.RankInfo = encode(rankInfo)
	return 0
}

func CheckAndChangeFormation(formation []int64) []int64 {
	if formation == nil {
		return nil
	}
	// 检测是否有重复的
	formationMap := make(map[int64]int)
	for _, idx := range formation {
		if idx != -1 {
			formationMap[idx] = formationMap[idx] + 1
		}
	}
	newFormation := make([]int64, 9)
	for i := range newFormation {
		newFormation[i] = -1
	}
	tempFormation := make([]int64, 0)
	duplicate := false
	for key, value := range formationMap {
		tempFormation = append(tempFormation, int64(key))
		if value > 1 {
			duplicate = true
		}
	}
	if duplicate {
		if len(tempFormation) < 3 {
			for i := 0; i < 3-len(tempFormation); i++ {
				tempFormation = append(tempFormation, tempFormation[i])
			}
		}
		count := 0
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if count >= len(tempFormation) {
					return newFormation
				}
				newFormation[3*j+i] = tempFormation[count]
				count++
			}
		}
		return newFormation
	} else {
		return formation
	}
}
