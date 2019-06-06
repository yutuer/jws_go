package guild

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GuildModule) ChangeMemberPosition(guildUUID, memberAcID, kickAcID string, pos int, channel string) GuildRet {
	c := guildCommand{
		Type:    Command_ChangeMemberPosition,
		Channel: channel,
	}

	c.BaseInfo.GuildUUID = guildUUID
	c.Player1.AccountID = memberAcID
	c.Player2.AccountID = kickAcID
	c.Player2.GuildPosition = pos
	res := r.guildCommandExec(c)
	return res.ret
}

func (g *GuildWorker) changeMemberPosition(c *guildCommand) {
	guildInfo := g.guild

	memberAcID := c.Player1.AccountID
	acIDToSet := c.Player2.AccountID

	if acIDToSet == memberAcID {
		logs.Error("can not change myself postion")
		c.resChan <- genErrRes(Err_Guild_Position)
		return
	}

	m := guildInfo.GetMember(memberAcID)
	k := guildInfo.GetMember(acIDToSet)

	if m == nil || k == nil {
		logs.Warn("changeMemberPosition no member %s %s",
			memberAcID, acIDToSet)
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}

	if c.Player2.GuildPosition == k.GuildPosition {
		logs.Warn("changeMemberPosition pos eq %s %s",
			memberAcID, acIDToSet)
		//c.resChan <- genWarnRes(errCode.GuildPosChangeFailedBySamePosition)
		c.resChan <- guildCommandRes{}
		return
	}

	if !g.canChangeMemberPosition(m.GuildPosition, k.GuildPosition, c.Player2.GuildPosition) {
		logs.Error("changeMemberPosition permission denied %s %d, %d, %d",
			memberAcID, m.GuildPosition, k.GuildPosition, c.Player2.GuildPosition)
		c.resChan <- genWarnRes(errCode.GuildPositionErr)
		return
	}

	if c.Player2.GuildPosition < 0 || c.Player2.GuildPosition > gamedata.Guild_Pos_Count {
		logs.Error("changeMemberPosition no position %s %d",
			memberAcID, m.GuildPosition)
		c.resChan <- genErrRes(Err_Guild_Position)
		return
	}

	oldPos := k.GuildPosition
	// 公会会长禅让
	if c.Player2.GuildPosition == gamedata.Guild_Pos_Chief {
		logs.Info("changeMemberPosition Guild_Pos_Chief down %s %s", memberAcID, acIDToSet)
		if k.CorpLv < gamedata.GetCommonCfg().GetPresidentLevelLimit() {
			c.resChan <- genWarnRes(errCode.GuildMemLevelNotEnough)
			return
		}
		guildInfo.posNum[k.GuildPosition] -= 1
		guildInfo.posNum[gamedata.Guild_Pos_Mem] += 1
		m.GuildPosition = gamedata.Guild_Pos_Mem
		k.GuildPosition = gamedata.Guild_Pos_Chief
		guildInfo.Base.LeaderAcid = k.AccountID
		guildInfo.Base.LeaderName = k.Name
		guildInfo.GuildLog.AddLog(guild_info.IDS_GUILD_LOG_7, []string{k.Name})
		mail_sender.SendGangPosChg(memberAcID, gamedata.Guild_Pos_Chief, gamedata.Guild_Pos_Mem)
		mail_sender.SendGangPosChg(acIDToSet, oldPos, gamedata.Guild_Pos_Chief)
		guildInfo.UpdateChiefInfo(k.AccountID, k.Name)
		// 同步会长的职位
		syncGuild2Player(m.AccountID,
			guildInfo.Base.GuildUUID,
			guildInfo.Base.Name,
			m.GuildPosition,
			int(guildInfo.Base.Level))
	} else {
		numMax := gamedata.GetGuildPosMaxNum(c.Player2.GuildPosition)
		if numMax > 0 && guildInfo.posNum[c.Player2.GuildPosition] >= numMax {
			c.resChan <- genErrRes(Err_Guild_Full)
			return
		} else {
			guildInfo.posNum[k.GuildPosition] -= 1
			guildInfo.posNum[c.Player2.GuildPosition] += 1
			k.GuildPosition = c.Player2.GuildPosition
			guildInfo.GuildLog.AddLog(guild_info.IDS_GUILD_LOG_5,
				[]string{m.Name, k.Name, fmt.Sprintf("%d", k.GuildPosition)})
			mail_sender.SendGangPosChg(acIDToSet, oldPos, c.Player2.GuildPosition)
		}
	}
	g.m.updateGuildInfo2AW(guildInfo.Base)

	syncGuild2Player(acIDToSet,
		guildInfo.Base.GuildUUID,
		guildInfo.Base.Name,
		k.GuildPosition,
		int(guildInfo.Base.Level))

	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	// log
	logicLog(memberAcID, c.Channel, logiclog.LogicTag_GuildPosChg, guildInfo, acIDToSet,
		gamedata.GuildPositionString(oldPos), gamedata.GuildPositionString(k.GuildPosition))
	c.resChan <- guildCommandRes{}
	return
}

func (g *GuildWorker) canChangeMemberPosition(operatorPos, memberPos, changePos int) bool {
	if operatorPos == gamedata.Guild_Pos_Chief {
		return true // 会长拥有所有权限
	}
	operatorPosWeight := g.convertGuildPosWeight(operatorPos)
	memberPosWeight := g.convertGuildPosWeight(memberPos)
	if operatorPosWeight >= memberPosWeight {
		return false // 同级或低职位的不能操作
	}
	operatorConfig := gamedata.GetGuildPosData(operatorPos)
	if operatorConfig == nil {
		return false // 错误的职位
	}
	// 执行到这里的最高权限是副军团长， 所以只需要判断更低级的2个位置
	switch changePos {
	case gamedata.Guild_Pos_Elite:
		return operatorConfig.GetAppointElite() == 1
	case gamedata.Guild_Pos_Mem:
		return operatorConfig.GetAppointMember() == 1
	}
	return false
}

// 将职位换算成数值  职位越高， 数值越小
func (g *GuildWorker) convertGuildPosWeight(position int) int {
	if position == gamedata.Guild_Pos_Mem {
		return gamedata.Guild_Pos_Count
	} else {
		return position
	}
}

// 比较职位，1 left > right, -1 left < right 0 left = right
func (g *GuildWorker) ComparePos(leftPos, rightPos int) int {
	leftPosWeight := g.convertGuildPosWeight(leftPos)
	rightPosWeight := g.convertGuildPosWeight(rightPos)
	if leftPosWeight < rightPosWeight {
		return 1
	} else if leftPosWeight > rightPosWeight {
		return -1
	} else {
		return 0
	}
}
