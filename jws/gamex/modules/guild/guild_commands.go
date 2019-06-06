package guild

import (
	"math/rand"

	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	warnCode "vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	Command_Null = iota
	Command_GetGuildInfo
	Command_AddMem
	Command_UpdateAccountInfo
	Command_KickMember
	Command_ChangeMemberPosition
	Command_Quit
	Command_Dismiss
	Command_ChangeNotice
	Command_NoticeHasApply
	Command_NoticeGuild_Offline // 10
	Command_SetGuildApplySetting
	Command_AddGuildInventory
	Command_ApplyListGuildInventory
	Command_ApplyGuildInventory
	Command_ApproveGuildInventory
	Command_ExchangeGuildInventory
	Command_AssignGuildInventory
	Command_DebugSetGuildInventory
	Command_DebugResetGuildInventory
	Command_GetGSTBonus //20
	Command_GSTLevelUp
	Command_GetGSTDay
	Command_GetGSTWeek
	Command_DebugSetGSTLevel
	Command_DebugResetGateEnemy

	Command_AddMemSyncReceiver // 开始通知活动人员变动
	Command_DelMemSyncReceiver // 结束通知活动人员变动
	Command_UseGateEnemyCount
	Command_OnGateEnemyStop    // 兵临城下结束时的通知
	Command_SetGateEnemyReward // 30
	Command_Sign
	Command_AddXP
	Command_AddSP
	Command_RenameGuild

	GuildMgr_Cmd_GetRandomGuild
	GuildMgr_Cmd_FindGuild
	GuildMgr_Cmd_Dismiss_CallBack

	Command_GuildActGetStat

	Command_GuildActBossLock
	Command_GuildActBossUnLock // 40
	Command_GuildActBossBeginFight
	Command_GuildActBossEndFight
	Command_GuildActBossSendActNotify
	Command_GuildActBossDebugClean
	Command_GuildActBossIsPassed

	Command_GVGBalanceForPlayer
	Command_GVGBalanceForInventory
	Command_DebugAutoChangeChief

	Command_SendGuildRedPacket
	Command_GrabRedPacket // 50
	Command_DebugResetRedPacketForGuild
	Command_DebugResetRedPacketForPlayer
	Command_DebugAllSendRedPacket

	Command_AddWorshipLogInfo
	Command_AddWorshipIndex
	Command_UpdateGVGScore
	Command_DebugResetGuildWorship

	Command_Debug
	Command_GetLog

	Command_Count
)

type guildCommandRes struct {
	guildCommandResWithActDatas

	ret       GuildRet
	guildInfo GuildInfo
	guilds    []GuildSimpleInfo // 一定要复制，因为要跨rountine

	playerApply      []PlayerApplyInfo2Client
	guildApply       []GuildApplyInfo
	ResInt           []int64
	ResInt1          []int64
	ResInt2          []int64
	ResStr           []string
	ResStr2          []string
	ResItemC         map[string]uint32
	ResFloat         []float32
	isApplySyncGuild bool
	IsDismiss        bool
}

type guildCommand struct {
	Type              int
	Player1           helper.AccountSimpleInfo
	Player2           helper.AccountSimpleInfo
	BaseInfo          GuildSimpleInfo
	HasApply          bool
	Channel           string
	resChan           chan guildCommandRes
	memSyncReceiverID int
	memSyncReceiver   helper.IGuildMemSyncReceiver
	gateEnemyData     *player_msg.GatesEnemyData
	inventoryLoot     []GuildInventoryLoot
	LootId            string
	ParamInts         []int64
	ParamStrs         []string
	ParamBools        []bool
	DebugTime         int64
	Rand              *rand.Rand
	Reason            string
	AddXP             int64
	AddGB             int64
}

const (
	Code_Err = iota
	Code_Warn
	Code_Inner_Msg
)
const (
	Success = iota
	Err_Guild_New
	Err_Guild_Already_Exist
	Err_Guild_Not_Exist
	Err_Guild_Apply_Full
	Err_Guild_Position
	Err_Guild_Applicant_Not_Found
	Err_Guild_Full
	Err_DB
	Err_Player_In_Other_Guild
	Err_Name_Repeat // 10
	Err_Chief_Not_Quit
	Err_Mem_Not_Found
	Err_Inner
	Err_Pos_Appoint
	Err_Player_Apply_Max
	Err_Player_Apply_Repeat
	Err_CODE_ERR_Name_Len
	Err_No_Gate_Enemy_Count
	Err_No_Gate_Enemy_Reward
	Err_Gate_Enemy_Already_Start // 20
	Err_Param
	Err_Guild_SP_Not_Enough
	Err_Guild_Lvl_Not_Enough
	Err_No_Gate_Enemy_Not_Join
	Err_Unknown_Err
	Err_Count
)

