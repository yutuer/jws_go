package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/clienttag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/gates_enemy"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestCreateGuild struct {
	Req
	Name         string `codec:"na"`
	Icon         string `codec:"icon"`
	ApplyGsLimit int    `codec:"aply_gs"`
	ApplyAuto    bool   `codec:"aply_auto"`
}

type ResponseCreateGuild struct {
	SyncResp
}

func (p *Account) CreateGuildRequest(r servers.Request) *servers.Response {
	req := &RequestCreateGuild{}
	resp := &ResponseCreateGuild{}
	initReqRsp(
		"PlayerGuild/CreateGuildResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Corp_Lvl
		Err_Vip
		Err_Hc
		Err_LeaveGuild_CD
	)
	// 等级检查
	lvl, _ := p.Profile.GetCorp().GetXpInfo()
	if lvl < gamedata.GetCommonCfg().GetPresidentLevelLimit() {
		return rpcError(resp, uint32(Err_Corp_Lvl))
	}
	// vip
	vip, _ := p.Profile.Vip.GetVIP()
	if vip < gamedata.GetCommonCfg().GetPresidentVIPLimit() {
		return rpcError(resp, uint32(Err_Vip))
	}
	// 是否在公会里
	if p.GuildProfile.InGuild() {
		return rpcWarn(resp, errCode.GuildAlreadyInGuild)
	}
	// check hc and sc
	if p.Profile.GetSC().GetSC(helper.SC_Money) < int64(gamedata.GetCommonCfg().GetGuildFondSC()) ||
		p.Profile.GetHC().GetHC() < int64(gamedata.GetCommonCfg().GetGuildFondHC()) {
		return rpcError(resp, uint32(Err_Hc))
	}
	// 离帮cd检查
	if p.Profile.GetProfileNowTime() < p.GuildProfile.NextEnterGuildTime {
		logs.Warn("%s CreateGuildRequest Err_LeaveGuild_CD", p.AccountID.String())
		return rpcWarn(resp, errCode.ClickTooQuickly)
		//return rpcError(resp, uint32(Err_LeaveGuild_CD))
	}
	// new guild
	memInfo := p.Account.GetSimpleInfo()
	memInfo.GuildPosition = gamedata.Guild_Pos_Chief
	gUuid, _, errRet := guild.GetModule(p.AccountID.ShardId).NewGuild(guild_info.GuildSimpleInfo{
		Name:         req.Name,
		Icon:         req.Icon,
		Level:        1,
		ApplyGsLimit: req.ApplyGsLimit,
		ApplyAuto:    req.ApplyAuto,
		LeaderAcid:   memInfo.AccountID,
		LeaderName:   memInfo.Name,
	}, memInfo, p.Profile.ChannelId, p.Account.GuildProfile.GuildAssignInfo.AssignID,
		p.Account.GuildProfile.GuildAssignInfo.AssignTimes, p.Account.GuildProfile.LastLeaveGuildTime)

	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	// cost hc and sc
	c := &account.CostGroup{}
	c.AddHc(p.Account, int64(gamedata.GetCommonCfg().GetGuildFondHC()))
	c.AddSc(p.Account, helper.SC_Money, int64(gamedata.GetCommonCfg().GetGuildFondSC()))
	c.CostBySync(p.Account, resp, "CreateGuild")

	// 成功
	p.GuildProfile.GuildUUID = gUuid
	p.GuildProfile.GuildPosition = gamedata.Guild_Pos_Chief
	p.Profile.GetClientTagInfo().SetTag(clienttag.Tag_GuildIn, 1)
	p.Profile.GetGatesEnemy().CleanDatas()

	csrob.GetModule(p.AccountID.ShardId).PlayerMod.JoinGuild(p.AccountID.String(), p.GuildProfile.GuildUUID)

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.OnChangePlayerGuildApply()
	resp.OnChangeGuildScience()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type RequestGetRandomGuildList struct {
	Req
}

type ResponseGetRandomGuildList struct {
	SyncResp
	GuildList [][]byte `codec:"gl"`
}

func (p *Account) GetRandomGuildListRequest(r servers.Request) *servers.Response {
	req := &RequestGetRandomGuildList{}
	resp := &ResponseGetRandomGuildList{}
	initReqRsp(
		"PlayerGuild/GetGuildListResp",
		r.RawBytes,
		req, resp, p)

	// 是否在公会里
	if p.GuildProfile.InGuild() {
		return rpcWarn(resp, errCode.GuildAlreadyInGuild)
	}

	// 回复协议
	guilds := guild.GetModule(p.AccountID.ShardId).
		GetRandGuild(p.AccountID.String())
	resp.GuildList = make([][]byte, len(guilds))
	for i, g := range guilds {
		resp.GuildList[i] = encode(guildBasicInfoToClient{
			GuildUUID:       g.GuildUUID,
			GuildID:         guild.GuildItoa(g.GuildID),
			Name:            g.Name,
			Level:           g.Level,
			ApplyGsLimit:    g.ApplyGsLimit,
			ApplyAuto:       g.ApplyAuto,
			Icon:            g.Icon,
			Exp:             g.XpCurr,
			NextExp:         gamedata.GetGuildXpNeedNext(g.Level),
			Notice:          g.Notice,
			GuildGSSum:      g.GuildGSSum,
			MemNum:          uint32(g.MemNum),
			MaxMem:          uint32(g.MaxMemNum),
			GatesEnemyCount: g.GetGateEnemyCount(),
			RenameTimes:     g.RenameTimes,
		})
	}
	resp.OnChangePlayerGuild()
	resp.OnChangePlayerGuildApply()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type RequestFindGuild struct {
	Req
	GuildId string `codec:"id"`
}

type ResponseFindGuild struct {
	SyncResp
	FoundGuildInfo []byte   `codec:"gi"` // guildBasicInfoToClient
	FoundGuildMem  [][]byte `codec:"gm"` // guildBasicInfoToClient
}

func (p *Account) FindGuildRequest(r servers.Request) *servers.Response {
	req := &RequestFindGuild{}
	resp := &ResponseFindGuild{}
	initReqRsp(
		"PlayerGuild/FindGuildResp",
		r.RawBytes,
		req, resp, p)

	guildIndex := guild.GuildIdAtoi(req.GuildId)
	if guildIndex <= 0 {
		return rpcWarn(resp, errCode.GuildIndexIllegal)
	}
	guildInfo, errRet := guild.GetModule(p.AccountID.ShardId).
		FindGuild(p.AccountID.String(), guildIndex)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	g := guildInfo.GuildInfoBase.Base
	pos, _ := rank.GetModule(p.AccountID.ShardId).
		RankGuildGs.GetPos(g.GuildUUID)

	data := guildBasicInfoToClient{
		GuildUUID:       g.GuildUUID,
		GuildID:         guild.GuildItoa(g.GuildID),
		Name:            g.Name,
		Level:           g.Level,
		ApplyGsLimit:    g.ApplyGsLimit,
		ApplyAuto:       g.ApplyAuto,
		Icon:            g.Icon,
		Rank:            pos,
		Exp:             g.XpCurr,
		NextExp:         gamedata.GetGuildXpNeedNext(g.Level),
		Notice:          g.Notice,
		GuildGSSum:      g.GuildGSSum,
		MemNum:          uint32(g.MemNum),
		MaxMem:          uint32(g.MaxMemNum),
		GatesEnemyCount: g.GetGateEnemyCount(),
		RenameTimes:     g.RenameTimes,
	}
	resp.FoundGuildInfo = encode(data)

	resp.FoundGuildMem = make([][]byte, 0, g.MemNum)
	for i := 0; i < g.MemNum; i++ {
		member := guildInfo.Members[i]
		guildBossDamage := int(guildInfo.ActBoss.LastDayDamages.GetSorce(member.AccountID))
		m := guildMemToClient{
			Acid:                    member.AccountID,
			Name:                    member.Name,
			Level:                   member.CorpLv,
			Gs:                      member.CurrCorpGs,
			Vip:                     member.Vip,
			CurrAvatar:              member.CurrAvatar,
			Position:                member.GuildPosition,
			LastLoginTime:           member.LastLoginTime,
			IsGettedGateEnemyReward: guildInfo.GatesEnemyData.GetGetRewardTime(member.AccountID) > 0,
			GuildContribution:       member.Contribution[0],
			GuildSp:                 member.GuildSp,
			Online:                  member.GetOnline(),
			GuildBossDamage:         guildBossDamage,
		}
		resp.FoundGuildMem = append(resp.FoundGuildMem, encode(m))
	}
	return rpcSuccess(resp)
}

func (p *Account) GetGuildInfoRequest(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/GetGuildInfoResp",
		r.RawBytes,
		req, resp, p)

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.OnChangeClientTagInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) SetGuildApplySettingRequest(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ApplyGsLimit int  `codec:"aply_gs"`
		ApplyAuto    bool `codec:"aply_auto"`
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/SetGuildApplySettingResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	// 职位检查
	if !gamedata.CheckApprovePosition(p.GuildProfile.GuildPosition) {
		return rpcWarn(resp, errCode.GuildPositionErr)
	}

	guildUuid := p.GuildProfile.GuildUUID
	errRet := guild.GetModule(p.AccountID.ShardId).SetGuildApplySetting(guildUuid,
		p.AccountID.String(), req.ApplyGsLimit, req.ApplyAuto)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type RequestApplyGuild struct {
	Req
	GuildUUid string `json:"guuid"`
}

type ResponseApplyGuild struct {
	SyncResp
	IsSuccess bool `json:"is_su"`
}

func (p *Account) ApplyGuildRequest(r servers.Request) *servers.Response {
	req := &RequestApplyGuild{}
	resp := &ResponseApplyGuild{}
	initReqRsp(
		"PlayerGuild/ApplyGuildResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_LeaveGuild_CD
	)

	// 自己是否已经在公会中
	if p.GuildProfile.InGuild() {
		return rpcWarn(resp, errCode.GuildHasIn)
	}
	// 离帮cd检查
	if p.Profile.GetProfileNowTime() < p.GuildProfile.NextEnterGuildTime {
		logs.Warn("%s ApplyGuildRequest Err_LeaveGuild_CD", p.AccountID.String())
		return rpcWarn(resp, errCode.ClickTooQuickly)
		//return rpcError(resp, uint32(Err_LeaveGuild_CD))
	}
	// 公会申请
	logs.Debug("AssignInfo: %v", p.Account.GuildProfile.GuildAssignInfo)
	errRet, isSyncGuld := guild.GetModule(p.AccountID.ShardId).ApplyGuild(req.GuildUUid, p.Account.GetSimpleInfo(),
		p.Account.GuildProfile.GuildAssignInfo.AssignID, p.Account.GuildProfile.GuildAssignInfo.AssignTimes,
		p.Account.GuildProfile.LastLeaveGuildTime)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	if isSyncGuld {
		resp.OnChangeGuildInfo()
		resp.OnChangeGuildMemsInfo()
		resp.OnChangeGuildApplyMemsInfo()
		csrob.GetModule(p.AccountID.ShardId).PlayerMod.JoinGuild(p.AccountID.String(), req.GuildUUid)
		p.updateCSRobPlayerRank(true)
	} else {
		resp.IsSuccess = true
	}

	resp.OnChangePlayerGuild()
	resp.OnChangePlayerGuildApply()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) CancelApplyGuildRequest(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GuildUUid string `json:"guuid"`
	}{}
	resp := &struct {
		SyncResp
		IsSuccess bool `json:"is_su"`
	}{}
	initReqRsp(
		"PlayerGuild/CancelApplyGuildResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_LeaveGuild_CD
	)

	// 自己是否已经在公会中
	if p.GuildProfile.InGuild() {
		return rpcWarn(resp, errCode.GuildHasIn)
	}
	// 公会申请
	errRet := guild.GetModule(p.AccountID.ShardId).CancelApplyGuild(req.GuildUUid, p.Account.AccountID.String())
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	resp.IsSuccess = true

	resp.OnChangePlayerGuild()
	resp.OnChangePlayerGuildApply()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type RequestApproveGuildApplicant struct {
	Req
	Applicant string `json:"acid"`
	Oper      int    `json:"oper"` // 0: 拒绝，1：同意
}

type ResponseApproveGuildApplicant struct {
	SyncResp
}

func (p *Account) ApproveGuildApplicant(r servers.Request) *servers.Response {
	req := &RequestApproveGuildApplicant{}
	resp := &ResponseApproveGuildApplicant{}
	initReqRsp(
		"PlayerGuild/ApproveGuildApplicantResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否已经不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, errCode.GuildPlayerNotFound)
	}
	guildUuid := p.GuildProfile.GuildUUID
	// oper
	if req.Oper == 0 { // 拒绝
		errRet := guild.GetModule(p.AccountID.ShardId).ApplyOper(guildUuid,
			p.AccountID.String(), req.Applicant, false, p.Profile.ChannelId)
		if rsp := guildErrRet(errRet, resp); rsp != nil {
			return rsp
		}
	} else { // 同意
		errRet := guild.GetModule(p.AccountID.ShardId).ApplyOper(guildUuid,
			p.AccountID.String(), req.Applicant, true, p.Profile.ChannelId)
		if rsp := guildErrRet(errRet, resp); rsp != nil {
			return rsp
		}
		csrob.GetModule(p.AccountID.ShardId).PlayerMod.JoinGuild(req.Applicant, guildUuid)
		//通知刷新劫营夺粮里面的榜
		player_msg.Send(req.Applicant, player_msg.PlayerMsgCSRobAddPlayerRank,
			player_msg.PlayerCSRobAddPlayerRank{})
	}

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.OnChangeGuildApplyMemsInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) GetPlayerGuildApplyList(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/GetPlayerGuildApplyListResp",
		r.RawBytes,
		req, resp, p)

	resp.OnChangePlayerGuild()
	resp.OnChangePlayerGuildApply()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

type RequestGetGuildApplyList struct {
	Req
}

type ResponseGetGuildApplyList struct {
	SyncResp
}

func (p *Account) GetGuildApplyList(r servers.Request) *servers.Response {
	req := &RequestGetGuildApplyList{}
	resp := &ResponseGetGuildApplyList{}
	initReqRsp(
		"PlayerGuild/GetGuildApplyListResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildApplyMemsInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 自行退出公会
func (p *Account) GuildQuit(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/GuildQuitResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}
	oldGuildID := p.GuildProfile.GuildUUID
	// 离开
	errRet, isDismiss := guild.GetModule(p.AccountID.ShardId).QuitGuild(
		p.GuildProfile.GuildUUID,
		p.AccountID.String(),
		p.Profile.ChannelId)

	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}
	if isDismiss {
		dismissGuildDo(p)
		resp.OnChangeGuildApplyMemsInfo()

	}
	// player guild
	now_t := p.Profile.GetProfileNowTime()
	p.GuildProfile.GuildUUID = ""
	p.GuildProfile.GuildPosition = 0
	p.GuildProfile.NextEnterGuildTime =
		util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(now_t), now_t)

	p.Profile.GetGatesEnemy().CleanDatas()

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.mkInfo(p)

	//通知劫营夺粮模块
	csrob.GetModule(p.AccountID.ShardId).PlayerMod.LeaveGuild(p.AccountID.String())
	csrob.GetModule(p.AccountID.ShardId).PlayerRanker.RemovePlayerRank(p.AccountID.String())
	if isDismiss {
		csrob.GetModule(p.AccountID.ShardId).GuildMod.Dismiss(oldGuildID)
	}
	return rpcSuccess(resp)
}

// 踢出成员
func (p *Account) GuildKick(r servers.Request) *servers.Response {
	req := &struct {
		Req
		KickMemAcid string `json:"ka"`
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/GuildKickResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ uint32 = iota
		Err_Position
		Err_Param
	)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	// 权限检查
	posCfg := gamedata.GetGuildPosData(p.GuildProfile.GuildPosition)
	if posCfg.GetKickMember() == 0 {
		return rpcWarn(resp, uint32(errCode.GuildPositionErr))
	}

	if req.KickMemAcid == "" {
		return rpcError(resp, Err_Param)
	}

	guildUuid := p.GuildProfile.GuildUUID
	errRet := guild.GetModule(p.AccountID.ShardId).KickMember(guildUuid,
		p.AccountID.String(), req.KickMemAcid, p.Profile.ChannelId)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.mkInfo(p)

	//通知劫营夺粮模块
	csrob.GetModule(p.AccountID.ShardId).PlayerMod.LeaveGuild(req.KickMemAcid)
	csrob.GetModule(p.AccountID.ShardId).PlayerRanker.RemovePlayerRank(req.KickMemAcid)
	return rpcSuccess(resp)
}

// 职位任命
func (p *Account) GuildPositionAppoint(r servers.Request) *servers.Response {
	req := &struct {
		Req
		MemAcid  string `json:"ma"`
		Position int    `json:"po"`
	}{}

	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerGuild/GuildPosAppointResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ uint32 = iota
		Err_Param
	)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	if req.MemAcid == "" {
		return rpcError(resp, Err_Param)
	}

	guildUuid := p.GuildProfile.GuildUUID
	errRet := guild.GetModule(p.AccountID.ShardId).ChangeMemberPosition(guildUuid,
		p.AccountID.String(), req.MemAcid, req.Position, p.Profile.ChannelId)

	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	csrob.GetModule(p.AccountID.ShardId).PlayerMod.ChangeGuildPos(req.MemAcid, p.GuildProfile.GuildUUID, req.Position)
	if gamedata.Guild_Pos_Chief == req.Position {
		csrob.GetModule(p.AccountID.ShardId).PlayerMod.ChangeGuildPos(p.AccountID.String(), p.GuildProfile.GuildUUID, gamedata.Guild_Pos_Mem)
		csrob.GetModule(p.AccountID.ShardId).GuildMod.MasterChange(p.GuildProfile.GuildUUID)
	}

	resp.OnChangeGuildMemsInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 公会解散
func (p *Account) GuildDismiss(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/GuildDismissResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	// 职位检查
	if p.GuildProfile.GuildPosition != gamedata.Guild_Pos_Chief {
		return rpcWarn(resp, errCode.GuildPositionErr)
	}

	// 解散
	errRet := guild.GetModule(p.AccountID.ShardId).DismissGuild(p.GuildProfile.GuildUUID,
		p.AccountID.String(), p.Profile.ChannelId)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}

	dismissGuildDo(p)

	p.GuildProfile.GuildUUID = ""
	p.GuildProfile.GuildPosition = 0
	p.Profile.GetGatesEnemy().CleanDatas()

	//通知其他模块
	csrob.GetModule(p.AccountID.ShardId).GuildMod.Dismiss(p.GuildProfile.GuildUUID)

	resp.OnChangePlayerGuild()
	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 变更公会日志
func (p *Account) ChangeGuildNotice(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Notice string `json:"notc"`
	}{}
	resp := &struct {
		SyncResp
	}{}
	initReqRsp(
		"PlayerGuild/ChangeGuildNoticeResp",
		r.RawBytes,
		req, resp, p)

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(resp, uint32(errCode.GuildPlayerNotIn))
	}

	// 职位检查
	posCfg := gamedata.GetGuildPosData(p.GuildProfile.GuildPosition)
	if posCfg.GetRenNewsPower() == 0 {
		return rpcWarn(resp, uint32(errCode.GuildPositionErr))
	}

	// 字数检查 TODO
	// 敏感词检查 TODO

	errRet := guild.GetModule(p.AccountID.ShardId).ChangeGuildNotice(p.GuildProfile.GuildUUID, p.AccountID.String(), req.Notice)
	if rsp := guildErrRet(errRet, resp); rsp != nil {
		return rsp
	}
	resp.OnChangeGuildInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 公会祈福
