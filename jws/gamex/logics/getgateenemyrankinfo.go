package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetGateEnemyRankInfo : 获取兵临城下排行信息
// 请求兵临城下排行信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetGateEnemyRankInfo 获取兵临城下排行信息请求消息定义
type reqMsgGetGateEnemyRankInfo struct {
	Req
}

// rspMsgGetGateEnemyRankInfo 获取兵临城下排行信息回复消息定义
type rspMsgGetGateEnemyRankInfo struct {
	SyncResp
	RankNames  []string `codec:"ranknames"`  // 角色名字
	RankPoints []int64  `codec:"rankpoints"` // 角色积分
}

// GetGateEnemyRankInfo 获取兵临城下排行信息: 请求兵临城下排行信息
func (p *Account) GetGateEnemyRankInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetGateEnemyRankInfo)
	rsp := new(rspMsgGetGateEnemyRankInfo)

	initReqRsp(
		"Attr/GetGateEnemyRankInfoRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.RankNames = p.Profile.GetGatesEnemy().GetPushData().Names
	rsp.RankPoints = make([]int64, len(p.Profile.GetGatesEnemy().GetPushData().Points))
	for i, point := range p.Profile.GetGatesEnemy().GetPushData().Points {
		rsp.RankPoints[i] = int64(point)
	}
	return rpcSuccess(rsp)
}
