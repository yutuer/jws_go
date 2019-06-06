package logics

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) GetAllDailyAwardsInfo(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/GetAllDailyAwardsInfoResp",
		r.RawBytes,
		req, resp, p)

	p.Profile.GetDailyAwards().UpdateDailyAwards(p.Profile.GetProfileNowTime())

	resp.OnChangeDailyAward()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) AwardDailyAward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Id    uint32 `codec:"id"`
		SubId uint32 `codec:"subid"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/AwardDailyAwardResp",
		r.RawBytes,
		req, resp, p)

	p.Profile.GetDailyAwards().UpdateDailyAwards(p.Profile.GetProfileNowTime())

	errCode, errMsg, warnCode := p.Profile.GetDailyAwards().AwardDailyAward(p.Account,
		req.Id, req.SubId, resp)
	if warnCode > 0 {
		logs.Warn("AwardDailyAward %d", errCode)
		return rpcWarn(resp, uint32(warnCode))
	}
	if errCode > 0 {
		return rpcErrorWithMsg(resp, uint32(errCode)+20, fmt.Sprintf("%s %d %d", errMsg, req.Id, req.SubId))
	}

	resp.OnChangeDailyAward()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
