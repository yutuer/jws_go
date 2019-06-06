package logics

import (

	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/redeem_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redeem_code"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"fmt"

	"strings"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
)

type responseRedeemCodeExchange struct {
	SyncRespWithRewards
	Title string `codec:"title"`
}

func (p *Account) redeemCodeExchange(r servers.Request) *servers.Response {
	req := &struct {
		Req
		RedeemCode string `json:"rcode"`
	}{}
	resp := &responseRedeemCodeExchange{}

	initReqRsp(
		"PlayerAttr/RedeemCodeExchangeRsp",
		r.RawBytes,
		req, resp, p)

	code := strings.ToUpper(req.RedeemCode)

	if code == "" {
		return rpcWarn(resp, errCode.RecodeGiftCodeFormatErr)
	}
	acID := p.AccountID.String()
	logs.Trace("[%s]redeemCodeExchange %s", acID, code)

	// 这步处理可以排除大部分胡乱输入
	bID, _, _, isNoLimit, ok := redeemCode.Parse(code)
	if !ok {
		return rpcWarn(resp, errCode.RecodeGiftCodeFormatErr)
	}

	//在领过该批礼包后尝试使用同一批另一组的兑换码，会被告知“你已经领取过XXX大礼包，不能再领
	if p.Profile.GetRedeemCode().IsHasToken(bID) {
		return rpcWarn(resp, errCode.RecodeGiftCodeBatchHasExchange)
	}

	data := redeemCodeModule.GetCodeData(p.AccountID.ShardId, code)
	if data.BatchID == "" {
		// data为空
		return rpcWarn(resp, errCode.RecodeGiftCodeDataErr)
	}

	if data.DoneBy != "" {
		return rpcWarn(resp, errCode.RecodeGiftCodeUsed)
	}

	if data.Bind != "" && data.Bind != fmt.Sprintf("%d", p.AccountID.ShardId) {
		//logs.SentryLogicCritical(acID, "redeemCode Bind Err By %s - %s", data.Bind, game.Cfg.ShardId)
		return rpcWarn(resp, errCode.RecodeGiftCodeBindErr)
	}

	now_t := p.Profile.GetProfileNowTime()

	if now_t < data.Begin {
		return rpcWarn(resp, errCode.RecodeGiftCodeTimeNoStart)
	}

	if now_t > data.End {
		return rpcWarn(resp, errCode.RecodeGiftCodeTimeout)
	}

	if !isNoLimit {
		redeemCodeModule.SetCodeUsed(p.AccountID.ShardId, acID, code)
		logs.Info("[%s]redeemCodeExchange Limit %s", acID, code)
	} else {
		logs.Info("[%s]redeemCodeExchange NoLimit %s", acID, code)
	}
	var gives gamedata.CostData
	for idx, r := range data.ItemIDs {
		gives.AddItem(r, data.Counts[idx])
	}

	resp.Title = data.Title
	p.Profile.GetRedeemCode().SetHasToken(bID)
	account.GiveBySync(p.Account, &gives, resp, "RedeemCode")
	resp.mkInfo(p)

	// logiclog
	logiclog.LogRedeemCode(acID, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, p.Profile.Name, code, bID, isNoLimit,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}
