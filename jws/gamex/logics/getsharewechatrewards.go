package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetShareWeChatRewards : 微信分享获取奖励协议
// 每天微信分享可以获得奖励，奖励与每天分享的次数有关

// reqMsgGetShareWeChatRewards 微信分享获取奖励协议请求消息定义
type reqMsgGetShareWeChatRewards struct {
	Req
	Type   int64 `codec:"type"` // 分享类型 type=1 得到新主将， type=2 主将合影
	HeroID int   `codec:"heroID"`
}

// rspMsgGetShareWeChatRewards 微信分享获取奖励协议回复消息定义
type rspMsgGetShareWeChatRewards struct {
	SyncRespWithRewards
}

// GetShareWeChatRewards 微信分享获取奖励协议: 每天微信分享可以获得奖励，奖励与每天分享的次数有关
func (p *Account) GetShareWeChatRewards(r servers.Request) *servers.Response {
	req := new(reqMsgGetShareWeChatRewards)
	rsp := new(rspMsgGetShareWeChatRewards)

	initReqRsp(
		"Attr/GetShareWeChatRewardsRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		NoTypeErr
		CanNotGet
	)
	// logic imp begin
	shareInfo := p.Profile.GetShareWeChatInfo()
	shareInfo.UpdateTimesAndRest(p.Profile.GetProfileNowTime())
	shareInfo.AddTimesByType(int(req.Type))
	times, ok := shareInfo.GetTimesByType(int(req.Type))
	if !ok {
		rpcError(rsp, NoTypeErr)
	}
	data := gamedata.GetShareWeChatDataByKey(uint32(req.Type), uint32(times))
	logs.Debug("ShareWeChat: type: %d, times: %d/%d", req.Type, times, data.GetShareCount())

	// BI log
	logiclog.LogShareWeChat(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		int(p.GetCorpLv()), int(req.Type), req.HeroID, p.Profile.GetHero().GetOwnedHeroCount(),
		helper.AVATAR_NUM_CURR, p.Profile.GetProfileNowTime(), p.Profile.GetVipLevel(),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	if uint32(times) <= data.GetShareCount() {
		// can share and get rewards
		rewards := data.GetFixed_Loot()
		data := &gamedata.CostData{}
		for _, reward := range rewards {
			data.AddItem(reward.GetFixedLootID(), reward.GetFixedLootNumber())
		}
		account.GiveBySync(p.Account, data, rsp, "ShareWeChat")

	} else {
		rsp.OnChangeShareWeChat()
		rsp.mkInfo(p)
		return rpcSuccess(rsp)
	}
	// logic imp end
	rsp.OnChangeShareWeChat()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