var (
	err_code = []string{
		Success:                       "Success",
		Err_Guild_New:                 "Err_Guild_New",
		Err_Guild_Already_Exist:       "Err_Guild_Already_Exist",
		Err_Guild_Not_Exist:           "Err_Guild_Not_Exist",
		Err_Guild_Apply_Full:          "Err_Guild_Apply_Full",
		Err_Guild_Position:            "Err_Guild_Position",
		Err_Guild_Applicant_Not_Found: "Err_Guild_Applicant_Not_Found",
		Err_Guild_Full:                "Err_Guild_Full",
		Err_DB:                        "Err_DB",
		Err_Player_In_Other_Guild:    "Err_Player_In_Other_Guild",
		Err_Name_Repeat:              "Err_Name_Repeat",
		Err_Chief_Not_Quit:           "Err_Chief_Not_Quit",
		Err_Mem_Not_Found:            "Err_Mem_Not_Found",
		Err_Inner:                    "Err_Inner",
		Err_Pos_Appoint:              "Err_Pos_Appoint",
		Err_Player_Apply_Max:         "Err_Player_Apply_Max",
		Err_Player_Apply_Repeat:      "Err_Player_Apply_Repeat",
		Err_CODE_ERR_Name_Len:        "Err_CODE_ERR_Name_Len",
		Err_No_Gate_Enemy_Count:      "Err_No_Gate_Enemy_Count",
		Err_No_Gate_Enemy_Reward:     "Err_No_Gate_Enemy_Reward",
		Err_Gate_Enemy_Already_Start: "Err_Gate_Enemy_Already_Start",
		Err_Param:                    "Err_Param",
		Err_Guild_SP_Not_Enough:      "Err_Guild_SP_Not_Enough",
		Err_Guild_Lvl_Not_Enough:     "Err_Guild_Lvl_Not_Enough",
		Err_Unknown_Err:              "Err_Unknown_Err",
	}
)

func (g *GuildWorker) processCommand(c *guildCommand) {
	g.guild.TryRefresh(g.guild.shardId)

	switch c.Type {
	case Command_GetGuildInfo:
		g.getGuildInfo(c)
	case Command_NoticeGuild_Offline:
		g.noticeWhenOffline(c)
	case Command_SetGuildApplySetting:
		g.setGuildApplySetting(c)
	case Command_AddGuildInventory:
		g.addGuildInventory(c)
	case Command_ApplyListGuildInventory:
		g.getApplyListGuildInventoryItem(c)
	case Command_ApplyGuildInventory:
		g.applyGuildInventoryItem(c)
	case Command_ApproveGuildInventory:
		g.approveGuildInventoryItem(c)
	case Command_ExchangeGuildInventory:
		g.exchangeGuildInventoryItem(c)
	case Command_AssignGuildInventory:
		g.assignGuildInventoryItem(c)
	case Command_DebugSetGuildInventory:
		g.debugSetGuildInventoryTime(c)
	case Command_DebugResetGuildInventory:
		g.debugResetGuildInventoryTime(c)
	case Command_AddMem:
		g.addMem(c)
	case Command_UpdateAccountInfo:
		g.updateAccountInfo(c)
	case Command_Quit:
		g.quitGuildOrDismiss(c)
	case Command_KickMember:
		g.kickMember(c)
	case Command_Dismiss:
		g.dismissGuild(c)
	case Command_ChangeMemberPosition:
		g.changeMemberPosition(c)
	case Command_ChangeNotice:
		g.changeGuildNotice(c)
	case Command_NoticeHasApply:
		g.noticeHasApply(c)
	case Command_AddMemSyncReceiver:
		g.addMemSyncReceiver(c)
	case Command_DelMemSyncReceiver:
		g.delMemSyncReceiver(c)
	case Command_UseGateEnemyCount:
		g.useGateEnemyCount(c)
	case Command_OnGateEnemyStop:
		g.onGateEnemyStop(c)
	case Command_SetGateEnemyReward:
		g.setGateEnemyReward(c)
	case Command_AddXP:
		g.addXP(c)
	case Command_AddSP:
		g.addSP(c)
	case Command_RenameGuild:
		g.renameGuild(c)
	case Command_Sign:
		g.sign(c)
	case Command_Debug:
		g.debugOp(c)
	case Command_GetGSTBonus:
		g.getGuildScienceBonus(c)
	case Command_GSTLevelUp:
		g.addGuildSciencePoint(c)
	case Command_GetGSTDay:
		g.scienceDayLog(c, false)
	case Command_GetGSTWeek:
		g.scienceDayLog(c, true)
	case Command_DebugSetGSTLevel:
		g.debugSetGuildScienceLevel(c)
	case Command_DebugResetGateEnemy:
		g.debugResetGateEnemy(c)
	case Command_GVGBalanceForPlayer:
		g.gvgBalanceForPlayer(c)
	case Command_GVGBalanceForInventory:
		g.gvgBalanceForInventory(c)
	case Command_DebugAutoChangeChief:
		g.guild.debugAutoChangeChief()
	case Command_SendGuildRedPacket:
		g.sendGuildRedPacket(c)
	case Command_GrabRedPacket:
		g.grabRedPacket(c)
	case Command_DebugResetRedPacketForGuild:
		g.guild.GuildRedPacket.DailyReset(g.guild.GetDebugNowTime(g.guild.shardId))
	case Command_DebugResetRedPacketForPlayer:
		g.guild.GuildRedPacket.DebugCleanPlayer(c.ParamStrs[0])
	case Command_DebugAllSendRedPacket:
		g.DebugAllSendRedPacket()
	case Command_GetLog:
		g.getGuildLog(c)
	case Command_AddWorshipLogInfo:
		g.addWorshipLogInfo(c)
	case Command_UpdateGVGScore:
		g.updateGVGScore(c)
	case Command_AddWorshipIndex:
		g.addWorshipIndex(c)
	case Command_DebugResetGuildWorship:
		g.guild.GuildWorship.LastDailyResetTime = 0
		g.guild.Base.Guild2LvlTimes = 0
		g.guild.TryResetGuildWorship(g.guild.GetDebugNowTime(g.guild.shardId))
	default:
		if g.processActCmd(c) {
			return
		}
	}

}

