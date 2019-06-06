package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GuildWorshipAvatarInfo : 军团膜拜信息
// 获得当天膜拜成员信息

// reqMsgGuildWorshipAvatarInfo 军团膜拜信息请求消息定义
type reqMsgGuildWorshipAvatarInfo struct {
	Req
}

// rspMsgGuildWorshipAvatarInfo 军团膜拜信息回复消息定义
type rspMsgGuildWorshipAvatarInfo struct {
	SyncResp
	GuildWorshipAvatarInfo [][]byte `codec:"gws_a_i"` // 成员信息
}

// GuildWorshipAvatarInfo 军团膜拜信息: 获得当天膜拜成员信息
func (p *Account) GuildWorshipAvatarInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGuildWorshipAvatarInfo)
	rsp := new(rspMsgGuildWorshipAvatarInfo)

	initReqRsp(
		"Attr/GuildWorshipAvatarInfoRsp",
		r.RawBytes,
		req, rsp, p)

	warnCode := p.GuildWorshipAvatarInfoHandler(req, rsp)

	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WorshipAvatarInfo 军团膜拜信息
type WorshipAvatarInfo struct {
	PlayerNames  string    `codec:"py_name"` // 玩家姓名
	PlayerGs     int64     `codec:"py_gs"`   // 战力
	HeroId       int64     `codec:"h_id"`    // 主将ID
	HeroSwing    int64     `codec:"h_s"`     // 翅膀ID
	HeroMagicPet int64     `codec:"h_m"`     // 灵宠形象
	HeroFashion  [2]string `codec:"H_f"`     // 时装
}

// GuildWorshipAvatarInfo : 军团膜拜信息
// 获得当天膜拜成员信息
func (p *Account) GuildWorshipAvatarInfoHandler(req *reqMsgGuildWorshipAvatarInfo, resp *rspMsgGuildWorshipAvatarInfo) uint32 {
	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return errCode.ClickTooQuickly
	}

	res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if code.HasError() {
		logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
			code, p.GuildProfile.GuildUUID)
		return errCode.ClickTooQuickly
	}
	eq := make([]WorshipAvatarInfo, len(res.GuildInfoBase.GuildWorship.WorshipMember))
	for i, info := range res.GuildInfoBase.GuildWorship.WorshipMember {
		eq[i].HeroFashion = info.HeroFashion
		eq[i].HeroId = int64(info.HeroId)
		eq[i].HeroSwing = int64(info.HeroSwing)
		eq[i].HeroMagicPet = int64(info.MagicPetfigure)
		eq[i].PlayerGs = int64(info.MemberGs)
		eq[i].PlayerNames = info.MemberNames

	}
	for _, x := range eq {
		resp.GuildWorshipAvatarInfo = append(resp.GuildWorshipAvatarInfo, encode(x))
	}

	return 0
}

// WorshipPlayer : 膜拜的玩家ID
// 膜拜的玩家ID

const (
	Multipling       = 1.5
	DoubleMultipling = 2
	GoodSign         = 2
)

// reqMsgWorshipPlayer 膜拜的玩家ID请求消息定义
type reqMsgWorshipPlayer struct {
	Req
	WorshipPlayer int64 `codec:"ws_p"` // 膜拜的玩家ID 从 1 开始
	WorshipNum    int64 `codec:"ws_n"` // 膜拜的次数
}

// rspMsgWorshipPlayer 膜拜的玩家ID回复消息定义
type rspMsgWorshipPlayer struct {
	SyncRespWithRewards
	RewardTyp int64 `codec:"rewardtyp"` // 回复的暴击奖励类型,0普通，1暴击
	SignTyp   int64 `codec:"signtyp"`   // 回复的抽签类型
}

