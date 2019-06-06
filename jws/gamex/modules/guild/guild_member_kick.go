package guild

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/account_info"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
)

func (r *GuildModule) KickMember(guildUUID, memberAcID, kickAcID, channel string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_KickMember,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: memberAcID},
		Player2: helper.AccountSimpleInfo{AccountID: kickAcID},
		Channel: channel,
	})
	return res.ret
}

func (g *GuildWorker) kickMember(c *guildCommand) {
	res := guildCommandRes{}

	guildInfo := g.guild
	memberAcID := c.Player1.AccountID
	kickAcID := c.Player2.AccountID

	m := guildInfo.GetMember(memberAcID)
	k := guildInfo.GetMember(kickAcID)

	if m == nil || k == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}

	// 权限检查
	posCfg := gamedata.GetGuildPosData(m.GuildPosition)
	if posCfg.GetKickMember() == 0 {
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}
	// 低职位的不能T同等级或更高职位的
	if g.ComparePos(m.GuildPosition, k.GuildPosition) != 1 {
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}

	kickPos := k.GuildPosition

	k.GuildPosition = gamedata.Guild_Pos_Mem

	errRes := guildInfo.delMember(kickAcID, m.Name)
	if errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	guildInfo.ActBoss.OnMemKick(kickAcID)

	nowDebugTime := g.guild.GetDebugNowTime(g.guild.shardId)
	//account_info.SaveLeaveGuildTime(kickAcID, nowDebugTime)
	nextEnterGuildTime := util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(nowDebugTime), nowDebugTime)
	assignID, assignTimes := g.guild.Inventory.GetAssignTimesByAcID(kickAcID)

	account_info.SaveInfoOnLeaveGuild(kickAcID, nextEnterGuildTime, assignID, assignTimes, nowDebugTime)

	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := delGuildMem(kickAcID, cb); err != nil {
			return err
		}
		if err := guildInfo.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		c.resChan <- genErrRes(errCode)
		return
	}
	guildInfo.posNum[kickPos] -= 1
	// 成功发邮件
	mail_sender.SendGangKick(kickAcID, m.Name, guildInfo.Base.Name)

	g.m.updateGuildInfo2AW(guildInfo.Base)
	// log
	logicLog(memberAcID, c.Channel, logiclog.LogicTag_GuildDelMem, guildInfo, kickAcID, "", "")
	c.resChan <- res
	return
}
