package logics

import (
	"strconv"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// AwardMarketActivity : 运营活动领奖
// 运营活动领奖

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgAwardMarketActivity 运营活动领奖请求消息定义
type reqMsgAwardMarketActivity struct {
	Req
	ActivityId    int64 `codec:"aid"`   // 本批次活动id
	ActivitySubId int64 `codec:"subid"` // 本批次活动子id
}

// rspMsgAwardMarketActivity 运营活动领奖回复消息定义
type rspMsgAwardMarketActivity struct {
	SyncRespWithRewards
	IsOver int64 `codec:"isover"` // 是否已经结束
}

// AwardMarketActivity 运营活动领奖: 运营活动领奖
func (p *Account) AwardMarketActivity(r servers.Request) *servers.Response {
	req := new(reqMsgAwardMarketActivity)
	rsp := new(rspMsgAwardMarketActivity)

	initReqRsp(
		"PlayerAttr/AwardMarketActivityRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
		Err_Got
		Err_Give
	)

	now_t := p.Profile.GetProfileNowTime()
	ma := p.Profile.GetMarketActivitys()
	ma.UpdateMarketActivity(p.AccountID.String(), now_t)
	act := ma.GetMarketActivityById(uint32(req.ActivityId))
	if act == nil {
		rsp.IsOver = 1
		return rpcSuccess(rsp)
	}
	actData := gamedata.GetHotDatas().Activity
	cfg := actData.GetActivitySimpleInfoById(act.ActivityId)
	if cfg == nil {
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}
	if now_t >= cfg.EndTime || now_t < cfg.StartTime {
		rsp.IsOver = 1
		return rpcSuccess(rsp)
	}

	subId := int(req.ActivitySubId - 1)
	if subId >= len(act.State) || act.State[subId] == market_activity.MA_ST_GOT {
		logs.Warn("AwardMarketActivity Err_Got")
		return rpcSuccess(rsp)
	}
	act.State[subId] = market_activity.MA_ST_GOT

	activityCfg := gamedata.GetHotDatas().Activity
	simpleCfg := activityCfg.GetActivitySimpleInfoById(uint32(req.ActivityId))

	if int(simpleCfg.ActivityType) == gamedata.ActOnlyPay {
		act.TmpValue[req.ActivitySubId*2-1] += 1
		ipaCount := act.TmpValue[req.ActivitySubId*2-2]    //充值次数
		rewardCount := act.TmpValue[req.ActivitySubId*2-1] //领奖次数
		subCfg := activityCfg.GetMarketActivitySubConfig(uint32(req.ActivityId))
		maxIapCount, _ := strconv.ParseFloat(subCfg[uint32(req.ActivitySubId)].GetSFCValue1(), 32)

		if ipaCount > rewardCount && rewardCount < int64(maxIapCount) {
			act.State[req.ActivitySubId-1] = market_activity.MA_ST_ACT
		} else if rewardCount >= int64(maxIapCount) {
			act.State[req.ActivitySubId-1] = market_activity.MA_ST_GOT
		} else {
			act.State[req.ActivitySubId-1] = market_activity.MA_ST_INIT
		}
	}

	// 发奖
	subCfgs := actData.GetMarketActivitySubConfig(act.ActivityId)
	subCfg, ok := subCfgs[uint32(req.ActivitySubId)]
	if !ok {
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}

	data := &gamedata.CostData{}
	for _, v := range subCfg.GetItem_Table() {
		data.AddItem(v.GetItemID(), v.GetItemCount())
	}
	itemCounts := make(map[string]uint32, 0)
	for i, item := range data.Item2Client {
		itemCounts[item] = data.Count2Client[i]
	}
	logiclog.LogHgrHotActivity(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		uint32(p.AccountID.ShardId),
		uint32(req.ActivityId),
		uint32(req.ActivitySubId),
		itemCounts,
		simpleCfg.ActivityType,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	if !account.GiveBySync(p.Account, data, rsp, "AwardMarketActivity") {
		return rpcErrorWithMsg(rsp, Err_Give, "Err_Give")
	}

	rsp.OnChangeMarketActivity()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
