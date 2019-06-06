package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetOfflineRecoverInfo : 获取离线资源相关信息
// 离线资源

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetOfflineRecoverInfo 获取离线资源相关信息请求消息定义
type reqMsgGetOfflineRecoverInfo struct {
	Req
}

// rspMsgGetOfflineRecoverInfo 获取离线资源相关信息回复消息定义
type rspMsgGetOfflineRecoverInfo struct {
	SyncResp
	OfflineResources [][]byte `codec:"or_list"` // 所有离线资源奖励
}

// GetOfflineRecoverInfo 获取离线资源相关信息: 离线资源
func (p *Account) GetOfflineRecoverInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetOfflineRecoverInfo)
	rsp := new(rspMsgGetOfflineRecoverInfo)

	initReqRsp(
		"Attr/GetOfflineRecoverInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetOfflineRecoverInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OfflineResource 获取离线资源相关信息
type OfflineResource2Client struct {
	ScId          string `codec:"or_id"`       // 资源ID
	ScOfflineDays int64  `codec:"or_off_days"` // 累计的离线天数
}

// ClaimOfflineRecoverReward : 领取离线资源的奖励
// 领取离线资源的奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgClaimOfflineRecoverReward 领取离线资源的奖励请求消息定义
type reqMsgClaimOfflineRecoverReward struct {
	Req
	ScId   string `codec:"or_id"`   // 资源ID
	IsFree bool   `codec:"or_free"` // true=免费领取 false=花钱领取
}

// rspMsgClaimOfflineRecoverReward 领取离线资源的奖励回复消息定义
type rspMsgClaimOfflineRecoverReward struct {
	SyncRespWithRewards
}

// ClaimOfflineRecoverReward 领取离线资源的奖励: 领取离线资源的奖励
func (p *Account) ClaimOfflineRecoverReward(r servers.Request) *servers.Response {
	req := new(reqMsgClaimOfflineRecoverReward)
	rsp := new(rspMsgClaimOfflineRecoverReward)

	initReqRsp(
		"Attr/ClaimOfflineRecoverRewardRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ClaimOfflineRecoverRewardHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
