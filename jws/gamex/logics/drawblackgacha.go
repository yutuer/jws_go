package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// DrawBlackGacha : 黑盒宝箱抽奖
// 黑盒宝箱抽奖

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgDrawBlackGacha 黑盒宝箱抽奖请求消息定义
type reqMsgDrawBlackGacha struct {
	Req
	BlackGachaId    int64 `codec:"bg_id"`    // 想要抽奖的ID
	BlackGachaSubId int64 `codec:"bg_subid"` // 想要抽奖的SubID
	IsTen           bool  `codec:"is_ten"`   // 是否是十连抽
}

// rspMsgDrawBlackGacha 黑盒宝箱抽奖回复消息定义
type rspMsgDrawBlackGacha struct {
	SyncRespWithRewards
	GiveRewardId    string   `codec:"bg_grid"`       // 想要抽奖的ID
	GiveRewardCount int64    `codec:"bg_grc"`        // 想要抽奖的ID
	ExtRewardId     []string `codec:"bg_extid"`      // 想要抽奖的ID
	ExtRewardCount  []int64  `codec:"bg_extc"`       // 想要抽奖的ID
	ExtRewardData   []string `codec:"bg_extd"`       // 想要抽奖的ID
	BlackGachaId    int64    `codec:"bg_id"`         // 想要抽奖的ID
	GachaInfo       []byte   `codec:"bg_gacha_info"` // 武将巡礼的活动内容
}

// DrawBlackGacha 黑盒宝箱抽奖: 黑盒宝箱抽奖
func (p *Account) DrawBlackGacha(r servers.Request) *servers.Response {
	req := new(reqMsgDrawBlackGacha)
	rsp := new(rspMsgDrawBlackGacha)

	initReqRsp(
		"Attr/DrawBlackGachaRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.DrawBlackGachaHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetBlackGachaInfo : 黑盒宝箱信息
// 黑盒宝箱信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetBlackGachaInfo 黑盒宝箱信息请求消息定义
type reqMsgGetBlackGachaInfo struct {
	Req
}

// rspMsgGetBlackGachaInfo 黑盒宝箱信息回复消息定义
type rspMsgGetBlackGachaInfo struct {
	SyncResp
	HeroActivityId   int64    `codec:"bg_hero_id"`     // 武将巡礼的activityId
	HeroInfo         [][]byte `codec:"bg_hero_info"`   // 武将巡礼的活动内容
	WeaponActivityId int64    `codec:"bg_weapon_id"`   // 神兵降临的activityId
	WeaponInfo       [][]byte `codec:"bg_weapon_info"` // 神兵再临的活动内容
}

// GetBlackGachaInfo 黑盒宝箱信息: 黑盒宝箱信息
func (p *Account) GetBlackGachaInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetBlackGachaInfo)
	rsp := new(rspMsgGetBlackGachaInfo)

	initReqRsp(
		"Attr/GetBlackGachaInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetBlackGachaInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BlackGachaActivity 黑盒宝箱信息
type BlackGachaActivity struct {
	TodayFreeUsedCount int64   `codec:"bg_free"`   // 今天免费次数的使用次数
	GachaCount         int64   `codec:"bg_count"`  // 累计抽奖次数
	HasClaimedReward   []int64 `codec:"bg_reward"` // 已经领取的进度宝箱
	SubActivityId      int64   `codec:"bg_subid"`  // 活动ID
}

// ClaimBlackGachaExtraReward : 获取黑盒宝箱的额外奖励
// 累计抽奖一定次数后会有额外的奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgClaimBlackGachaExtraReward 获取黑盒宝箱的额外奖励请求消息定义
type reqMsgClaimBlackGachaExtraReward struct {
	Req
	BlackGachaActivityId    int64 `codec:"bg_id"`        // 黑盒活动具体ID
	BlackGachaActivitySubId int64 `codec:"bg_sub_id"`    // 黑盒活动具体ID
	RewardId                int64 `codec:"bg_reward_id"` // 奖励ID
}

// rspMsgClaimBlackGachaExtraReward 获取黑盒宝箱的额外奖励回复消息定义
type rspMsgClaimBlackGachaExtraReward struct {
	SyncRespWithRewards
	BlackGachaActivityId    int64  `codec:"bg_id"`         // 黑盒活动具体ID
	BlackGachaActivitySubId int64  `codec:"bg_sub_id"`     // 黑盒活动具体ID
	GachaInfo               []byte `codec:"bg_gacha_info"` // 活动内容
}

// ClaimBlackGachaExtraReward 获取黑盒宝箱的额外奖励: 累计抽奖一定次数后会有额外的奖励
func (p *Account) ClaimBlackGachaExtraReward(r servers.Request) *servers.Response {
	req := new(reqMsgClaimBlackGachaExtraReward)
	rsp := new(rspMsgClaimBlackGachaExtraReward)

	initReqRsp(
		"Attr/ClaimBlackGachaExtraRewardRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ClaimBlackGachaExtraRewardHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
