package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) getActGiftByCondReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ID  int `codec:"id"`
		Typ int `codec:"typ"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/GetActivityGiftByCondRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                = iota
		CODE_No_Gift_Err // 失败:没奖可领
		CODE_No_Cond_Err // 失败:没可领条件
		CODE_Give_Err    // 失败:发奖错误
	)
	act := p.Profile.GetActGiftByCond()
	actData := act.GetData(req.ID, req.Typ)
	nowT := p.Profile.GetProfileNowTime()
	data := gamedata.GetActivityGiftByCondData(uint32(req.ID), uint32(req.Typ), nowT)

	//logs.Trace("data %v %v", actData, data)

	if actData == nil || data == nil {
		return rpcError(resp, CODE_No_Gift_Err)
	}

	progress, all := account.GetConditionProgress(
		&actData.Cond,
		p.Account, data.Cond.Ctyp,
		data.Cond.Param1, data.Cond.Param2,
		data.Cond.Param3, data.Cond.Param4)

	logs.Trace("getActGiftByCondReward %d-%d %d-%d, %v",
		req.ID, req.Typ, progress, all, actData)

	if progress < all {
		return rpcError(resp, CODE_No_Cond_Err)
	}

	if actData.IsHasGet == 1 {
		logs.Warn("%s getActGiftByCondReward repeat", p.AccountID.String())
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	actData.SetHasGet()

	pRander := p.GetRand()
	givesData := &gamedata.CostData{}
	srcData := data.Reward.GetData()
	logs.Debug("srcData: %v", srcData.Item2Client)
	for i := 0; i < len(srcData.Item2Client); i++ {
		id := srcData.Item2Client[i]
		c := srcData.Count2Client[i]
		if !gamedata.IsFixedIDItemID(id) {
			for j := 0; j < int(c); j++ {
				data := gamedata.MakeItemData(p.AccountID.String(), pRander, id)
				givesData.AddItemWithData(id, *data, 1)
			}
		} else {
			givesData.AddItem(id, c)
		}
	}
	logs.Debug("givesData: %v", givesData.Items)
	isOk := account.GiveBySync(
		p.Account,
		givesData,
		resp,
		fmt.Sprintf("ActGiftCond-%d", req.ID))
	if !isOk {
		return rpcError(resp, CODE_Give_Err)
	}

	resp.OnChangeActivityByCond()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) getActGiftByTimeReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ID int `codec:"id"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/GetActivityGiftByTimeRsp",
		r.RawBytes,
		req, resp, p)

	act := p.Profile.GetActGiftByTime()
	actData := act.GetData(req.ID)
	//nowT := p.Profile.GetProfileNowTime()
	data := gamedata.GetActivityGiftByTime()

	if actData == nil || data == nil {
		return rpcWarn(resp, errCode.GiftByTimeErrNoGift)
	}

	if actData.IsHasGet == 1 {
		return rpcWarn(resp, errCode.GiftByTimeErrNoGift)
	}

	progress, all := account.GetConditionProgress(
		&actData.Cond,
		p.Account, data[req.ID].Cond.Ctyp,
		data[req.ID].Cond.Param1,
		data[req.ID].Cond.Param2,
		data[req.ID].Cond.Param3,
		data[req.ID].Cond.Param4)

	logs.Trace("GetActivityGiftByTime %d %d-%d, %v",
		req.ID, progress, all, actData)

	if progress < all-30 { // 略微宽松的条件判断,防止时间不同步
		return rpcWarn(resp, errCode.GiftByTimeErrNoCond)
	}

	actData.SetHasGet()
	isOk := account.GiveBySync(
		p.Account,
		data[req.ID].Reward.GetData(),
		resp,
		fmt.Sprintf("ActGiftTime-%d", req.ID))
	if !isOk {
		return rpcWarn(resp, errCode.GiftByTimeErrGive)
	}

	resp.OnChangeActivityByTime()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
