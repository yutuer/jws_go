package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// Expeditionrest : 远征重置协议
// 用来传输远征重置信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExpeditionrest 远征重置协议请求消息定义
type reqMsgExpeditionrest struct {
	Req
}

// rspMsgExpeditionrest 远征重置协议回复消息定义
type rspMsgExpeditionrest struct {
	SyncResp
	ExpeditionIds   []string `codec:"expeditionids"` // 被远征的玩家ID
	ExpeditionNames []string `codec:"ednames"`       // 被远征玩家姓名
	ExpeditionState int64    `codec:"edstate"`       // 当前最远关卡
	ExpeditionAvard int64    `codec:"edavard"`       // 当前最远宝箱
	ExpeditionNum   int64    `codec:"ednum"`         // 远征通关总计次数
	ExpeditionStep  bool     `codec:"eds"`           // 远征是否通过9关
}

// Expeditionrest 远征重置协议: 用来传输远征重置信息
func (p *Account) Expeditionrest(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionrest)
	rsp := new(rspMsgExpeditionrest)

	initReqRsp(
		"Attr/ExpeditionrestRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Times_Not_Enough
	)

	// 次数检查
	if !p.Profile.GetCounts().Use(counter.CounterTypeExpedition, p.Account) {
		logs.Warn("Expeditionrest Err_Times_Not_Enough")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	//初始化数据
	p._initExpeditionInfo()
	if !p.setExpeditionEnmy() {
		// 通过让玩家多次点击来实现
		logs.Warn("there is no Imp for ExpeditionInfo")
		p.Profile.GetExpeditionInfo().LoadEnemyToday(p.AccountID.String(),
			int64(p.Profile.GetData().CorpCurrGS_HistoryMax), p.GetProfileNowTime())
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	rsp.ExpeditionIds = p.Profile.GetExpeditionInfo().ExpeditionIds
	rsp.ExpeditionNames = p.Profile.GetExpeditionInfo().ExpeditionNames
	rsp.ExpeditionState = int64(p.Profile.GetExpeditionInfo().ExpeditionState)
	rsp.ExpeditionAvard = int64(p.Profile.GetExpeditionInfo().ExpeditionAward)
	rsp.ExpeditionNum = int64(p.Profile.GetExpeditionInfo().ExpeditionNum)
	rsp.ExpeditionStep = p.Profile.GetExpeditionInfo().ExpeditionStep
	logiclog.LogExpeditionRest(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		int(p.Profile.GetData().CorpCurrGS),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	rsp.OnChangeGameMode(counter.CounterTypeExpedition)
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
