package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetMyRedPacketLog : 获得自己当天的领取列表
// 获得自己当天的领取列表

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetMyRedPacketLog 获得自己当天的领取列表请求消息定义
type reqMsgGetMyRedPacketLog struct {
	Req
	Id string `codec:"id"` // 红包ID
}

// rspMsgGetMyRedPacketLog 获得自己当天的领取列表回复消息定义
type rspMsgGetMyRedPacketLog struct {
	SyncResp
	GrabLog [][]byte `codec:"grablog"` // 抢红包记录
}

// GetMyRedPacketLog 获得自己当天的领取列表: 获得自己当天的领取列表
func (p *Account) GetMyRedPacketLog(r servers.Request) *servers.Response {
	req := new(reqMsgGetMyRedPacketLog)
	rsp := new(rspMsgGetMyRedPacketLog)

	initReqRsp(
		"Attr/GetMyRedPacketLogRsp",
		r.RawBytes,
		req, rsp, p)

	if ok, _ := p.Profile.MarketActivitys.HasRedPacketActivity(p.AccountID.String(), p.GetProfileNowTime()); !ok {
		logs.Warn("claim rp ipa: no available red packet activity")
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	// logic imp begin
	reset := p.GuildProfile.RedPacketInfo.CheckDailyReset(p.GetProfileNowTime())
	p.getMyRpLogList(rsp)
	// logic imp end
	if reset {
		rsp.OnChangeGuildRedPacket()
	}
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// RedPacketLog 获得自己当天的领取列表
type RedPacketLog struct {
	PlayerName string   `codec:"playername"` // 玩家名字
	ItemId     []string `codec:"itemid"`     // 奖励ID
	ItemCount  []int64  `codec:"itemcount"`  // 奖励数量
}

func (p *Account) getMyRpLogList(resp *rspMsgGetMyRedPacketLog) *servers.Response {
	rpLog := make([]RedPacketLog, len(p.GuildProfile.RedPacketInfo.GrabLogList))
	for i, logItem := range p.GuildProfile.RedPacketInfo.GrabLogList {
		rpLog[i] = RedPacketLog{PlayerName: logItem.SenderName}
		rpLog[i].ItemId = make([]string, len(logItem.ItemList))
		rpLog[i].ItemCount = make([]int64, len(logItem.ItemList))
		for j, reward := range logItem.ItemList {
			rpLog[i].ItemId[j] = reward.Id
			rpLog[i].ItemCount[j] = int64(reward.Count)
		}
	}
	resp.GrabLog = make([][]byte, len(rpLog))
	for i, logItem := range rpLog {
		resp.GrabLog[i] = encode(logItem)
	}
	return nil
}
