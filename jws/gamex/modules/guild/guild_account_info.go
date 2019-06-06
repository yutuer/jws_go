package guild

/*
公会中需要更新玩家当前状态的接口。
全部都是内存操作
*/

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 申请列表中，以及公会成员列表中的玩家信息属性更新。
func (r *GuildModule) UpdateAccountInfo(info helper.AccountSimpleInfo) {
	r.applyCommandExecAsyn(applyCommand{
		Type:      Apply_Cmd_UpdateAccountInfoApply,
		Applicant: info,
	})

	guid := GetPlayerGuild(info.AccountID)
	if guid != "" {
		r.guildCommandExecAsyn(guildCommand{
			Type:    Command_UpdateAccountInfo,
			Player1: info,
			BaseInfo: GuildSimpleInfo{
				GuildUUID: guid,
			},
		})
	}
}

func (g *GuildWorker) updateAccountInfo(c *guildCommand) {
	guildInfo := g.guild
	isChg := false
	for i := 0; i < len(guildInfo.Members); i++ {
		if guildInfo.Members[i].AccountID == c.Player1.AccountID {
			// 下面的信息是从玩家存档向公会更新的
			guildInfo.UpdateGs(int64(c.Player1.CurrCorpGs - guildInfo.Members[i].CurrCorpGs))

			guildInfo.Members[i].CorpLv = c.Player1.CorpLv
			guildInfo.Members[i].CurrAvatar = c.Player1.CurrAvatar
			guildInfo.Members[i].CurrCorpGs = c.Player1.CurrCorpGs
			guildInfo.Members[i].LastLoginTime = c.Player1.LastLoginTime
			guildInfo.Members[i].Name = c.Player1.Name
			guildInfo.Members[i].FashionEquips = c.Player1.FashionEquips
			guildInfo.Members[i].Swing = c.Player1.Swing
			guildInfo.Members[i].MagicPetfigure = c.Player1.MagicPetfigure
			guildInfo.Members[i].WeaponStartLvl = c.Player1.WeaponStartLvl
			guildInfo.Members[i].EqStartLvl = c.Player1.EqStartLvl
			guildInfo.Members[i].TitleOn = c.Player1.TitleOn
			guildInfo.Members[i].InfoUpdateTime = c.Player1.InfoUpdateTime
			guildInfo.Members[i].Vip = c.Player1.Vip
			guildInfo.Members[i].SetOnline(true)
			isChg = true
		}
	}
	// 用作军团长改名字
	if isChg && guildInfo.Base.LeaderAcid == c.Player1.AccountID {
		guildInfo.Base.LeaderName = c.Player1.Name
	}

	g.guild.CheckGuildChief()
	g.guild.ActBoss.UpdateInfo(&c.Player1)

	if isChg {
		if err := g.saveGuild(guildInfo); err != nil {
			c.resChan <- genErrRes(Err_DB)
			return
		}
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}

	return
}

// 玩家等级，活跃度等信息同步到公会，实时影响公会列表成员信息状态的显示
// 同时会触发公会活跃度排行榜的更新
func (r *GuildModule) Sign(guildUUID, acID string, contribution int64, nowT int64) (bool, *GuildInfo) {
	c := guildCommand{
		Type: Command_Sign,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	c.Player1.AccountID = acID
	c.Player1.Contribution[0] = contribution
	c.Player1.Contribution[1] = nowT
	res := r.guildCommandExec(c)
	return !res.ret.HasError(), &res.guildInfo
}

func (r *GuildModule) AddXp(guildUUID, acID string, xp int64) {
	c := guildCommand{
		Type: Command_AddXP,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	c.Player1.AccountID = acID
	c.AddXP = xp
	r.guildCommandExecAsyn(c)
}

func (g *GuildWorker) addXP(c *guildCommand) {
	guildInfo := g.guild
	logs.Trace("guild addXP %v", c)

	guildInfo.AddXP(c.AddXP)
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}

func (r *GuildModule) AddSp(guildUUID, acID string, xp int64) {
	c := guildCommand{
		Type: Command_AddSP,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	c.Player1.AccountID = acID
	c.AddXP = xp
	r.guildCommandExecAsyn(c)
}

func (g *GuildWorker) addSP(c *guildCommand) {
	guildInfo := g.guild
	logs.Trace("guild addSP %v", c)

	guildInfo.AddSP(c.Player1.AccountID, c.AddXP)
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}

func (g *GuildWorker) sign(c *guildCommand) {
	guildInfo := g.guild
	logs.Trace("sign %v", c)

	for i := 0; i < len(guildInfo.Members); i++ {
		if guildInfo.Members[i].AccountID == c.Player1.AccountID {
			m := &guildInfo.Members[i]
			if !gamedata.IsSameDayCommon(m.Contribution[1], c.Player1.Contribution[1]) {
				m.Contribution[0] = 0
			}
			m.Contribution[1] = c.Player1.Contribution[1]
			m.Contribution[0] += c.Player1.Contribution[0]
		}
	}
	// 表结构更改, 此处不续加XP,增加都有givebysync实现
	//guildInfo.AddXP(c.Player1.Contribution[0])
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}

func (r *GuildModule) DebugOp(guildUUID, acID string, p1, p2, p3 int64, p4 string) (bool, *GuildInfo) {
	c := guildCommand{
		Type: Command_Debug,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	c.Player1.AccountID = acID
	c.Player1.Contribution[0] = p1
	c.Player1.Contribution[1] = p2
	c.Player1.InfoUpdateTime = p3
	c.Player1.Name = p4
	res := r.guildCommandExec(c)
	return !res.ret.HasError(), &res.guildInfo
}

func (g *GuildWorker) debugOp(c *guildCommand) {
	guildInfo := g.guild
	logs.Trace("sign %v", c)

	p1 := c.Player1.Contribution[0]
	p2 := c.Player1.Contribution[1]
	p3 := c.Player1.InfoUpdateTime
	p4 := c.Player1.Name

	logs.Trace("debugOp %d, %d, %d, %s", p1, p2, p3, p4)
	switch p1 {
	case 0:
		// Add Contribution
		guildInfo.AddXP(p2)
	}

	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}
