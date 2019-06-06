package logics

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/secure"
)

type LogRequestReturnTown struct {
	Req
	Param    string `codec:"param"`
	ParamCrc string `codec:"crc"`
}

type LogResponseReturnTown struct {
	Resp
}

func (a *Account) LogReturnTownRequest(r servers.Request) *servers.Response {
	req := &LogRequestReturnTown{}
	resp := &LogResponseReturnTown{}

	initReqRsp(
		"Log/LogReturnTownRsp",
		r.RawBytes,
		req, resp, a)

	_param, _ := secure.DefaultEncode.Decode64FromNet(req.ParamCrc)
	param := string(_param)
	if req.Param != param {
		return rpcErrorWithMsg(resp, 1, "ReturnTownRequest Err, may be attacked !!!")
	}
	a.Profile.GetHero().HeroStarActivity(a.Account)
	a.Profile.AutomationQuest(a.Account)

	// 记log
	logiclog.LogReturnTown_c(a.AccountID.String(), a.Profile.GetCurrAvatar(),
		a.Profile.GetCorp().GetLvlInfo(), a.Profile.ChannelId, "")
	return rpcSuccess(resp)
}

type LogRequestEnterBoss struct {
	Req
	BossIdx  int    `codec:"bid"`
	Param    string `codec:"param"`
	ParamCrc string `codec:"crc"`
}

type LogResponseEnterBoss struct {
	Resp
}

func (p *Account) LogEnterBossRequest(r servers.Request) *servers.Response {
	req := &LogRequestEnterBoss{}
	resp := &LogResponseEnterBoss{}

	initReqRsp(
		"Log/LogEnterBossRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		CODE_Err_IDX
		CODE_Err_Crc
	)

	_param, _ := secure.DefaultEncode.Decode64FromNet(req.ParamCrc)
	param := string(_param)
	if req.Param != param {
		return rpcErrorWithMsg(resp, CODE_Err_Crc, "EnterBossRequest Err, may be attacked !!!")
	}
	player_boss := p.Profile.GetBoss()
	boss_to_fight := player_boss.GetBoss(req.BossIdx)

	if boss_to_fight == nil {
		return rpcError(resp, CODE_Err_IDX)
	}

	now_time := time.Now().Unix()
	p.Tmp.SetLevelEnterTime(now_time)
	// 记log
	logiclog.LogBoss_c(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		boss_to_fight.BossTyp, true, now_time, "")
	return rpcSuccess(resp)
}

type RequestLogClientTimeEvent struct {
	Req
	Param          string `codec:"param"`
	ParamCrc       string `codec:"crc"`
	AvatarId       int    `codec:"avatarid"`
	Data           string `codec:"data"`
	Time           int64  `codec:"time"`
	Avatars        []int  `codec:"avs"`
	ChgAvatarCount uint32 `codec:"chgavc"`
	DeadAvatars    []int  `codec:"dedavs"`
}

type ResponseLogClientTimeEvent struct {
	Resp
}

func (p *Account) LogClientTimeEventRequest(r servers.Request) *servers.Response {
	req := &RequestLogClientTimeEvent{}
	resp := &ResponseLogClientTimeEvent{}

	initReqRsp(
		"Log/LogClientTimeEventRsp",
		r.RawBytes,
		req, resp, p)

	_param, _ := secure.DefaultEncode.Decode64FromNet(req.ParamCrc)
	param := string(_param)
	if req.Param != param {
		return rpcErrorWithMsg(resp, 1, "LogClientDataRequest Err, may be attacked !!!")
	}
	p.Profile.LastClientTimeEvent = req.Data
	// 记log
	logiclog.LogClientData_c(p.AccountID.String(), req.AvatarId,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, req.Data, req.Time,
		req.Avatars, req.ChgAvatarCount, req.DeadAvatars,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		func(last string) string { return p.Profile.GetLastSetCurClientEvent(last) }, "")
	return rpcSuccess(resp)
}

// FenghuoClientLog : 烽火埋点事件
// 处理烽火燎原的所有埋点

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFenghuoClientLog 烽火埋点事件请求消息定义
type reqMsgFenghuoClientLog struct {
	Req
	Param           string `codec:"param"`           // 校验原始字符串
	Crc             string `codec:"crc"`             // 加密字符串
	AvatarID        int64  `codec:"avatarid"`        // 当前所用avatarID
	EventName       string `codec:"eventname"`       // 事件id
	SelfUid         string `codec:"selfuid"`         // 自己的uid
	SelfGs          int64  `codec:"selfgs"`          // 自己的战力
	OtherUid        string `codec:"otheruid"`        // 另一个玩家的uid
	OtherGs         int64  `codec:"othergs"`         // 另一个玩家的战力
	Difficult       string `codec:"difficult"`       // 难度
	RoomWaitTime    int64  `codec:"roomwaittime"`    // 解散房间时间 - 点击创建房间时间（总等待时间）
	Rate            string `codec:"rate"`            // 几倍
	IsWin           int64  `codec:"iswin"`           // 是否胜利  0是输，1是胜
	BattleTotalTime int64  `codec:"battletotaltime"` // 战斗总消耗时间
}

// rspMsgFenghuoClientLog 烽火埋点事件回复消息定义
type rspMsgFenghuoClientLog struct {
	SyncResp
}

// FenghuoClientLog 烽火埋点事件: 处理烽火燎原的所有埋点
func (p *Account) FenghuoClientLog(r servers.Request) *servers.Response {
	req := new(reqMsgFenghuoClientLog)
	rsp := new(rspMsgFenghuoClientLog)

	initReqRsp(
		"Attr/FenghuoClientLogRsp",
		r.RawBytes,
		req, rsp, p)

	_param, _ := secure.DefaultEncode.Decode64FromNet(req.Crc)
	param := string(_param)
	if req.Param != param {
		return rpcErrorWithMsg(rsp, 1, "LogClientDataRequest Err, may be attacked !!!")
	}

	logiclog.LogClientFenghuoData_c(p.AccountID.String(), req.AvatarID,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		req.EventName,
		req.SelfUid,
		req.SelfGs,
		req.OtherUid,
		req.OtherGs,
		req.Difficult,
		req.RoomWaitTime,
		req.Rate,
		req.IsWin,
		req.BattleTotalTime,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		func(last string) string { return p.Profile.GetLastSetCurClientEvent(last) }, "")
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
