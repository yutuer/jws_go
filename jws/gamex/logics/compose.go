package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"

	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestCompose struct {
	Req
	FormulaID []int `codec:"fid"`
}

type ResponseCompose struct {
	SyncRespWithRewards
}

func (p *Account) ItemCompose(fid int, sync interfaces.ISyncRspWithRewards) uint32 {
	const (
		_                  = iota
		CODE_FormulaID_Err // 失败:配方ID不存在
		CODE_Cost_Err      // 失败:配方消耗错误，玩家原料不全
		CODE_Give_Err      // 失败:添加合成结果错误
	)

	acid := p.AccountID.String()

	ok, need, give := gamedata.GetFormulaData(fid)
	if ok {
		cost_group := account.CostGroup{}
		is_has := cost_group.AddCostData(p.Account, need)
		if !is_has {
			logs.SentryLogicCritical(acid, "ItemCompose_Cost_Error:%d,%v,%v",
				fid, need, give)
			return mkCode(CODE_ERR, CODE_Cost_Err)
		}

		is_cost := cost_group.CostBySync(p.Account, sync, "Compose")
		if !is_cost {
			logs.SentryLogicCritical(acid, "ItemCompose_Cost_Run_Error:%d,%v,%v",
				fid, need, give)
			return mkCode(CODE_ERR, CODE_Cost_Err)
		}

		give_group := account.GiveGroup{}
		give_group.AddCostData(give)
		is_give := give_group.GiveBySyncAuto(p.Account, sync, "Compose")

		if !is_give {
			logs.SentryLogicCritical(acid, "ItemCompose_Give_Error:%d,%v,%v",
				fid, need, give)
			// TBD 这里按理应该返还消耗的，不过考虑到防止刷物品所以只记一个ErrorLog
			return mkCode(CODE_ERR, CODE_Give_Err)
		}

	} else {
		return mkCode(CODE_ERR, CODE_FormulaID_Err)
	}

	return 0
}

func (p *Account) Compose(r servers.Request) *servers.Response {
	req := &RequestCompose{}
	resp := &ResponseCompose{}

	initReqRsp(
		"PlayerBag/ComposeResponse",
		r.RawBytes,
		req, resp, p)

	for i := 0; i < len(req.FormulaID); i++ {
		code := p.ItemCompose(req.FormulaID[i], resp)
		resp.Code = code
		if code != 0 {
			logs.SentryLogicCritical(p.AccountID.String(), "ItemComposeErr:%d,%d",
				req.FormulaID, code)
			break
		}
	}

	resp.mkInfo(p)

	return rpcSuccess(resp)
}
