package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// WheelShowInfo : 幸运转盘展示
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgWheelShowInfo 幸运转盘展示请求消息定义
type reqMsgWheelShowInfo struct {
	Req
}

// rspMsgWheelShowInfo 幸运转盘展示回复消息定义
type rspMsgWheelShowInfo struct {
	SyncResp
	ItemID     []string `codec:"item_id"`    // 道具ID
	ItemCount  []int64  `codec:"item_c"`     // 道具数量
	HcBase     int64    `codec:"hc_base"`    // 充值基数
	GetCoin    int64    `codec:"get_coin_c"` // 充值可得积分
	GetCoinMax int64    `codec:"coin_max_c"` // 充值可得积分上限
	Index      int64    `codec:"index"`      // 兑换次数
	Special    []int64  `codec:"special"`    // 是否流光
	CurCoin    int64    `codec:"cur_coin_c"` // 当前代币数
	GachaCoin  string   `codec:"gacha_coin"` // 兑换所需代币种类
	CoinCost   int64    `codec:"coin_cost"`  // 消耗代币的数量
	DontShow   []int64  `codec:"dont_show"`  // 已经被抽取
	RareLevel  []int64  `codec:"rare_level"` // 稀有度
}

// WheelShowInfo 幸运转盘展示:
func (p *Account) WheelShowInfo(r servers.Request) *servers.Response {
	req := new(reqMsgWheelShowInfo)
	rsp := new(rspMsgWheelShowInfo)

	initReqRsp(
		"Attr/WheelShowInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.WheelShowInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// UseWheelOne : 转动一次转盘
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgUseWheelOne 转动一次转盘请求消息定义
type reqMsgUseWheelOne struct {
	Req
}

// rspMsgUseWheelOne 转动一次转盘回复消息定义
type rspMsgUseWheelOne struct {
	SyncRespWithRewards
	ItemID    string `codec:"item_id"`   // 道具ID
	ItemCount int64  `codec:"item_c"`    // 道具数量
	Unusual   int64  `codec:"unusual"`   // 稀有烟花特效
	NextCost  int64  `codec:"next_cost"` // 下一次的消耗
}

// UseWheelOne 转动一次转盘:
func (p *Account) UseWheelOne(r servers.Request) *servers.Response {
	req := new(reqMsgUseWheelOne)
	rsp := new(rspMsgUseWheelOne)

	initReqRsp(
		"Attr/UseWheelOneRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.UseWheelOneHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
