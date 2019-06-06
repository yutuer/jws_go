package logics

import (
	"fmt"
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	guild2 "vcs.taiyouxi.net/jws/gamex/models/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ClaimRedPacketReward : 抢红包或者领取宝箱
// 抢红包或者领取宝箱

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgClaimRedPacketReward 抢红包或者领取宝箱请求消息定义
type reqMsgClaimRedPacketReward struct {
	Req
	ClaimType int64  `codec:"claimtype"` // 0 抢红包, 1 宝箱
	Id        string `codec:"id"`        // 红包ID或者宝箱ID
}

// rspMsgClaimRedPacketReward 抢红包或者领取宝箱回复消息定义
type rspMsgClaimRedPacketReward struct {
	SyncRespWithRewards
}

const (
	Claim_Type_Grab_RP = iota
	Claim_Type_Claim_Box
	Claim_Type_Ipa
)

// ClaimRedPacketReward 抢红包或者领取宝箱: 抢红包或者领取宝箱
func (p *Account) ClaimRedPacketReward(r servers.Request) *servers.Response {
	req := new(reqMsgClaimRedPacketReward)
	rsp := new(rspMsgClaimRedPacketReward)

	initReqRsp(
		"Attr/ClaimRedPacketRewardRsp",
		r.RawBytes,
		req, rsp, p)

	if ok, _ := p.Profile.MarketActivitys.HasRedPacketActivity(p.AccountID.String(), p.GetProfileNowTime()); !ok {
		logs.Warn("claim rp ipa: no available red packet activity")
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}
	_, act := p.Profile.MarketActivitys.HasRedPacketActivity(p.AccountID.String(), p.GetProfileNowTime())
	p.GuildProfile.RedPacketInfo.CheckDailyReset(p.GetProfileNowTime())

	// logic imp begin
	switch req.ClaimType {
	case Claim_Type_Grab_RP:
		if errResp := p.doGrabRP(act, req.Id, rsp); errResp != nil {
			return errResp
		}
	case Claim_Type_Claim_Box:
		p.doClaimRPBox(req.Id, act, rsp)
	case Claim_Type_Ipa:
		p.doClaimIpaReward(act, rsp)
	}
	// logic imp end
	rsp.OnChangeGuildRedPacket()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// 领取红包宝箱
func (p *Account) doClaimRPBox(boxId string, act uint32, resp *rspMsgClaimRedPacketReward) *servers.Response {
	boxIdInt, err := strconv.ParseInt(boxId, 10, 64)
	if err != nil {
		return rpcWarn(resp, errCode.RedPacketErrorBoxId)
	}
	if p.Account.GuildProfile.RedPacketInfo.IsClaimed(boxIdInt) {
		return rpcWarn(resp, errCode.RedPacketHasClaimed)
	}
	guildInfo, ret := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if rsp := guildErrRet(ret, resp); rsp != nil {
		return rsp
	}
	rpCount := guildInfo.GuildRedPacket.GetRpCount()
	boxConfig := gamedata.GetHotDatas().RedPacketConfig.GetRpBoxConfig(int(boxIdInt), act)
	if boxConfig == nil || rpCount < int(boxConfig.GetFCValue2()) {
		return rpcWarn(resp, errCode.RedPacketCannotClaim)
	}
	p.GuildProfile.RedPacketInfo.OnClaim(boxIdInt)
	rewardData := &gamedata.CostData{}
	for _, item := range boxConfig.GetItem_Table() {
		rewardData.AddItem(item.GetItemID(), item.GetItemCount())
	}
	reason := fmt.Sprintf("claim red packet box %s", boxIdInt)
	account.GiveBySync(p.Account, rewardData, resp, reason)
	return nil
}

// 抢红包
func (p *Account) doGrabRP(act uint32, rpId string, resp *rspMsgClaimRedPacketReward) *servers.Response {
	if len(p.GuildProfile.RedPacketInfo.TodayGrabList) >= int(gamedata.GetCommonCfg().GetREDPACKETLIMIT()) {
		return rpcWarn(resp, errCode.RedPacketReachMax)
	}
	ret, senderName, items := guild.GetModule(p.AccountID.ShardId).GrabRedPacket(p.GuildProfile.GuildUUID,
		p.AccountID.String(),
		p.Profile.Name,
		rpId,
		act)
	if rsp := guildErrRet(ret, resp); rsp != nil {
		return rsp
	} else {
		rewardData := &gamedata.CostData{}
		for itemId, count := range items {
			rewardData.AddItem(itemId, count)
		}
		// TODO 检查背包是否已满
		reason := fmt.Sprintf("grab red packet %s", rpId)
		account.GiveBySync(p.Account, rewardData, resp, reason)
		p.Account.GuildProfile.RedPacketInfo.OnGrab(rpId, senderName, items)
		return nil
	}
}

// 领取充值额外奖励
func (p *Account) doClaimIpaReward(act uint32, resp *rspMsgClaimRedPacketReward) *servers.Response {
	if !p.GuildProfile.RedPacketInfo.CanClaimIpa() {
		logs.Warn("claim rp ipa: can not claim: status:", p.GuildProfile.RedPacketInfo.IpaStatus)
		return rpcWarn(resp, errCode.RedPacketCannotClaim)
	}
	p.GuildProfile.RedPacketInfo.IpaStatus = guild2.RP_IPA_HAS_CLAIM
	boxConfig := gamedata.GetHotDatas().RedPacketConfig.GetIpaConfig(act)
	rewardData := &gamedata.CostData{}
	for _, item := range boxConfig.GetItem_Table() {
		rewardData.AddItem(item.GetItemID(), item.GetItemCount())
	}
	reason := fmt.Sprintf("claim red packet ipa %s")
	account.GiveBySync(p.Account, rewardData, resp, reason)
	return nil
}
