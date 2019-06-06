package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) heroStarUp(r servers.Request) *servers.Response {
	req := &struct {
		Req
		HeroIdx int `codec:"heroIdx"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/HeroStarUpRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Cfg
		Err_VerClose
		Err_Cost
	)
	logs.Trace("[%s]heroStarUp %d", p.AccountID.String(), req.HeroIdx)

	info := gamedata.GetHeroData(req.HeroIdx)
	if info == nil {
		return rpcErrorWithMsg(resp, Err_Cfg, "Err_Cfg")
	}

	if !info.IsInCurrVersion {
		return rpcErrorWithMsg(resp, Err_VerClose, "Err_VerClose")
	}

	ok, isUnlock, oldStar, newStar := p.Profile.GetHero().StarUp(
		p.Account, req.HeroIdx, resp)

	if !ok {
		logs.Warn("heroStarUp StarUp failed")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	resp.OnChangeUpdateHeroStarPiece(int(req.HeroIdx))
	resp.OnChangeUpdateHeroStarLevel(int(req.HeroIdx))

	if isUnlock {
		onAvatarUnlock(p, req.HeroIdx)
	}
	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 主将升星

	resp.OnChangeHeroTalent()

	// 检查是否能激活情缘
	if p.willOpenCompanion(req.HeroIdx) {
		resp.OnChangeUpdateHeroCompanion(int(req.HeroIdx))
	}

	resp.mkInfo(p)

	// log
	logiclog.LogHeroStarUp(
		p.AccountID.String(),
		p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, req.HeroIdx,
		oldStar, newStar, p.Profile.GetData().CorpCurrGS, p.Profile.Hero.HeroStarPiece[req.HeroIdx],
		func(last string) string {
			return p.Profile.GetLastSetCurLogicLog(last)
		},
		"")

	return rpcSuccess(resp)
}

func (p *Account) willOpenCompanion(heroIdx int) bool {
	return p.IsHeroCompanionOpen(heroIdx) && !p.Profile.GetHero().HeroCompanionInfos[heroIdx].HasCompanions(heroIdx)
}

// HeroTalentLevelUp : 主将天赋升级
// 主将天赋升级的协议

// reqMsgHeroTalentLevelUp 主将天赋升级请求消息定义
type reqMsgHeroTalentLevelUp struct {
	Req
	HeroId   int64 `codec:"hid"` // 主将ID
	TalentId int64 `codec:"tid"` // 天赋ID
}

// rspMsgHeroTalentLevelUp 主将天赋升级回复消息定义
type rspMsgHeroTalentLevelUp struct {
	SyncResp
}

// HeroTalentLevelUp 主将天赋升级: 主将天赋升级的协议
func (p *Account) HeroTalentLevelUp(r servers.Request) *servers.Response {
	req := new(reqMsgHeroTalentLevelUp)
	rsp := new(rspMsgHeroTalentLevelUp)

	initReqRsp(
		"Attr/HeroTalentLevelUpRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
		Err_Unlock
		Err_Full_Lvl
		Err_Cost
	)

	tal := p.Profile.GetHeroTalent()
	now_t := p.Profile.GetProfileNowTime()
	tal.UpdateTalentPoint(now_t)

	// 判断解锁
	talCfg := gamedata.GetHeroTalentConfig(uint32(req.TalentId))
	if talCfg == nil {
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}
	if int(req.HeroId) >= len(p.Profile.GetHero().HeroStarLevel) ||
		p.Profile.GetHero().HeroStarLevel[int(req.HeroId)] <
			talCfg.GetUnlockStarLevel() {
		return rpcErrorWithMsg(rsp, Err_Unlock, "Err_Unlock")
	}

	// 升级
	curLvl := tal.HeroTalentLevel[int(req.HeroId)][int(req.TalentId)]
	if curLvl <= 0 { // 保护代码，防止天赋技能没在星级改变的时候解锁的情况
		star := p.Profile.GetHero().HeroStarLevel[req.HeroId]
		tal.ActTalentByStar(int(req.HeroId), star)
		curLvl = tal.HeroTalentLevel[int(req.HeroId)][int(req.TalentId)]
	}
	lvlCfg := gamedata.GetHeroTalentLevelConfig(curLvl)
	if lvlCfg == nil {
		logs.Warn("HeroTalentLevelUp Err_Full_Lvl %d", curLvl)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	// 消耗
	if !p.Profile.GetSC().HasSC(helper.SCId(lvlCfg.GetHeroTalentCoin()),
		int64(lvlCfg.GetHeroTalentCost())) {
		logs.Warn("%s Err_Cost sc %s %d", p.AccountID.String(),
			lvlCfg.GetHeroTalentCoin(), lvlCfg.GetHeroTalentCost())
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if tal.TalentPoint <= 0 {
		logs.Warn("%s Err_Cost tal point %d", p.AccountID.String(),
			tal.TalentPoint)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	data := &gamedata.CostData{}
	data.AddItem(lvlCfg.GetHeroTalentCoin(), lvlCfg.GetHeroTalentCost())
	if !account.CostBySync(p.Account, data, rsp, "HeroTalentLevelUp") {
		logs.Warn("%s Err_Cost sc %s %d", p.AccountID.String(),
			lvlCfg.GetHeroTalentCoin(), lvlCfg.GetHeroTalentCost())
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	tal.UseTalentPoint(now_t)

	tal.HeroTalentLevel[int(req.HeroId)][int(req.TalentId)] = curLvl + 1

	// gs change
	p.Profile.GetData().SetNeedCheckMaxGS()

	rsp.OnChangeHeroTalent()
	rsp.mkInfo(p)

	logiclog.LogTalentLvUp(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		uint32(req.TalentId),
		curLvl,
		tal.HeroTalentLevel[int(req.HeroId)][int(req.TalentId)],
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return rpcSuccess(rsp)
}

func (p *Account) debugSetHeroTalentLevel(heroId int, talId uint32, lvl uint32) {
	tal := p.Profile.GetHeroTalent()
	now_t := p.Profile.GetProfileNowTime()
	tal.UpdateTalentPoint(now_t)

	talCfg := gamedata.GetHeroTalentConfig(talId)
	if talCfg == nil {
		return
	}
	if heroId >= len(p.Profile.GetHero().HeroStarLevel) ||
		p.Profile.GetHero().HeroStarLevel[int(heroId)] <
			talCfg.GetUnlockStarLevel() {
		return
	}
	tal.HeroTalentLevel[heroId][int(talId)] = lvl
}

// HeroSoulLevelUp : 主将武魂升级
// 主将武魂升级的协议
// reqMsgHeroSoulLevelUp 主将武魂升级请求消息定义
type reqMsgHeroSoulLevelUp struct {
	Req
}

// rspMsgHeroSoulLevelUp 主将武魂升级回复消息定义
type rspMsgHeroSoulLevelUp struct {
	SyncResp
}

// HeroSoulLevelUp 主将武魂升级: 主将武魂升级的协议
func (p *Account) HeroSoulLevelUp(r servers.Request) *servers.Response {
	req := new(reqMsgHeroSoulLevelUp)
	rsp := new(rspMsgHeroSoulLevelUp)

	initReqRsp(
		"Attr/HeroSoulLevelUpRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Hero_GS_Not_Enough
	)
	lv := p.Profile.GetHeroSoul().HeroSoulLevel
	cfg := gamedata.GetHeroSoulLvlConfig(lv + 1)
	if cfg == nil {
		return rpcWarn(rsp, errCode.QuestErrFinish)
	}

	if p.Profile.Data.HeroBaseGSSum_Max < int(cfg.GetHeroSoulValue()) {
		return rpcSuccess(rsp)
	}
	p.Profile.GetHeroSoul().HeroSoulLevel = lv + 1

	// gs变化
	p.Profile.GetData().SetNeedCheckMaxGS()

	rsp.OnChangeHeroSoul()
	rsp.mkInfo(p)

	logiclog.LogHeroSoulLvUp(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		lv,
		p.Profile.GetHeroSoul().HeroSoulLevel,
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return rpcSuccess(rsp)
}

// ChangeHeroTeam : 改变主将阵容
// 改变主将阵容的协议
// reqMsgChangeHeroTeam 改变主将阵容请求消息定义
type reqMsgChangeHeroTeam struct {
	Req
	Typ      int64   `codec:"typ"`    // 玩法类型ID
	HeroTeam []int64 `codec:"herotm"` // 主将阵容
}

// rspMsgChangeHeroTeam 改变主将阵容回复消息定义
type rspMsgChangeHeroTeam struct {
	SyncResp
}

// ChangeHeroTeam 改变主将阵容: 改变主将阵容的协议
func (p *Account) ChangeHeroTeam(r servers.Request) *servers.Response {
	req := new(reqMsgChangeHeroTeam)
	rsp := new(rspMsgChangeHeroTeam)

	initReqRsp(
		"Attr/ChangeHeroTeamRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
	)

	if int(req.Typ) >= len(p.Profile.GetHeroTeams().HeroTeams) {
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}

	tm := &p.Profile.GetHeroTeams().HeroTeams[req.Typ]
	tm.Team = make([]int, len(req.HeroTeam))
	for i, h := range req.HeroTeam {
		tm.Team[i] = int(h)
	}
	rsp.OnChangeHeroTeam()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
