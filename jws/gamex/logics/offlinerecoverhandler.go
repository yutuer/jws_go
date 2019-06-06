package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetOfflineRecoverInfo : 获取离线资源相关信息
// 离线资源
func (p *Account) GetOfflineRecoverInfoHandler(req *reqMsgGetOfflineRecoverInfo, resp *rspMsgGetOfflineRecoverInfo) uint32 {
	resources := p.Profile.OfflineRecoverInfo.Resources
	resp.OfflineResources = make([][]byte, 0)
	for _, res := range resources {
		res2Client := buildOfflineResource2Client(res)
		resp.OfflineResources = append(resp.OfflineResources, encode(res2Client))
	}
	return 0
}

func buildOfflineResource2Client(res account.OfflineResource) OfflineResource2Client {
	return OfflineResource2Client{
		ScId:          res.ScId,
		ScOfflineDays: int64(res.OfflineDays),
	}
}

// ClaimOfflineRecoverReward : 领取离线资源的奖励
// 领取离线资源的奖励
func (p *Account) ClaimOfflineRecoverRewardHandler(req *reqMsgClaimOfflineRecoverReward, resp *rspMsgClaimOfflineRecoverReward) uint32 {
	if p.Profile.GetCorp().Level < gamedata.GetCommonCfg().GetRecoverLevel() {
		logs.Warn("<ClaimOfflineRecover> level is not enough, %d", p.Profile.GetCorp().Level)
		return errCode.CommonConditionFalse
	}

	scId := req.ScId
	isFree := req.IsFree

	res := p.Profile.OfflineRecoverInfo.GetResource(scId)
	if res.OfflineDays == 0 {
		return errCode.CommonCountLimit
	}

	config := gamedata.GetOfflineRecoverConfig(scId)
	if config == nil {
		logs.Warn("<ClaimOfflineRecover> config err, %s", scId)
		return errCode.CommonInvalidParam
	}

	if res.OfflineDays == 0 {
		logs.Warn("<ClaimOfflineRecover> no reward days, %s", scId)
		return errCode.CommonConditionFalse
	}

	var rewardPerDay uint32
	days := uint32(res.OfflineDays)

	if !isFree {
		cost := &gamedata.CostData{}
		cost.AddItem(gamedata.VI_Hc, config.GetPayDiamonds()*days)
		if ok := account.CostBySync(p.Account, cost, resp, "claimOfflineRecover"); !ok {
			logs.Warn("<ClaimOfflineRecover> cost err, %v", cost)
			return errCode.ClickTooQuickly
		}
		rewardPerDay = config.GetPayRecover()
	} else {
		rewardPerDay = config.GetFreeRecover()
	}

	res.OfflineDays = 0

	give := &gamedata.CostData{}
	give.AddItem(config.GetResourcesID(), rewardPerDay*days)
	if ok := account.GiveBySync(p.Account, give, resp, "claimOfflineRecover"); !ok {
		logs.Warn("<ClaimOfflineRecover> give err, %v", give)
		return errCode.ClickTooQuickly
	}
	if !p.Profile.OfflineRecoverInfo.HasRewards() {
		p.Profile.OfflineRecoverInfo.LastClaimAllTime = p.Profile.GetProfileNowTime()
	}
	resp.OnChangeOfflineRecover()
	resp.mkInfo(p)
	return 0
}
