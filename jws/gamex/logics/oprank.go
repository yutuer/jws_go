package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetOpRank : 获取运营活动排行榜
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetOpRank 获取运营活动排行榜请求消息定义
type reqMsgGetOpRank struct {
	Req
	ActivityType int64 `codec:"activity_id"` // 请求某个类型的运营活动排行榜
}

// rspMsgGetOpRank 获取运营活动排行榜回复消息定义
type rspMsgGetOpRank struct {
	SyncResp
	RankInfo  [][]byte `codec:"rank_info"`  // 运营活动排行榜排行信息
	SelfRank  int64    `codec:"self_rank"`  // 个人排名
	SelfScore int64    `codec:"self_score"` // 个人分数
}

// GetOpRank 获取运营活动排行榜:
func (p *Account) GetOpRank(r servers.Request) *servers.Response {
	req := new(reqMsgGetOpRank)
	rsp := new(rspMsgGetOpRank)

	initReqRsp(
		"Attr/GetOpRankRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetOpRankHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OpRankInfo 获取运营活动排行榜
type OpRankInfo struct {
	Acid  string `codec:"acid"`  // 角色acid
	Rank  int64  `codec:"rank"`  // 排名
	Name  string `codec:"name"`  // 名字
	Score int64  `codec:"score"` // 分数
}

// GetOpRankRewardInfo : 获取运营活动排行榜可获得的奖励
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetOpRankRewardInfo 获取运营活动排行榜可获得的奖励请求消息定义
type reqMsgGetOpRankRewardInfo struct {
	Req
	ActivityType int64 `codec:"activity_id"` // 请求某个类型的运营活动排行榜
}

// rspMsgGetOpRankRewardInfo 获取运营活动排行榜可获得的奖励回复消息定义
type rspMsgGetOpRankRewardInfo struct {
	SyncResp
	OpRankReward [][]byte `codec:"rank_reward_info"` // 运营活动排行榜奖励信息
	SelfRank     int64    `codec:"self_rank"`        // 个人排名
	SelfScore    int64    `codec:"self_score"`       // 个人分数
}

// GetOpRankRewardInfo 获取运营活动排行榜可获得的奖励:
func (p *Account) GetOpRankRewardInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetOpRankRewardInfo)
	rsp := new(rspMsgGetOpRankRewardInfo)

	initReqRsp(
		"Attr/GetOpRankRewardInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetOpRankRewardInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OpRankRewardInfo 获取运营活动排行榜可获得的奖励
type OpRankRewardInfo struct {
	RewardID    []string `codec:"reward_id"` // 奖励道具ID
	RewardCount []int64  `codec:"reward_c"`  // 奖励数量
	RankMin     int64    `codec:"rank_min"`  // 排名下限
	RankMax     int64    `codec:"rank_max"`  // 排名上限
	MinCond     int64    `codec:"min_cond"`  // 条件限制
}
