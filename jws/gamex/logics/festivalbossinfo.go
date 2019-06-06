package logics

import (
	"vcs.taiyouxi.net/jws/gamex/modules/festivalboss"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// FestivalBossInfo : 节日boss基本信息
// 节日Boss基本信息

// reqMsgFestivalBossInfo 节日boss基本信息请求消息定义
type reqMsgFestivalBossInfo struct {
	Req
}

// rspMsgFestivalBossInfo 节日boss基本信息回复消息定义
type rspMsgFestivalBossInfo struct {
	SyncResp
	PlayerNames []string `codec:"py_name"`   // 击杀Boss的玩家姓名
	PlayerTimes []int64  `codec:"py_time"`   // 击杀Boss的时间
	AttackCount int64    `codec:"ack_count"` // 击杀Bosszong次数
}

// FestivalBossInfo 节日boss基本信息: 节日Boss基本信息
func (p *Account) FestivalBossInfo(r servers.Request) *servers.Response {
	req := new(reqMsgFestivalBossInfo)
	rsp := new(rspMsgFestivalBossInfo)

	initReqRsp(
		"Attr/FestivalBossInfoRsp",
		r.RawBytes,
		req, rsp, p)

	bossInfo, ackCount := festivalboss.GetModule(p.AccountID.ShardId).GetFestivalBossInfo()
	for i := 0; i < len(bossInfo); i++ {
		rsp.PlayerNames = append(rsp.PlayerNames, bossInfo[i].Player_names)
		rsp.PlayerTimes = append(rsp.PlayerTimes, bossInfo[i].Player_time)
	}
	rsp.AttackCount = ackCount

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
