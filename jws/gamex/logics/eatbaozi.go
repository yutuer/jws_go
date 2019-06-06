package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// EatBaozi : 吃包子结算协议
// 吃包子结算，发送奖励，当isHc为true时为扫荡

// reqMsgEatBaozi 吃包子结算协议请求消息定义
type reqMsgEatBaozi struct {
	Req
	IsUseHc int64 `codec:"_p1_"` // 是否是用HC扫荡（0-否 1-是）
	Score   int64 `codec:"_p2_"` // 吃包子的积分，扫荡时无效
}

// rspMsgEatBaozi 吃包子结算协议回复消息定义
type rspMsgEatBaozi struct {
	SyncRespWithRewards
	Count int64 `codec:"count"` // 本次最终获得的包子数量
}

// EatBaozi 吃包子结算协议: 吃包子结算，发送奖励，当isHc为true时为扫荡
func (p *Account) EatBaozi(r servers.Request) *servers.Response {
	req := new(reqMsgEatBaozi)
	rsp := new(rspMsgEatBaozi)

	initReqRsp(
		"Attr/EatBaoziRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin

	const (
		_ = iota
		Err_Give
		Err_Cost
	)
	const (
		UseHC    = int64(1)
		NotUseHC = int64(0)
	)
	const Code_Min = 30

	data := &gamedata.CostData{}

	itemID, minCount, maxCount, _ := gamedata.GetEatBaoziData()
	getScore := uint32(0)
	ts := 1
	if req.IsUseHc == NotUseHC {
		getScore = uint32(req.Score)
		if getScore < minCount {
			getScore = minCount
		} else if getScore > maxCount {
			getScore = maxCount
		}
		success, errCode, warnCode, _ := p.Profile.GetGameMode().GameModeCheckAndNoCost(p.Account, counter.CounterTypeEatBaozi, 1, rsp)
		if warnCode != 0 || !success {
			logs.Warn("EatBaozi GameModeCheckAndNoCost %d", errCode)
			return rpcWarn(rsp, warnCode)
		}
		if errCode != 0 || !success {
			return rpcError(rsp, errCode+Code_Min)
		}

		// update quest
		p.updateCondition(account.COND_TYP_EatBaozi,
			1, 0, "", "", rsp)

		// 更新玩家单次最大吃包数目
		p.Profile.GetEatBaozi().UpdateMaxEatBaoziCount(getScore)

		// 称号
		p.Profile.GetTitle().OnEatBaozi(p.Account)
	} else if req.IsUseHc == UseHC {
		ok, errcode, warnCode, times := p.Profile.GetGameMode().GameModeLevelSweep(p.Account, counter.CounterTypeEatBaozi, rsp)
		if !ok {
			if warnCode > 0 {
				return rpcWarn(rsp, errCode.ClickTooQuickly)
			}
			return rpcError(rsp, errcode+Code_Min)
		}
		ts = times
		getScore = uint32(times) * maxCount

		// update quest
		p.updateCondition(account.COND_TYP_EatBaozi,
			times, 0, "", "", rsp)
	}
	data.AddItem(itemID, getScore)

	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, rsp, "EatBaozi") {
		return rpcErrorWithMsg(rsp, Err_Give,
			fmt.Sprintf("Err_Give count %v item_id %v", getScore, itemID))
	}

	rsp.Count = int64(getScore)
	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeEatBaozi,
		ts,
		p.Profile.GetProfileNowTime())

	// logic imp end
	rsp.OnChangeGameMode(counter.CounterTypeEatBaozi)
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
