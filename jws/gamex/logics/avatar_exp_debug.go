package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models/account"
	//"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	//"vcs.taiyouxi.net/jws/gamex/models/helper"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// Just Debug
type RequestAvatarExpOp struct {
	Req
	OpType int    `codec:"op"`
	SCType int    `codec:"typ"`
	Value  uint32 `codec:"value"`
}

type ResponseAvatarExpOp struct {
	Resp
	Exps []uint32 `codec:"exps"`
}

func (p *Account) AvatarExpOp(r servers.Request) *servers.Response {
	req := &RequestAvatarExpOp{}
	resp := &ResponseAvatarExpOp{}

	initReqRsp(
		"Debug/AvatarExpOpResponse",
		r.RawBytes,
		req, resp, p)

	logs.Error("[%s] Debug Op AvatarExp %d,%d,%d.",
		p.AccountID, req.OpType, req.SCType, req.Value)

	if req.OpType == 1 {
		p.Profile.GetHero().AddHeroExp(p.Account, req.SCType, req.Value)
	} else if req.OpType == 2 {
		if req.SCType < len(p.Profile.GetAvatarExp().Avatars) {
			p.Profile.GetAvatarExp().Avatars[req.SCType].Level = req.Value
			p.Profile.GetAvatarExp().Avatars[req.SCType].Xp = 0
		}
	}

	resp.Exps = p.Profile.GetAvatarExp().GetAll()
	//p.updateCondition(account.COND_TYP_Hero_Star_Together,
	//	0, 0, "", "", resp)
	//p.updateCondition(account.COND_TYP_Hero_Lvl_Together,
	//	0, 0, "", "", resp)
	//logs.Trace("sc %v", p.Profile.GetAvatarExp())

	return rpcSuccess(resp)
}
