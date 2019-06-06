package logics

import (
	//"time"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	//"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"

	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/global_info"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) BossFightBegin(r servers.Request) *servers.Response {
	const (
		_ = iota
		CODE_Err_IDX
		CODE_Err_Damage
		CODE_Err_CostErr
	)

	player_boss := p.Profile.GetBoss()
	player_count := p.Profile.GetCounts()

	req := &struct {
		Req
		BossIdx int `codec:"bid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/BossBeginRsp",
		r.RawBytes,
		req, resp, p)

	boss_to_fight := player_boss.GetBoss(req.BossIdx)

	if boss_to_fight == nil {
		return rpcError(resp, CODE_Err_IDX)
	}

	if boss_to_fight.IsNil() {
		return rpcError(resp, CODE_Err_IDX)
	}

	if !player_count.UseJustDayBegin(counter.CounterTypeBoss, p.Account) {
		return rpcWarn(resp, errCode.CommonConditionFalse)
	}

	p.Tmp.SetLevelEnterTime(time.Now().Unix())

	resp.OnChangeGameMode(counter.CounterTypeBoss)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) BossFightEnd(r servers.Request) *servers.Response {
	const (
		_ = iota
		CODE_Err_IDX
		CODE_Err_Damage
		CODE_Err_CostErr
		CODE_Err_GiveErr
	)

	player_boss := p.Profile.GetBoss()
	player_count := p.Profile.GetCounts()

	acid := p.AccountID.String()
	req := &struct {
		ReqWithAnticheat
		AvatarId int   `codec:"aid"`
		BossIdx  int   `codec:"bid"`
		HpDamage int64 `codec:"hp"`
	}{}
	resp := &struct {
		SyncRespWithRewardsAnticheat
		PointAdd uint32 `codec:"padd"`
	}{}

	initReqRsp(
		"PlayerAttr/BossEndRsp",
		r.RawBytes,
		req, resp, p)
	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&resp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_BossFight); cheatRsp != nil {
		return cheatRsp
	}
	boss_to_fight := player_boss.GetBoss(req.BossIdx)

	if boss_to_fight == nil {
		return rpcError(resp, CODE_Err_IDX)
	}

	if boss_to_fight.IsNil() {
		return rpcError(resp, CODE_Err_IDX)
	}

	/*
		check_ok := CheckBossFight(
			p, boss_to_fight.BossTyp, req.HpDamage)

		if !check_ok {
			logs.SentryLogicCritical(acid,
				"CheckBossFight Err by %v to %d %d",
				*boss_to_fight, req.HpDamage, req.AvatarID)
			return rpcError(resp, CODE_Err_Damage)
		}
	*/

	err, isSuccess := player_boss.BossFightDamage(acid, req.BossIdx, req.HpDamage)

	if err != nil {
		logs.Error("BossFightDamage Err By %d %v", req.BossIdx, player_boss)
		return rpcError(resp, CODE_Err_IDX)
	}

	if isSuccess {
		if !player_count.UseJustDayEnd(counter.CounterTypeBoss, p.Account) {
			logs.Warn("BossFightEnd CODE_Err_CostErr")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}

		if !account.GiveBySync(p.Account, &boss_to_fight.Rewards.Cost, resp, "boss") {
			return rpcError(resp, CODE_Err_GiveErr)
		}
		// sysnotice
		global_info.OnBossFinish(p.AccountID.ShardId, boss_to_fight.BossTyp,
			p.AccountID.String(), p.Profile.Name)
	}

	p.updateCondition(account.COND_TYP_Boss_Fight_Count, 1, 0, "", "", resp)

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeBoss,
		1,
		p.Profile.GetProfileNowTime())

	resp.OnChangeBoss()
	resp.OnChangeSC()
	resp.OnChangeGameMode(counter.CounterTypeBoss)

	resp.mkInfo(p)
	// 记log
	costTime := time.Now().Unix() - p.Tmp.GetLevelEnterTime()
	gs := GetCurrGS(p.Account)
	logiclog.LogPveBoss(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		gs, isSuccess,
		costTime, boss_to_fight.BossTyp, boss_to_fight.Degree, boss_to_fight.GS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) BossSweep(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/BossSweepRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Boss
		CODE_Err_IDX
		Err_Give
	)

	player_boss := p.Profile.GetBoss()
	if player_boss.MaxDegree <= 0 {
		return rpcErrorWithMsg(resp, Err_Boss, "Err_Boss")
	}
	boss_to_fight := player_boss.GetBoss(player_boss.MaxDegree2BossIdx)

	if boss_to_fight == nil {
		return rpcErrorWithMsg(resp, CODE_Err_IDX, "CODE_Err_IDX")
	}

	if boss_to_fight.IsNil() {
		return rpcErrorWithMsg(resp, CODE_Err_IDX, "CODE_Err_IDX")
	}

	// 次数减少
	ok, errcode, warnCode, leftC := p.Profile.GetGameMode().GameModeLevelSweep(p.Account, counter.CounterTypeBoss, resp)
	if !ok {
		if warnCode > 0 {
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		return rpcError(resp, errcode+20)
	}

	give := &account.GiveGroup{}
	for i := 0; i < leftC; i++ {
		give.AddCostData(&boss_to_fight.Rewards.Cost)
	}
	if !give.GiveBySyncAuto(p.Account, resp, "BossSweep") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeBoss,
		leftC,
		p.Profile.GetProfileNowTime())

	p.updateCondition(account.COND_TYP_Boss_Fight_Count, leftC, 0, "", "", resp)

	resp.OnChangeBoss()
	resp.OnChangeGameMode(counter.CounterTypeBoss)
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
