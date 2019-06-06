package guild

func (g *GuildWorker) addWorshipLogInfo(c *guildCommand) {
	g.guild.GuildWorship.AddWorshipLogInfo(c.ParamStrs[0], c.ParamInts[0], c.DebugTime)
	if err := g.saveGuild(g.guild); err != nil {
		c.resChan <- genErrRes(Err_DB)
	}

}

func (g *GuildWorker) addWorshipIndex(c *guildCommand) {
	g.guild.GuildWorship.UpdateWorshipIndex(int(c.ParamInts[0]))
	if err := g.saveGuild(g.guild); err != nil {
		c.resChan <- genErrRes(Err_DB)
	}

}
