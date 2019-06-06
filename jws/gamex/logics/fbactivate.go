package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FBactivate : facebook激活
// facebook激活

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFBactivate facebook激活请求消息定义
type reqMsgFBactivate struct {
	Req
}

// rspMsgFBactivate facebook激活回复消息定义
type rspMsgFBactivate struct {
	SyncRespWithRewards
}

// FBactivate facebook激活: facebook激活
func (p *Account) FBactivate(r servers.Request) *servers.Response {
	req := new(reqMsgFBactivate)
	rsp := new(rspMsgFBactivate)

	initReqRsp(
		"Attr/FBactivateRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_               = iota
		CODE_Has_active // 失败:没奖可领
	)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_FaceBookFocus) {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	if p.Profile.IsFaceBook {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	p.Profile.IsFaceBook = true
	data := &gamedata.CostData{}
	data.AddItem(gamedata.GetCommonCfg().GetFollowUpReward(), gamedata.GetCommonCfg().GetFollowUpRewardCount())

	if !account.GiveBySync(p.Account, data, rsp, "FaceBook") {
		logs.Error("FaceBook GiveBySync Err")
	}
	rsp.OnChangeFaceBook()

	return rpcSuccess(rsp)
}