// WorshipPlayer 膜拜的玩家ID: 膜拜的玩家ID
func (p *Account) WorshipPlayer(r servers.Request) *servers.Response {
	req := new(reqMsgWorshipPlayer)
	rsp := new(rspMsgWorshipPlayer)

	initReqRsp(
		"Attr/WorshipPlayerRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_Err
		COED_Time_Not_Enough
	)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	p.GuildProfile.WorshipInfo.CheckDailyReset(p.GetProfileNowTime())
	gg := &p.GuildProfile.WorshipInfo

	if gamedata.GetVIPCfg(int(p.Profile.GetVipLevel())).GuildWorshipLimit <= uint32(gg.PersionTakeNum) {
		return rpcErrorWithMsg(rsp, COED_Time_Not_Enough, "COED_Time_Not_Enough")
	}
	if gamedata.GetGuildWorshipRewardCost(req.WorshipNum) != 0 {
		data1 := &gamedata.CostData{}
		data1.AddItem(gamedata.VI_Hc, uint32(gamedata.GetGuildWorshipRewardCost(req.WorshipNum)))
		if !account.CostBySync(p.Account, data1, rsp, "GuildWorship Cost") {
			return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
		}
	}
	res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if code.HasError() {
		logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
			code, p.GuildProfile.GuildUUID)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if int(req.WorshipPlayer)-1 >= len(res.GuildInfoBase.GuildWorship.WorshipMember) {
		logs.SentryLogicCritical(p.AccountID.String(), "WorshipPlayer WorshipPlayer Err %v %d %d",
			p.GuildProfile.GuildUUID, req.WorshipPlayer-1, len(res.GuildInfoBase.GuildWorship.WorshipMember))
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	if gg.TakeId == 0 {
		info := res.GuildInfoBase.GuildWorship.WorshipMember[req.WorshipPlayer-1]
		gg.SetWorshipInfo(req.WorshipPlayer)
		gg.TakeSign = info.MemberSign

		//发放奖励
		rewardData := &gamedata.CostData{}
		data := gamedata.GetguildWorshipReward(int(info.MemberSign))
		multipling, isCrit := getCrit(req.WorshipNum)
		reward := make([]string, len(data))
		oneTime := make([]int64, len(data))
		doubleTime := make([]int64, len(data))
		for i, x := range data {
			if x.GetWorshipBaseReward() != "" {
				reward[i] = x.GetWorshipBaseReward()
				oneTime[i] = int64(float64(x.GetBaseRewardNum()) * multipling * Multipling)
				doubleTime[i] = int64(float64(x.GetBaseRewardNum()) * multipling * float64(DoubleMultipling))
				rewardData.AddItem(x.GetWorshipBaseReward(), uint32(float64(x.GetBaseRewardNum())*multipling))
			}
		}

		gg.Reward = reward
		gg.OneTime = oneTime
		gg.DoubleTime = doubleTime
		gg.WorshipAccoundID = info.MemberAccountId

		rsp.RewardTyp = isCrit
		rsp.SignTyp = gg.TakeSign

		guild.GetModule(p.AccountID.ShardId).AddWorshipLogInfo(p.GuildProfile.GuildUUID, p.Profile.Name,
			gg.TakeSign, p.GetProfileNowTime())

		if !account.GiveBySync(p.Account, rewardData, rsp, "guildWorshipGive") {
			logs.Error("guildWorshipGive GiveBySync Err")
		}

		if gg.TakeSign == GoodSign {
			memberDates := make([]string, res.Base.MemNum) // 军团所有成员
			for i, member := range res.Members[:res.Base.MemNum] {
				memberDates[i] = member.AccountID
			}

			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_GUILD_WORSHIPCRIT_MARQUEE).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", gg.TakeId)).
				AddSids(memberDates).SendGuild()
		}
		logiclog.LogGuildWorship(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			p.Profile.GetVipLevel(),
			gg.WorshipAccoundID,
			info.CorpLv,
			info.Vip,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")
	} else {
		info := res.GuildInfoBase.GuildWorship.WorshipMember[req.WorshipPlayer-1]

		if info.MemberAccountId != gg.WorshipAccoundID {
			return rpcWarn(rsp, errCode.ChangeGuildInSameDay)
		}

		gg.SetWorshipInfo(req.WorshipPlayer)
		rewardData := &gamedata.CostData{}

		if gamedata.IsGuildWorshipCrit(req.WorshipNum) {
			for i, x := range gg.DoubleTime {
				rewardData.AddItem(gg.Reward[i], uint32(x))
				gg.DoubleTime[i] = x * int64(DoubleMultipling)
				gg.OneTime[i] = int64(float64(x) * Multipling)
			}
			rsp.RewardTyp = 1

		} else {
			for i, x := range gg.OneTime {
				rewardData.AddItem(gg.Reward[i], uint32(x))
				gg.DoubleTime[i] = x * int64(DoubleMultipling)
				gg.OneTime[i] = int64(float64(x) * Multipling)
			}
			rsp.RewardTyp = 0

		}

		rsp.SignTyp = gg.TakeSign
		if !account.GiveBySync(p.Account, rewardData, rsp, "guildWorshipGive") {
			logs.Error("guildWorshipGive GiveBySync Err")
		}
		logiclog.LogGuildWorship(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			p.Profile.GetVipLevel(),
			gg.WorshipAccoundID,
			info.CorpLv,
			info.Vip,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")

	}

	guild.GetModule(p.AccountID.ShardId).AddWorshipIndex(p.GuildProfile.GuildUUID, req.WorshipPlayer-1)
	p.updateCondition(account.COND_TYP_GUILD_WORSHIP,
		1, 0, "", "", rsp)

	rsp.OnChangeGuildWorshipInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func getCrit(num int64) (float64, int64) {
	if gamedata.IsGuildWorshipCrit(num) {
		return float64(DoubleMultipling), 1
	} else {
		return Multipling, 0
	}
}

// WorshipBox : 膜拜的宝箱
// 宝箱

// reqMsgWorshipBox 膜拜的宝箱请求消息定义
type reqMsgWorshipBox struct {
	Req
	WorshipBoxID int64 `codec:"wsb_id"` // 膜拜的宝箱ID
}

// rspMsgWorshipBox 膜拜的宝箱回复消息定义
type rspMsgWorshipBox struct {
	SyncRespWithRewards
}

// WorshipBox 膜拜的宝箱: 宝箱
func (p *Account) WorshipBox(r servers.Request) *servers.Response {
	req := new(reqMsgWorshipBox)
	rsp := new(rspMsgWorshipBox)

	initReqRsp(
		"Attr/WorshipBoxRsp",
		r.RawBytes,
		req, rsp, p)

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(rsp, warnCode)
	}

	p.GuildProfile.WorshipInfo.CheckDailyReset(p.GetProfileNowTime())
	boxReward := gamedata.GetguildWorshipBoxReward(int(req.WorshipBoxID))
	rewardData := &gamedata.CostData{}

	for _, x := range boxReward {
		rewardData.AddItem(x.GetBoxRewardID(), x.GetBoxRewardNum())
	}

	if !account.GiveBySync(p.Account, rewardData, rsp, "guildWorshipBoxGive") {
		logs.Error("guildWorshipBoxGive GiveBySync Err")
	}

	p.GuildProfile.WorshipInfo.UpdateWorshipHasReward(req.WorshipBoxID)
	rsp.OnChangeGuildWorshipInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// WorshipLog : 膜拜的log
// 回该军团今日logList（人名+什么签）

// reqMsgWorshipLog 膜拜的log请求消息定义
type reqMsgWorshipLog struct {
	Req
}

// rspMsgWorshipLog 膜拜的log回复消息定义
type rspMsgWorshipLog struct {
	SyncResp
	ResAccountID []string `codec:"resaccountid"` // 膜拜玩家人名
	WorshipSign  []int64  `codec:"worshipsign"`  // 膜拜到的什么签
	Worshiptime  []int64  `codec:"time"`
}

// WorshipLog 膜拜的log: 回该军团今日logList（人名+什么签）
func (p *Account) WorshipLog(r servers.Request) *servers.Response {
	req := new(reqMsgWorshipLog)
	rsp := new(rspMsgWorshipLog)

	initReqRsp(
		"Attr/WorshipLogRsp",
		r.RawBytes,
		req, rsp, p)

	warnCode := p.CheckGuildStatus(true)
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, warnCode)
	}

	p.GuildProfile.WorshipInfo.CheckDailyReset(p.GetProfileNowTime())
	res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if code.HasError() {
		logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
			code, p.GuildProfile.GuildUUID)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}
	var resAccountID []string
	var worshipSign []int64
	var worshipTime []int64

	for _, x := range res.GuildInfoBase.GuildWorship.WorshipLog {
		resAccountID = append(resAccountID, x.PlayerNames)
		worshipSign = append(worshipSign, x.SignId)
		worshipTime = append(worshipTime, x.Worshiptiem)
	}
	for a := len(res.GuildInfoBase.GuildWorship.WorshipLog) - 1; a >= 0; a-- {
		rsp.ResAccountID = append(rsp.ResAccountID, resAccountID[a])
		rsp.WorshipSign = append(rsp.WorshipSign, worshipSign[a])
		rsp.Worshiptime = append(rsp.Worshiptime, worshipTime[a])
	}

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
