package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) TitleActivate(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Id string `codec:"id"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/TitleActivateRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Title_Not_Activate
		Err_Cfg_Not_Found
		Err_Cost
	)

	if req.Id == "" {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}

	mTitle := p.Profile.GetTitle()
	mTitle.UpdateTitle(p.Account, p.Profile.GetProfileNowTime())
	if !mTitle.IsCanActivate(req.Id) {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	cfg := gamedata.GetTitleCfg(req.Id)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_Cfg_Not_Found, "Err_Cfg_Not_Found")
	}

	// 花费
	data := &gamedata.CostData{}
	data.AddItem(cfg.GetCostItem(), cfg.GetCostNum())
	cost := &account.CostGroup{}
	if !cost.AddCostData(p.Account, data) || !cost.CostBySync(p.Account, resp, "TitleActivate") {
		logs.Error("Err_Cost")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	mTitle.ActivateTitle(req.Id)
	p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号

	resp.OnChangeTitle()
	resp.mkInfo(p)
	logiclog.LogTitleChange(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		req.Id,
		"",
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return rpcSuccess(resp)
}

func (p *Account) TitleTakeOnOff(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Id string `codec:"id"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/TitleTakeOnOffRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Title_Not_Found
	)

	mTitle := p.Profile.GetTitle()
	mTitle.UpdateTitle(p.Account, p.Profile.GetProfileNowTime())
	if req.Id == "" {
		mTitle.SetTitleOn("")
	} else {
		found := false
		titles := mTitle.GetTitles()
		for _, t := range titles {
			if t == req.Id {
				found = true
			}
		}
		if !found {
			resp.OnChangeTitle()
			return rpcSuccess(resp)
		}
		mTitle.SetTitleOn(req.Id)
	}

	resp.OnChangeTitle()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) TitleClearHint(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/TitleClearHintRsp",
		r.RawBytes,
		req, resp, p)

	profile := &p.Profile
	mTitle := profile.GetTitle()
	mTitle.TitleForClient = make(map[string]struct{}, 2)

	resp.OnChangeTitle()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
