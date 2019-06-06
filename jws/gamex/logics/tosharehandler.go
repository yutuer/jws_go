package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TwitterShare : Twitter分享
//
func (p *Account) TwitterShareHandler(req *reqMsgTwitterShare, resp *rspMsgTwitterShare) uint32 {

	if p.Profile.IsTwitterShared {
		return errCode.ClickTooQuickly
	}
	p.Profile.IsTwitterShared = true
	logs.Debug("Twitter share reward")
	if !account.GiveBySync(p.Account, gamedata.GetActivitySpecRewards(3).Gives(), resp, "Twtter") {
		logs.Error("Twtter GiveBySync Err")
	}

	resp.OnChangeTwitter()
	return 0
}

// LineShare : Line分享
//
func (p *Account) LineShareHandler(req *reqMsgLineShare, resp *rspMsgLineShare) uint32 {
	if p.Profile.IsLineShared {
		return errCode.ClickTooQuickly
	}
	p.Profile.IsLineShared = true
	logs.Debug("Line share reward")
	if !account.GiveBySync(p.Account, gamedata.GetActivitySpecRewards(4).Gives(), resp, "Line") {
		logs.Error("Line GiveBySync Err")
	}

	resp.OnChangeLine()
	return 0
}
