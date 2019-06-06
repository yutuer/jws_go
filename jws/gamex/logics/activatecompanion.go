package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ActivateCompanion : 激活武将羁绊
// 请求激活下一个羁绊

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgActivateCompanion 激活武将羁绊请求消息定义
type reqMsgActivateCompanion struct {
	Req
	HeroId      int64 `codec:"heroid"`      // 武将ID
	CompanionId int64 `codec:"companionid"` // 伙伴武将ID
}

// rspMsgActivateCompanion 激活武将羁绊回复消息定义
type rspMsgActivateCompanion struct {
	SyncResp
}

// ActivateCompanion 激活武将羁绊: 请求激活下一个羁绊
func (p *Account) ActivateCompanion(r servers.Request) *servers.Response {
	req := new(reqMsgActivateCompanion)
	rsp := new(rspMsgActivateCompanion)

	initReqRsp(
		"Attr/ActivateCompanionRsp",
		r.RawBytes,
		req, rsp, p)

	errCode := p.doActivateCompanion(int(req.HeroId), int(req.CompanionId))
	if errCode != 0 {
		return rpcWarn(rsp, uint32(errCode))
	}
	rsp.OnChangeUpdateHeroCompanion(int(req.HeroId))
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) doActivateCompanion(heroIdx, companionId int) int {
	if !p.IsHeroCompanionOpen(heroIdx) {
		logs.Warn("companion is not open, %v, %d, %d", p.AccountID, heroIdx, companionId)
		return errCode.HeroCompanionNotOpen
	}

	companionInfo := p.Profile.GetHero().GetCompanion(heroIdx)
	companion := companionInfo.GetCompanion(heroIdx, companionId)

	if companion == nil {
		logs.Warn("companion is nil, %v, %d, %d", p.AccountID, heroIdx, companionId)
		return errCode.HeroCompanionNotOpen
	}

	// 激活的时候判断是否有下一等级
	activeConfig := gamedata.GetCompanionActiveConfigById(companion.Id)
	if activeConfig == nil || activeConfig.Config.GetRelationArray() == 0 {
		logs.Warn("fail to active for config, %v, %d, %d", p.AccountID, heroIdx, companionId)
		return errCode.ClickTooQuickly
	}
	nextId := companion.Id
	// 大于1级的处理
	if companion.GetLevel() > 0 {
		nextCfg := gamedata.GetCompanionActiveConfigById(int(activeConfig.Config.GetRelationArray()))
		if nextCfg == nil {
			logs.Warn("fail to active for config, %v, %d, %d", p.AccountID, heroIdx, companionId)
			return errCode.ClickTooQuickly
		}
		nextId = int(nextCfg.Config.GetUniqueID())
		activeConfig = nextCfg
	}
	if p.canActive(companionId, activeConfig) {
		companion.Id = nextId
		companion.Active = true
		companion.UpdateLevelAndCompanion()
		p.Profile.GetData().SetNeedCheckMaxGS()
	} else {
		logs.Warn("fail to active, %v, %d, %d", p.AccountID, heroIdx, companionId)
		return errCode.ClickTooQuickly
	}
	return 0
}

const (
	active_companion_hero_level = iota + 1
	active_companion_star_level
	active_companion_wing_level
)

func (p *Account) canActive(companionId int, config *gamedata.CompanionActiveConfig) bool {
	switch int(config.Config.GetActiveType()) {
	case active_companion_hero_level:
		return p.Profile.GetHero().HeroLevel[companionId] >= config.Config.GetActivePara()
	case active_companion_star_level:
		return p.Profile.GetHero().HeroStarLevel[companionId] >= config.Config.GetActivePara()
	case active_companion_wing_level:
		return p.Profile.GetHero().HeroSwings[companionId].StarLv >= int(config.Config.GetActivePara())
	}
	return false
}

// EvolveCompanion : 进化羁绊
// 请求进化下一级羁绊

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgEvolveCompanion 进化羁绊请求消息定义
type reqMsgEvolveCompanion struct {
	Req
	HeroId int64 `codec:"heroid"` // 要进阶的武将
}

// rspMsgEvolveCompanion 进化羁绊回复消息定义
type rspMsgEvolveCompanion struct {
	SyncResp
}

// EvolveCompanion 进化羁绊: 请求进化下一级羁绊
func (p *Account) EvolveCompanion(r servers.Request) *servers.Response {
	req := new(reqMsgEvolveCompanion)
	rsp := new(rspMsgEvolveCompanion)

	initReqRsp(
		"Attr/EvolveCompanionRsp",
		r.RawBytes,
		req, rsp, p)

	errCode := p.doEvolveCompanion(int(req.HeroId))
	if errCode != 0 {
		return rpcWarn(rsp, uint32(errCode))
	}
	rsp.OnChangeUpdateHeroCompanion(int(req.HeroId))
	p.Profile.GetData().SetNeedCheckMaxGS()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) doEvolveCompanion(heroIdx int) int {
	companionInfo := p.Profile.GetHero().GetCompanion(heroIdx)
	errCode := companionInfo.CanEvolve()
	if errCode != 0 {
		return errCode
	}
	companionInfo.IncEvolveLevel()
	return 0
}
