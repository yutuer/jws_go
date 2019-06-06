package guild

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/gate_enemy_push"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildWorker struct {
	m     *GuildModule
	guild *GuildInfo

	waitter      util.WaitGroupWrapper
	command_chan chan guildCommand
}

func (g *GuildWorker) Start() {
	g.command_chan = make(chan guildCommand, 64)

	g.waitter.Wrap(func() {
		tn := 0
		th := 0 // 1 hour
		timer := time.After(time.Second)
		for {
			select {
			case cmd, ok := <-g.command_chan:
				if !ok {
					return
				}

				if g.guild.Base.GuildUUID == "" {
					// 说明已解散，不处理任何命令，并等待GuildMgrWorker关闭
					continue
				}
				func(command guildCommand) {
					logs.Trace("guild %s command %v", g.guild.Base.GuildUUID, command)
					//TODO: by YZH 这个让parent never dead, 应该如此吗？
					defer logs.PanicCatcherWithInfo("GuildWorker Panic")
					g.processCommand(&command)
				}(cmd)
			case <-timer:
				nowT := g.guild.GetDebugNowTime(g.guild.shardId)
				g.Tick(nowT)
				// 兵临活动检测
				s, e := gamedata.GetGETime(nowT)
				if nowT >= s && nowT < e { // 检测兵临活动开启，1s一次
					err_code, ok := g.startGateEnemy()
					if err_code <= 0 && ok {
						gate_enemy_push.GateEnemyStart(g.m.sid,
							g.guild.Base.GuildUUID, e, g.guild.Members[:g.guild.Base.MemNum])
					}
				}
				tn++
				if tn >= 10 { // 检测兵临准备状态，10s一次推送
					if nowT >= s-gamedata.GetGEWaitTimeSec() && nowT < s {
						gate_enemy_push.GateEnemyReady(g.m.sid,
							g.guild.Base.GuildUUID, e, g.guild.Members[:g.guild.Base.MemNum])
					}
					tn = 0
				}
				th++
				if th >= 3600 {
					logs.Debug("autoChangeChief per hour")
					g.guild.TryAutoChangeGuildChief(g.guild.GetDebugNowTime(g.guild.shardId))
					th = 0
				}
				timer = time.After(time.Second)
			}
		}
		logs.Info("GuildWorker command_chan close!")
	})
}

func (g *GuildWorker) Stop() {
	close(g.command_chan)
	g.waitter.Wait()
	if g.guild.saveReqCount > 0 {
		errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			if err := g.guild.ForceDBSave(cb); err != nil {
				return err
			}
			return nil
		})
		if errCode != 0 {
			logs.Error("save guild db when stop %d", errCode)
		} else {
			logs.Info("save guild db ok when stop %s", g.guild.Base.GuildUUID)
		}
	}
}

func (g *GuildWorker) Tick(nowT int64) {
	g.guild.Tick(nowT)
	if g.guild.IsNeedSave2DB() {
		g.guild.SetNoNeedSave2DB()
		if err := g.saveGuild(g.guild); err != nil {
			logs.Error("save guild %v err By %s",
				g.guild.Base.GuildUUID, err.Error())
		}
	}
}

func (g *GuildWorker) saveGuild(guild *GuildInfo) error {
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		err := guild.DBSave(cb)
		if err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		return fmt.Errorf("guild_worker saveGuild err")
	}
	return nil
}

func (g *GuildWorker) delGuild(guuid string) error {
	guildInfo := g.guild
	guid := guildInfo.Base.GuildUUID
	if guid == "" {
		logs.Error("GuildWorker.delGuild2 %v not found", guuid)
		return nil
	}

	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := delGuild(guildInfo, cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		return fmt.Errorf("GuildWorker delGuild err")
	}
	g.m.guildMgrWorker.delGuildUuid(guuid)

	mems := make([]string, len(guildInfo.Members))
	for i := 0; i < len(guildInfo.Members) && i < guildInfo.Base.MemNum; i++ {
		mem := guildInfo.Members[i]
		guildInfo.Members[i] = helper.AccountSimpleInfo{}
		mems[i] = mem.AccountID
	}

	syncGuild2Players(mems, "", "", 0, 0)

	guildInfo.Base.MemNum = 0
	guildInfo.Base.GuildUUID = ""

	go func() {
		g.m.guildCommandExecAsyn(guildCommand{
			Type:     GuildMgr_Cmd_Dismiss_CallBack,
			BaseInfo: guild_info.GuildSimpleInfo{GuildUUID: guid},
		})
	}()
	g.m.applyCommandExecAsyn(applyCommand{
		Type:     Apply_Cmd_Dismiss_Callback,
		BaseInfo: guild_info.GuildSimpleInfo{GuildUUID: guid},
	})
	return nil
}
