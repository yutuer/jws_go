package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// AddGuildSciencePoint : 加工会科技树点
// 加工会科技树点协议

// reqMsgAddGuildSciencePoint 加工会科技树点请求消息定义
type reqMsgAddGuildSciencePoint struct {
	Req
	ScienceIdx int64 `codec:"sidx"` // 科技点类型id
	Point      int64 `codec:"sp"`   // 加多少科技点
}

// rspMsgAddGuildSciencePoint 加工会科技树点回复消息定义
type rspMsgAddGuildSciencePoint struct {
	SyncResp
}

// AddGuildSciencePoint 加工会科技树点: 加工会科技树点协议
func (p *Account) AddGuildSciencePoint(r servers.Request) *servers.Response {
	req := new(reqMsgAddGuildSciencePoint)
	rsp := new(rspMsgAddGuildSciencePoint)

	initReqRsp(
		"Guild/AddGuildSciencePointRsp",
		r.RawBytes,
		req, rsp, p)

	if req.Point <= 0 {
		return rpcSuccess(rsp)
	}

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	oldV := p.Profile.GetSC().GetSC(gamedata.SC_GuildSp)
	if oldV < req.Point {
		logs.Warn("AddGuildSciencePoint SC_GuildSp not enough")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	errRet := guild.GetModule(
		p.AccountID.ShardId).AddGuildSciencePoint(p.GuildProfile.GuildUUID,
		p.AccountID.String(), req.ScienceIdx, req.Point, oldV-req.Point)
	if rsp := guildErrRet(errRet, rsp); rsp != nil {
		return rsp
	}

	cost := &gamedata.CostData{}
	cost.AddItem(gamedata.VI_GuildSP, uint32(req.Point))
	if !account.CostBySync(p.Account, cost, rsp, "AddGuildSciencePoint") {
		logs.Error("AddGuildSciencePoint CostBySync fail")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	// 条件更新
	p.updateCondition(account.COND_TYP_Add_Guild_Science_Point, 1, 0, "", "", rsp)

	rsp.OnChangeGuildInfo()
	rsp.OnChangeGuildMemsInfo()
	rsp.OnChangeGuildScience()
	rsp.mkInfo(p)

	// log
	logiclog.LogCostCurrency(p.AccountID.String(), p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, fmt.Sprintf("GuildScience-%d", req.ScienceIdx),
		gamedata.VI_GuildSP, oldV, req.Point, p.Profile.GetVipLevel(),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(rsp)
}

// GuildSciencePointLog : 工会科技树点日志
// 工会科技树点日志协议
// reqMsgGuildSciencePointLog 工会科技树点日志请求消息定义
type reqMsgGuildSciencePointLog struct {
	Req
	LogTyp int64 `codec:"logtyp"` // 日志类型,1:每日log, 2:每周log
}

// rspMsgGuildSciencePointLog 工会科技树点日志回复消息定义
type rspMsgGuildSciencePointLog struct {
	SyncResp
	Names    []string `codec:"names"` // 昵称名
	SP       []int64  `codec:"sp"`    // 捐献点
	LastTime []int64  `codec:"lst"`   // 上次捐献时间
}

// GuildSciencePointLog 工会科技树点日志: 工会科技树点日志协议
func (p *Account) GuildSciencePointLog(r servers.Request) *servers.Response {
	req := new(reqMsgGuildSciencePointLog)
	rsp := new(rspMsgGuildSciencePointLog)

	initReqRsp(
		"Guild/GuildSciencePointLogRsp",
		r.RawBytes,
		req, rsp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	bGstWeek := false
	if req.LogTyp == 2 {
		bGstWeek = true
	}

	var errRet guild.GuildRet
	errRet, rsp.Names, rsp.SP, rsp.LastTime = guild.GetModule(
		p.AccountID.ShardId).GetGuildScienceLog(p.GuildProfile.GuildUUID,
		p.AccountID.String(), bGstWeek)
	if rsp := guildErrRet(errRet, rsp); rsp != nil {
		return rsp
	}

	return rpcSuccess(rsp)
}
