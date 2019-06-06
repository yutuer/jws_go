package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// OpenSimplePvpDayChest : 每日1v1竞技场获取宝箱奖励协议
// 每日挑战1v1竞技场一定次数，可以领取宝箱奖励

// reqMsgOpenSimplePvpDayChest 每日1v1竞技场获取宝箱奖励协议请求消息定义
type reqMsgOpenSimplePvpDayChest struct {
	Req
	ChestID int64 `codec:"_p1_"` // 想要领取的宝箱的ID
}

// rspMsgOpenSimplePvpDayChest 每日1v1竞技场获取宝箱奖励协议回复消息定义
type rspMsgOpenSimplePvpDayChest struct {
	SyncRespWithRewards
}

// OpenSimplePvpDayChest 每日1v1竞技场获取宝箱奖励协议: 每日挑战1v1竞技场一定次数，可以领取宝箱奖励
func (p *Account) OpenSimplePvpDayChest(r servers.Request) *servers.Response {
	req := new(reqMsgOpenSimplePvpDayChest)
	rsp := new(rspMsgOpenSimplePvpDayChest)

	initReqRsp(
		"Attr/OpenSimplePvpDayChestRsp",
		r.RawBytes,
		req, rsp, p)
	const (
		_ = iota
		Err_AlreadyOpen
		Err_CanNotOpen
		Err_Give_Fail
	)
	// logic imp begin
	simplePvpInfo := p.Profile.GetSimplePvp()
	simplePvpInfo.UpdateChestInfo(p.Profile.GetProfileNowTime())
	logs.Debug("req id: %d, fight count: %d", req.ChestID, simplePvpInfo.PvpCountToday)
	if !simplePvpInfo.CanOpenChest(uint32(req.ChestID)) {
		logs.Warn("OpenSimplePvpDayChest Err_AlreadyOpen")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if rewards, ok := gamedata.GetPvpDayReward(int(req.ChestID)); ok && int(req.ChestID) <= simplePvpInfo.PvpCountToday {
		simplePvpInfo.SetChestOpen(uint32(req.ChestID))
		// 给予奖励
		give := account.GiveGroup{}
		give.AddCostData(&rewards.Cost)
		if !give.GiveBySyncAuto(p.Account, rsp, "1v1PvpChestAward") {
			return rpcErrorWithMsg(rsp, Err_Give_Fail, "1v1PvpChestAward give award fail")
		}
	} else {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	// logic imp end
	rsp.OnChangeSimplePvp()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
