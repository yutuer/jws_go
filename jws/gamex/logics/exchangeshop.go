package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ExchangeProp : 兑换商品道具
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExchangeProp 兑换商品道具请求消息定义
type reqMsgExchangeProp struct {
	Req
	ExchangeID int64 `codec:"exchange_id"` // 请求兑换的商品ID
	ActivityID int64 `codec:"activity_id"` // 兑换商店活动ID
}

// rspMsgExchangeProp 兑换商品道具回复消息定义
type rspMsgExchangeProp struct {
	SyncRespWithRewards
	AlreadyExchangeTimes int64 `codec:"a_exchange_t"` // 已经兑换的次数
}

// ExchangeProp 兑换商品道具:
func (p *Account) ExchangeProp(r servers.Request) *servers.Response {
	req := new(reqMsgExchangeProp)
	rsp := new(rspMsgExchangeProp)

	initReqRsp(
		"Attr/ExchangePropRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ExchangePropHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetExchangeShopInfo : 获取兑换商店信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetExchangeShopInfo 获取兑换商店信息请求消息定义
type reqMsgGetExchangeShopInfo struct {
	Req
	ActivityID int64 `codec:"activity_id"` // 兑换商店活动ID
}

// rspMsgGetExchangeShopInfo 获取兑换商店信息回复消息定义
type rspMsgGetExchangeShopInfo struct {
	SyncResp
	ExchangePropInfo [][]byte `codec:"exchange_prop_info"` // 兑换商店商品信息
	HasAutoProp      int64    `codec:"has_auto_prop"`      // 是否有道具自动转换，0代表无，1代表有
}

// GetExchangeShopInfo 获取兑换商店信息:
func (p *Account) GetExchangeShopInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetExchangeShopInfo)
	rsp := new(rspMsgGetExchangeShopInfo)

	initReqRsp(
		"Attr/GetExchangeShopInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetExchangeShopInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ExchangePropInfo 获取兑换商店信息
type ExchangePropInfo struct {
	ExchangeID           []string `codec:"exchange_id"`          // 兑换商品所需道具ID
	ExchangeCount        []int64  `codec:"exchange_c"`           // 兑换商品所需道具数量
	ShowPropID           string   `codec:"show_prop_id"`         // 展示物品ID
	ShowPropCount        int64    `codec:"show_prop_c"`          // 展示物品数量
	ExchangeLimitTimes   int64    `codec:"exchange_limit_times"` // 兑换限制次数
	ExchangePropIndex    int64    `codec:"exchange_prop_i"`      // 兑换商店商品索引
	AlreadyExchangeTimes int64    `codec:"a_exchange_t"`         // 已经兑换的次数
}
