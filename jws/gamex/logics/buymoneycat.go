package logics

import (
	"fmt"
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/moneycat_marquee"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// BuyMoneyCat : 招财进宝购买协议
// 招财进宝购买

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBuyMoneyCat 招财进宝购买协议请求消息定义
type reqMsgBuyMoneyCat struct {
	Req
}

// rspMsgBuyMoneyCat 招财进宝购买协议回复消息定义
type rspMsgBuyMoneyCat struct {
	SyncRespWithRewards
	ResultHCNum int64 `codec:"resulthcnum"` // 返回值为招财进宝结果,获得的钻石数,范围应是0-99999
	BoughtTimes int64 `codec:"boughttimes"` // 已参与活动次数
}

// BuyMoneyCat 招财进宝购买协议: 招财进宝购买
func (p *Account) BuyMoneyCat(r servers.Request) *servers.Response {
	req := new(reqMsgBuyMoneyCat)
	rsp := new(rspMsgBuyMoneyCat)

	initReqRsp(
		"Attr/BuyMoneyCatRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_Err
		CODE_Active_Closed
	)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_Money_Cat) {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	// 活动时间检查
	pg := p.Profile.GetMoneyCatInfo()
	now_t := p.Profile.GetProfileNowTime()
	ga := gamedata.GetHotDatas().Activity
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActMoneyCat, p.Profile.ChannelQuickId)
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	pg.UpdateMoneyCatActId(actInfo.ActivityId)

	MoneyCatCost := gamedata.GetHotDatas().Activity.GetMoneyCatCost(pg.GetMoneyCatTime())
	data1 := &gamedata.CostData{}
	data1.AddItem(gamedata.VI_Hc, MoneyCatCost)
	if !account.CostBySync(p.Account, data1, rsp, Pay_Result_MonyerCat) {
		return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
	}

	randWeight := RandInt64(1, int64(ga.GetMoneyCatWeight(pg.GetMoneyCatTime())))
	section := RandSection(pg.GetMoneyCatTime(), randWeight)

	MinNum, MaxNum := ga.GetMoneyCatNum(pg.GetMoneyCatTime(), section)

	rsp.ResultHCNum = RandInt64(MinNum, MaxNum)

	data := &gamedata.CostData{}
	data.AddItem(gamedata.VI_Hc, uint32(rsp.ResultHCNum))

	if !account.GiveBySync(p.Account, data, rsp, "MneyCatGive") {
		logs.Error("Moneycat GiveBySync Err")
	}
	if gamedata.GetHotDatas().Activity.GetMoneyCatMarquee(pg.MoneyCatTime) == 1 {
		if moneycat_marquee.GetModule(p.AccountID.ShardId).TryAddMoneyCatInfo(gamedata.ActMoneyCat,
			p.Profile.ChannelQuickId, p.Profile.GetProfileNowTime(), p.Profile.Name, rsp.ResultHCNum) {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_ROLLINFO_PLAYER_MONEYGOD).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", rsp.ResultHCNum)).Send()
		}
	}
	//更新招财猫招财次数
	p.Profile.GetMoneyCatInfo().UpdateMoneyCatTime()
	rsp.BoughtTimes = p.Profile.GetMoneyCatInfo().GetMoneyCatTime()
	rsp.onChangeMoneyCat()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}

func RandSection(step int64, randweight int64) int {
	subData := gamedata.GetHotDatas().Activity.GetMoneyCatSubData(step)
	var section int = 0
	weight := subData[0].GetWeight()
	for i := 0; i < len(subData)-1; i++ {
		if randweight < int64(weight) {
			return section
		}
		weight += subData[i+1].GetWeight()
		section += 1
	}
	return section
}
