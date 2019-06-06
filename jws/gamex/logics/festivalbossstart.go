package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
)

// FestivalBossStart : 节日Boss挑战协议
// 节日Boss挑战，返回挑战倍率

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgFestivalBossStart 节日Boss挑战协议请求消息定义
type reqMsgFestivalBossStart struct {
	Req
	FestivalId int64 `codec:"fes_id"`  //节日类型ID
	StageId    int64 `codec:"s_id"`    //关卡类型ID
	BossLvl    int64 `codec:"bosslvl"` // 挑战倍率0无倍率,1二倍,2五倍
}

// rspMsgFestivalBossStart 节日Boss挑战协议回复消息定义
type rspMsgFestivalBossStart struct {
	SyncResp
}

// FestivalBossStart 节日Boss挑战协议: 节日Boss挑战，返回挑战倍率
func (p *Account) FestivalBossStart(r servers.Request) *servers.Response {
	req := new(reqMsgFestivalBossStart)
	rsp := new(rspMsgFestivalBossStart)

	initReqRsp(
		"Attr/FestivalBossStartRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_SC_Err
		CODE_Cost_HC_Err
		CODE_Err_CostErr
	)

	now_t := p.Profile.GetProfileNowTime()
	onlyShopTime := gamedata.GetFestivallBossCfg(uint32(req.FestivalId)).GetOnlyShop()
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(int(req.FestivalId), p.Profile.ChannelQuickId)
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < (v.EndTime-int64(onlyShopTime)*int64(util.MinSec)) {
			actInfo = v
		}
	}

	if actInfo == nil {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	costTyp, costNum := gamedata.GetFestivalBossCostChallengeCost(req.FestivalId)
	if !p.Profile.GetSC().HasSC(helper.SCId(costTyp),
		int64(costNum)) {
		return rpcErrorWithMsg(rsp, CODE_Cost_SC_Err, "CODE_Cost_Er")
	}
	if req.BossLvl != 0 {
		_, _, costNum1 := gamedata.GetFestivalBossReward(uint32(req.FestivalId), req.BossLvl)
		if !p.Profile.GetHC().HasHC(int64(costNum1)) {
			return rpcErrorWithMsg(rsp, CODE_Cost_HC_Err, "CODE_Cost_Er")
		}
	}
	// 次数检查
	if p.Profile.GetCounts().Counts[counter.CounterTypeFestivalBoss] == 0 {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	rsp.OnChangeGameMode(counter.CounterTypeFestivalBoss)
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