func (g *GuildWorker) getGuildInfo(c *guildCommand) {
	guildNowTime := g.guild.GetDebugNowTime(g.m.sid)
	changeChiefOk := g.guild.TryAutoChangeGuildChief(guildNowTime)
	updateInventoryOk := g.guild.Inventory.UpdateGuildInventory(guildNowTime)
	updateLostInventoryOk := g.guild.LostInventory.UpdateGuildInventory(guildNowTime)
	if changeChiefOk || updateInventoryOk || updateLostInventoryOk {
		// save guild
		errC := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			if err := g.guild.DBSave(cb); err != nil {
				return err
			}
			return nil
		})
		if errC != 0 {
			c.resChan <- genErrRes(errC)
			return
		}
	}
	c.resChan <- guildCommandRes{
		guildInfo: *(g.guild),
	}
}

func (g *GuildWorker) noticeWhenOffline(c *guildCommand) {
	for i := 0; i < len(g.guild.Members); i++ {
		if g.guild.Members[i].AccountID == c.Player1.AccountID {
			g.guild.Members[i].SetOnline(false)
			g.guild.Members[i].LastLoginTime = time.Now().Unix()
		}
	}
}

func (g *GuildWorker) setGuildApplySetting(c *guildCommand) {
	res := guildCommandRes{}

	// 职位检查
	mem := g.guild.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil || !gamedata.CheckApprovePosition(mem.GuildPosition) {
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}

	g.guild.Base.ApplyGsLimit = c.BaseInfo.ApplyGsLimit
	g.guild.Base.ApplyAuto = c.BaseInfo.ApplyAuto

	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := g.guild.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		c.resChan <- genErrRes(errCode)
		return
	}
	g.m.updateGuildInfo2AW(g.guild.Base)

	c.resChan <- res
	return
}

