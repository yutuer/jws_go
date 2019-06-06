package guild

import (
	"fmt"
	"sort"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/notifycsrob"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (g *GuildInfo) TryAutoChangeGuildChief(nowTime int64) bool {
	logs.Debug("autoChangeChief: try auto change guild, %d, %v, %d, %d, %d", g.Base.GuildID,
		g.GuildChangeChief, gamedata.ChiefMaxAbsentTime, gamedata.GuildSafeTimeOnAwake, gamedata.ChiefMaxAbsentTime)
	if awakeFromSleep := g.AwakeIfSleep(nowTime); awakeFromSleep {
		return true
	}
	if g.canChangeChief(nowTime) {
		logs.Debug("autoChangeChief: do change chief %s", g.Base.GuildID)
		ok := g.doChangeChief(nowTime)
		if !ok {
			g.GuildChangeChief.IsSleep = true // 没有合适的军团长, 公会进入睡眠状态
			logs.Info("autoChangeChief: change guild status from awake to sleep %s", g.Base.GuildID)
		}
		return true
	}
	return false
}

func (g *GuildInfo) AwakeIfSleep(nowTime int64) bool {
	if g.GuildChangeChief.IsSleep {
		// 有人登陆可以唤醒公会
		g.GuildChangeChief.IsSleep = false
		g.GuildChangeChief.LastUpdateTime = nowTime
		logs.Info("autoChangeChief: change guild status from sleep to awake, %s", g.Base.GuildID)
		return true
	}
	return false
}

func (g *GuildInfo) canChangeChief(nowTime int64) bool {
	chief := g.GetGuildChief()
	if chief == nil {
		logs.Error("autoChangeChief: no chief in guild %s, %s", g.GuildInfoBase.Base.Name, g.GuildInfoBase.Base.GuildUUID)
		return false
	}
	return !g.GuildChangeChief.IsSleep &&
		nowTime-g.GuildChangeChief.LastUpdateTime > gamedata.GuildSafeTimeOnAwake &&
		!chief.GetOnline() &&
		nowTime-chief.LastLoginTime > gamedata.ChiefMaxAbsentTime
}

func (g *GuildInfo) doChangeChief(nowTime int64) bool {
	newChief, ok := g.findNewChief(nowTime)
	if !ok {
		return false
	}
	logs.Debug("autoChangeChief: find new chief ", g.Base.GuildID, newChief.Name)
	oldChief := g.swapChief(newChief)
	g.sendChangeChiefMail(oldChief)
	return true
}

func (g *GuildInfo) findNewChief(nowTime int64) (*helper.AccountSimpleInfo, bool) {
	candidates := make(guild_info.GuildSortMember, 0, len(g.Members)) // 候选人
	for i, member := range g.Members[:g.Base.MemNum] {
		if member.GuildPosition != gamedata.Guild_Pos_Chief && (member.GetOnline() || nowTime-member.LastLoginTime <= gamedata.ChiefMaxAbsentTime) {
			candidates = append(candidates, &g.Members[i])
		} else {
			logs.Debug("autoChangeChief, not a candidate: %s, position=%d, absent=%d", member.Name, member.GuildPosition, nowTime-member.LastLoginTime)
		}
	}
	if len(candidates) < 1 {
		logs.Debug("autoChangeChief: fail to find any candidates, guilduuid: %s", g.Base.GuildUUID)
		return nil, false
	} else {
		logs.Debug("autoChangeChief: find %d candidates", len(candidates))
	}
	sort.Sort(candidates)
	return candidates[0], true
}

// 返回oldChief
func (g *GuildInfo) swapChief(newChief *helper.AccountSimpleInfo) *helper.AccountSimpleInfo {
	oldChief := g.GetGuildChief()
	newPosition := newChief.GuildPosition // 新会长之前的位置
	newChief.GuildPosition = gamedata.Guild_Pos_Chief
	oldChief.GuildPosition = gamedata.Guild_Pos_Mem
	g.posNum[newPosition]--
	g.posNum[gamedata.Guild_Pos_Mem]++
	g.onSwapChief(oldChief, newChief)
	g.GuildLog.AddLog(guild_info.IDS_GUILD_LOG_7, []string{newChief.Name})
	return oldChief
}

func (g *GuildInfo) sendChangeChiefMail(oldChief *helper.AccountSimpleInfo) {
	for i, member := range g.Members {
		if member.AccountID == "" {
			continue
		}
		sid, _ := guild_info.GetShardIdByGuild(g.Base.GuildUUID)
		oldChiefAbsentHour := (time.Now().Unix() - oldChief.LastLoginTime) / 3600
		params := []string{oldChief.Name, fmt.Sprintf("%d", oldChiefAbsentHour), g.GetGuildChief().Name}
		mail_sender.BatchSendChangeChiefMail(sid, member.AccountID, params, int64(i))
	}
}

func (g *GuildInfo) OnGuildBossDied(bossName string, itemIds []string, count []uint32) uint32 {
	var realGbCount uint32 = 0
	scienceIdx := gamedata.GST_Typ(gamedata.GST_BossFight)
	sc := &g.Sciences[int(scienceIdx)]
	bonusArray := gamedata.GetGuildScienceBonus(scienceIdx, sc.Lvl)
	bonus := bonusArray[0]
	// 军魂奖励受科技加成
	resCount := make([]uint32, len(count))
	for i := 0; i < len(itemIds); i++ {
		if itemIds[i] == gamedata.VI_GuildBoss {
			resCount[i] = uint32(float32(count[i]) * (1 + bonus))
			realGbCount += resCount[i]
		}
	}
	for i, member := range g.Members[:g.Base.MemNum] {
		if member.AccountID == "" {
			continue
		}
		mail_sender.BatchSendGuildBossDeath(g.shardId, member.AccountID, bossName, itemIds, resCount, int64(i))
	}
	return realGbCount
}

func (g *GuildInfo) onSwapChief(oldChief, newChief *helper.AccountSimpleInfo) {
	syncGuild2Player(oldChief.AccountID, g.Base.GuildUUID, g.Base.Name, gamedata.Guild_Pos_Mem, int(g.Base.Level))
	syncGuild2Player(newChief.AccountID, g.Base.GuildUUID, g.Base.Name, gamedata.Guild_Pos_Chief, int(g.Base.Level))
	g.UpdateChiefInfo(newChief.AccountID, newChief.Name)
}

func (g *GuildInfo) UpdateChiefInfo(acid, name string) {
	g.Base.LeaderAcid = acid
	g.Base.LeaderName = name
	g.GuildChangeChief.LastUpdateTime = g.GetDebugNowTime(g.shardId)

	notifycsrob.Call(g.Base.GuildUUID)
}

func (g *GuildModule) DebugAutoChangeChief(guildUUID string, acID string) GuildRet {
	res := g.guildCommandExec(guildCommand{
		Type: Command_DebugAutoChangeChief,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
	})
	return res.ret
}

func (g *GuildInfo) debugAutoChangeChief() {
	logs.Debug("debug auto change chief, %s", g.Base.Name)
	g.TryAutoChangeGuildChief(g.GetDebugNowTime(g.shardId))
}
