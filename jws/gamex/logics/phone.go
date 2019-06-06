package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) TryGetPhoneCode(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Phone string `codec:"phone"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/TryGetPhoneCodeRsp",
		r.RawBytes,
		req, resp, p)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_Phone) {
		return rpcWarn(resp, errCode.ActivityTimeOut)
	}

	acID := p.AccountID.String()
	nowT := p.Profile.GetProfileNowTime()

	playerPhone := p.Profile.GetPhone()
	err := playerPhone.IsCanGetCode(nowT)
	switch err {
	case account.ErrPhoneSmsTooMuchByAccount:
		return rpcWarn(resp, errCode.PhoneRegCannotGetCodeTooMuch)
	case account.ErrPhoneSmsTooFastByAccount:
		return rpcWarn(resp, errCode.PhoneRegCannotGetCodeTooFast)
	case account.ErrPhoneHasGetReward:
		return rpcWarn(resp, errCode.PhoneRegCannotGetCodeHasGotReward)
	}

	err = playerPhone.GetCode(req.Phone, nowT, p.GetRand())
	switch err {
	case account.ErrPhoneSmsTooMuch:
		return rpcWarn(resp, errCode.PhoneRegCannotGetCodeByPhone)
	case account.ErrPhoneFormat:
		return rpcWarn(resp, errCode.PhoneRegPhoneFormatErr)
	}

	if err != nil {
		logs.SentryLogicCritical(acID, "playerPhone GetCode Err By %s", err.Error())
		return rpcWarn(resp, errCode.PhoneRegUnknownErr)
	}

	resp.OnChangePhone()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) UsePhoneCodeForReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		PhoneCode string `codec:"pcode"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"Attr/UsePhoneCodeForRewardRsp",
		r.RawBytes,
		req, resp, p)

	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_Phone) {
		return rpcWarn(resp, errCode.ActivityTimeOut)
	}
	playerPhone := p.Profile.GetPhone()
	if req.PhoneCode != "" && playerPhone.PhoneCode == req.PhoneCode && (!playerPhone.HasGot) {
		rewards := gamedata.GetActivitySpecRewards(gamedata.ActSpecRewardIDXPhone)
		if rewards == nil {
			return rpcError(resp, 1)
		}
		if !account.GiveBySync(p.Account, &rewards.Cost, resp, "PhoneCode") {
			logs.Error("PhoneCode GiveBySync Err %v", rewards)
		}
		logs.Info("[%s]UsePhoneCodeForReward %s", p.AccountID.String(), playerPhone.Phone)
		playerPhone.SetHasBindPhone(playerPhone.Phone)
	} else {
		return rpcWarn(resp, errCode.PhoneRegRegCodeErr)
	}

	resp.OnChangePhone()
	resp.mkInfo(p)

	// log
	logiclog.LogPhone(p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, p.Profile.Name, playerPhone.Phone,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}
