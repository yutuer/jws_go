package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ExpeditionAward : 远征领奖协议
// 用来远征打开宝箱的协议

// reqMsgExpeditionAward 远征领奖协议请求消息定义
type reqMsgExpeditionAward struct {
	Req
	ExpeditionAward int64 `codec:"edad"`  // 第几个宝箱
	ExpeditionType  int64 `codec:"atype"` //宝箱品质 0：普通 1：巨额 2：超额
}

// rspMsgExpeditionAward 远征领奖协议回复消息定义
type rspMsgExpeditionAward struct {
	SyncRespWithRewards
	ExpeditionState int64 `codec:"es"`
}

//一共9个宝箱,传过来的宝箱范围应在1~9之间
const maxAwardNum = 9
const minAwardNum = 1

// ExpeditionAward 远征领奖协议: 用来远征打开宝箱的协议
func (p *Account) ExpeditionAward(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionAward)
	rsp := new(rspMsgExpeditionAward)

	initReqRsp(
		"Attr/ExpeditionAwardRsp",
		r.RawBytes,
		req, rsp, p)
	const (
		_ = iota
		CODE_Cost_Err
	)

	ps := p.Profile.GetExpeditionInfo()
	if req.ExpeditionAward == int64(ps.ExpeditionState) {

		if req.ExpeditionAward < minAwardNum || req.ExpeditionAward > maxAwardNum {
			logs.Error("ExpeditionAward param err %d", req.ExpeditionAward)
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}

		//领取第九个宝箱后将是否通关变成true,将通关关次+1
		if req.ExpeditionAward == 9 {
			ps.ExpeditionStep = true
			ps.ExpeditionNum = ps.ExpeditionNum + 1
		}
		awardcost := gamedata.GetExpeditionAwardCost(int(req.ExpeditionType))
		data1 := &gamedata.CostData{}
		data1.AddItem(helper.VI_Hc, awardcost)
		if !account.CostBySync(p.Account, data1, rsp, "Expedition Award Cost") {
			return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
		}

		Award1, Award2, Award3 := gamedata.GetAwardByStep(req.ExpeditionAward - 1)
		ItemNum1, ItemNum2, ItemNum3 := gamedata.GetAwardNumByStep(req.ExpeditionAward - 1)
		addition := gamedata.GetExpeditionAwardGive(int(req.ExpeditionType))
		passinfo := gamedata.GetExpeditionPassAwardCfgs(ps.ExpeditionLvl)

		data := &gamedata.CostData{}
		data.AddItem(Award1, uint32(float32(ItemNum1)*addition))
		data.AddItem(Award2, uint32(float32(ItemNum2)*addition))
		data.AddItem(Award3, uint32(float32(ItemNum3)*addition))
		data.AddItem(helper.VI_Hc, passinfo.GetGetNum())

		if !account.GiveBySync(p.Account, data, rsp, "ExpeditionAward") {
			return rpcErrorWithMsg(rsp, Err_Give,
				fmt.Sprintf("Err_Give count %v item_id %v", Award1, ItemNum1))
		}

		p.Profile.GetExpeditionInfo().ExpeditionState += 1
		passinfoNext := gamedata.GetExpeditionPassAwardCfgs(ps.ExpeditionLvl + 1)
		gpa := gamedata.GetExpeditionPassAward()
		if ps.ExpeditionNum >= int32(passinfoNext.GetPassTime()) {
			ps.ExpeditionLvl = ps.ExpeditionLvl + 1
			if ps.ExpeditionLvl > gpa[len(gpa)-1].GetPassLevel() {
				ps.ExpeditionLvl = gpa[len(gpa)-1].GetPassLevel()
			}
		}
		rsp.ExpeditionState = int64(p.Profile.GetExpeditionInfo().ExpeditionState)
	}
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
