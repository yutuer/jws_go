package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestActiveGiftActivity struct {
	Req
	Id uint32 `codec:"id"`
}

type ResponseActiveGiftActivity struct {
	SyncRespWithRewards
}

// 此协议用来激活活动（只限根据玩家行为激活的活动），目前客户端每次打开界面都发（如：7天领奖）
func (p *Account) ActiveGiftActivity(r servers.Request) *servers.Response {
	req := &RequestActiveGiftActivity{}
	resp := &ResponseActiveGiftActivity{}

	initReqRsp(
		"PlayerAttr/ActiveGiftActivityResponse", r.RawBytes,
		req, resp, p)

	const (
		_                   = iota
		CODE_GIFT_NOT_FOUND // 失败：活动未找到
	)

	player_gift := p.Profile.GetGifts()
	flag, err := player_gift.ActiveGift(req.Id, p.Profile.GetProfileNowTime())
	//	if err != nil {
	//		return rpcError(resp, CODE_GIFT_NOT_FOUND)
	//	}
	if err == nil && flag {
		resp.OnChangeGiftStateChange()
		resp.mkInfo(p)
	}

	return rpcSuccess(resp)
}

type RequestGetGiftActivity struct {
	Req
	Id uint32 `codec:"id"`
}

type ResponseGetGiftActivity struct {
	SyncRespWithRewards
}

func (p *Account) GetGiftActivity(r servers.Request) *servers.Response {
	req := &RequestGetGiftActivity{}
	resp := &ResponseGetGiftActivity{}

	initReqRsp(
		"PlayerAttr/GetGiftActivityResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                = iota
		CODE_No_Gift_Err // 失败:没奖可领
		CODE_Give_Err    // 失败:发奖错误
	)

	player_gift := p.Profile.GetGifts()
	player_vip := p.Profile.GetVipLevel()

	has_gift, data := player_gift.GetGiftToGet(req.Id, player_vip, p.Profile.GetProfileNowTime())

	if has_gift && data != nil {
		ok := account.GiveBySync(p.Account, &data.Cost, resp, fmt.Sprintf("ActGiftDaily-%d", req.Id))
		if !ok {
			return rpcError(resp, CODE_Give_Err)
		}

		player_gift.SetHasGet(req.Id, player_vip, p.Profile.GetProfileNowTime())

		resp.OnChangeGiftStateChange()
		resp.OnChangeBuy()
		resp.mkInfo(p)

	} else {
		// 正常情况下客户端不应该明知没奖励还要申请领奖
		logs.Warn("GetGiftActivity CODE_No_Gift_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	return rpcSuccess(resp)
}

// 活动红点，目前就只有一个活动（七日领奖）就先写死
func (p *Account) ActGiftRedPoint() bool {
	player_gift := p.Profile.GetGifts()
	player_vip := p.Profile.GetVipLevel()
	has_gift, data := player_gift.GetGiftToGet(1, player_vip, p.Profile.GetProfileNowTime())
	if has_gift && data != nil {
		return true
	}
	return false
}
