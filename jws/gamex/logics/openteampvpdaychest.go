package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// OpenTeamPvpDayChest : 每日3v3竞技场获取宝箱奖励协议
// 每日挑战3v3竞技场一定次数，可以领取宝箱奖励

// reqMsgOpenTeamPvpDayChest 每日3v3竞技场获取宝箱奖励协议请求消息定义
type reqMsgOpenTeamPvpDayChest struct {
	Req
	ChestID int64 `codec:"_p1_"` // 想要领取的宝箱的ID
}

// rspMsgOpenTeamPvpDayChest 每日3v3竞技场获取宝箱奖励协议回复消息定义
type rspMsgOpenTeamPvpDayChest struct {
	SyncRespWithRewards
}

// OpenTeamPvpDayChest 每日3v3竞技场获取宝箱奖励协议: 每日挑战3v3竞技场一定次数，可以领取宝箱奖励
func (p *Account) OpenTeamPvpDayChest(r servers.Request) *servers.Response {
	req := new(reqMsgOpenTeamPvpDayChest)
	rsp := new(rspMsgOpenTeamPvpDayChest)

	initReqRsp(
		"Attr/OpenTeamPvpDayChestRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	const (
		_ = iota
		Err_AlreadyOpen
		Err_CanNotOpen
		Err_Give_Fail
	)
	// logic imp begin
	teamPvpInfo := p.Profile.GetTeamPvp()
	teamPvpInfo.UpdateChestInfo(p.Profile.GetProfileNowTime())
	logs.Debug("req id: %d, fight count: %d", req.ChestID, teamPvpInfo.PvpCountToday)
	if !teamPvpInfo.CanOpenChest(uint32(req.ChestID)) {
		return rpcError(rsp, Err_AlreadyOpen)
	}
	if rewards, ok := gamedata.GetTPvpDayReward(int(req.ChestID)); ok && int(req.ChestID) <= teamPvpInfo.PvpCountToday {
		teamPvpInfo.SetChestOpen(uint32(req.ChestID))
		// 给予奖励
		give := account.GiveGroup{}
		give.AddCostData(&rewards.Cost)
		if !give.GiveBySyncAuto(p.Account, rsp, "3v3PvpChestAward") {
			return rpcErrorWithMsg(rsp, Err_Give_Fail, "3v3PvpChestAward give award fail")
		}
	} else {
		logs.Warn("OpenTeamPvpDayChest Err_CanNotOpen")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	rsp.OnChangeTeamPvp()
	// logic imp end
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
