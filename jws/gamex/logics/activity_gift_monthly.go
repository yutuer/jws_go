package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
)

type RequestGetGiftMonthly struct {
	Req
	ReplenishSign bool `codec:"re_sign"`
}

type ResponseGetGiftMonthly struct {
	SyncRespWithRewards
}

func (p *Account) GetGiftMonthly(r servers.Request) *servers.Response {
	req := &RequestGetGiftMonthly{}
	resp := &ResponseGetGiftMonthly{}

	initReqRsp(
		"PlayerAttr/GetGiftMonthlyResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                = iota
		CODE_No_Gift_Err // 失败:没奖可领
		CODE_Give_Err    // 失败:发奖错误
	)

	player_gift := p.Profile.GetGiftMonthly()
	player_vip := p.Profile.GetVipLevel()

	has_gift, data := player_gift.GetGiftToGet(player_vip, p.Profile.GetProfileNowTime(), req.ReplenishSign)

	if has_gift && data != nil {
		if req.ReplenishSign {
			// 补签扣钱
			costItem := &gamedata.CostData{}
			repairSignCost := gamedata.GetCommonCfg().GetRepairSignCost()
			costItem.AddItem(gamedata.VI_Hc, repairSignCost) // 读配置
			ok := account.CostBySync(p.Account, costItem, resp, "MonthlyGift")
			if !ok {
				return rpcWarn(resp, errCode.ClickTooQuickly)
			}
		}
		setOk := player_gift.SetHasGet(player_vip, p.Profile.GetProfileNowTime(), req.ReplenishSign)
		if setOk != 0 {
			return rpcWarn(resp, setOk)
		}

		for _, dataIt := range data {
			ok := account.GiveBySyncWithoutMerge(p.Account, dataIt, resp, "MonthlyGift")
			if !ok {
				return rpcError(resp, CODE_Give_Err)
			}
		}

		resp.OnChangeMonthlyGiftStateChange()
		resp.mkInfo(p)
	} else {
		// 正常情况下客户端不应该明知没奖励还要申请领奖
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	return rpcSuccess(resp)
}
