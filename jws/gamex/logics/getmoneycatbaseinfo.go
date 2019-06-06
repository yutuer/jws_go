package logics

import (
	"vcs.taiyouxi.net/jws/gamex/modules/moneycat_marquee"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetMoneyCatBaseInfo : 招财进宝信息协议
// 招财进宝信息协议

const Max_Cat_Money_Info = 10

// reqMsgGetMoneyCatBaseInfo 招财进宝信息协议请求消息定义
type reqMsgGetMoneyCatBaseInfo struct {
	Req
}

// rspMsgGetMoneyCatBaseInfo 招财进宝信息协议回复消息定义
type rspMsgGetMoneyCatBaseInfo struct {
	SyncResp
	BoughtTimes int64    `codec:"boughttimes"` // 玩家已参与活动次数
	PlayerNames []string `codec:"p_name"`      //  10个玩家姓名
	PlayerHc    []int64  `codec:"p_hc"`        //10个玩家得到的钻石
}

// GetMoneyCatBaseInfo 招财进宝信息协议: 招财进宝信息协议
func (p *Account) GetMoneyCatBaseInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetMoneyCatBaseInfo)
	rsp := new(rspMsgGetMoneyCatBaseInfo)

	initReqRsp(
		"Attr/GetMoneyCatBaseInfoRsp",
		r.RawBytes,
		req, rsp, p)
	firstInfo := moneycat_marquee.GetModule(p.AccountID.ShardId).GetMoneyCatInfo()
	for i := 0; i < len(firstInfo); i++ {
		rsp.PlayerNames = append(rsp.PlayerNames, firstInfo[i].Player_names)
		rsp.PlayerHc = append(rsp.PlayerHc, firstInfo[i].Player_GetHc)
	}
	rsp.BoughtTimes = p.Profile.GetMoneyCatInfo().GetMoneyCatTime()

	rsp.onChangeMoneyCat()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
