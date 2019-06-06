package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// RedPacket7days : 七日红包
// 七日红包

// reqMsgRedPacket7days 七日红包请求消息定义
type reqMsgRedPacket7days struct {
	Req
	RedPacketId int64 `codec:"rp_id"` // 红包ID
}

// rspMsgRedPacket7days 七日红包回复消息定义
type rspMsgRedPacket7days struct {
	SyncRespWithRewards
}

// RedPacket7days 七日红包: 七日红包
func (p *Account) RedPacket7days(r servers.Request) *servers.Response {
	req := new(reqMsgRedPacket7days)
	rsp := new(rspMsgRedPacket7days)

	initReqRsp(
		"Attr/RedPacket7daysRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_RedPacket_NOT_TODAY
		Err_RedPacket_Has_Get
	)

	rpd := p.Profile.GetRedPacket7day()

	if rpd.GetDay(p.GetProfileNowTime())+1 != int64(gamedata.GetCommonCfg().GetReceiveDay()) {
		return rpcErrorWithMsg(rsp, Err_RedPacket_NOT_TODAY,
			fmt.Sprintf("Err_RedPacket_NOT_TODAY %d", rpd.GetDay(p.GetProfileNowTime())))
	}

	if rpd.SaveHc[int(req.RedPacketId)] == -1 {
		return rpcErrorWithMsg(rsp, Err_RedPacket_Has_Get,
			fmt.Sprintf("Err_RedPacket_Has_Get %d", rpd.SaveHc[int(req.RedPacketId)]))
	}

	hcNum := rpd.GetDayHcNum(int(req.RedPacketId)) / gamedata.GetCommonCfg().GetRedpackeTratio()
	data := &gamedata.CostData{}
	data.AddItem(gamedata.VI_Hc, hcNum)

	if !account.GiveBySync(p.Account, data, rsp, "RedPacket7Days") {
		logs.Error("RedPacket7Days GiveBySync Err")
	}

	rpd.SetPacket2Done(int(req.RedPacketId))

	rsp.OnChangeRedPacket7Days()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
