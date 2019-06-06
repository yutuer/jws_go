package logics

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) guildVisit(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Type int `codec:"typ"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerGuild/GuildVisitRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]guildVisit %s", req.Type)

	const (
		_ = iota
		CodeErrNoGeneralID
		CodeErrGeneralLvData
		CodeErrCost
	)

	resp.mkInfo(p)

	return rpcSuccess(resp)
}

// 新的模式下，此协议没用了
//func (p *Account) guildGatesEnemyStart(r servers.Request) *servers.Response {
//	req := &struct {
//		Req
//	}{}
//	resp := &struct {
//		SyncResp
//	}{}
//
//	initReqRsp(
//		"PlayerGuild/GuildGatesEnemyStartRsp",
//		r.RawBytes,
//		req, resp, p)
//
//	acID := p.AccountID.String()
//	guildID := p.GuildProfile.GetCurrGuildUUID()
//
//	logs.Trace("[%s]guildGatesEnemyStart", acID)
//
//	isStarted := gates_enemy.GetModule(p.AccountID.ShardId).IsActHasStart(guildID)
//	if isStarted {
//		return rpcWarn(resp, errCode.GuildGateEnemyHasStarted)
//	}
//
//	guildData, ret := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
//	if ret.HasError() || guildData == nil {
//		return rpcWarn(resp, errCode.GuildGateEnemyStartErr)
//	}
//
//	ok := guild.GetModule(p.AccountID.ShardId).UseGateEnemyCount(guildID)
//	if !ok {
//		return rpcWarn(resp, errCode.GuildGateEnemyStartErr)
//	}
//
//	p.Profile.GetGatesEnemy().CleanDatas()
//
//	ok = gates_enemy.GetModule(p.AccountID.ShardId).StartGatesEnemyAct(guildID, 0, guildData.Members)
//	if !ok {
//		return rpcWarn(resp, errCode.GuildGateEnemyStartErr)
//	}
//
//	resp.OnChangePlayerGuild()
//	resp.OnChangeGuildInfo()
//	resp.OnChangeGuildMemsInfo()
//	resp.OnChangeGatesEnemyData()
//	resp.OnChangeGatesEnemyPushData()
//
//	resp.mkInfo(p)
//
//	return rpcSuccess(resp)
//}

func (p *Account) guildGatesEnemyEnterAct(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	acID := p.AccountID.String()
	initReqRsp(
		"PlayerGuild/GuildGatesEnemyEnterActRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Not_Start
	)
	logs.Trace("[%s]guildGatesEnemyEnterAct", acID)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	// 如果活动没开，先开
	errNotStart, warnCode := p.startGateEnemy(acID)
	if errNotStart {
		return rpcWarn(resp, errCode.GateEnemyActivityOver)
	}
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	res := p.Profile.GetGatesEnemy().OnEnterAct(
		acID,
		p.Account.GetSimpleInfo())

	if res != 0 {
		return rpcWarn(resp, res)
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyLeaveAct(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	acID := p.AccountID.String()
	initReqRsp(
		"PlayerGuild/GuildGatesEnemyLeaveActRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]guildGatesEnemyLeaveAct", acID)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}
	res := p.Profile.GetGatesEnemy().OnLeaveAct(acID)

	if res != 0 {
		return rpcWarn(resp, res)
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyFightBegin(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ETyp int `codec:"et"`
		EIDx int `codec:"eid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerGuild/GuildGatesEnemyFightBRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]guildGatesEnemyFight %d", req.ETyp)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	res := p.Profile.GetGatesEnemy().OnFightBegin(
		p.AccountID.String(),
		p.Account.GetSimpleInfo(),
		req.ETyp, req.EIDx)

	if res != 0 {
		return rpcWarn(resp, res)
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyFightEnd(r servers.Request) *servers.Response {
	req := &struct {
		ReqWithAnticheat
		ETyp      int  `codec:"et"`
		EIDx      int  `codec:"eid"`
		IsSuccess bool `codec:"su"`
	}{}
	resp := &struct {
		SyncRespWithRewardsAnticheat
		KillPoint int `codec:"kp"`
		GEPoint   int `codec:"gp"`
	}{}

	initReqRsp(
		"PlayerGuild/GuildGatesEnemyFightERsp",
		r.RawBytes,
		req, resp, p)

	acID := p.AccountID.String()
	if cheatCode := p.AntiCheatCheckWithRewards(&resp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_GateEnemy); cheatCode != 0 {
		return rpcWarn(resp, cheatCode)
	}
	logs.Trace("[%s]guildGatesEnemyFightEnd %d", req.ETyp)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	res := p.Profile.GetGatesEnemy().OnFightEnd(
		acID,
		p.Account.GetSimpleInfo(),
		req.ETyp, req.EIDx, req.IsSuccess)

	if res != 0 {
		return rpcWarn(resp, res)
	}

	if req.IsSuccess {
		resp.KillPoint, resp.GEPoint = p.guildGatesEnemyFightGive(acID, req.ETyp, resp)
		// 条件更新
		p.updateCondition(account.COND_TYP_GateEnemy_Finish, 1, 0, "", "", resp)
		p.Profile.GetMarketActivitys().OnGameMode(acID, gamedata.CounterGateEnemy, 1, p.Profile.GetProfileNowTime())
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyFightBossBegin(r servers.Request) *servers.Response {
	req := &struct {
		Req
		BID int `codec:"bid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerGuild/GuildGatesEnemyFightBossBRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]GuildGatesEnemyFightBossBRsp %d", req.BID)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	res := p.Profile.GetGatesEnemy().OnFightBossBegin(
		p.AccountID.String(),
		p.Account.GetSimpleInfo(),
		req.BID)

	if res != 0 {
		return rpcWarn(resp, res)
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyFightGive(acID string, enemyTyp int, resp interfaces.ISyncRspWithRewards) (kp, gp int) {
	data := gamedata.GetAllGEEnemyGroupCfg()
	kp, gp = 0, 0
	if enemyTyp < len(data) {
		enemyID := data[enemyTyp].GetEGLevelID()
		lootData := gamedata.GetGEEnemyLootCfg(enemyID)
		if lootData != nil {
			kp = int(lootData.GetKillingValue())
			gp = int(lootData.GetGEPoint())
			// 固定掉落
			giveData := gamedata.CostData{}
			giveData.SetGST(gamedata.GST_GateEnemy)
			cfg := gamedata.GetGEConfig()
			buffc := p.Profile.GetGatesEnemy().GetPushData().GetBuffCount(acID)
			giveData.SetGateEnemyBonue(float32(1 + cfg.GetEncouragerLootAddition()*float32(buffc)))
			for _, d := range lootData.Fixed_Loot {
				giveData.AddItem(d.GetFixedLootID(), d.GetFixedLootNumber())
			}
			// 随机掉落
			for _, d := range lootData.Random_Loot {
				randNum := p.GetRand().Float32()
				logs.Trace("lootData.Random_Loot %v - %v",
					randNum, d.GetLootProbability())
				if d.GetLootProbability() >= randNum {
					gives, err := p.GetGivesByItemGroup(d.GetLootGroupID())
					if err == nil {
						giveData.AddGroup(gives.Gives())
					} else {
						logs.SentryLogicCritical(acID,
							"guildGatesEnemyFightEnd Loot %s Err By %s",
							d.GetLootGroupID(), err.Error())
					}
				}
			}

			if !account.GiveBySync(p.Account, &giveData, resp, "GatesEnemyFight") {
				logs.SentryLogicCritical(acID,
					"guildGatesEnemyFightEnd GiveBySync %v Err",
					giveData)
			}
		}
	}
	return
}

func (p *Account) guildGatesEnemyFightBossEnd(r servers.Request) *servers.Response {
	req := &struct {
		ReqWithAnticheat
		BID       int  `codec:"bid"`
		IsSuccess bool `codec:"su"`
	}{}
	resp := &struct {
		SyncRespWithRewardsAnticheat
		KillPoint int `codec:"kp"`
		GEPoint   int `codec:"gp"`
	}{}

	initReqRsp(
		"PlayerGuild/GuildGatesEnemyFightBossERsp",
		r.RawBytes,
		req, resp, p)
	if cheatCode := p.AntiCheatCheckWithRewards(&resp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_GateEnemy); cheatCode != 0 {
		return rpcWarn(resp, cheatCode)
	}
	acID := p.AccountID.String()

	logs.Trace("[%s]GuildGatesEnemyFightBossERsp %d", req.BID)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	res := p.Profile.GetGatesEnemy().OnFightBossEnd(
		p.AccountID.String(),
		p.Account.GetSimpleInfo(),
		req.BID, req.IsSuccess)

	if res != 0 {
		return rpcWarn(resp, res)
	}

	if req.IsSuccess {
		resp.KillPoint, resp.GEPoint = p.guildGatesEnemyFightGive(acID, req.BID, resp)
		// 条件更新
		p.updateCondition(account.COND_TYP_GateEnemy_Finish, 1, 0, "", "", resp)
		p.Profile.GetMarketActivitys().OnGameMode(acID, gamedata.CounterGateEnemy, 1, p.Profile.GetProfileNowTime())
	}

	resp.OnChangeGatesEnemyData()
	resp.OnChangeGatesEnemyPushData()

	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) guildGatesEnemyGetReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ReType int64 `codec:"retype"` // 0：普通，1：巨额，2：超额
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerGuild/GuildGatesEnemyGetRewardRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]GuildGatesEnemyGetRewardRsp")

	const (
		_ = iota
		CodeRewardErr
		Err_Cost
		Err_Param
	)

	acID := p.AccountID.String()
	res, pointAll := guild.GetModule(p.AccountID.ShardId).SetGateEnemyReward(p.GuildProfile.GuildUUID, acID)

	if res.HasError() {
		var code uint32
		switch res.ErrCode {
		case guild.Err_Guild_Not_Exist:
			code = errCode.GuildGatesEnemyGetRewardErrByNoGuild
		case guild.Err_No_Gate_Enemy_Reward:
			code = errCode.GuildGatesEnemyGetRewardErrByNoReward
		case guild.Err_No_Gate_Enemy_Not_Join:
			code = errCode.GuildGatesEnemyGetRewardErrByNoJoin
		default:
			code = errCode.GuildGatesEnemyGetRewardErrByNoReward
		}
		return rpcWarn(resp, code)
	}

	// cost
	cfg := gamedata.GetGEConfig()
	var addition float32
	addition = 1.0
	data := &gamedata.CostData{}
	switch req.ReType {
	case 0:
	case 1:
		data.AddItem(cfg.GetGiftLittleCoin(), cfg.GetGiftLittleCoinCount())
		addition = 1 + cfg.GetGiftLittleAddition()
	case 2:
		data.AddItem(cfg.GetGiftHugeCoin(), cfg.GetGiftHugeCoinCount())
		addition = 1 + cfg.GetGiftHugeAddition()
	default:
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	if !account.CostBySync(p.Account, data, resp, "GuildGatesEnemyGetReward") {
		return rpcErrorWithMsg(resp, Err_Cost, "Err_Cost")
	}

	gift := gamedata.GetGEEnemyGiftCfg(pointAll)
	if gift != nil {
		logs.Trace("GetGEEnemyGiftCfg gift %v", gift)
		gift.Gifts.Cost.SetGST(gamedata.GST_GateEnemy)
		cfg := gamedata.GetGEConfig()
		buffc := p.Profile.GetGatesEnemy().GetPushData().GetBuffCount(acID)
		gift.Gifts.Cost.SetGateEnemyBonue(
			(1 + cfg.GetEncouragerGiftAddition()*float32(buffc)) * addition)
		if !account.GiveBySync(p.Account, &gift.Gifts.Cost, resp, "GateEnemyGift") {
			logs.SentryLogicCritical(acID, "GetGEEnemyGiftCfg Err By %v", gift.Gifts)
			return rpcError(resp, CodeRewardErr)
		}
	}

	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.OnChangeGatesEnemyData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

// GuildGatesEnemyInspire : 兵临城下鼓舞协议
// 兵临城下鼓舞相关协议
// reqMsgGuildGatesEnemyInspire 兵临城下鼓舞协议请求消息定义
type reqMsgGuildGatesEnemyInspire struct {
	Req
	AccountID string `codec:"acid"` // 鼓舞的玩家ID
}

// rspMsgGuildGatesEnemyInspire 兵临城下鼓舞协议回复消息定义
type rspMsgGuildGatesEnemyInspire struct {
	SyncResp
	CurrentInspireBuffLevel int64    `codec:"buffLv"`   // 当前的鼓舞buff等级
	InspireNames            []string `codec:"insNames"` // 鼓舞的玩家名字列表
}

// GuildGatesEnemyInspire 兵临城下鼓舞协议: 兵临城下鼓舞相关协议
func (p *Account) GuildGatesEnemyInspire(r servers.Request) *servers.Response {
	req := new(reqMsgGuildGatesEnemyInspire)
	rsp := new(rspMsgGuildGatesEnemyInspire)

	initReqRsp(
		"Attr/GuildGatesEnemyInspireRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Not_Start
		Err_Vip
		Err_HC
	)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	// vip check
	if !gamedata.GetVIPCfg(int(p.Profile.GetVipLevel())).GateEnemyBuff {
		return rpcErrorWithMsg(rsp, Err_Vip, "Err_Vip")
	}

	// 如果活动没开，先开
	errNotStart, warnCode := p.startGateEnemy(p.AccountID.String())
	if errNotStart {
		return rpcWarn(rsp, errCode.GateEnemyActivityOver)
	}
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	cfg := gamedata.GetGEConfig()
	if !p.Profile.GetHC().HasHC(int64(cfg.GetGEBuffCount())) {
		return rpcErrorWithMsg(rsp, Err_HC, "Err_HC")
	}

	res := p.Profile.GetGatesEnemy().OnAddBuff(p.AccountID.String(),
		p.Profile.Name)

	if res.Code != 0 {
		return rpcWarn(rsp, res.Code)
	}
	rsp.InspireNames = res.RetStrParam
	for _, r := range rsp.InspireNames {
		if r != "" {
			rsp.CurrentInspireBuffLevel++
		}
	}

	cost := account.CostGroup{}
	cost.AddHc(p.Account, int64(cfg.GetGEBuffCount()))
	if !cost.CostBySync(p.Account, rsp, "GuildGatesEnemyInspire") {
		return rpcErrorWithMsg(rsp, Err_HC, "Err_HC")
	}

	// 跑马灯
	sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_GUILDBUFF).
		AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
		AddParam(sysnotice.ParamType_Value, p.GuildProfile.GuildName).Send()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) startGateEnemy(acID string) (
	errNotStart bool, warnCode uint32) {
	now_t := time.Now().Unix()
	s, e := gamedata.GetGETime(now_t)
	if now_t < s || now_t >= e {
		return true, 0
	}
	guildID := p.GuildProfile.GetCurrGuildUUID()
	//isStarted := gates_enemy.GetModule(p.AccountID.ShardId).IsActHasStart(guildID)
	//if !isStarted {
	//	guildData, ret := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(guildID)
	//	if ret.HasError() || guildData == nil {
	//		return false, errCode.GuildGateEnemyStartErr
	//	}
	//	ok := guild.GetModule(p.AccountID.ShardId).UseGateEnemyCount(guildID)
	//	if ok {
	//		// 为了让同时点进入的玩家都能进，并不报错，所以这里比检查返回值，谁开启都行，只能开一次
	//		gates_enemy.GetModule(p.AccountID.ShardId).StartGatesEnemyAct(guildID,
	//			e, guildData.Members)
	//	}
	//}
	// 初始化玩家channel
	p.Profile.GetGatesEnemy().InitGetChannel(p.AccountID.ShardId,
		p.AccountID.String(), guildID, p.GetSimpleInfo())

	return false, 0
}
