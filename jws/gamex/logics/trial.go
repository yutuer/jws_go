package logics

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/modules/global_info"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 爬塔

// 手动进爬塔关卡发协议，服务器用来记录通关时间
func (p *Account) TrialEnterLvl(r servers.Request) *servers.Response {
	req := &struct {
		Req
		LvlId int32 `codec:"lvl"`
	}{}
	resp := &struct {
		Resp
	}{}
	initReqRsp(
		"PlayerAttr/TrialEnterLvlResp",
		r.RawBytes,
		req, resp, p)

	errResp, _ := p.trialFightLvlCommonCheck(req.LvlId, resp)
	if errResp != nil {
		return errResp
	}

	now_time := time.Now().Unix()
	p.Tmp.SetLevelEnterTime(now_time)
	// log
	logiclog.LogStage_c(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		fmt.Sprintf("%d", req.LvlId),
		"EnterTrial", now_time,
		p.Profile.GetData().CorpCurrGS,
		"")
	return rpcSuccess(resp)
}

// 手动通过爬塔的某关
func (p *Account) TrialFight(r servers.Request) *servers.Response {
	req := &struct {
		ReqWithAnticheat
		LvlId     int32 `codec:"lvl"`
		IsSuccess bool  `codec:"is_success"`
	}{}
	resp := &struct {
		SyncRespWithRewardsAnticheat
	}{}

	initReqRsp(
		"PlayerAttr/TrialFightResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_AntiCheat_Unmarshal_Err
		Err_Give
	)

	trial := p.Profile.GetPlayerTrial()
	errResp, lvlCfg := p.trialFightLvlCommonCheck(req.LvlId, resp)
	if errResp != nil {
		return errResp
	}

	levelCostTime := time.Now().Unix() - p.Tmp.GetLevelEnterTime()
	if !req.IsSuccess {
		logs.Trace("Trial acid %s lvl %d Failed", p.AccountID.String(), req.LvlId)
		// logiclog
		logiclog.LogTrialLvlFinish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			req.LvlId, false, levelCostTime, p.Profile.Data.CorpCurrGS,
			p.Profile.GetDestinyGeneral().SkillGenerals,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		return rpcSuccess(resp)
	}

	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&resp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_Trial); cheatRsp != nil {
		return cheatRsp
	}

	// 进入下一关
	isFirstPassLvl := trial.NextLvl()
	p.updateCondition(account.COND_TYP_Trial,
		1, 0, "", "", resp)

	// 发奖
	give := &account.GiveGroup{}
	if isFirstPassLvl { // 首次通关
		p._trialAddSC(give, gamedata.VI_Sc0, lvlCfg.GetFirstSC(), resp)
		p._trialAddSC(give, gamedata.VI_DC, lvlCfg.GetFirstDC(), resp)
		p._trialAddSC(give, gamedata.VI_Sc1, lvlCfg.GetFirstFI(), resp)
		p._trialAddSC(give, gamedata.VI_StarBlessCoin, lvlCfg.GetFirstSB(), resp)
		info := p.GetSimpleInfo()
		rank.GetModule(p.AccountID.ShardId).RankByCorpTrial.Add(&info)
		//43.最高通关爬塔第N层
		p.updateCondition(account.COND_TYP_TrialMaxLv,
			0, 0, "", "", resp)
	} else { // 非首次通关
		p._trialAddSC(give, gamedata.VI_Sc0, lvlCfg.GetRewardSC(), resp)
		p._trialAddSC(give, gamedata.VI_DC, lvlCfg.GetRewardDC(), resp)
		p._trialAddSC(give, gamedata.VI_Sc1, lvlCfg.GetRewardFI(), resp)
		p._trialAddSC(give, gamedata.VI_StarBlessCoin, lvlCfg.GetRewardSB(), resp)
		if p.GetRand().Float32() < lvlCfg.GetLootProbability() {
			p._trialLootGroupAndGive(lvlCfg.GetDropGroup(), give, resp)
			p._trialLootGroupAndGive(lvlCfg.GetDropGroup1(), give, resp)
			p._trialLootGroupAndGive(lvlCfg.GetDropGroup2(), give, resp)
		}
	}
	if !give.GiveBySyncAuto(p.Account, resp, "TrialFight") {
		return rpcErrorWithMsg(resp, Err_Give, fmt.Sprintf("Err_Give lvl %d", req.LvlId))
	}

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeTrial, 1, p.Profile.GetProfileNowTime())

	resp.OnChangeTrial()
	resp.mkInfo(p)
	// logiclog
	logiclog.LogTrialLvlFinish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		req.LvlId, true, levelCostTime, p.Profile.Data.CorpCurrGS,
		p.Profile.GetDestinyGeneral().SkillGenerals,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	logiclog.LogStage_c(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		fmt.Sprintf("%d", req.LvlId), "LeaveTrial", p.Tmp.GetLevelEnterTime(),
		p.Profile.GetData().CorpCurrGS,
		"")
	// sysnotice
	if gamedata.IsTrialLevelSysNotice(uint32(req.LvlId)) {
		global_info.OnTrialFinish(p.AccountID.ShardId,
			fmt.Sprintf("%d", req.LvlId), p.AccountID.String(), p.Profile.Name)
	}
	return rpcSuccess(resp)
}

