package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	ApplyWorker公会外玩家进入公会的操作都在这里，创建公会，申请公会
	1、申请导致的红点变化，会发给GuildWorker处理，异步的
	2、ApplyWorker需要查询公会信息时，会去GuildWorker要并同步等待，同步
	3、ApplyWorker缓存了公会的简单信息，需要GuildWorker变化时更新ApplyWorker，异步的
	4、公会解散GuildWorker会更新ApplyWorker，异步的
	5、创建公会直接在db里；加人要同步调用GuildWorker，同步
*/
type ApplyWorker struct {
	m           *GuildModule
	playerApply map[string]*PlayerApply // 玩家->公会
	guildApply  map[string]*GuildApply  // 公会的申请列表

	waitter      util.WaitGroupWrapper
	command_chan chan applyCommand
}

type applyCommand struct {
	shard         uint
	Type          int
	Applicant     helper.AccountSimpleInfo
	Approver      helper.AccountSimpleInfo
	BaseInfo      guild_info.GuildSimpleInfo
	Channel       string
	AssignID      []string
	AssignTimes   []int64
	LastLeaveTime int64
	resChan       chan guildCommandRes
}

func (aw *ApplyWorker) Start(shard uint) {
	aw.playerApply = make(map[string]*PlayerApply, 2048)
	aw.guildApply = make(map[string]*GuildApply, 2048)
	aw.command_chan = make(chan applyCommand, 2048)

	aw.waitter.Wrap(func() {
		for command := range aw.command_chan {
			logs.Trace("guild apply command %v", command)
			command.shard = shard
			func() {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("GuildApplyWorker Panic")
				aw.processCommand(&command)
			}()
		}
		logs.Warn("GuildApplyWorker command_chan close!")
	})
}

func (aw *ApplyWorker) Stop() {
	//	close(aw.command_chan)
	//	aw.waitter.Wait()
}

const (
	Apply_Cmd_Null = iota
	Apply_Cmd_NewGuild
	Apply_Cmd_ApplyGuild
	Apply_Cmd_CancelApplyGuild
	Apply_Cmd_ApproveApply
	Apply_Cmd_DelApply
	Apply_Cmd_GetPlayerApplyList
	Apply_Cmd_GetGuildApplyList
	Apply_Cmd_UpdateAccountInfoApply
	Apply_Cmd_GuildInfo_Update
	Apply_Cmd_Dismiss_Callback

	Apply_Cmd_Count
)

func (aw *ApplyWorker) processCommand(c *applyCommand) {
	// load data
	if c.Type == Apply_Cmd_Dismiss_Callback {
		if c.BaseInfo.GuildUUID != "" {
			if err := aw.loadGuildApply(c.BaseInfo.GuildUUID); err != nil {
				logs.Error("ApplyWorker processCommand load guildApply err %v", err)
				c.resChan <- genErrRes(Err_DB)
				return
			}
			aw.loadGuildApplyPlayer(c.BaseInfo.GuildUUID)
		}
	} else {
		if c.Applicant.AccountID != "" {
			if _, ok := aw.playerApply[c.Applicant.AccountID]; !ok {
				if err := aw.loadPlayerApply(c.Applicant.AccountID); err != nil {
					logs.Error("ApplyWorker processCommand load playerApply err %v", err)
					c.resChan <- genErrRes(Err_DB)
					return
				}
				aw.loadPlayerApplyGuild(c.Applicant.AccountID)
			}
		}
		if c.BaseInfo.GuildUUID != "" {
			if _, ok := aw.guildApply[c.BaseInfo.GuildUUID]; !ok {
				if err := aw.loadGuildApply(c.BaseInfo.GuildUUID); err != nil {
					logs.Error("ApplyWorker processCommand load guildApply err %v", err)
					c.resChan <- genErrRes(Err_DB)
					return
				}
				aw.updateGuildInfo(c.BaseInfo.GuildUUID)
			}
		}
	}
	switch c.Type {
	case Apply_Cmd_NewGuild:
		aw.newGuild(c)
	case Apply_Cmd_ApplyGuild:
		aw.applyGuild(c)
	case Apply_Cmd_CancelApplyGuild:
		aw.cancelApplyGuild(c)
	case Apply_Cmd_ApproveApply:
		aw.approveApply(c)
	case Apply_Cmd_DelApply:
		aw.delApply(c)
	case Apply_Cmd_GetPlayerApplyList:
		aw.getPlayerApplyList(c)
	case Apply_Cmd_GetGuildApplyList:
		aw.getGuildApplyList(c)
	case Apply_Cmd_UpdateAccountInfoApply:
		aw.updateAccountInfoApply(c)
	case Apply_Cmd_Dismiss_Callback:
		aw.dismissGuildCallBack(c)
	case Apply_Cmd_GuildInfo_Update:
		aw.updateGuildInfoCmd(c)
	}
}

func (aw *ApplyWorker) loadGuildApply(guid string) error {
	if _, ok := aw.guildApply[guid]; ok {
		return nil
	}

	guildApply := GuildApply{}
	guildApply.Guild.GuildUUID = guid
	err := guildApply.DBLoad()
	if err != nil {
		logs.Error("Load GuildApply %s Err By %s", guid, err.Error())
		return err
	}

	aw.guildApply[guid] = &guildApply
	return nil
}

func (aw *ApplyWorker) loadGuildApplyPlayer(guid string) {
	if ga, ok := aw.guildApply[guid]; ok {
		for i := ga.ApplyNum - 1; i >= 0; i-- {
			aply := ga.ApplyList[i]
			aw.loadPlayerApply(aply.PlayerInfo.AccountID)
		}
	}
}

func (aw *ApplyWorker) updateGuildInfo(guid string) {
	if ga, ok := aw.guildApply[guid]; ok {
		guild, errRet := aw.m.GetGuildInfo(guid)
		if errRet.HasError() {
			return
		}

		ga.updateGuildInfo(guild.Base)
	}
}

func (aw *ApplyWorker) loadPlayerApply(acid string) error {
	if _, ok := aw.playerApply[acid]; ok {
		return nil
	}

	playerApply := PlayerApply{
		AccountId: acid,
	}

	err := playerApply.DBLoad()
	if err != nil {
		logs.Error("Load PlayerApplyInfo %s Err By %s", acid, err.Error())
		return err
	}
	aw.playerApply[acid] = &playerApply
	return nil
}

func (aw *ApplyWorker) loadPlayerApplyGuild(acid string) {
	if playerApply, ok := aw.playerApply[acid]; ok {
		// 将此玩家申请过的公会也load出来
		for i := playerApply.ApplyNum - 1; i >= 0; i-- {
			aply := playerApply.ApplyList[i]
			aw.loadGuildApply(aply.GuildUuid)
		}
	}
}
