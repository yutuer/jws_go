package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// Starhero : 将星激活
// 用来传输激活的武将以及所激活武将的技能

// reqMsgStarhero 将星激活请求消息定义
type reqMsgStarhero struct {
	Req
	HeroId  int64  `codec:"heroid"`      // 主将ID
	SkillID string `codec:"heroskillid"` // 主将技能ID
}

// rspMsgStarhero 将星激活回复消息定义
type rspMsgStarhero struct {
	SyncRespWithRewards
}

// Starhero 将星激活: 用来传输激活的武将以及所激活武将的技能
func (p *Account) Starhero(r servers.Request) *servers.Response {
	req := new(reqMsgStarhero)
	rsp := new(rspMsgStarhero)

	initReqRsp(
		"Attr/StarheroRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_Err
		Err_Hero_star_No
		Err_Hero_Skill_Actived
		Err_Hero_Not_Found
	)

	info := gamedata.GetHeroData(int(req.HeroId))
	if info == nil || req.HeroId >= account.AVATAR_NUM_MAX {
		return rpcErrorWithMsg(rsp, Err_Hero_Not_Found, "Err_Hero_Not_Found")
	}
	pHero := p.Profile.GetHero()
	starLv := pHero.HeroStarLevel[int64(req.HeroId)]

	idid1 := info.LvData[starLv].IdId
	pskill := gamedata.GetpskillByidid(idid1)
	cskill := gamedata.GetcskillByidid(idid1)
	tskill := gamedata.GettskillByidid(idid1)
	allskill := pskill + cskill + tskill

	if !strings.Contains(allskill, req.SkillID) {
		return rpcErrorWithMsg(rsp, Err_Hero_star_No, "Err_Hero_Not_Enough")
	}
	if pHero.IsHeroSkillActive(int(req.HeroId), req.SkillID) {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	data := gamedata.GetSkillCostData(req.SkillID)

	if !account.CostBySync(p.Account, data, rsp, "StarSkillCost") {
		return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
	}
	pHero.AddSkill2Hero(int(req.HeroId), req.SkillID)
	// log
	logiclog.LogPassiveSkillAdd(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		req.SkillID,
		int(p.Profile.GetData().CorpCurrGS),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	p.Profile.GetHero().SyncObj.SetNeedSync()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
