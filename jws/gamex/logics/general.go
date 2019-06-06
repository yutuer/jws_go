package logics

import (
	"fmt"

	"math"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) addGeneralStarLevel(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GeneralId string `codec:"gid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/AddGeneralStarlLvlRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]addGeneralStarLevel %s", req.GeneralId)

	const (
		_ = iota
		CodeErrNoGeneralID
		CodeErrGeneralLvData
		CodeErrCost
		CodeErrStarCfg
		CodeErrCostSC
	)

	g := p.GeneralProfile.GetGeneral(req.GeneralId)
	if g == nil {
		return rpcError(resp, CodeErrNoGeneralID)
	}

	wantStar := g.StarLv + 1
	star_cfg := gamedata.GeneralStarCfg(g.Id, wantStar)
	if star_cfg == nil {
		return rpcErrorWithMsg(resp, CodeErrStarCfg, fmt.Sprintf("CodeErrStarCfg %s %d", g.Id, wantStar))
	}

	cost := &account.CostGroup{}
	if !cost.AddSc(p.Account, helper.SC_Money, int64(star_cfg.GetPieceSC())) ||
		!cost.CostBySync(p.Account, resp, "AddGeneralStarLevel") {
		return rpcErrorWithMsg(resp, CodeErrCostSC, "CodeErrCostSC")
	}

	ok, star_aft := g.AddStarLevel()
	if !ok {
		return rpcError(resp, CodeErrCost)
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 4. 副将

	resp.OnChangeGeneralAllChange()

	// 强制刷新任务条件支持红点
	p.updateCondition(account.COND_TYP_GeneralCount,
		0, 0, "", "", resp)

	// 51.激活N个主将, P1主将数
	p.updateCondition(account.COND_TYP_General_Active,
		0, 0, "", "", resp)

	//44.拥有N个品质大于等于X的副将
	p.updateCondition(account.COND_TYP_GeneralCountWithRare,
		0, 0, "", "", resp)

	// 11.拥有P3与P4副将,P3或P4为空时不起作用
	p.updateCondition(account.COND_TYP_General_Has,
		0, 0, "", "", resp)

	resp.mkInfo(p)

	logiclog.LogGeneralStarUp(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		req.GeneralId,
		star_aft,
		"GeneralStarLvlUp",
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(resp)
}

func (p *Account) levelupGeneralRelation(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GeneralRelId string `codec:"grid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/LevelupGeneralRelRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		CodeErrRelationCfgNotFound
		CodeErrLevelUpCondDissatisfy // 升级条件不满足
	)
	rel := p.GeneralProfile.GetGeneralRelation(req.GeneralRelId)
	if rel == nil {
		return rpcErrorWithMsg(resp, CodeErrRelationCfgNotFound, fmt.Sprintf("general relation %s not found cfg", req.GeneralRelId))
	}
	ok, lvl_aft := rel.RelationLevelup(&p.GeneralProfile.PlayerGenerals, p.Profile.GetHero().HeroStarLevel[:])
	if !ok {
		return rpcErrorWithMsg(resp, CodeErrLevelUpCondDissatisfy, fmt.Sprintf("general relation %s cond dissatisfy", req.GeneralRelId))
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 4. 副将

	resp.OnChangeGeneralRelAllChange()
	resp.mkInfo(p)

	logiclog.LogGeneralRelLevelUp(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, req.GeneralRelId, lvl_aft, "GeneralRelationLvlUp",
		p.Profile.GetData().CorpCurrGS, func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

// General Quest

func (p *Account) generalQuestReceive(r servers.Request) *servers.Response {
	req := &struct {
		Req
		QuestId    int64    `codec:"qid"`
		GeneralIds []string `codec:"gids"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/GeneralQuestRevRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Same_Time_Rec_Quest_Max
		Err_Quest_Not_Found
		Err_General_Used
		Err_General_Match_Require
		Err_Times_Not_Enough
		Err_General_Not_Found
		Err_Quest_Already_Rec
	)
	now_time := p.Profile.GetProfileNowTime()
	pg := &p.GeneralProfile
	pg.GQListUpdate(now_time)
	// 同时接取任务数检查
	settig := gamedata.GeneralQuestSetting()
	if len(p.GeneralProfile.QuestList) >= int(settig.GetNGQAccParallel()) {
		return rpcErrorWithMsg(resp, Err_Same_Time_Rec_Quest_Max,
			"Err_Same_Time_Rec_Quest_Max")
	}
	// 每日接取上限检查
	if !p.Profile.GetCounts().Use(counter.CounterTypeGeneralQuest, p.Account) {
		return rpcErrorWithMsg(resp, Err_Times_Not_Enough, "Err_Times_Not_Enough")
	}
	// 是否有任务可接
	q := pg.GetQuestInList(req.QuestId)
	if q == nil {
		return rpcErrorWithMsg(resp, Err_Quest_Not_Found,
			"Err_Quest_Not_Found")
	}
	if q.IsRec {
		return rpcErrorWithMsg(resp, Err_Quest_Already_Rec,
			"Err_Quest_Already_Rec")
	}
	// 副将是否有
	for _, g := range req.GeneralIds {
		gen := pg.GetGeneral(g)
		if gen == nil || !gen.IsHas() {
			return rpcErrorWithMsg(resp, Err_General_Not_Found,
				"Err_General_Not_Found")
		}
	}
	// 副将是否都可用
	if pg.GeneralUsedByQuest(req.GeneralIds) {
		return rpcErrorWithMsg(resp, Err_General_Used,
			"Err_General_Used")
	}
	// 任务条件是否符合
	qCfg := gamedata.GeneralQuestCfg(q.QuestCfgId)
	if len(req.GeneralIds) != int(qCfg.GetGeneralCountRequired()) {
		return rpcErrorWithMsg(resp, Err_General_Match_Require,
			"Err_General_Match_Require")
	}
	for _, g := range req.GeneralIds {
		gCfg := gamedata.GetGeneralInfo(g)
		if gCfg.GetRareLevel() < qCfg.GetRareLimit() {
			return rpcErrorWithMsg(resp, Err_General_Match_Require,
				"Err_General_Match_Require")
		}
	}
	// 接取任务
	pg.ReceiveQuest(q, req.GeneralIds,
		now_time+int64(qCfg.GetFinishTime())*util.MinSec)

	// 条件
	p.updateCondition(account.COND_TYP_General_Quest_Count,
		1, 0, "", "", resp)

	resp.OnChangeGameMode(counter.CounterTypeGeneralQuest)
	resp.OnChangeGeneralQuest()
	resp.mkInfo(p)

	// log
	logiclog.LogGeneralQuestReceive(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, q.QuestCfgId, req.GeneralIds,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) generalQuestRefresh(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/GeneralQuestRefRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Hc_Not_Enough
	)

	// 扣钱
	settig := gamedata.GeneralQuestSetting()
	cost := &account.CostGroup{}
	if !cost.AddHc(p.Account, int64(settig.GetRefreshPrice())) ||
		!cost.CostBySync(p.Account, resp, "GeneralQuestRefresh") {
		return rpcErrorWithMsg(resp, Err_Hc_Not_Enough,
			"Err_Hc_Not_Enough")
	}
	// 刷新
	pg := &p.GeneralProfile
	pg.GQListUpdateForce()

	resp.OnChangeGeneralQuest()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) generalQuestFinish(r servers.Request) *servers.Response {
	req := &struct {
		Req
		QuestId  int64 `codec:"qid"`
		HcFinish bool  `codec:"hc_f"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/GeneralQuestFinishRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Quest_Not_Found
		Err_Hc_Not_Enough
		Err_Finish_Time_Not_Over
		Err_General_Not_Found
		Err_Give
	)

	now_time := p.Profile.GetProfileNowTime()
	pg := &p.GeneralProfile
	// 得到任务
	q, idx := pg.GetQuestInRec(req.QuestId)
	if q == nil {
		return rpcErrorWithMsg(resp, Err_Quest_Not_Found,
			"Err_Quest_Not_Found")
	}
	qCfg := gamedata.GeneralQuestCfg(q.QuestCfgId)
	// 检查完成条件
	if req.HcFinish {
		cost := &account.CostGroup{}
		if !cost.AddHc(p.Account, int64(qCfg.GetQuickFinishPrice())) ||
			!cost.CostBySync(p.Account, resp, "GeneralQuestFinish") {
			return rpcErrorWithMsg(resp, Err_Hc_Not_Enough,
				"Err_Hc_Not_Enough")
		}
	} else {
		if now_time < q.FinishTime {
			return rpcErrorWithMsg(resp, Err_Finish_Time_Not_Over,
				"Err_Finish_Time_Not_Over")
		}
	}

	// 算奖励加成
	settig := gamedata.GeneralQuestSetting()
	var bonusRate float32
	if qCfg.GetBonusType() == 0 { // 星级加成
		var bonus uint32
		for _, g := range q.GeneralIds {
			gen := pg.GetGeneral(g)
			if !gen.IsHas() {
				return rpcErrorWithMsg(resp, Err_General_Not_Found,
					"Err_General_Not_Found")
			}
			bonus += gen.StarLv - 1
		}
		bonusRate = settig.GetStarBonus() * float32(bonus)
	} else { // 羁绊加成
		var bonus int
		for _, g := range q.GeneralIds {
			bonus += pg.GetGeneralActRelNum(g)
		}
		bonusRate = settig.GetRelationBonus() * float32(bonus)
	}

	// 给奖励
	data := &gamedata.CostData{}
	rc := make(map[string]uint32, 16)
	for _, r := range qCfg.GetNGQAward_Template() {
		if r.GetAwardCount() <= 0 {
			continue
		}
		c := r.GetAwardCount()
		if r.GetCanBeBonus() > 0 {
			c = uint32(math.Ceil(float64(float32(c) * (1 + bonusRate))))
		}
		data.AddItem(r.GetAwardItem(), c)
		rc[r.GetAwardItem()] = c
	}
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "GeneralQuestFinish") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}
	// 删除任务
	pg.DelQuestInRec(idx)

	resp.OnChangeGeneralQuest()
	resp.mkInfo(p)

	// log
	logiclog.LogGeneralQuestFinish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, q.QuestCfgId, q.GeneralIds, rc, req.HcFinish,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}
