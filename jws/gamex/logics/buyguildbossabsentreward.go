package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// BuyGuildBossAbsentReward : 购买公会BOSS未参与的奖励
// 购买公会BOSS未参与的奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBuyGuildBossAbsentReward 购买公会BOSS未参与的奖励请求消息定义
type reqMsgBuyGuildBossAbsentReward struct {
	Req
	BossType int64 `codec:"boss_type"` // BOSS 0 小BOSS， 1 大BOSS
	BuyType  int64 `codec:"buy_type"`  // 0免费 1 购买
}

// rspMsgBuyGuildBossAbsentReward 购买公会BOSS未参与的奖励回复消息定义
type rspMsgBuyGuildBossAbsentReward struct {
	SyncRespWithRewards
}

// BuyGuildBossAbsentReward 购买公会BOSS未参与的奖励: 购买公会BOSS未参与的奖励
func (p *Account) BuyGuildBossAbsentReward(r servers.Request) *servers.Response {
	req := new(reqMsgBuyGuildBossAbsentReward)
	rsp := new(rspMsgBuyGuildBossAbsentReward)

	initReqRsp(
		"Attr/BuyGuildBossAbsentRewardRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BuyGuildBossAbsentRewardHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
