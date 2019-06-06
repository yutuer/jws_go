package guild

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/account_info"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util"
)

func (r *GuildModule) DismissGuild(guildUUID, acid, channel string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_Dismiss,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acid},
		Channel: channel,
	})
	return res.ret
}

// 解散公会
// note：涉及删除各级缓冲和表，和搜索、查找公会操作可能存在多线程问题
func (g *GuildWorker) dismissGuild(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild
	// 职位检查
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil || mem.GuildPosition != gamedata.Guild_Pos_Chief {
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}
	guildName := info.Base.Name
	var chielfName string
	mems := make([]string, info.Base.MemNum)
	for i := 0; i < info.Base.MemNum && info.Members[i].AccountID != ""; i++ {
		mem := info.Members[i]
		mems[i] = mem.AccountID
		if mem.GuildPosition == gamedata.Guild_Pos_Chief {
			chielfName = mem.Name
		}
	}
	// log
	logicLog(c.Player1.AccountID, c.Channel, logiclog.LogicTag_GuildDismiss, info, c.Player1.AccountID, "", "")

	nowDebugTime := g.guild.GetDebugNowTime(g.guild.shardId)
	nextEnterGuildTime := util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(nowDebugTime), nowDebugTime)
	assignID, assignTimes := g.guild.Inventory.GetAssignTimesByAcID(c.Player1.AccountID)
	syncLeaveGuild2Player(c.Player1.AccountID, nowDebugTime, nextEnterGuildTime, assignID, assignTimes)

	// 记录公会每个人离开公会的时间
	acids := make([]string, g.guild.Base.MemNum)
	for i := 0; i < g.guild.Base.MemNum; i++ {
		acids[i] = g.guild.Members[i].AccountID
	}
	account_info.BatchSaveNextJoinGuildTime(acids, nextEnterGuildTime, assignID, assignTimes, nowDebugTime)

	// 解散
	if err := g.delGuild(info.Base.GuildUUID); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	// 成功发邮件
	for _, memId := range mems {
		mail_sender.SendGangDismiss(memId, chielfName, guildName)
	}
	res.IsDismiss = true
	c.resChan <- res
	return
}
