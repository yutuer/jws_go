package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ExchangeHeroPiece : 请求兑换武将碎片
// 武将碎片兑换令牌

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExchangeHeroPiece 请求兑换武将碎片请求消息定义
type reqMsgExchangeHeroPiece struct {
	Req
	HeroAvatar int64 `codec:"hero_avatar"` // 武将int类型的ID
	IsTen      bool  `codec:"is_ten"`      // 是否是兑换10次
}

// rspMsgExchangeHeroPiece 请求兑换武将碎片回复消息定义
type rspMsgExchangeHeroPiece struct {
	SyncRespWithRewards
}

// ExchangeHeroPiece 请求兑换武将碎片: 武将碎片兑换令牌
func (p *Account) ExchangeHeroPiece(r servers.Request) *servers.Response {
	req := new(reqMsgExchangeHeroPiece)
	rsp := new(rspMsgExchangeHeroPiece)

	initReqRsp(
		"Attr/ExchangeHeroPieceRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ExchangeHeroPieceHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// DrawHeroPieceGacha : 武将碎片抽奖
// 武将碎片抽奖

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgDrawHeroPieceGacha 武将碎片抽奖请求消息定义
type reqMsgDrawHeroPieceGacha struct {
	Req
	HeroPieceGachaId    int64 `codec:"hpg_id"`    // 想要抽奖的ID
	HeroPieceGachaSubId int64 `codec:"hpg_subid"` // 想要抽奖的SubID
	IsTen               bool  `codec:"is_ten"`    // 是否是十连抽
}

// rspMsgDrawHeroPieceGacha 武将碎片抽奖回复消息定义
type rspMsgDrawHeroPieceGacha struct {
	SyncRespWithRewards
	GiveRewardId    string   `codec:"bg_grid"`  // 想要抽奖的ID
	GiveRewardCount int64    `codec:"bg_grc"`   // 想要抽奖的ID
	ExtRewardId     []string `codec:"bg_extid"` // 想要抽奖的ID
	ExtRewardCount  []int64  `codec:"bg_extc"`  // 想要抽奖的ID
	ExtRewardData   []string `codec:"bg_extd"`  // 想要抽奖的ID
}

// DrawHeroPieceGacha 武将碎片抽奖: 武将碎片抽奖
func (p *Account) DrawHeroPieceGacha(r servers.Request) *servers.Response {
	req := new(reqMsgDrawHeroPieceGacha)
	rsp := new(rspMsgDrawHeroPieceGacha)

	initReqRsp(
		"Attr/DrawHeroPieceGachaRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.DrawHeroPieceGachaHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
