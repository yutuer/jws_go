package guild

import (
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (g *GuildWorker) processActCmd(c *guildCommand) bool {
	var res *base.ActCommand
	res = g.processActBossCommand(c)
	if res != nil {
		g.processActCmdRes(c, res)
		return true
	}

	return false
}

func (g *GuildWorker) processActCmdRes(c *guildCommand, res *base.ActCommand) {
	if res != nil {
		if res.ErrCode != 0 {
			logs.Warn("processActCmdRes Err By %d", res.ErrCode)
			logs.Debug("processActCmdRes %d, %d", c.Type, res.ErrCode)
			c.resChan <- genErrRes(res.ErrCode)
		} else {
			resCmd := guildCommandRes{
				ResInt:   res.ParamInts,
				ResStr:   res.ParamStrs,
				ResItemC: res.ParamItemC,
			}
			resCmd.OnActRes(g.guild, res)
			c.resChan <- resCmd
		}
	} else {
		c.resChan <- genErrRes(Err_Unknown_Err)
	}
}

func (r *GuildModule) GetActStat(guildID string, actType int) *guildCommandResWithActDatas {
	res := r.guildCommandExec(guildCommand{
		Type:              Command_GuildActGetStat,
		memSyncReceiverID: actType,
	})
	return &res.guildCommandResWithActDatas
}

func (g *GuildWorker) processActGetStatCommand(c *guildCommand) *base.ActCommand {
	needSyncTyp := c.memSyncReceiverID

	switch c.Type {
	case Command_GuildActGetStat:
		res := new(base.ActCommand)
		if needSyncTyp == -1 {
			res.SetNeedAll()
		} else {
			res.SetNeedSync(needSyncTyp)
		}
		return res
	}

	return nil
}

func (g *GuildWorker) processActBossCommand(c *guildCommand) *base.ActCommand {
	actCmd := new(base.ActCommand)

	actCmd.ParamAccountInfo = &c.Player1
	actCmd.ParamInts = c.ParamInts[:]
	actCmd.ParamStrs = c.ParamStrs[:]

	switch c.Type {
	case Command_GuildActBossLock:
		return g.guild.ActBoss.LockBoss(actCmd)
	case Command_GuildActBossUnLock:
		return g.guild.ActBoss.UnLockBoss(actCmd)
	case Command_GuildActBossBeginFight:
		return g.guild.ActBoss.BeginBossFight(actCmd)
	case Command_GuildActBossEndFight:
		return g.guild.ActBoss.EndBossFight(actCmd)
	case Command_GuildActBossSendActNotify:
		return g.guild.ActBoss.SendActNotify(actCmd)
	case Command_GuildActBossIsPassed:
		return g.guild.ActBoss.IsBossAllPassed(actCmd)
	case Command_GuildActBossDebugClean:
		return g.guild.ActBoss.DebugClean(actCmd)
	}

	return nil
}
