package logics

import (
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// BuyItem : 购买部分道具
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBuyItem 购买部分道具请求消息定义
type reqMsgBuyItem struct {
	Req
	PropID    string `codec:"prop_id"` // 想要购买的物品ID
	PropCount int64  `codec:"prop_c"`  // 想要购买的物品数量
}

// rspMsgBuyItem 购买部分道具回复消息定义
type rspMsgBuyItem struct {
	SyncRespWithRewards
}

// BuyItem 购买部分道具:
func (p *Account) BuyItem(r servers.Request) *servers.Response {
	req := new(reqMsgBuyItem)
	rsp := new(rspMsgBuyItem)

	initReqRsp(
		"Attr/BuyItemRsp",
		r.RawBytes,
		req, rsp, p)

	if req.PropCount < 0 || req.PropCount > uutil.CHEAT_INT_MAX {
		return rpcErrorWithMsg(rsp, 99, "BuyItem Count cheat")
	}

	// logic imp begin
	warnCode := p.BuyItemHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