func (p *Account) GuildSign(r servers.Request) *servers.Response {
	req := &struct {
		Req
		SignID int `codec:"id"`
	}{}
	resp := &struct {
		SyncRespWithRewards
		GuildXpAdd int `codec:"gxpadd"`
	}{}

	initReqRsp(
		"PlayerGuild/GuildSignResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ uint32 = iota
		ErrNoData
		ErrNoCost
		ErrNoCount
	)

	vipLv := int(p.Profile.GetVipLevel())
	nowT := p.Profile.GetProfileNowTime()
	acID := p.AccountID.String()

	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return rpcWarn(resp, warnCode)
	}

	guildData, ret := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if ret.HasError() || guildData == nil {
		return rpcWarn(resp, errCode.GuildNotFound)
	}

	ok, cost, give, xp := gamedata.GetGuildSignNewInfo(req.SignID - 1)
	if !ok {
		return rpcError(resp, ErrNoData)
	}

	c := p.GuildProfile.GetGuildSignCount(vipLv, nowT)
	if c <= 0 {
		logs.Warn("GuildSign GetGuildSignCount %d", c)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	ok = account.CostBySync(p.Account, &cost.Cost, resp, "GuildSign")
	if !ok {
		logs.Warn("GuildSign ErrNoCost")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	p.GuildProfile.UseGuildSignCount(vipLv, nowT)

	guild.GetModule(p.AccountID.ShardId).Sign(p.GuildProfile.GuildUUID, acID, int64(xp), nowT)

	if !account.GiveBySync(p.Account, &give.Cost, resp, "GuildSign") {
		logs.SentryLogicCritical(acID, "GiveBySync Err by GuildSign %d", req.SignID-1)
	}
	logs.Trace("guild add xp %d", int(xp))
	resp.GuildXpAdd = int(xp)

	p.updateCondition(account.COND_TYP_GuildSign,
		1, 0, "", "", resp)
	p.updateCondition(account.COND_TYP_GUILD_COLLECTION,
		1, 0, "", "", resp)

	resp.OnChangeGuildInfo()
	resp.OnChangeGuildMemsInfo()
	resp.OnChangePlayerGuild()
	resp.OnChangeGuildApplyMemsInfo()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) GetGuildLog(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Ts        []int64  `codec:"tss"`
		Ids       []int64  `codec:"ids"`
		ParamNums []int64  `codec:"pns"`
		Params    []string `codec:"pms"`
	}{}

	initReqRsp(
		"PlayerGuild/GetGuildLogResp",
		r.RawBytes,
		req, resp, p)

	if p.GuildProfile.GuildUUID != "" {
		ret, ts, ids, pns, pms := guild.GetModule(p.AccountID.ShardId).
			GetGuildLog(p.GuildProfile.GuildUUID, p.AccountID.String())
		if rsp := guildErrRet(ret, resp); rsp != nil {
			return rsp
		}
		resp.Ts = ts
		resp.Ids = ids
		resp.ParamNums = pns
		resp.Params = pms
	}
	return rpcSuccess(resp)
}

func guildErrRet(ret guild.GuildRet, r respInterface) *servers.Response {
	if !ret.HasError() {
		return nil
	}
	if ret.CodeLevel == guild.Code_Warn {
		return rpcWarn(r, uint32(ret.ErrCode))
	} else if ret.CodeLevel == guild.Code_Err {
		return rpcErrorWithMsg(r, uint32(40+ret.ErrCode), ret.ErrMsg)
	}
	return nil
}

func dismissGuildDo(p *Account) {
	gates_enemy.GetModule(p.AccountID.ShardId).StopGatesEnemyAct(p.GuildProfile.GuildUUID)
	rank.GetModule(p.AccountID.ShardId).RankGuildGs.Del(p.GuildProfile.GuildUUID)
	rank.GetModule(p.AccountID.ShardId).RankGuildGateEnemy.Del(p.GuildProfile.GuildUUID)
	gvg.GetModule(p.AccountID.ShardId).CommandExecAsync(gvg.GVGCmd{
		Typ:  gvg.Cmd_Typ_Remove_Guild,
		GuID: p.GuildProfile.GuildUUID,
	})
}
