package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetWBInfo : 获取世界boss信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWBInfo 获取世界boss信息请求消息定义
type reqMsgGetWBInfo struct {
	Req
}

// rspMsgGetWBInfo 获取世界boss信息回复消息定义
type rspMsgGetWBInfo struct {
	SyncResp
	BossID      string  `codec:"boss_id"`      // boss ID
	BossLevel   int64   `codec:"boss_level"`   // boss等级
	BossScene   string  `codec:"BossScene"`    // boss场景
	BossDamaged int64   `codec:"boss_damaged"` // boss目前受到的伤害量
	SelfDamage  int64   `codec:"self_damage"`  // 自己对boss造成的伤害
	SelfRank    int64   `codec:"self_rank"`    // 自己对boss造成伤害的排名
	BuffLevel   int64   `codec:"buff_level"`   // 当前上古之力buff层数
	LeftTimes   int64   `codec:"left_times"`   // 剩余挑战次数
	GotRewards  []int64 `codec:"got_rewards"`  // 已经领取的奖励档位
}

// GetWBInfo 获取世界boss信息:
func (p *Account) GetWBInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetWBInfo)
	rsp := new(rspMsgGetWBInfo)

	initReqRsp(
		"Attr/GetWBInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWBInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BeginWB : 开始挑战世界boss
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBeginWB 开始挑战世界boss请求消息定义
type reqMsgBeginWB struct {
	Req
	BossID    string `codec:"boss_id"`    // boss ID
	BossLevel int64  `codec:"boss_level"` // boss等级
}

// rspMsgBeginWB 开始挑战世界boss回复消息定义
type rspMsgBeginWB struct {
	SyncResp
	BossHP         int64    `codec:"boss_hp"`      // boss当前生命值
	Seq            int64    `codec:"seq"`          // boss序列
	BossID         string   `codec:"boss_id"`      // boss ID
	BossLevel      int64    `codec:"boss_level"`   // boss等级
	BossDamaged    int64    `codec:"boss_damaged"` // boss受到的总伤害
	WBFewRankInfos [][]byte `codec:"rank_info"`    // 世界boss少量排行榜排行信息
	SelfRank       int64    `codec:"self_rank"`    // 个人排名
	SelfScore      int64    `codec:"self_score"`   // 个人分数
}

// BeginWB 开始挑战世界boss:
func (p *Account) BeginWB(r servers.Request) *servers.Response {
	req := new(reqMsgBeginWB)
	rsp := new(rspMsgBeginWB)

	initReqRsp(
		"Attr/BeginWBRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BeginWBHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// EndWB : 结束挑战世界boss
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgEndWB 结束挑战世界boss请求消息定义
type reqMsgEndWB struct {
	Req
	BossID    string `codec:"boss_id"`    // boss ID
	BossLevel int64  `codec:"boss_level"` // boss等级
}

// rspMsgEndWB 结束挑战世界boss回复消息定义
type rspMsgEndWB struct {
	SyncRespWithRewards
	BuffLevel     int64 `codec:"buff_level"`      // 当前上古之力buff层数
	LeftTimes     int64 `codec:"left_times"`      // 剩余挑战次数
	RemindBuyBuff bool  `codec:"remind_buy_buff"` // 是否提醒购买buff
	RoundDamage   int64 `codec:"round_damage"`    // 本场战斗伤害
}

// EndWB 结束挑战世界boss:
func (p *Account) EndWB(r servers.Request) *servers.Response {
	req := new(reqMsgEndWB)
	rsp := new(rspMsgEndWB)

	initReqRsp(
		"Attr/EndWBRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.EndWBHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetWBRankInfo : 获取世界boss伤害排行榜信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWBRankInfo 获取世界boss伤害排行榜信息请求消息定义
type reqMsgGetWBRankInfo struct {
	Req
	Typ int64 `codec:"typ"` // 1代表总榜,2代表最佳阵容榜
}

// rspMsgGetWBRankInfo 获取世界boss伤害排行榜信息回复消息定义
type rspMsgGetWBRankInfo struct {
	SyncResp
	WBRankInfos    [][]byte `codec:"rank_info"`      // 世界boss排行榜排行信息
	SelfWBRankInfo []byte   `codec:"self_rank_info"` // 世界boss排行榜个人排行信息
}

// GetWBRankInfo 获取世界boss伤害排行榜信息:
func (p *Account) GetWBRankInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetWBRankInfo)
	rsp := new(rspMsgGetWBRankInfo)

	initReqRsp(
		"Attr/GetWBRankInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWBRankInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WBRankInfo 获取世界boss伤害排行榜信息
type WBRankInfo struct {
	Acid      string  `codec:"acid"`       // 角色acid
	Rank      int64   `codec:"rank"`       // 排名
	Name      string  `codec:"name"`       // 名字
	Score     int64   `codec:"score"`      // 分数
	HeroID    []int64 `codec:"hero_id"`    // 最佳阵容主将ID
	HeroStar  []int64 `codec:"hero_star"`  // 最佳阵容主将星级
	BuffLevel int64   `codec:"buff_level"` // buff level
}

// UpdateBattleInfo : 战斗中更新自己的伤害量、boss血量、伤害排行榜
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgUpdateBattleInfo 战斗中更新自己的伤害量、boss血量、伤害排行榜请求消息定义
type reqMsgUpdateBattleInfo struct {
	Req
	BossID     string  `codec:"boss_id"`     // boss ID
	BossLevel  int64   `codec:"boss_level"`  // boss等级
	SelfDamage int64   `codec:"self_damage"` // 个人造成的伤害量
	CheatParam []int64 `codec:"cheat_param"` // cheat的参数
}

// rspMsgUpdateBattleInfo 战斗中更新自己的伤害量、boss血量、伤害排行榜回复消息定义
type rspMsgUpdateBattleInfo struct {
	SyncResp
	BossHP         int64    `codec:"boss_hp"`      // boss当前生命值
	BossID         string   `codec:"boss_id"`      // boss ID
	BossLevel      int64    `codec:"boss_level"`   // boss等级
	BossDamaged    int64    `codec:"boss_damaged"` // boss受到的总伤害
	WBFewRankInfos [][]byte `codec:"rank_info"`    // 世界boss少量排行榜排行信息
	SelfRank       int64    `codec:"self_rank"`    // 个人排名
	SelfScore      int64    `codec:"self_score"`   // 个人分数
	Seq            int64    `codec:"seq"`          // boss序列
	IsCheat        bool     `codec:"is_cheat"`     // 是否作弊
}

// UpdateBattleInfo 战斗中更新自己的伤害量、boss血量、伤害排行榜:
func (p *Account) UpdateBattleInfo(r servers.Request) *servers.Response {
	req := new(reqMsgUpdateBattleInfo)
	rsp := new(rspMsgUpdateBattleInfo)

	initReqRsp(
		"Attr/UpdateBattleInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.UpdateBattleInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	return rpcSuccess(rsp)
}

// WBFewRankInfo 战斗中更新自己的伤害量、boss血量、伤害排行榜
type WBFewRankInfo struct {
	Acid  string `codec:"acid"`  // 角色acid
	Rank  int64  `codec:"rank"`  // 排名
	Name  string `codec:"name"`  // 名字
	Score int64  `codec:"score"` // 分数
}

// UseBuff : 使用上古之力
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgUseBuff 使用上古之力请求消息定义
type reqMsgUseBuff struct {
	Req
	UseCount int64 `codec:"use_count"` // 使用的数量
}

// rspMsgUseBuff 使用上古之力回复消息定义
type rspMsgUseBuff struct {
	SyncResp
	BuffLevel int64 `codec:"buff_level"` // 当前上古之力buff层数
}

// UseBuff 使用上古之力:
func (p *Account) UseBuff(r servers.Request) *servers.Response {
	req := new(reqMsgUseBuff)
	rsp := new(rspMsgUseBuff)

	initReqRsp(
		"Attr/UseBuffRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.UseBuffHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetWBRankRewards : 领取排行奖励
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWBRankRewards 领取排行奖励请求消息定义
type reqMsgGetWBRankRewards struct {
	Req
	LevelID int64 `codec:"level_id"` // 领取的奖励档位
}

// rspMsgGetWBRankRewards 领取排行奖励回复消息定义
type rspMsgGetWBRankRewards struct {
	SyncRespWithRewards
	GotRewards []int64 `codec:"got_rewards"` // 已经领取的奖励档位
}

// GetWBRankRewards 领取排行奖励:
func (p *Account) GetWBRankRewards(r servers.Request) *servers.Response {
	req := new(reqMsgGetWBRankRewards)
	rsp := new(rspMsgGetWBRankRewards)

	initReqRsp(
		"Attr/GetWBRankRewardsRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWBRankRewardsHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetWBPlayerDetail : 查看世界boss玩家信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWBPlayerDetail 查看世界boss玩家信息请求消息定义
type reqMsgGetWBPlayerDetail struct {
	Req
	AcID string `codec:"aicd"` // 所查询玩家的acid
}

// rspMsgGetWBPlayerDetail 查看世界boss玩家信息回复消息定义
type rspMsgGetWBPlayerDetail struct {
	SyncResp
	DetailInfo []byte `codec:"detail_info"` // 玩家详细信息
}

// GetWBPlayerDetail 查看世界boss玩家信息:
func (p *Account) GetWBPlayerDetail(r servers.Request) *servers.Response {
	req := new(reqMsgGetWBPlayerDetail)
	rsp := new(rspMsgGetWBPlayerDetail)

	initReqRsp(
		"Attr/GetWBPlayerDetailRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWBPlayerDetailHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// SetBuyBuffReminder : 设置是否提醒购买buff
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgSetBuyBuffReminder 设置是否提醒购买buff请求消息定义
type reqMsgSetBuyBuffReminder struct {
	Req
	RemindBuyBuff bool `codec:"remind_buy_buff"` // 是否提醒购买buff
}

// rspMsgSetBuyBuffReminder 设置是否提醒购买buff回复消息定义
type rspMsgSetBuyBuffReminder struct {
	SyncResp
	RemindBuyBuff bool `codec:"remind_buy_buff"` // 是否提醒购买buff
}

// SetBuyBuffReminder 设置是否提醒购买buff:
func (p *Account) SetBuyBuffReminder(r servers.Request) *servers.Response {
	req := new(reqMsgSetBuyBuffReminder)
	rsp := new(rspMsgSetBuyBuffReminder)

	initReqRsp(
		"Attr/SetBuyBuffReminderRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.SetBuyBuffReminderHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
