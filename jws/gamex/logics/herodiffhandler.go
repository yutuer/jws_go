package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// StartHeroDiffFight : 开始出奇制胜战斗
//
func (p *Account) StartHeroDiffFightHandler(req *reqMsgStartHeroDiffFight, resp *rspMsgStartHeroDiffFight) uint32 {
	curStageID := p.Profile.GetHeroDiff().GetCurStageID()
	team := p.Profile.GetHeroTeams().GetHeroTeam(getHeroDiffTeamByIDTyp(hero_diff.HeroDiffID2Index(curStageID)))
	if len(team) < 1 {
		return errCode.HeroTeamWarn
	}
	logs.Debug("heroteam: herodiff: %d, curStageID: %d", team[0], curStageID)
	if p.Profile.GetHeroDiff().IsUsedHero(team[0]) {
		return errCode.HeroDiffHeroAlreadyUse
	}
	logiclog.LogHeroDiffStart(p.AccountID.String(), team[0], p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, p.Profile.GetData().CorpCurrGS, curStageID,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return 0
}

// OverHeroDiffFight : 出奇制胜战斗结束
//
func (p *Account) OverHeroDiffFightHandler(req *reqMsgOverHeroDiffFight, resp *rspMsgOverHeroDiffFight) uint32 {
	if cheatCode := p.AntiCheatCheckWithRewards(&resp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_HeroDiff); cheatCode != 0 {
		return cheatCode
	}
	heroDiff := p.Profile.GetHeroDiff()

	lastStageID := heroDiff.GetCurStageID()

	team := p.Profile.GetHeroTeams().GetHeroTeam(getHeroDiffTeamByIDTyp(hero_diff.HeroDiffID2Index(lastStageID)))
	if len(team) < 1 {
		logs.Warn("heroteam warn")
		resp.OnChangeHeroDiff()
		return 0
	}

	logs.Debug("heroteam: herodiff: %d, curStageID: %d", team[0], lastStageID)
	if heroDiff.IsUsedHero(team[0]) {
		logs.Warn("hero used")
		resp.OnChangeHeroDiff()
		return 0
	}
	// new anticheat
	if req.Score > 0 {
		ok := gamedata.CheckHeroDiffFZBScore(
			hero_diff.HeroDiffID2Index(lastStageID),
			uint32(p.Profile.GetData().HeroGs[team[0]]),
			p.Profile.GetCorp().Level,
			uint32(req.Score))
		if !ok {
			logs.Warn("fetch cheat player: %v, levelType: %v, gs: %v, score: %v, level: %v",
				p.AccountID.String(), hero_diff.HeroDiffID2Index(lastStageID), uint32(p.Profile.GetData().HeroGs[team[0]]),
				uint32(req.Score), p.Profile.GetCorp().Level)
			return errCode.YouCheat
		}
	}

	heroDiff.OnPassStage(team[0], int(req.Score))

	extraData := hero_diff.HeroDiffRankData{
		AcID:       p.AccountID.String(),
		FreqAvatar: heroDiff.GetTopNFreqHero(hero_diff.HeroDiffID2Index(lastStageID), gamedata.HeroDiffRankShowAvatarCount),
	}
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[hero_diff.HeroDiffID2Index(lastStageID)].AddWithExtraData(&info, extraData)
	p.updateCondition(account.COND_TYP_HERODIFF_FINISH, 1, 0, "", "", resp)
	resp.OnChangeHeroDiff()

	priceData, ok := p.genHeroDiffReward(int(req.Score), lastStageID)
	if !ok {
		return errCode.RewardFail
	}
	if !account.GiveBySync(p.Account, &priceData.Cost, resp, "HeroDiff") {
		return errCode.RewardFail
	}
	logiclog.LogHeroDiffFinish(p.AccountID.String(), team[0], p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, p.Profile.GetData().CorpCurrGS, lastStageID, int(req.Score), false,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return 0
}

func getHeroDiffTeamByIDTyp(stageID int) int {
	switch stageID {
	case gamedata.HeroDiff_TU:
		return gamedata.LEVEL_TYPE_HERODIFF_TU
	case gamedata.HeroDiff_ZHAN:
		return gamedata.LEVEL_TYPE_HERODIFF_ZHAN
	case gamedata.HeroDiff_HU:
		return gamedata.LEVEL_TYPE_HERODIFF_HU
	case gamedata.HeroDiff_SHI:
		return gamedata.LEVEL_TYPE_HERODIFF_SHI
	default:
		return gamedata.LEVEL_TYPE_HERODIFF_TU
	}
}

func (p *Account) genHeroDiffReward(curScore int, curStageID int) (gamedata.PriceDatas, bool) {
	lootTemple := gamedata.GetHeroDiffRewardData(int(curScore), curStageID, p.Profile.GetCorp().Level)
	priceData := gamedata.PriceDatas{}
	for _, temple := range lootTemple {
		if p.GetRand().Float32() < temple.GetLootChance() {
			for i := temple.GetLootTime(); i > 0; i-- {
				data, err := p.GetGivesByTemplate(temple.GetLootTemplateID())
				if err != nil {
					return gamedata.PriceDatas{}, false
				}
				priceData.AddOther(&data)
			}
		}
	}
	logs.Debug("hero diff get price data: %v", priceData)
	return priceData, true
}
