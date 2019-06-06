package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/error_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GuildBossHeartBeatReq : 战斗中心跳包，带有当前打掉的Boss血量
// 请求挑战boss，锁定boss

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGuildBossHeartBeatReq 战斗中心跳包，带有当前打掉的Boss血量请求消息定义
type reqMsgGuildBossHeartBeatReq struct {
	Req
	BossID    string `codec:"bossid"`    // 挑战的bossID
	BossGroup string `codec:"bossgroup"` // 要挑战的bossGroup
	BossHp    int64  `codec:"bosshp"`    // 挑战boss当前血量
}

// rspMsgGuildBossHeartBeatReq 战斗中心跳包，带有当前打掉的Boss血量回复消息定义
type rspMsgGuildBossHeartBeatReq struct {
	SyncResp
}

// GuildBossHeartBeatReq 战斗中心跳包，带有当前打掉的Boss血量: 请求挑战boss，锁定boss
func (p *Account) GuildBossHeartBeatReq(r servers.Request) *servers.Response {
	req := new(reqMsgGuildBossHeartBeatReq)
	rsp := new(rspMsgGuildBossHeartBeatReq)

	initReqRsp(
		"Attr/GuildBossHeartBeatReqRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}
	simpleInfo := p.GetSimpleInfo()
	res := guild.GetModule(p.AccountID.ShardId).ActBossNotify(
		p.GuildProfile.GuildUUID,
		req.BossID,
		req.BossGroup, &simpleInfo)
	if res != 0 {
		return rpcWarn(rsp, error_code.GetWarnCodeFromErr(res))
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GuildBossLockReq : 点击挑战，锁定boss
// 请求挑战boss，锁定boss

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGuildBossLockReq 点击挑战，锁定boss请求消息定义
type reqMsgGuildBossLockReq struct {
	Req
	BossID    string `codec:"bossid"`    // 要挑战的bossID
	BossGroup string `codec:"bossgroup"` // 要挑战的bossGroup
}

// rspMsgGuildBossLockReq 点击挑战，锁定boss回复消息定义
type rspMsgGuildBossLockReq struct {
	SyncResp
}

// GuildBossLockReq 点击挑战，锁定boss: 请求挑战boss，锁定boss
func (p *Account) GuildBossLockReq(r servers.Request) *servers.Response {
	req := new(reqMsgGuildBossLockReq)
	rsp := new(rspMsgGuildBossLockReq)

	initReqRsp(
		"Attr/GuildBossLockReqRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	simpleInfo := p.GetSimpleInfo()
	res := guild.GetModule(p.AccountID.ShardId).ActBossLockFight(
		p.GuildProfile.GuildUUID,
		req.BossID,
		req.BossGroup, &simpleInfo)
	if res != 0 {
		return rpcWarn(rsp, error_code.GetWarnCodeFromErr(res))
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GuildBossFinishReq : 挑战boss结束
// 请求挑战boss，锁定boss

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGuildBossFinishReq 挑战boss结束请求消息定义
type reqMsgGuildBossFinishReq struct {
	ReqWithAnticheat
	BossID    string `codec:"bossid"`    // 开始挑战的bossID
	BossGroup string `codec:"bossgroup"` // 要挑战的bossGroup
	IsSuccess int64  `codec:"issuccess"` // 是否挑战成功
	HpChange  int64  `codec:"hpchange"`  // 本次挑战消耗的BossHP
}

// rspMsgGuildBossFinishReq 挑战boss结束回复消息定义
type rspMsgGuildBossFinishReq struct {
	SyncRespWithRewardsAnticheat
	ShowItemId    string `codec:"show_item_id"`    // 军团奖励界面显示的物品掉落
	ShowItemCount int64  `codec:"show_item_count"` // 军团奖励界面显示的物品掉落
}

// GuildBossFinishReq 挑战boss结束: 请求挑战boss，锁定boss
func (p *Account) GuildBossFinishReq(r servers.Request) *servers.Response {
	req := new(reqMsgGuildBossFinishReq)
	rsp := new(rspMsgGuildBossFinishReq)

	initReqRsp(
		"Attr/GuildBossFinishReqRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	// 反作弊检查
	if cheatRsp := p.AntiCheatCheck(&rsp.SyncRespWithRewardsAnticheat, &req.ReqWithAnticheat, 0,
		account.Anticheat_Typ_GuildBoss); cheatRsp != nil {
		return cheatRsp
	}
	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	costTyp := counter.CounterTypeFreeGuildBoss
	if p.Tmp.BossIdx == 3 {
		costTyp = counter.CounterTypeFreeGuildBigBoss
	}

	if !p.Profile.GetCounts().Has(
		costTyp,
		p.Account) {
		return rpcWarn(rsp, errCode.GuildBossCount)
	}

	acId := p.AccountID.String()

	simpleInfo := p.GetSimpleInfo()
	res, sc, loot, leftHp, itemC, realGbCount := guild.GetModule(p.AccountID.ShardId).ActBossEndFight(
		p.GuildProfile.GuildUUID,
		req.BossID,
		req.BossGroup, req.HpChange, &simpleInfo)
	if res != 0 {
		return rpcWarn(rsp, error_code.GetWarnCodeFromErr(res))
	}

	if !p.Profile.GetCounts().Use(
		costTyp,
		p.Account) {
		return rpcWarn(rsp, errCode.GuildBossCount)
	}

	logs.Trace("ActBossEndFight %v %v", sc, loot)

	priceGive, err := p.GetGivesByTemplate(loot)
	if err != nil {
		logs.SentryLogicCritical(acId, "GetGivesByTemplate Err By %s", loot)
	}
	priceGive.Cost.SetGST(gamedata.GST_BossFight)
	for item, c := range itemC {
		priceGive.AddItem(item, c)
	}

	logs.Trace("priceGive %v", priceGive)

	if !account.GiveBySync(p.Account, &priceGive.Cost, rsp, "GuildBoss") {
		logs.Error("Give Player Err %v", priceGive)
		return rpcError(rsp, 1)
	}

	// logic imp end

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeFreeGuildBoss,
		1,
		p.Profile.GetProfileNowTime())

	rsp.OnChangeGameMode(uint32(costTyp))
	rsp.mkInfo(p)
	rsp.ShowItemId = gamedata.VI_GuildBoss
	rsp.ShowItemCount = realGbCount

	// log
	acData, ok := gamedata.GetAcData(req.BossID)
	if ok {
		totalHp := float64(acData.GetHitPoint())
		logiclog.LogGuildBossFight(p.AccountID.String(), p.Profile.CurrAvatar,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			p.GuildProfile.GuildUUID, float64(req.HpChange)/totalHp, float64(leftHp)/totalHp,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}
	return rpcSuccess(rsp)
}

// GuildBossBeginReq : loading结束正式开始打boss，开始倒计时，其他玩家表现为显示倒计时
// 请求挑战boss，锁定boss

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGuildBossBeginReq loading结束正式开始打boss，开始倒计时，其他玩家表现为显示倒计时请求消息定义
type reqMsgGuildBossBeginReq struct {
	Req
	BossID    string `codec:"bossid"`    // 开始挑战的bossID
	BossGroup string `codec:"bossgroup"` // 要挑战的bossGroup
}

// rspMsgGuildBossBeginReq loading结束正式开始打boss，开始倒计时，其他玩家表现为显示倒计时回复消息定义
type rspMsgGuildBossBeginReq struct {
	SyncResp
}

// GuildBossBeginReq loading结束正式开始打boss，开始倒计时，其他玩家表现为显示倒计时: 请求挑战boss，锁定boss
func (p *Account) GuildBossBeginReq(r servers.Request) *servers.Response {
	req := new(reqMsgGuildBossBeginReq)
	rsp := new(rspMsgGuildBossBeginReq)

	initReqRsp(
		"Attr/GuildBossBeginReqRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}
	simpleInfo := p.GetSimpleInfo()
	res, bossIdx := guild.GetModule(p.AccountID.ShardId).ActBossBeginFight(
		p.GuildProfile.GuildUUID,
		req.BossID,
		req.BossGroup, &simpleInfo)

	if res != 0 {
		return rpcWarn(rsp, error_code.GetWarnCodeFromErr(res))
	}
	p.Tmp.BossIdx = bossIdx
	// logic imp end

	// 条件更新
	p.updateCondition(account.COND_TYP_Guild_Boss, 1, 0, "", "", rsp)

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
