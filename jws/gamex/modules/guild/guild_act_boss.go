package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
)

func (r *GuildModule) ActBossBeginFight(
	guildID, bossID, groupID string,
	player *helper.AccountSimpleInfo) (int, int64) {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Type:      Command_GuildActBossBeginFight,
		ParamStrs: []string{bossID, groupID},
		Player1:   *player,
	})
	return res.ret.ErrCode, res.ResInt[0]
}

func (r *GuildModule) ActBossEndFight(
	guildID, bossID, groupID string, damage int64,
	player *helper.AccountSimpleInfo) (int, int64, string, int64, map[string]uint32, int64) {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Type:      Command_GuildActBossEndFight,
		ParamStrs: []string{bossID, groupID},
		ParamInts: []int64{damage},
		Player1:   *player,
	})

	return res.ret.ErrCode, res.ResInt[0], res.ResStr[0], res.ResInt[1], res.ResItemC, res.ResInt[2]
}

func (r *GuildModule) ActBossLockFight(
	guildID, bossID, groupID string,
	player *helper.AccountSimpleInfo) int {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Type:      Command_GuildActBossLock,
		ParamStrs: []string{bossID, groupID},
		Player1:   *player,
	})
	return res.ret.ErrCode
}

func (r *GuildModule) ActBossUnlockFight(
	guildID, bossID, groupID string,
	player *helper.AccountSimpleInfo) int {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Type:      Command_GuildActBossUnLock,
		ParamStrs: []string{bossID, groupID},
		Player1:   *player,
	})
	return res.ret.ErrCode
}

func (r *GuildModule) ActBossNotify(
	guildID, bossID, groupID string,
	player *helper.AccountSimpleInfo) int {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Type:      Command_GuildActBossSendActNotify,
		ParamStrs: []string{bossID, groupID},
		Player1:   *player,
	})
	return res.ret.ErrCode
}

func (r *GuildModule) ActBossDebugClean(
	guildID string, accountID string, degree int64) int {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		Player1: helper.AccountSimpleInfo{
			AccountID: accountID,
		},
		ParamInts: []int64{degree},
		Type:      Command_GuildActBossDebugClean,
	})
	return res.ret.ErrCode
}

func syncGuildBoss2Players(acids []string, g *GuildInfo, nowT int64) {
	pInfo := g.ActBoss.ToClient(nowT)
	player_msg.SendToPlayers(acids, player_msg.PlayerMsgGuildBossSyncCode,
		player_msg.PlayerGuildBossUpdate{
			GuildUUID: g.Base.GuildUUID,
			Info:      *pInfo,
		})
}

func (r *GuildModule) ActBossIsAllPassed(guildID string, bossType int) (int, int) {
	res := r.guildCommandExec(guildCommand{
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildID,
		},
		ParamInts: []int64{int64(bossType)},
		Type:      Command_GuildActBossIsPassed,
	})
	if res.ret.HasError() {
		return 0, res.ret.ErrCode
	} else {
		return int(res.ResInt[0]), res.ret.ErrCode
	}
}