func (g *GuildWorker) renameGuild(c *guildCommand) {
	res := guildCommandRes{}
	code, isError := renameGuildName(g.guild.Base.Name, c.BaseInfo.Name, TableGuildName(g.m.sid), g.guild.Base.GuildUUID)
	if code != 0 {
		if isError {
			c.resChan <- genErrRes(code)
		} else {
			c.resChan <- genWarnRes(code)
		}
		return
	}
	g.guild.Base.Name = c.BaseInfo.Name
	g.guild.Base.RenameTimes++
	// 发送邮件通知换军团长
	for _, member := range g.guild.Members {
		if member.AccountID == "" {
			continue
		}
		mail_sender.BatchSendMail2Account(member.AccountID, timail.Mail_send_By_Guild,
			mail_sender.IDS_MAIL_REGUILDNAME_TITLE,
			[]string{c.BaseInfo.Name}, nil,
			"RenameGuild", true)
	}
	g.guild.GuildLog.AddLog(IDS_GUILD_LOG_11, []string{c.Player1.Name})
	c.resChan <- res
	g.m.updateGuildInfo2AW(g.guild.Base)
	rank.GetModule(g.m.sid).RankGuildGs.OnGuildOrLeaderRename(&g.guild.Base)

}
func (g *GuildWorker) noticeHasApply(c *guildCommand) {
	g.guild.SyncApply2AllMem(c.HasApply)
}

func (g *GuildWorker) addMem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild
	// 加公会
	playerInfo := &c.Player1
	playerInfo.GuildPosition = gamedata.Guild_Pos_Mem
	info.addMem(playerInfo)
	info.posNum[gamedata.Guild_Pos_Mem] += 1
	g.guild.AwakeIfSleep(g.guild.GetDebugNowTime(g.m.sid))
	if c.Player2.Name != "" {
		g.guild.GuildLog.AddLog(IDS_GUILD_LOG_2, []string{c.Player2.Name, playerInfo.Name})
	} else {
		g.guild.GuildLog.AddLog(IDS_GUILD_LOG_10, []string{playerInfo.Name})
	}
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := addGuildMem(playerInfo.AccountID, info.Base.GuildUUID, cb); err != nil {
			return err
		}
		if err := info.DBSave(cb); err != nil {
			return err
		}
		return nil
	})

	g.guild.Inventory.SetAssignTimesByAcID(c.Player1.AccountID, c.ParamStrs, c.ParamInts)
	if errCode != 0 {
		c.resChan <- genErrRes(errCode)
		return
	}
	res.guildInfo.Base = info.Base
	c.resChan <- res
	return
}

func (g *GuildWorker) addMemSyncReceiver(c *guildCommand) {
	info := g.guild
	info.AddMemSyncReceiver(c.memSyncReceiver)
	c.resChan <- guildCommandRes{}
}

func (g *GuildWorker) delMemSyncReceiver(c *guildCommand) {
	info := g.guild
	info.DelMemSyncReceiver(c.memSyncReceiverID)
	c.resChan <- guildCommandRes{}
}

func genErrRes(errcode int) guildCommandRes {
	msg := ""
	if errcode < len(err_code) {
		msg = err_code[errcode]
	} else {
		msg = fmt.Sprintf("%d", errcode)
	}
	return guildCommandRes{
		ret:    GuildRet{ErrCode: errcode, ErrMsg: msg},
		ResInt: []int64{0, 0, 0, 0, 0},
		ResStr: []string{"", "", "", ""},
	}
}

func genWarnRes(code int) guildCommandRes {
	msg := ""
	if code < len(warnCode.Warn_Str) {
		msg = warnCode.Warn_Str[code]
	}
	return guildCommandRes{
		ret:    GuildRet{CodeLevel: Code_Warn, ErrCode: code, ErrMsg: msg},
		ResInt: []int64{0, 0, 0, 0, 0},
		ResStr: []string{"", "", "", ""},
	}
}

func logicLog(accountid, channel string, guildOper string, info *GuildInfo, acid, befPos, aftPos string) {
	logicInfo := &logiclog.LogicInfo_Guild{
		GuildUUID: info.Base.GuildUUID,
		GuildID:   info.Base.GuildID,
		Name:      info.Base.Name,
		Level:     info.Base.Level,
		MemNum:    info.Base.MemNum,
		Acid:      acid,
		BefPos:    befPos,
		AftPos:    aftPos,
	}
	logicInfo.Mems = make([]logiclog.LogicInfo_GuildMem, 0, info.Base.MemNum)
	for i := 0; i < len(info.Members) && i < info.Base.MemNum; i++ {
		mem := info.Members[i]
		logicInfo.Mems = append(logicInfo.Mems, logiclog.LogicInfo_GuildMem{
			Name:          mem.Name,
			AccountID:     mem.AccountID,
			GuildPosition: gamedata.GuildPositionString(mem.GuildPosition),
		})
	}

	logiclog.LogGuildOper(accountid, channel, guildOper, logicInfo, "")
}
