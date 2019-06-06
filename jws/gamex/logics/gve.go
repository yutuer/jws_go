package logics

import (
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/city_broadcast"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Gve_BroadCast_IDS_Count = 3
)

func (p *Account) StartMatchGVE(r servers.Request) *servers.Response {
	req := &struct {
		Req
		IsHard     bool `codec:"isHard"`
		IsUseHc    bool `codec:"isHc"`
		NeedCancel bool `codec:"Cancel"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/StartMatchGVERsp",
		r.RawBytes,
		req, resp, p)

	acId := p.AccountID.String()

	isCan := gamedata.IsCanGVE(p.Profile.GetProfileNowTime())
	if !req.NeedCancel && !isCan {
		return rpcWarn(resp, errCode.CurrNoGVE)
	}

	playerCounter := p.Profile.GetCounts()
	if !req.NeedCancel && !playerCounter.UseJustDayBegin(counter.CounterTypeGVE, p.Account) {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	if !req.NeedCancel {
		p.Tmp.CleanGVEData()
	}

	_, doubleC := gamedata.GetGVEGameCfg()

	gveCounts, _ := playerCounter.Get(counter.CounterTypeGVE, p.Account)
	gveCount := playerCounter.GetDailyMax(counter.CounterTypeGVE)
	p.Tmp.CurrCount = (gveCount - gveCounts)
	p.Tmp.GameIsUseHc = req.IsUseHc
	if (gveCount - gveCounts) < doubleC {
		p.Tmp.GameIsDouble = true
		p.Tmp.GameIsUseHc = false
	}
	//p.Tmp.GameIsDouble = (gveCount - gveCounts) < doubleC
	p.Tmp.GameIsHard = req.IsHard

	if !req.NeedCancel {
		p.Tmp.MatchBeginTime = time.Now().Unix()
	}

	err := p.SendMatchReq(req.IsHard, req.NeedCancel)
	if err != nil {
		logs.SentryLogicCritical(acId,
			"StartMatchGVE Err By %s", err.Error())
		return rpcWarn(resp, 99)
	}

	if req.NeedCancel {
		p.Tmp.CurrWaitting = false
	} else {
		p.Tmp.CurrWaitting = true
	}

	resp.mkInfo(p)

	// broadcast match
	if !req.NeedCancel && gve_notify.GetModule(p.AccountID.ShardId).TryBroadCastMatchMsg() {
		city_broadcast.Pool.UseRes2Send(
			city_broadcast.CBC_Typ_Gve,
			p.AccountID.ServerString(),
			fmt.Sprintf("%d", p.GetRand().Intn(Gve_BroadCast_IDS_Count)),
			nil,
		)
	}

	// logic log
	logiclog.LogGveMatch(p.AccountID.String(), p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.AccountID.ServerString(), req.NeedCancel, time.Now().Unix()-p.Tmp.MatchBeginTime,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) GetGVEState(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		IsCurr  bool   `codec:"is_game"`
		GameUrl string `codec:"game_url"`
		GameID  string `codec:"game_id"`
		Secert  string `codec:"game_secert"` //TODO: 还没用上,因为有ROOMID,随机产生
		IsBot   bool   `codec:"is_bot"`
	}{}

	initReqRsp(
		"Attr/GetGVEStateRsp",
		r.RawBytes,
		req, resp, p)

	//true, tmp.GameID, tmp.GameSecret, tmp.GameUrl, tmp.IsBot
	resp.IsCurr, resp.GameID, resp.Secert, resp.GameUrl, resp.IsBot = p.Tmp.GetGVEData()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
