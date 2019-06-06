package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// SetWSPVPDefenseFormation : 布阵无双争霸防守阵容
// 无双争霸防守阵容

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgSetWSPVPDefenseFormation 布阵无双争霸防守阵容请求消息定义
type reqMsgSetWSPVPDefenseFormation struct {
	Req
	AvatarId []int64 `codec:"avatar"` // 武将ID
}

// rspMsgSetWSPVPDefenseFormation 布阵无双争霸防守阵容回复消息定义
type rspMsgSetWSPVPDefenseFormation struct {
	SyncResp
}

// SetWSPVPDefenseFormation 布阵无双争霸防守阵容: 无双争霸防守阵容
func (p *Account) SetWSPVPDefenseFormation(r servers.Request) *servers.Response {
	req := new(reqMsgSetWSPVPDefenseFormation)
	rsp := new(rspMsgSetWSPVPDefenseFormation)

	initReqRsp(
		"Attr/SetWSPVPDefenseFormationRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.SetWSPVPDefenseFormationHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetMatchOpponent : 获取匹配对手
// 随机4个对手

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetMatchOpponent 获取匹配对手请求消息定义
type reqMsgGetMatchOpponent struct {
	Req
	ForceRefresh bool `codec:"force_refresh"` // 刷新对手列表
}

// rspMsgGetMatchOpponent 获取匹配对手回复消息定义
type rspMsgGetMatchOpponent struct {
	SyncRespWithRewards
	Opponent        [][]byte `codec:"avatar"`         // 匹配对手信息
	IsMyRankChanged bool     `codec:"is_rank_change"` // 匹配对手信息
}

// GetMatchOpponent 获取匹配对手: 随机4个对手
func (p *Account) GetMatchOpponent(r servers.Request) *servers.Response {
	req := new(reqMsgGetMatchOpponent)
	rsp := new(rspMsgGetMatchOpponent)

	initReqRsp(
		"Attr/GetMatchOpponentRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetMatchOpponentHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WSPVPOpp 获取匹配对手
type WSPVPOpp struct {
	Acid      string `codec:"acid"`      // 角色acid
	Rank      int64  `codec:"rank"`      // 排名
	ServerId  string `codec:"sid"`       // 区服
	Name      string `codec:"name"`      // 名字
	GuildName string `codec:"gname"`     // 军团名字
	TitleId   string `codec:"title_id"`  // 当前称号
	VipLevel  int64  `codec:"vip_level"` // VIP等级
}

// LockWSPVPBattle : 锁定无双争霸战斗
// 锁定单个武将

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgLockWSPVPBattle 锁定无双争霸战斗请求消息定义
type reqMsgLockWSPVPBattle struct {
	Req
	Acid string `codec:"acid"` // 角色ACID
}

// rspMsgLockWSPVPBattle 锁定无双争霸战斗回复消息定义
type rspMsgLockWSPVPBattle struct {
	SyncResp
	Formation   []int64 `codec:"formation"`    // 返回每个位置对应的武将idx，固定长度9
	HeroStar    []int64 `codec:"hero_star"`    // 返回每个位置武将的升星等级
	CorpGs      []int64 `codec:"corpgs"`       // 每个队伍的战力，固定长度3
	MyFormation []int64 `codec:"my_formation"` // 返回己方上次阵容
}

// LockWSPVPBattle 锁定无双争霸战斗: 锁定单个武将
func (p *Account) LockWSPVPBattle(r servers.Request) *servers.Response {
	req := new(reqMsgLockWSPVPBattle)
	rsp := new(rspMsgLockWSPVPBattle)

	initReqRsp(
		"Attr/LockWSPVPBattleRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.LockWSPVPBattleHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BeginWSPVPBattle : 开始无双争霸战斗
// 选好己方阵容，开始挑战

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBeginWSPVPBattle 开始无双争霸战斗请求消息定义
type reqMsgBeginWSPVPBattle struct {
	Req
	Formation []int64 `codec:"formation"` // 己方阵容
}

// rspMsgBeginWSPVPBattle 开始无双争霸战斗回复消息定义
type rspMsgBeginWSPVPBattle struct {
	SyncResp
	OpponentBattleInfo      [][]byte `codec:"opp_bat_info"` // 返回对手每个武将的详细信息
	CurrDestinyGeneralSkill []int64  `codec:"dgss"`         // 神兽技能
}

// BeginWSPVPBattle 开始无双争霸战斗: 选好己方阵容，开始挑战
func (p *Account) BeginWSPVPBattle(r servers.Request) *servers.Response {
	req := new(reqMsgBeginWSPVPBattle)
	rsp := new(rspMsgBeginWSPVPBattle)

	initReqRsp(
		"Attr/BeginWSPVPBattleRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BeginWSPVPBattleHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// OpponentInfo 开始无双争霸战斗
type OpponentInfo struct {
	Idx            int64    `codec:"idx"`        // 武将数字ID
	Attr           []byte   `codec:"attr"`       // 武将属性
	Gs             int64    `codec:"gs"`         // 武将战力
	Skills         []int64  `codec:"skills"`     // 技能等级
	Fashions       []string `codec:"fashions"`   // 时装ID
	PassiveSkillId []string `codec:"p_skill_id"` // 被动技能ID
	CounterSkillId []string `codec:"c_skill_id"` // 被动技能ID
	TriggerSkillId []string `codec:"t_skill_id"` // 被动技能ID
	Star           int64    `codec:"star"`       // 武将星级
	HeroWing       int64    `codec:"hero_wing"`  // 翅膀ID
}

// EndWSPVPBattle : 结束无双争霸战斗
// 3场战斗结束一起发送, 返回内容走sync

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgEndWSPVPBattle 结束无双争霸战斗请求消息定义
type reqMsgEndWSPVPBattle struct {
	ReqWithAnticheat
	BattleResult bool `codec:"bat_res"` // 战斗结果，true表示己方胜,
}

// rspMsgEndWSPVPBattle 结束无双争霸战斗回复消息定义
type rspMsgEndWSPVPBattle struct {
	RespWithAnticheat
	ServerResult int64 `codec:"server_res"` // 战斗结果确认, 0, 表示服务器确认, 1 超时
	BattleResult bool  `codec:"bat_res"`    // 战斗结果，true表示己方胜,
}

// EndWSPVPBattle 结束无双争霸战斗: 3场战斗结束一起发送, 返回内容走sync
func (p *Account) EndWSPVPBattle(r servers.Request) *servers.Response {
	req := new(reqMsgEndWSPVPBattle)
	rsp := new(rspMsgEndWSPVPBattle)

	initReqRsp(
		"Attr/EndWSPVPBattleRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.EndWSPVPBattleHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// UnlockWSPVPBattle : 取消无双战斗的锁定
// 角色点击返回，需要发送解锁信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgUnlockWSPVPBattle 取消无双战斗的锁定请求消息定义
type reqMsgUnlockWSPVPBattle struct {
	Req
}

// rspMsgUnlockWSPVPBattle 取消无双战斗的锁定回复消息定义
type rspMsgUnlockWSPVPBattle struct {
	SyncResp
}

// UnlockWSPVPBattle 取消无双战斗的锁定: 角色点击返回，需要发送解锁信息
func (p *Account) UnlockWSPVPBattle(r servers.Request) *servers.Response {
	req := new(reqMsgUnlockWSPVPBattle)
	rsp := new(rspMsgUnlockWSPVPBattle)

	initReqRsp(
		"Attr/UnlockWSPVPBattleRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.UnlockWSPVPBattleHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetWSPVPBattleLog : 获取无双争霸的战斗日志
// 无双争霸的战斗日志

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWSPVPBattleLog 获取无双争霸的战斗日志请求消息定义
type reqMsgGetWSPVPBattleLog struct {
	Req
}

// rspMsgGetWSPVPBattleLog 获取无双争霸的战斗日志回复消息定义
type rspMsgGetWSPVPBattleLog struct {
	SyncResp
	WSPVPLog [][]byte `codec:"wspvp_log"` // 日志
}

// GetWSPVPBattleLog 获取无双争霸的战斗日志: 无双争霸的战斗日志
func (p *Account) GetWSPVPBattleLog(r servers.Request) *servers.Response {
	req := new(reqMsgGetWSPVPBattleLog)
	rsp := new(rspMsgGetWSPVPBattleLog)

	initReqRsp(
		"Attr/GetWSPVPBattleLogRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWSPVPBattleLogHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WSPVPLogInfo 获取无双争霸的战斗日志
type WSPVPLogInfo struct {
	Attack            bool   `codec:"att"`            // 是否是进攻方
	Result            bool   `codec:"result"`         // 己方是否胜利
	RankChange        int64  `codec:"rank_chg"`       // 排名变化值
	OpponentName      string `codec:"opp_name"`       // 对手名字
	OpponentGuildName string `codec:"opp_guild_name"` // 对手公会名字
	Time              int64  `codec:"time"`           // 挑战时间
}

// ClaimWSPVPReward : 领取无双争霸相关奖励
// 领取历史最高排名奖励和每小时累计的奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgClaimWSPVPReward 领取无双争霸相关奖励请求消息定义
type reqMsgClaimWSPVPReward struct {
	Req
	RewardType       int64 `codec:"reward_type"`  // 0=历史最高排名奖励 1=累计奖励 2=宝箱
	BestRankRewardId int64 `codec:"best_rank_id"` // 领取历史最高排名的奖励ID
}

// rspMsgClaimWSPVPReward 领取无双争霸相关奖励回复消息定义
type rspMsgClaimWSPVPReward struct {
	SyncRespWithRewards
}

// ClaimWSPVPReward 领取无双争霸相关奖励: 领取历史最高排名奖励和每小时累计的奖励
func (p *Account) ClaimWSPVPReward(r servers.Request) *servers.Response {
	req := new(reqMsgClaimWSPVPReward)
	rsp := new(rspMsgClaimWSPVPReward)

	initReqRsp(
		"Attr/ClaimWSPVPRewardRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ClaimWSPVPRewardHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetWSPVPPlayerInfo : 查看无双争霸排行榜上的人的信息
// 简短信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetWSPVPPlayerInfo 查看无双争霸排行榜上的人的信息请求消息定义
type reqMsgGetWSPVPPlayerInfo struct {
	Req
	AccountId string `codec:"acid"` // 排行榜人的ID
}

// rspMsgGetWSPVPPlayerInfo 查看无双争霸排行榜上的人的信息回复消息定义
type rspMsgGetWSPVPPlayerInfo struct {
	SyncResp
	RankInfo []byte `codec:"wspvp_info"` // 排行榜人的信息
}

// GetWSPVPPlayerInfo 查看无双争霸排行榜上的人的信息: 简短信息
func (p *Account) GetWSPVPPlayerInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetWSPVPPlayerInfo)
	rsp := new(rspMsgGetWSPVPPlayerInfo)

	initReqRsp(
		"Attr/GetWSPVPPlayerInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetWSPVPPlayerInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WSPVPRankInfo 查看无双争霸排行榜上的人的信息
type WSPVPRankInfo struct {
	Sid               string  `codec:"sid"`                // 服务器ID
	Name              string  `codec:"name"`               // 个人名字
	VipLevel          int64   `codec:"vip_lv"`             // VIP等级
	CorpLevel         int64   `codec:"corp_lv"`            // 战队等级
	AllGs             int64   `codec:"all_gs"`             // 总战力
	GuildName         string  `codec:"guild_name"`         // 公会名字
	BestHeroIdx       []int64 `codec:"best_hero_id"`       // 最强阵容武将ID
	BestHeroLevel     []int64 `codec:"best_hero_lv"`       // 最强阵容武将等级
	BestHeroStarLevel []int64 `codec:"best_hero_star_lv"`  // 最强阵容武将星级
	BestHeroBaseGs    []int64 `codec:"best_hero_base_gs"`  // 最强阵容战力
	BestHeroExtraGs   []int64 `codec:"best_hero_extra_gs"` // 最强阵容额外战力
	EquipAttr         []int64 `codec:"equip_attr"`         // 装备附加攻防血
	DestinyAttr       []int64 `codec:"destiny_attr"`       // 神兽附加攻防血
	JadeAttr          []int64 `codec:"jade_attr"`          // 宝石附加攻防血
}
