package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetRedPacketLog : 获得某个红包的领取列表
// 获得某个红包的领取列表

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetRedPacketLog 获得某个红包的领取列表请求消息定义
type reqMsgGetRedPacketLog struct {
	Req
	Id string `codec:"id"` // 红包ID
}

// rspMsgGetRedPacketLog 获得某个红包的领取列表回复消息定义
type rspMsgGetRedPacketLog struct {
	SyncResp
	RedPacketLog [][]byte `codec:"rplog"` // 红包领取记录
}

// GetRedPacketLog 获得某个红包的领取列表: 获得某个红包的领取列表
func (p *Account) GetRedPacketLog(r servers.Request) *servers.Response {
	req := new(reqMsgGetRedPacketLog)
	rsp := new(rspMsgGetRedPacketLog)

	initReqRsp(
		"Attr/GetRedPacketLogRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if ok, _ := p.Profile.MarketActivitys.HasRedPacketActivity(p.AccountID.String(), p.GetProfileNowTime()); !ok {
		logs.Warn("claim rp ipa: no available red packet activity")
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}
	p.GuildProfile.RedPacketInfo.CheckDailyReset(p.GetProfileNowTime())
	p.getRedPacketLogList(req.Id, rsp)
	// logic imp end
	rsp.OnChangeGuildRedPacket()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) getRedPacketLogList(rpId string, resp *rspMsgGetRedPacketLog) *servers.Response {
	guildInfo, ret := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.Account.GuildProfile.GuildUUID)
	if rsp := guildErrRet(ret, resp); rsp != nil {
		return rsp
	}
	redPacket, ok := guildInfo.GuildRedPacket.Get(rpId)
	if !ok {
		return rpcWarn(resp, errCode.RedPacketNotFound)
	}
	rpLog := make([]RedPacketLog, len(redPacket.GrabRPRecord))
	for i, logItem := range redPacket.GrabRPRecord {
		rpLog[i] = RedPacketLog{PlayerName: logItem.PlayerName}
		rpLog[i].ItemId = make([]string, len(logItem.RewardList))
		rpLog[i].ItemCount = make([]int64, len(logItem.RewardList))
		for j, reward := range logItem.RewardList {
			rpLog[i].ItemId[j] = reward.ItemId
			rpLog[i].ItemCount[j] = int64(reward.Count)
		}
	}
	resp.RedPacketLog = make([][]byte, len(rpLog))
	for i, logItem := range rpLog {
		resp.RedPacketLog[i] = encode(logItem)
	}
	return nil
}
