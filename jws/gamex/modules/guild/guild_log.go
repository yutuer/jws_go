package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
)

func (r *GuildModule) GetGuildLog(guildUUID, acid string) (
	GuildRet, []int64, []int64, []int64, []string) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_GetLog,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{
			AccountID: acid,
		},
	})
	return res.ret, res.ResInt, res.ResInt1, res.ResInt2, res.ResStr
}

func (g *GuildWorker) getGuildLog(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild
	// 在工会检查
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}

	gLogs := g.guild.GuildInfoBase.GuildLog.Logs
	res.ResInt = make([]int64, 0, len(gLogs))
	res.ResInt1 = make([]int64, 0, len(gLogs))
	res.ResInt2 = make([]int64, 0, len(gLogs))
	res.ResStr = make([]string, 0, len(gLogs)*3)
	for _, log := range gLogs {
		res.ResInt = append(res.ResInt, log.TimeStamp)
		res.ResInt1 = append(res.ResInt1, int64(log.CfgId))
		res.ResInt2 = append(res.ResInt2, int64(len(log.Param)))
		res.ResStr = append(res.ResStr, log.Param...)
	}
	c.resChan <- res
	return
}
