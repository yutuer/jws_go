package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FBInvitation : FB好友邀请
// FB好友邀请

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFBInvitation FB好友邀请请求消息定义
type reqMsgFBInvitation struct {
	Req
}

// rspMsgFBInvitation FB好友邀请回复消息定义
type rspMsgFBInvitation struct {
	SyncRespWithRewards
}

// FBInvitation FB好友邀请: FB好友邀请
func (p *Account) FBInvitation(r servers.Request) *servers.Response {
	req := new(reqMsgFBInvitation)
	rsp := new(rspMsgFBInvitation)

	initReqRsp(
		"Attr/FBInvitationRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Times_Not_Enough
	)

	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_FaceBookInvite) {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}
	// 次数检查
	if !p.Profile.GetCounts().Use(counter.CounterTypeFBInvitation, p.Account) {
		logs.Warn("FBInvitation Err_Times_Not_Enough")
		return rpcErrorWithMsg(rsp, Err_Times_Not_Enough, "Err_Times_Not_Enough")
	}

	data := &gamedata.CostData{}
	data.AddItem(gamedata.GetCommonCfg().GetFriendInviteReward(), gamedata.GetCommonCfg().GetFriendInviteRewardCount())

	if !account.GiveBySync(p.Account, data, rsp, "FBInvitation") {
		logs.Error("FB Invitation GIveBySync Err")
	}

	rsp.OnChangeGameMode(counter.CounterTypeFBInvitation)
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
