package guild

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	GuildModule是公会总对外接口
	1、ApplyWorker是一个goroutine，处理未进入公会前的操作，如：创建， 申请相关
	2、GuildMgrWorker是一个goroutine，管理所有已有公会的goroutine；每个公会一个goroutine
*/

func genGuildModule(sid uint) *GuildModule {
	m := &GuildModule{
		sid: sid,
	}
	m.applyWorker.m = m
	m.guildMgrWorker.m = m
	return m
}

type GuildModule struct {
	sid            uint
	applyWorker    ApplyWorker
	guildMgrWorker GuildMgrWorker
}

func (r *GuildModule) GetGuildMgrChann() chan guildCommand {
	return r.guildMgrWorker.command_chan
}

func (r *GuildModule) GetApplyChann() chan applyCommand {
	return r.applyWorker.command_chan
}

func (r *GuildModule) AfterStart(g *gin.Engine) {
}

func (r *GuildModule) BeforeStop() {
}

func (r *GuildModule) Start() {
	r.guildMgrWorker.Start(r.sid)
	r.applyWorker.Start(r.sid)
	guildRankAward()
}

func (r *GuildModule) Stop() {
	r.guildMgrWorker.Stop()
	r.applyWorker.Stop()
}

type GuildRet struct {
	CodeLevel int
	ErrCode   int
	ErrMsg    string
}

func (gr GuildRet) HasError() bool {
	return (gr.CodeLevel == Code_Warn || gr.CodeLevel == Code_Err) && gr.ErrCode != 0
}

// apply oper
func (r *GuildModule) NewGuild(base GuildSimpleInfo, founder helper.AccountSimpleInfo, channel string,
	lootID []string, times []int64, lastLeaveTime int64) (string, *GuildInfo, GuildRet) {
	res := r.applyCommandExec(applyCommand{
		Type:          Apply_Cmd_NewGuild,
		BaseInfo:      base,
		Applicant:     founder,
		Channel:       channel,
		AssignID:      lootID,
		AssignTimes:   times,
		LastLeaveTime: lastLeaveTime,
	})

	if res.ret.HasError() {
		return "", nil, res.ret
	}

	return res.guildInfo.Base.GuildUUID, &res.guildInfo, GuildRet{}
}

func (r *GuildModule) ApplyGuild(guildUUID string, applicant helper.AccountSimpleInfo, lootID []string,
	times []int64, lastLeaveTime int64) (GuildRet, bool) {
	res := r.applyCommandExec(applyCommand{
		Type:          Apply_Cmd_ApplyGuild,
		Applicant:     applicant,
		BaseInfo:      GuildSimpleInfo{GuildUUID: guildUUID},
		AssignID:      lootID,
		AssignTimes:   times,
		LastLeaveTime: lastLeaveTime,
	})
	return res.ret, res.isApplySyncGuild
}

func (r *GuildModule) CancelApplyGuild(guildUUID string, applicant string) GuildRet {
	res := r.applyCommandExec(applyCommand{
		Type: Apply_Cmd_CancelApplyGuild,
		Applicant: helper.AccountSimpleInfo{
			AccountID: applicant,
		},
		BaseInfo: GuildSimpleInfo{GuildUUID: guildUUID},
	})
	return res.ret
}

func (r *GuildModule) ApplyOper(guildUUID, approveAcid, applicantAcid string,
	isApprove bool, channel string) GuildRet {
	var cmdType int
	if isApprove {
		cmdType = Apply_Cmd_ApproveApply
	} else {
		cmdType = Apply_Cmd_DelApply
	}
	res := r.applyCommandExec(applyCommand{
		Type: cmdType,
		Approver: helper.AccountSimpleInfo{
			AccountID: approveAcid,
		},
		Applicant: helper.AccountSimpleInfo{
			AccountID: applicantAcid,
		},
		BaseInfo: GuildSimpleInfo{GuildUUID: guildUUID},
		Channel:  channel,
	})
	return res.ret
}

func (r *GuildModule) GetPlayerApplyInfo(acid string) []PlayerApplyInfo2Client {
	res := r.applyCommandExec(applyCommand{
		Type:      Apply_Cmd_GetPlayerApplyList,
		Applicant: helper.AccountSimpleInfo{AccountID: acid},
	})
	return res.playerApply
}

func (r *GuildModule) GetGuildApplyInfo(guildUUID, acid string) []GuildApplyInfo {
	res := r.applyCommandExec(applyCommand{
		Type: Apply_Cmd_GetGuildApplyList,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Applicant: helper.AccountSimpleInfo{AccountID: acid},
	})
	return res.guildApply
}

// 更新公会信息到ApplyWorker，异步
func (r *GuildModule) updateGuildInfo2AW(gsInfo GuildSimpleInfo) {
	r.applyCommandExecAsyn(applyCommand{
		Type:     Apply_Cmd_GuildInfo_Update,
		BaseInfo: gsInfo,
	})
}

// guild oper
func (r *GuildModule) GetGuildInfo(guildUUID string) (*GuildInfo, GuildRet) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_GetGuildInfo,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	})

	if res.ret.HasError() {
		return nil, res.ret
	}

	return &res.guildInfo, GuildRet{}
}

