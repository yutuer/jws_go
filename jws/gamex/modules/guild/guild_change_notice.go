package guild

/*
公会公告修改
除了最后的saveGuild其他都是内存操作
*/
import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
)

func (r *GuildModule) ChangeGuildNotice(guildUUID, acid, notice string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_ChangeNotice,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
			Notice:    notice,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acid},
	})
	return res.ret
}

func (g *GuildWorker) changeGuildNotice(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild
	// 职位检查
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil || gamedata.GetGuildPosData(mem.GuildPosition).GetRenNewsPower() == 0 {
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}
	// 检查敏感词
	if gamedata.CheckSensitive(c.BaseInfo.Notice) {
		c.resChan <- genWarnRes(errCode.GuildWordSensitive)
		return
	}
	// 改公告
	info.Base.Notice = c.BaseInfo.Notice
	info.GuildLog.AddLog(IDS_GUILD_LOG_8, []string{mem.Name})

	if err := g.saveGuild(info); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	g.m.updateGuildInfo2AW(info.Base)

	c.resChan <- res
	return
}
