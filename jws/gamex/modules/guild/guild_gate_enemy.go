package guild

import (
	"time"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/common/guild_player_rank"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GuildModule) UseGateEnemyCount(guildUUID string) bool {
	res := r.guildCommandExec(guildCommand{
		Type: Command_UseGateEnemyCount,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	})
	return !res.ret.HasError()
}

func (g *GuildWorker) useGateEnemyCount(c *guildCommand) {
	err_code, ok := g._useGateEnemyCount()
	if err_code > 0 {
		c.resChan <- genErrRes(err_code)
		return
	}

	if !ok {
		c.resChan <- genErrRes(Err_No_Gate_Enemy_Count)
		return
	}
	c.resChan <- guildCommandRes{}
	return
}

func (g *GuildWorker) startGateEnemy() (int, bool) {
	info := g.guild
	if info.Base.GetInGE() {
		return 0, false
	}
	info.Base.SetInGE(true)
	if err := g.saveGuild(info); err != nil {
		return Err_DB, false
	}
	// 活动开始时，清兵临积分榜
	info.GatesEnemyData.PlayerRank = *(guild_player_rank.NewPlayerSimpleInfoRankByCap(MaxGuildMember))
	return 0, true
}

func (g *GuildWorker) _useGateEnemyCount() (int, bool) {
	info := g.guild
	if info.Base.GetInGE() {
		return 0, false
	}
	if info.UseGateEnemyCount() {
		logs.Trace("useGateEnemyCount %v", info.Base.GuildUUID)
		info.Base.SetInGE(true)
		if err := g.saveGuild(info); err != nil {
			return Err_DB, false
		}
		// 活动开始时，清兵临积分榜
		info.GatesEnemyData.PlayerRank = *(guild_player_rank.NewPlayerSimpleInfoRankByCap(MaxGuildMember))
		return 0, true
	}
	return 0, false
}

func (r *GuildModule) OnGateEnemyStop(guildUUID string,
	id int,
	gateEnemyData *player_msg.GatesEnemyData) (GuildRet, *GuildInfoBase) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_OnGateEnemyStop,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		memSyncReceiverID: id,
		gateEnemyData:     gateEnemyData,
	})
	return res.ret, &res.guildInfo.GuildInfoBase
}

func (g *GuildWorker) onGateEnemyStop(c *guildCommand) {
	info := g.guild
	info.DelMemSyncReceiver(c.memSyncReceiverID)
	if c.gateEnemyData != nil {
		info.GatesEnemyData = *c.gateEnemyData
		info.Base.AddGEPointWeek(int64(c.gateEnemyData.Point))
	}
	info.Base.SetInGE(false)
	// log
	var joinCount int
	for _, sc := range c.gateEnemyData.PlayerRank.Rank.Sorces {
		if sc > 0 {
			joinCount++
		}
	}

	logiclog.LogGuildGEOver(info.Base.GuildUUID, info.Base.GuildID,
		info.Base.Name, joinCount, info.GatesEnemyData.Point, "")
	if err := g.saveGuild(info); err != nil {
		c.resChan <- genErrRes(Err_DB)
	} else {
		c.resChan <- guildCommandRes{
			guildInfo: *info,
		}
	}
}

func (r *GuildModule) SetGateEnemyReward(guildUUID string, acID string) (GuildRet, int) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_SetGateEnemyReward,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
	})
	return res.ret, res.guildInfo.GatesEnemyData.Point
}

func (g *GuildWorker) setGateEnemyReward(c *guildCommand) {
	info := g.guild
	if errCode := setGetRewardTime(&info.GatesEnemyData, c.Player1.AccountID); errCode != 0 {
		logs.Warn("SetGetRewardTime Err By  %s %v, %d", c.Player1.AccountID, info.GatesEnemyData, errCode)
		c.resChan <- genErrRes(errCode)
	} else {
		if err := g.saveGuild(info); err != nil {
			c.resChan <- genErrRes(Err_DB)
		} else {
			c.resChan <- guildCommandRes{
				guildInfo: *info,
			}
		}
	}
}

func setGetRewardTime(p *player_msg.GatesEnemyData, acID string) int {
	nowT := time.Now().Unix()
	if !p.HasReward(acID) {
		return Err_No_Gate_Enemy_Not_Join
	}

	simpleInfo := p.PlayerRank.GetSimpleInfo(p.PlayerRank.GetRank(acID))
	if simpleInfo == nil {
		return Err_No_Gate_Enemy_Not_Join
	}
	if simpleInfo.Other.Pi[0] != 0 {
		return Err_No_Gate_Enemy_Reward // 已经领过奖励了
	}
	simpleInfo.Other.Pi[0] = nowT
	return 0

}

func (r *GuildModule) DebugResetGateEnemy(guildUUID string, acID string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_DebugResetGateEnemy,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
	})
	return res.ret
}

func (g *GuildWorker) debugResetGateEnemy(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	m := info.GetMember(c.Player1.AccountID)
	if m == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}
	info.Base.DebugResetGateEnemyCount()
	c.resChan <- res
}