func (r *GuildModule) SetGuildApplySetting(guildUUID string,
	acid string, applyGsLimit int, applyAuto bool) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_SetGuildApplySetting,
		BaseInfo: GuildSimpleInfo{
			GuildUUID:    guildUUID,
			ApplyGsLimit: applyGsLimit,
			ApplyAuto:    applyAuto,
		},
		Player1: helper.AccountSimpleInfo{
			AccountID: acid,
		},
	})

	if res.ret.HasError() {
		return res.ret
	}

	return GuildRet{}
}

func (r *GuildModule) NoticeGuildWhenOffline(guid string, acid string) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_NoticeGuild_Offline,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guid,
		},
		Player1: helper.AccountSimpleInfo{
			AccountID: acid,
		},
	})
}

func (r *GuildModule) GetRandGuild(acid string) []GuildSimpleInfo {
	res := r.guildCommandExec(guildCommand{
		Type: GuildMgr_Cmd_GetRandomGuild,
		Player1: helper.AccountSimpleInfo{
			AccountID: acid,
		},
	})
	return res.guilds[:]
}

func (r *GuildModule) FindGuild(acid string, guildId int64) (GuildInfo, GuildRet) {
	res := r.guildCommandExec(guildCommand{
		Type: GuildMgr_Cmd_FindGuild,
		BaseInfo: GuildSimpleInfo{
			GuildID: guildId,
		},
		Player1: helper.AccountSimpleInfo{
			AccountID: acid,
		},
	})
	return res.guildInfo, res.ret
}

func (r *GuildModule) AddMem(guildUuId string, mem *helper.AccountSimpleInfo, approveName string, assignID []string,
	assignTimes []int64) (
	GuildRet, GuildSimpleInfo) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_AddMem,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		Player1:   *mem,
		Player2:   helper.AccountSimpleInfo{Name: approveName},
		ParamStrs: append([]string{}, assignID[:]...),
		ParamInts: append([]int64{}, assignTimes[:]...),
	})
	return res.ret, res.guildInfo.Base
}

func (r *GuildModule) SendRedPacket(guildUuId string, playerName string) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_SendGuildRedPacket,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		Player1: helper.AccountSimpleInfo{
			Name: playerName,
		},
	})
}

func (r *GuildModule) GrabRedPacket(guildUuId, acid, playerName, rpId string, act uint32) (GuildRet, string, map[string]uint32) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_GrabRedPacket,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		Player1:   helper.AccountSimpleInfo{AccountID: acid, Name: playerName},
		ParamStrs: []string{rpId},
		ParamInts: []int64{int64(act)},
	})
	return res.ret, res.ResStr[0], res.ResItemC
}

//添加公会膜拜日志

func (r *GuildModule) AddWorshipLogInfo(guildUuId, playerName string, signId int64, time int64) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_AddWorshipLogInfo,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		ParamStrs: []string{playerName},
		ParamInts: []int64{signId},
		DebugTime: time,
	})
}

//公会膜拜index++

func (r *GuildModule) AddWorshipIndex(guildUuId string, i int64) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_AddWorshipIndex,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		ParamInts: []int64{i},
	})
}

func (r *GuildModule) DebugResetGuildWorship(guildUuId string) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_DebugResetGuildWorship,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
	})
}

// 通知公会的申请数量，异步的
func (r *GuildModule) noticeHasApply(guid string, hasApply bool) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_NoticeHasApply,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guid,
		},
		HasApply: hasApply,
	})
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *GuildModule) guildCommandExec(cmd guildCommand) guildCommandRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	errRet := genErrRes(Err_Inner)

	res_chan := make(chan guildCommandRes, 1)
	cmd.resChan = res_chan
	chann := r.GetGuildMgrChann()
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("guildCommandExec %d guild chann full, cmd put timeout", cmd.Type)
		return errRet
	}

	select {
	case res := <-res_chan:
		return res
	case <-ctx.Done():
		logs.Error("guildCommandExec %d guild <-res_chan timeout", cmd.Type)
		return errRet
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *GuildModule) guildCommandExecAsyn(cmd guildCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	res_chan := make(chan guildCommandRes, 1)
	cmd.resChan = res_chan
	chann := r.GetGuildMgrChann()
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("[guildCommandExec] guild chann full, cmd put timeout")
		return
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *GuildModule) applyCommandExec(cmd applyCommand) guildCommandRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	errRet := genErrRes(Err_Inner)

	res_chan := make(chan guildCommandRes, 1)
	cmd.resChan = res_chan
	chann := r.GetApplyChann()
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("[applyCommandExec] guild chann full, cmd put timeout")
		return errRet
	}

	select {
	case res := <-res_chan:
		return res
	case <-ctx.Done():
		logs.Error("[applyCommandExec] apply <-res_chan timeout")
		return errRet
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *GuildModule) applyCommandExecAsyn(cmd applyCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	res_chan := make(chan guildCommandRes, 1)
	cmd.resChan = res_chan
	chann := r.GetApplyChann()
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("[applyCommandExec] guild chann full, cmd put timeout")
	}
}

func (r *GuildModule) RenameGuild(guildID string, newName string, accountName string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_RenameGuild,
		BaseInfo: GuildSimpleInfo{
			Name:      newName,
			GuildUUID: guildID,
		},
		Player1: helper.AccountSimpleInfo{
			Name: accountName,
		},
	})
	return res.ret
}
