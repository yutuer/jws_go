package guild

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

func (r *GuildModule) QuitGuild(guildUUID, acid, channel string) (GuildRet, bool) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_Quit,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acid},
		Channel: channel,
	})
	return res.ret, res.IsDismiss
}

func (g *GuildWorker) quitGuildOrDismiss(c *guildCommand) {
	info := g.guild
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if g.guild.Base.MemNum == 1 && mem.GuildPosition == gamedata.Guild_Pos_Chief {
		g.dismissGuild(c)
	} else {
		g.quitGuild(c)
	}

}

func (g *GuildWorker) quitGuild(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild
	// 职位检查
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil || mem.GuildPosition == gamedata.Guild_Pos_Chief {
		c.resChan <- genWarnRes(errCode.GuildChiefNotQuit)
		return
	}
	quitPos := mem.GuildPosition
	// 退出
	if errRes := info.delMember(c.Player1.AccountID, ""); errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	info.ActBoss.OnMemKick(c.Player1.AccountID)

	info.posNum[quitPos] -= 1

	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := delGuildMem(c.Player1.AccountID, cb); err != nil {
			return err
		}
		if err := info.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		c.resChan <- genErrRes(errCode)
		return
	}

	g.m.updateGuildInfo2AW(info.Base)
	// log
	logicLog(c.Player1.AccountID, c.Channel, logiclog.LogicTag_GuildDelMem, info, c.Player1.AccountID, "", "")
	c.resChan <- res
	return
}
