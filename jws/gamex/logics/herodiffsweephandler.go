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

// HeroDiffSweep : 出奇制胜扫荡
//
func (p *Account) HeroDiffSweepHandler(req *reqMsgHeroDiffSweep, resp *rspMsgHeroDiffSweep) uint32 {
	heroDiff := p.Profile.GetHeroDiff()
	hdCf := gamedata.GetHeroDiffConfig()
	curStageID := heroDiff.GetCurStageID()
	team := p.Profile.GetHeroTeams().GetHeroTeam(getHeroDiffTeamByIDTyp(hero_diff.HeroDiffID2Index(curStageID)))
	if len(team) < 1 {
		return errCode.HeroTeamWarn
	}
	logs.Debug("heroteam: herodiff: %d, curStageID: %d", team[0], curStageID)
	if p.Profile.GetHeroDiff().IsUsedHero(team[0]) {
		return errCode.HeroDiffHeroAlreadyUse
	}
	curScore := heroDiff.GetLastMaxScore(team[0])

	if p.Profile.GetCorp().GetLvlInfo() < hdCf.GetSweepUnlockLevel() || curScore < int(hdCf.GetSweepNeedPoint()) {
		return errCode.CommonConditionFalse
	}
	if !p.Profile.GetSC().HasSC(gamedata.SC_VI_HDP_SD, int64(hdCf.GetSweepCost())) {
		return errCode.CommonLessMoney
	}

	heroDiff.OnPassStage(team[0], curScore)
	extraData := hero_diff.HeroDiffRankData{
		AcID:       p.AccountID.String(),
		FreqAvatar: heroDiff.GetTopNFreqHero(hero_diff.HeroDiffID2Index(curStageID), gamedata.HeroDiffRankShowAvatarCount),
	}
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[hero_diff.HeroDiffID2Index(curStageID)].AddWithExtraData(&info, extraData)
	p.updateCondition(account.COND_TYP_HERODIFF_FINISH, 1, 0, "", "", resp)
	resp.OnChangeHeroDiff()

	priceData, ok := p.genHeroDiffReward(curScore, curStageID)
	if !ok {
		return errCode.RewardFail
	}
	if !account.GiveBySync(p.Account, &priceData.Cost, resp, "HeroDiff") {
		return errCode.RewardFail
	}

	// 消耗扫荡卷
	if !p.Profile.GetSC().UseSC(gamedata.SC_VI_HDP_SD, int64(hdCf.GetSweepCost()), "HeroDiff Sweep") {
		return errCode.RewardFail
	}

	logiclog.LogHeroDiffFinish(p.AccountID.String(), team[0], p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, p.Profile.GetData().CorpCurrGS, curStageID, int(curScore), true,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return 0
}
