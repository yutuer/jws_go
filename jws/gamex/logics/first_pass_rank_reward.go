package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) firstPassReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		RewardTyp   int `codec:"fprt"`
		RewardCfgId int `codec:"fprid"`
	}{}

	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"Attr/FirstPassRewardRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		ErrParam
		ErrAlreadyReward
		ErrGive
		ErrRankNotEnough
	)

	fpr := p.Profile.GetFirstPassRank()

	cfg := fpr.GetReward(req.RewardTyp, req.RewardCfgId)
	if cfg == nil {
		return rpcErrorWithMsg(resp, ErrParam, "Err_Param")
	}

	// 名次是否够
	if !fpr.IsRankMaxCanGetReward(req.RewardTyp, cfg.Start) {
		return rpcErrorWithMsg(resp, ErrRankNotEnough, "Err_Rank_Not_Enough")
	}

	// 记录
	if !fpr.AddFirstPassReward(req.RewardTyp, req.RewardCfgId) {
		logs.Warn("firstPassReward Err_Already_Reward")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 发奖
	give := &account.GiveGroup{}
	give.AddCostData(&cfg.Reward.Cost)
	if !give.GiveBySyncAuto(p.Account, resp, "FirstPassReward") {
		return rpcErrorWithMsg(resp, ErrGive, "Err_Give")
	}
	resp.OnChangeFirstPassRewardInfo()

	resp.mkInfo(p)
	return rpcSuccess(resp)
}
