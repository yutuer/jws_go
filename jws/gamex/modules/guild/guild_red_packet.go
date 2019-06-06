package guild

import (
	"math"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (g *GuildWorker) sendGuildRedPacket(c *guildCommand) {
	logs.Debug("GuildRedPakcet: send red packet, %s, %s", c.BaseInfo.GuildUUID, c.Player1.Name)
	g.guild.GuildRedPacket.NewGuildRedPacket(c.BaseInfo.GuildUUID, c.Player1.Name)
	if err := g.saveGuild(g.guild); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	acids := make([]string, g.guild.Base.MemNum)
	for i, member := range g.guild.Members[:g.guild.Base.MemNum] {
		acids[i] = member.AccountID
	}
	g.syncGuildRedPacket2Players(acids)
}

func (g *GuildWorker) syncGuildRedPacket2Players(acids []string) {
	player_msg.SendToPlayers(acids, player_msg.PlayerMsgGuildRedPacketSyncCode,
		player_msg.DefaultMsg{})
}

func (g *GuildWorker) grabRedPacket(c *guildCommand) {
	guildInfo := g.guild
	rp, ok := guildInfo.GuildRedPacket.Get(c.ParamStrs[0]) // [0] id
	if !ok {
		c.resChan <- genWarnRes(errCode.RedPacketNotFound)
		return
	}

	grabPlayer := c.Player1 // 抢红包角色
	if rp.Contains(grabPlayer.Name) {
		c.resChan <- genWarnRes(errCode.RedPacketHasGrabbed)
		return
	}
	res := guildCommandRes{}
	grabRecord := rp.AddGrab(uint32(c.ParamInts[0]), grabPlayer.AccountID, grabPlayer.Name)
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	res.ResItemC = make(map[string]uint32, len(grabRecord.RewardList))
	for _, gr := range grabRecord.RewardList {
		res.ResItemC[gr.ItemId] = gr.Count
	}
	res.ResStr = make([]string, 1)
	res.ResStr[0] = rp.SenderName
	c.resChan <- res
}

func (g *GuildWorker) DebugAllSendRedPacket() {
	acids := make([]string, g.guild.Base.MemNum)
	for i, member := range g.guild.Members[:g.guild.Base.MemNum] {
		if !g.guild.GuildRedPacket.ContainsBySenderName(member.Name) {
			g.guild.GuildRedPacket.NewGuildRedPacket(g.guild.Base.GuildUUID, member.Name)
		}
		acids[i] = member.AccountID
	}
	g.syncGuildRedPacket2Players(acids)
}

func (g *GuildModule) DebugResetRedPacketForGuild(guildUUID string) GuildRet {
	res := g.guildCommandExec(guildCommand{
		Type: Command_DebugResetRedPacketForGuild,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	})
	return res.ret
}

func (g *GuildModule) DebugResetRedPacketForPlayer(guildUUID string, playerName string) GuildRet {
	res := g.guildCommandExec(guildCommand{
		Type: Command_DebugResetRedPacketForPlayer,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		ParamStrs: []string{playerName},
	})
	return res.ret
}

func (g *GuildModule) DebugAllSendRedPacket(guildUUID string) GuildRet {
	res := g.guildCommandExec(guildCommand{
		Type: Command_DebugAllSendRedPacket,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	})
	return res.ret
}

func (g *GuildModule) DebugAutoJoinGuild(guildUUID string, maxCount int64, topN []rank.CorpDataInRank) {
	guildInfo, _ := g.GetGuildInfo(guildUUID)
	if guildInfo != nil {
		needJoinCount := guildInfo.Base.MaxMemNum - guildInfo.Base.MemNum
		needJoinCount = int(math.Min(float64(needJoinCount), float64(maxCount)))
		for _, ranker := range topN {
			if needJoinCount <= 0 {
				break
			}
			if ranker.Info.AccountID != "" {
				ret := checkPlayerInGuild(ranker.Info.AccountID)
				if !ret.ret.HasError() {
					addRet, _ := g.AddMem(guildUUID, &ranker.Info, "", []string{}, []int64{})
					if !addRet.HasError() {
						needJoinCount--
					}
				}

			}
		}
	}
}