// 领取宝箱奖励
func (p *Account) TrialBonusAward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		LvlId int32 `codec:"lvl"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/TrialBonusAwardResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_No_Bonus
		Err_Lvl_Cfg_Not_Found
		Err_Give
	)

	trial := p.Profile.GetPlayerTrial()
	if trial.BonusLevelId != req.LvlId {
		logs.Warn("TrialBonusAward Err_No_Bonus %d", req.LvlId)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 领奖
	lvlCfg := gamedata.GetTrialLvlById(req.LvlId)
	if lvlCfg == nil {
		return rpcErrorWithMsg(resp, Err_Lvl_Cfg_Not_Found, fmt.Sprintf("Err_Lvl_Cfg_Not_Found lvl %d", req.LvlId))
	}
	give := &account.GiveGroup{}
	p._trialAddSC(give, gamedata.VI_Sc0, lvlCfg.GetBonusSC(), resp)
	p._trialAddSC(give, gamedata.VI_DC, lvlCfg.GetBonusDC(), resp)
	p._trialLootGroupAndGive(lvlCfg.GetBonusDropGroup(), give, resp)
	p._trialLootGroupAndGive(lvlCfg.GetBonusDropGroup1(), give, resp)
	p._trialLootGroupAndGive(lvlCfg.GetBonusDropGroup2(), give, resp)
	if !give.GiveBySyncAuto(p.Account, resp, "TrialFight") {
		return rpcErrorWithMsg(resp, Err_Give, fmt.Sprintf("Err_Give lvl %d", req.LvlId))
	}

	// 领完
	trial.BonusLevelId = 0

	resp.OnChangeTrial()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

const GiveClientFutureLvlCount = 10

// 爬塔扫荡开始
func (p *Account) TrialSweepStart(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		FutureLvlIds        []int32  `codec:"fut_lvls"`
		FutureLvlSc         []int32  `codec:"fut_lvlsc"`
		FutureLvlDc         []int32  `codec:"fut_lvldc"`
		FutureLvlFi         []int32  `codec:"fut_lvlfi"`
		FutureLvlSb         []int32  `codec:"fut_lvlsb"`
		FutureLvlItemC      []int    `codec:"fut_lvlitm_c"`
		FutureLvlItems      []string `codec:"fut_lvlitm"`
		FutureLvlItemCounts []uint32 `codec:"fut_lvlitmc"`
	}{}

	initReqRsp(
		"PlayerAttr/TrialSweepStartResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_First_Get_Bonus
		Err_Already_Sweep
		Err_No_Lvl_For_Sweep
	)

	trial := p.Profile.GetPlayerTrial()

	// 是否应该先领宝箱
	if trial.BonusLevelId > 0 {
		return rpcErrorWithMsg(resp, Err_First_Get_Bonus, fmt.Sprintf("Err_First_Get_Bonus bonusLvl %d", trial.BonusLevelId))
	}

	// 是否在扫荡中或扫荡奖励没领
	if p.Profile.GetProfileNowTime() < trial.SweepEndTime || trial.SweepBeginLvlId > 0 {
		logs.Warn("TrialSweepStart Err_Already_Sweep")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 没有可扫荡的关卡
	if trial.CurLevelId > trial.MostLevelId {
		return rpcErrorWithMsg(resp, Err_No_Lvl_For_Sweep, "Err_No_Lvl_For_Sweep")
	}

	trial.SweepBeginLvlId = trial.CurLevelId
	var costTime int32
	for lvlId := trial.SweepBeginLvlId; lvlId <= trial.MostLevelId; lvlId++ {
		lvlCfg := gamedata.GetTrialLvlById(lvlId)
		costTime += lvlCfg.GetTime()
	}
	trial.SweepStartTime = p.Profile.GetProfileNowTime()
	trial.SweepEndTime = trial.SweepStartTime + int64(costTime)

	// 奖励先都算出来，并保存
	trial.SweepAwards = make([]account.TrialAward, 0,
		trial.MostLevelId-trial.SweepBeginLvlId+1)
	for lvlId := trial.SweepBeginLvlId; lvlId <= trial.MostLevelId; lvlId++ {
		lvlCfg := gamedata.GetTrialLvlById(lvlId)
		award := &account.TrialAward{
			LevelId:   lvlId,
			SC:        lvlCfg.GetRewardSC(),
			DC:        lvlCfg.GetRewardDC(),
			FI:        lvlCfg.GetRewardFI(),
			SB:        lvlCfg.GetRewardSB(),
			ItemId:    make([]string, 0, 3),
			ItemCount: make([]uint32, 0, 3),
		}
		if p.GetRand().Float32() < lvlCfg.GetLootProbability() {
			p._trialLootAndSaveAward(lvlCfg.GetDropGroup(), award)
			p._trialLootAndSaveAward(lvlCfg.GetDropGroup1(), award)
			p._trialLootAndSaveAward(lvlCfg.GetDropGroup2(), award)
		}
		trial.SweepAwards = append(trial.SweepAwards, *award)
	}

	resp.OnChangeTrial()
	resp.mkInfo(p)

	// 刚开始给未来部分关的奖励
	lvlIdx := 0
	fc := 0 // 一次最多发给客户端，未来的奖励数量，目前10关吧
	n := trial.MostLevelId - trial.SweepBeginLvlId
	if n > GiveClientFutureLvlCount {
		n = GiveClientFutureLvlCount
	}
	resp.FutureLvlIds = make([]int32, 0, n)
	resp.FutureLvlSc = make([]int32, 0, n)
	resp.FutureLvlDc = make([]int32, 0, n)
	resp.FutureLvlFi = make([]int32, 0, n)
	resp.FutureLvlItems = make([]string, 0, n)
	resp.FutureLvlItemCounts = make([]uint32, 0, n)
	for lvlId := trial.SweepBeginLvlId; lvlId <= trial.MostLevelId && fc < GiveClientFutureLvlCount; lvlId++ {
		award := trial.SweepAwards[lvlIdx]
		resp.FutureLvlIds = append(resp.FutureLvlIds, award.LevelId)
		resp.FutureLvlSc = append(resp.FutureLvlSc, award.SC)
		resp.FutureLvlDc = append(resp.FutureLvlDc, award.DC)
		resp.FutureLvlFi = append(resp.FutureLvlFi, award.FI)
		resp.FutureLvlSb = append(resp.FutureLvlSb, award.SB)
		resp.FutureLvlItemC = append(resp.FutureLvlItemC, len(award.ItemId))
		for i := 0; i < len(award.ItemId); i++ {
			resp.FutureLvlItems = append(resp.FutureLvlItems, award.ItemId[i])
			resp.FutureLvlItemCounts = append(resp.FutureLvlItemCounts, award.ItemCount[i])
		}
		fc++
		lvlIdx++
	}

	// log
	logiclog.LogTrialSweep(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		"SweepStart", func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

// 定时请求扫荡奖励，用于显示
func (p *Account) TrialSweepAwardForShow(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		Resp
		HistorySc           int32    `codec:"his_sc"`
		HistoryDc           int32    `codec:"his_dc"`
		HistoryFi           int32    `codec:"his_fi"`
		HistoryItems        []string `codec:"his_itms"`
		HistoryItemCounts   []uint32 `codec:"his_itmc"`
		FutureLvlIds        []int32  `codec:"fut_lvls"`
		FutureLvlSc         []int32  `codec:"fut_lvlsc"`
		FutureLvlDc         []int32  `codec:"fut_lvldc"`
		FutureLvlFi         []int32  `codec:"fut_lvlfi"`
		FutureLvlSb         []int32  `codec:"fut_lvlsb"`
		FutureLvlItemC      []int    `codec:"fut_lvlitm_c"`
		FutureLvlItems      []string `codec:"fut_lvlitm"`
		FutureLvlItemCounts []uint32 `codec:"fut_lvlitmc"`
	}{}

	initReqRsp(
		"PlayerAttr/TrialSweepAwardForShowResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Sweep_No_Award
	)

	trial := p.Profile.GetPlayerTrial()
	if trial.SweepBeginLvlId <= 0 {
		return rpcSuccess(resp)
	}
	historyTime := p.Profile.GetProfileNowTime() - trial.SweepStartTime
	// 已扫荡完成的关卡
	var t int64
	var curSweepFinishLvlId int32
	for lvlId := trial.SweepBeginLvlId; lvlId <= trial.MostLevelId; lvlId++ {
		lvlCfg := gamedata.GetTrialLvlById(lvlId)
		if t+int64(lvlCfg.GetTime()) > historyTime {
			break
		}
		t += int64(lvlCfg.GetTime())
		curSweepFinishLvlId = lvlId
	}
	lvlIdx := 0
	if curSweepFinishLvlId > 0 { // 扫荡完的关卡奖励
		hisItems := make(map[string]uint32, curSweepFinishLvlId-trial.SweepBeginLvlId+1)
		for lvlId := trial.SweepBeginLvlId; lvlId <= curSweepFinishLvlId; lvlId++ {
			award := trial.SweepAwards[lvlIdx]
			resp.HistorySc += award.SC
			resp.HistoryDc += award.DC
			resp.HistoryFi += award.FI
			for i := 0; i < len(award.ItemId); i++ {
				c := hisItems[award.ItemId[i]]
				hisItems[award.ItemId[i]] = c + award.ItemCount[i]
			}
			lvlIdx++
		}
		resp.HistoryItems = make([]string, 0, len(hisItems))
		resp.HistoryItemCounts = make([]uint32, 0, len(hisItems))
		for k, v := range hisItems {
			resp.HistoryItems = append(resp.HistoryItems, k)
			resp.HistoryItemCounts = append(resp.HistoryItemCounts, v)
		}
	}
	// 未来一定数量关卡奖励
	fc := 0 // 一次最多发给客户端，未来的奖励数量，目前10关吧
	n := trial.MostLevelId - curSweepFinishLvlId
	if n > GiveClientFutureLvlCount {
		n = GiveClientFutureLvlCount
	}
	resp.FutureLvlIds = make([]int32, 0, n)
	resp.FutureLvlSc = make([]int32, 0, n)
	resp.FutureLvlDc = make([]int32, 0, n)
	resp.FutureLvlFi = make([]int32, 0, n)
	resp.FutureLvlSb = make([]int32, 0, n)
	resp.FutureLvlItems = make([]string, 0, n)
	resp.FutureLvlItemCounts = make([]uint32, 0, n)
	for lvlId := curSweepFinishLvlId + 1; lvlId <= trial.MostLevelId && fc < GiveClientFutureLvlCount; lvlId++ {
		if int(lvlId) >= len(trial.SweepAwards) {
			break
		}
		award := trial.SweepAwards[lvlIdx]
		resp.FutureLvlIds = append(resp.FutureLvlIds, award.LevelId)
		resp.FutureLvlSc = append(resp.FutureLvlSc, award.SC)
		resp.FutureLvlDc = append(resp.FutureLvlDc, award.DC)
		resp.FutureLvlFi = append(resp.FutureLvlFi, award.FI)
		resp.FutureLvlSb = append(resp.FutureLvlSb, award.SB)
		resp.FutureLvlItemC = append(resp.FutureLvlItemC, len(award.ItemId))
		for i := 0; i < len(award.ItemId); i++ {
			resp.FutureLvlItems = append(resp.FutureLvlItems, award.ItemId[i])
			resp.FutureLvlItemCounts = append(resp.FutureLvlItemCounts, award.ItemCount[i])
		}
		fc++
		lvlIdx++
	}
	return rpcSuccess(resp)
}

// 领取扫荡奖励
func (p *Account) TrialSweepAward(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/TrialSweepAwardResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Sweep_Not_Finish
		Err_Sweep_No_Award
		Err_Give
	)
	trial := p.Profile.GetPlayerTrial()

	if p.Profile.GetProfileNowTime() < trial.SweepEndTime {
		logs.Warn("TrialSweepAward Err_Sweep_Not_Finish")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	if trial.SweepBeginLvlId <= 0 {
		logs.Warn("TrialSweepAward Err_Sweep_No_Award")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	p.updateCondition(account.COND_TYP_Trial,
		int(trial.MostLevelId-trial.SweepBeginLvlId)+1, 0, "", "", resp)

	// 发奖
	give := &account.GiveGroup{}
	sc, dc, fi, sb, items := trial.MergeAward()
	p._trialAddSC(give, gamedata.VI_Sc0, sc, resp)
	p._trialAddSC(give, gamedata.VI_DC, dc, resp)
	p._trialAddSC(give, gamedata.VI_Sc1, fi, resp)
	p._trialAddSC(give, gamedata.VI_StarBlessCoin, sb, resp)
	for i, c := range items {
		p._trialAddItem(give, i, c, resp)
	}
	if !give.GiveBySyncAuto(p.Account, resp, "TrialSweepAward") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeTrial,
		int(trial.MostLevelId-trial.CurLevelId+1),
		p.Profile.GetProfileNowTime())

	trial.SweepStartTime = 0
	trial.SweepEndTime = 0
	trial.SweepBeginLvlId = 0
	trial.SweepAwards = []account.TrialAward{}
	trial.SetCurLvl2Most()

	resp.OnChangeTrial()
	resp.mkInfo(p)

	// log
	logiclog.LogTrialSweep(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		"SweepAward", func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

// 花hc立刻结束扫荡
func (p *Account) TrialSweepEndByHC(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/TrialSweepEndByHCResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Not_Sweeping
		Err_HC_Not_Enough
	)

	trial := p.Profile.GetPlayerTrial()
	if p.Profile.GetProfileNowTime() >= trial.SweepEndTime {
		logs.Warn("TrialSweepEndByHC Err_Not_Sweeping")
		return rpcSuccess(resp)
	}
	cost := account.CostGroup{}
	if !cost.AddHc(p.Account, int64(gamedata.GetCommonCfg().GetTrialCDTimeCostHC())) ||
		!cost.CostBySync(p.Account, resp, "TrialSweepEndByHC") {
		return rpcErrorWithMsg(resp, Err_HC_Not_Enough, "Err_HC_Not_Enough")
	}

	sc, dc, fi, sb, items := trial.MergeAward()
	data := &gamedata.CostData2Client{}

	data.AddItem2Client(gamedata.VI_Sc0, uint32(sc))
	data.AddItem2Client(gamedata.VI_DC, uint32(dc))
	data.AddItem2Client(gamedata.VI_Sc1, uint32(fi))
	data.AddItem2Client(gamedata.VI_StarBlessCoin, uint32(sb))
	for i, c := range items {
		data.AddItem2Client(i, c)
	}
	resp.AddResReward(data)

	trial.SweepEndTime = 0
	trial.SweepStartTime = 0

	resp.OnChangeTrial()
	resp.mkInfo(p)

	// log
	logiclog.LogTrialSweep(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		"SweepEndByHC", func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

// 重置爬塔
func (p *Account) TrialReset(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/TrialResetResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Times_Not_Enough
	)

	// 次数检查
	if !p.Profile.GetCounts().Use(counter.CounterTypeTrial, p.Account) {
		logs.Warn("TrialReset Err_Times_Not_Enough")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	trial := p.Profile.GetPlayerTrial()
	trial.CurLevelId = gamedata.GetTrialFirstLvlId()

	resp.OnChangeTrial()
	resp.OnChangeGameMode(counter.CounterTypeTrial)
	resp.mkInfo(p)

	// log
	logiclog.LogTrialReset(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		trial.MostLevelId, func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

// TrialFastPass : 过关斩将
// 过关斩将功能，快速通关， type=1 查询， type=2 执行

// reqMsgTrialFastPass 过关斩将请求消息定义
type reqMsgTrialFastPass struct {
	Req
	Type  int64 `codec:"_p1_"`  // 消息类型
	LvlId int64 `codec:"lvlid"` // _p2_
}

// rspMsgTrialFastPass 过关斩将回复消息定义
type rspMsgTrialFastPass struct {
	SyncResp
	CanPass      int64    `codec:"canpass"`      // 返回值为0时代表不可以，为1时代表可以过关斩将
	RewardsLevel []int64  `codec:"rewardslevel"` // 该索引下奖励所属的关卡ID
	RewardsType  []int64  `codec:"rewardstype"`  // 该索引下的奖励的类型，0-普通， 1-宝箱
	RewardsID    []string `codec:"rewardsid"`    // 该索引下奖励的ID
	RewardsCount []int64  `codec:"rewardscount"` // 该索引下奖励的数量
}

// TrialFastPass 过关斩将: 过关斩将功能，快速通关， type=1 查询， type=2 执行
func (p *Account) TrialFastPass(r servers.Request) *servers.Response {
	req := new(reqMsgTrialFastPass)
	rsp := new(rspMsgTrialFastPass)

	initReqRsp(
		"Attr/TrialFastPassRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		QueryType = int64(1)
		ExecType  = int64(2)
		CanPass   = int64(1)
		NotPass   = int64(0)
	)
	// logic imp begin
	trial := p.Profile.GetPlayerTrial()
	curLvl := trial.CurLevelId
	lvlCfg := gamedata.GetTrialLvlById(int32(curLvl))
	// 10关之前不可跳关
	if curLvl <= int32(gamedata.GetCommonCfg().GetTrailGSLevelCon()) {
		rsp.CanPass = NotPass
		//logs.Trace("can't fast pass until 10! curlevel: %d", curLvl)
		return rpcSuccess(rsp)
	}

	// 只有第一次打完某一关才会弹出
	if int32(curLvl) != trial.MostLevelId+1 {
		rsp.CanPass = NotPass
		//logs.Trace("not the first pass. curlevel: %d", curLvl)
		return rpcSuccess(rsp)
	}

	// 可跳关数大于一定数量方可跳关
	ID := int32(curLvl)
	var HeroID int
	heroIDs := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_TRIAL)
	if heroIDs == nil || len(heroIDs) <= 0 {
		HeroID = p.Profile.GetCurrAvatar()
	} else {
		HeroID = heroIDs[0]
	}
	heroGss := p.Profile.GetData().HeroGs
	if HeroID < 0 || HeroID >= len(heroGss) {
		return rpcError(rsp, 10)
	}
	gs := int(p.Profile.GetData().HeroGs[HeroID])
	// 需要高于的战力比
	x := gamedata.GetCommonCfg().GetTrailGSLevelGap()
	for float32(lvlCfg.GetLevelGS())*(1+x) <= float32(gs) && ID <= gamedata.GetTrialFinalLvlId() {
		ID++
		lvlCfg = gamedata.GetTrialLvlById(ID)
	}
	canPassCount := ID - int32(curLvl)
	if canPassCount < int32(gamedata.GetCommonCfg().GetTrailGSLevelNum()) {
		rsp.CanPass = NotPass
		//logs.Trace("level can pass too little, target level : %d, gs: %d", ID, gs)
		return rpcSuccess(rsp)
	}

	rsp.CanPass = CanPass
	// 请求跳关命令, 给予奖励
	if req.Type == ExecType {
		rsp.RewardsLevel = make([]int64, 0, 4*canPassCount /*default*/)
		rsp.RewardsType = make([]int64, 0, 4*canPassCount /*default*/)
		for lv := int32(curLvl); lv < ID; lv++ {

			lvlCfg := gamedata.GetTrialLvlById(lv)

			p._addFastPassReward(rsp, gamedata.VI_Sc0, uint32(lvlCfg.GetFirstSC()), lv, 0)
			p._addFastPassReward(rsp, gamedata.VI_DC, uint32(lvlCfg.GetFirstDC()), lv, 0)
			p._addFastPassReward(rsp, gamedata.VI_Sc1, uint32(lvlCfg.GetFirstFI()), lv, 0)
			p._addFastPassReward(rsp, gamedata.VI_StarBlessCoin, uint32(lvlCfg.GetFirstSB()), lv, 0)

			// 有宝箱奖励
			if lvlCfg.GetBonus() == 1 {

				ids := make([]string, 0, 5)
				counts := make([]uint32, 0, 5)
				bonusID, bonusC := p._trialLootGroup(lvlCfg.GetBonusDropGroup())
				ids = append(ids, bonusID)
				counts = append(counts, bonusC)
				bonusID1, bonusC1 := p._trialLootGroup(lvlCfg.GetBonusDropGroup1())
				ids = append(ids, bonusID1)
				counts = append(counts, bonusC1)
				if lvlCfg.GetBonusDropGroup2() != "" {
					bonusID2, bonusC2 := p._trialLootGroup(lvlCfg.GetBonusDropGroup2())
					ids = append(ids, bonusID2)
					counts = append(counts, bonusC2)
				}
				// 防止重复显示
				ids = append(ids, gamedata.VI_Sc0)
				counts = append(counts, uint32(lvlCfg.GetBonusSC()))
				ids = append(ids, gamedata.VI_DC)
				counts = append(counts, uint32(lvlCfg.GetBonusDC()))
				rewards := p._mergeGroupBonus(ids, counts)
				for id, count := range rewards {
					p._addFastPassReward(rsp, id, count, lv, 1)
				}
			}

			//43.最高通关爬塔第N层
			p.updateCondition(account.COND_TYP_TrialMaxLv,
				0, 0, "", "", rsp)
			// 领完更新trial
			trial.NextLvl()
			trial.BonusLevelId = 0
			// 更新排行榜
			info := p.GetSimpleInfo()
			rank.GetModule(p.AccountID.ShardId).RankByCorpTrial.Add(&info)
		}
	}

	p.updateCondition(account.COND_TYP_Trial,
		int(ID-curLvl), 0, "", "", rsp)

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeTrial,
		int(ID-curLvl),
		p.Profile.GetProfileNowTime())

	rsp.OnChangeTrial()
	// logic imp end
	rsp.mkInfo(p)
	//logs.Trace("Success!!! curlevel: %d, gs: %d", ID, gs)
	return rpcSuccess(rsp)
}

func (p *Account) trialFirstActivate() bool {
	if !p.Profile.GetPlayerTrial().IsActivate {
		if account.CondCheck(gamedata.Mod_Trial, p.Account) {
			p.Profile.GetPlayerTrial().IsActivate = true
			// 消耗一次重置次数
			if !p.Profile.GetCounts().Use(counter.CounterTypeTrial, p.Account) {
				logs.Error("trialFirstActivate use counter failed")
			}
			logs.Info("%s trial unlock", p.AccountID.String())
			return true
		}
	}
	return false
}

func (p *Account) trialFightLvlCommonCheck(lvlId int32, resp respInterface) (
	errResp *servers.Response, lvlCfg *ProtobufGen.LEVEL_TRIAL) {
	const (
		_ = iota + 20
		Err_Sweep
		Err_First_Get_Bonus
		Err_Fight_Not_Cur_Lvl
		Err_Already_Final_Lvl
		Err_Lvl_Cfg_Not_Found
		Err_Not_Activate
	)

	trial := p.Profile.GetPlayerTrial()
	curLvl := trial.CurLevelId
	if !p.Profile.GetPlayerTrial().IsActivate {
		p.trialFirstActivate()
		if !p.Profile.GetPlayerTrial().IsActivate {
			return rpcErrorWithMsg(resp, Err_Not_Activate,
				fmt.Sprintf("Err_Not_Activate lvl %d", lvlId)), nil
		}
	}
	// 是否在扫荡中或扫荡奖励没领
	if p.Profile.GetProfileNowTime() < trial.SweepEndTime || trial.SweepBeginLvlId > 0 {
		return rpcErrorWithMsg(resp, Err_Sweep, fmt.Sprintf("Err_Sweep lvl %d", lvlId)), nil
	}
	// 是否应该领宝箱
	if trial.BonusLevelId > 0 {
		logs.Warn("trialFightLvlCommonCheck Err_First_Get_Bonus bonusLvl %d", trial.BonusLevelId)
		return rpcWarn(resp, errCode.ClickTooQuickly), nil
	}
	// 是否已经过了最后一关
	if trial.CurLevelId > gamedata.GetTrialFinalLvlId() {
		return rpcErrorWithMsg(resp, Err_Already_Final_Lvl, fmt.Sprintf("Err_Already_Final_Lvl lvl %d", lvlId)), nil
	}
	if lvlId != curLvl {
		logs.Warn("trialFightLvlCommonCheck Err_Fight_Not_Cur_Lvl curLvl %d CurLvlId %d",
			lvlId, p.Profile.GetPlayerTrial().CurLevelId)
		return rpcWarn(resp, errCode.ClickTooQuickly), nil

	}
	lvlCfg = gamedata.GetTrialLvlById(lvlId)
	if lvlCfg == nil {
		return rpcErrorWithMsg(resp, Err_Lvl_Cfg_Not_Found, fmt.Sprintf("Err_Lvl_Cfg_Not_Found lvl %d", lvlId)), nil
	}
	return nil, lvlCfg
}

func (p *Account) _trialLootGroupAndGive(
	lootId string,
	give *account.GiveGroup,
	sync interfaces.ISyncRspWithRewards) {
	if lootId == "" {
		return
	}
	itemId, count := p._trialLootGroup(lootId)
	if itemId == "" || count <= 0 {
		return
	}

	logs.Trace("[%s]TrialSendReward:%s[%d]", p.AccountID, itemId, count)
	// 目前限制爬塔掉落物品只有宝石
	p._trialAddItem(give, itemId, count, sync)
	return
}

func (p *Account) _trialAddSC(
	give *account.GiveGroup,
	itemID string, count int32,
	sync interfaces.ISyncRspWithRewards) {
	if count > 0 {
		give.AddItem(itemID, uint32(count))
	}
}

func (p *Account) _trialLootGroup(lootId string) (itemId string, count uint32) {
	acid := p.AccountID.String()
	gives, err := p.GetGivesByItemGroup(lootId)
	if err != nil {
		logs.SentryLogicCritical(acid, "sendRewardByItemGroup %s Err %s.",
			lootId, err.Error())
		return "", 0
	}
	if !gives.IsNotEmpty() {
		return "", 0
	}

	// 假设这里只有一个
	return gives.Item2Client[0], gives.Count2Client[0]
}

func (p *Account) _trialAddItem(
	give *account.GiveGroup,
	itemId string, count uint32,
	sync interfaces.ISyncRspWithRewards) {
	data := &gamedata.CostData{}
	data.AddItem(itemId, count)
	give.AddCostData(data)
}

func (p *Account) _trialLootAndSaveAward(lootId string, award *account.TrialAward) {
	if lootId != "" {
		itemId, itemCount := p._trialLootGroup(lootId)
		if itemId != "" && itemCount > 0 {
			award.AddAward(itemId, itemCount)
		}
	}
}

func (p *Account) _addFastPassReward(rsp *rspMsgTrialFastPass, itemId string, count uint32, lvId int32, isBonus int32) bool {
	g := account.GiveGroup{}
	data := &gamedata.CostData{}
	data.AddItem(itemId, count)
	g.AddCostData(data)
	ok, res := g.GiveBySyncWithRes(p.Account, rsp, "FastPassTrial")
	p._addFastPassLevelType(rsp, res, lvId, isBonus)
	return ok
}

func (p *Account) _addFastPassLevelType(rsp *rspMsgTrialFastPass, g *gamedata.CostData2Client, lvId int32, isBonus int32) {
	if g == nil {
		return
	}
	for i := 0; i < g.Len(); i++ {
		ok, id, count, _, _ := g.GetItem(i)
		if ok {
			rsp.RewardsID = append(rsp.RewardsID, id)
			rsp.RewardsCount = append(rsp.RewardsCount, int64(count))
			rsp.RewardsLevel = append(rsp.RewardsLevel, int64(lvId))
			rsp.RewardsType = append(rsp.RewardsType, int64(isBonus))
		}
	}
}

func (p *Account) _mergeGroupBonus(ids []string, counts []uint32) map[string]uint32 {
	rewardMap := make(map[string]uint32, len(ids))
	for i, id := range ids {
		n, ok := rewardMap[id]
		if !ok {
			rewardMap[id] = counts[i]
		} else {
			rewardMap[id] = n + counts[i]
		}
	}

	return rewardMap
}
