package guild_boss

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/error_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	Command_GuildActBossGetStat
	Command_GuildActBossLock
	Command_GuildActBossUnLock
	Command_GuildActBossBeginFight
	Command_GuildActBossEndFight
	Command_GuildActBossSendActNotify
*/

func (a *ActivityState) LockBoss(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	if len(cmd.ParamStrs) < 2 || cmd.ParamAccountInfo == nil {
		return base.ReturnActCmdError(error_code.ErrCommonParam)
	}
	boss := a.GetBossById(cmd.ParamStrs[0], cmd.ParamStrs[1])
	if boss == nil {
		return base.ReturnActCmdError(error_code.ErrActBossNoFound)
	}
	if a.GetBossStat(boss.Idx) == BossStatUnlocked {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossUnLocked)
	}
	nowT := time.Now().Unix()
	boss.updateChallengeStat(nowT)
	if boss.CurrPlayerAcID != "" {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossHasLocked)
	}

	boss.CurrPlayerState = BossChallengeStatLocked
	boss.CurrPlayerAcID = cmd.ParamAccountInfo.AccountID
	boss.CurrPlayerName = cmd.ParamAccountInfo.Name
	boss.CurrPlayerAvatarID = cmd.ParamAccountInfo.CurrAvatar
	boss.CurrPlayerStopTime = nowT + int64(gamedata.GetGuildBossCfg().GetLoadingSafeTime())

	logs.Trace("LockBoss %v %d", boss, nowT)
	res.SetNeedSync(base.GuildActBoss)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	return res
}

func (a *ActivityState) UnLockBoss(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	if len(cmd.ParamStrs) < 2 || cmd.ParamAccountInfo == nil {
		return base.ReturnActCmdError(error_code.ErrCommonParam)
	}
	boss := a.GetBossById(cmd.ParamStrs[0], cmd.ParamStrs[1])
	if boss == nil {
		return base.ReturnActCmdError(error_code.ErrActBossNoFound)
	}
	if a.GetBossStat(boss.Idx) == BossStatUnlocked {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossUnLocked)
	}
	nowT := time.Now().Unix()
	boss.updateChallengeStat(nowT)
	if boss.CurrPlayerAcID != "" &&
		boss.CurrPlayerAcID != cmd.ParamAccountInfo.AccountID {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossNoLocked)
	}

	boss.CurrPlayerState = BossChallengeStatNormal
	boss.CurrPlayerAcID = ""
	boss.CurrPlayerName = ""
	boss.CurrPlayerAvatarID = 0
	boss.CurrPlayerStopTime = 0

	logs.Trace("UNLockBoss %v %d", boss, nowT)

	res.SetNeedSync(base.GuildActBoss)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	return res
}

func (a *ActivityState) BeginBossFight(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	res.ParamInts = []int64{0}

	if len(cmd.ParamStrs) < 2 || cmd.ParamAccountInfo == nil {
		return base.ReturnActCmdError(error_code.ErrCommonParam)
	}
	boss := a.GetBossById(cmd.ParamStrs[0], cmd.ParamStrs[1])
	if boss == nil {
		return base.ReturnActCmdError(error_code.ErrActBossNoFound)
	}
	if a.GetBossStat(boss.Idx) == BossStatUnlocked {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossUnLocked)
	}
	nowT := time.Now().Unix()
	boss.updateChallengeStat(nowT)
	if boss.CurrPlayerAcID != cmd.ParamAccountInfo.AccountID &&
		boss.CurrPlayerState != BossChallengeStatLocked {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossHasLocked)
	}

	boss.CurrPlayerState = BossChallengeStatFighting
	boss.CurrPlayerStopTime = nowT + boss.LevelTime + 10
	logs.Trace("boss %v %d", boss, nowT)

	res.ParamInts[0] = int64(boss.Idx)

	res.SetNeedSync(base.GuildActBoss)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	a.GetGuildHandler().SetNeedSave2DB()
	return res
}

func (a *ActivityState) EndBossFight(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	res.ParamStrs = []string{""}
	res.ParamInts = []int64{0, 0, 0}

	if len(cmd.ParamStrs) < 2 ||
		len(cmd.ParamInts) < 1 ||
		cmd.ParamAccountInfo == nil {
		return base.ReturnActCmdError(error_code.ErrCommonParam)
	}
	boss := a.GetBossById(cmd.ParamStrs[0], cmd.ParamStrs[1])
	hpDamage := cmd.ParamInts[0]
	if boss == nil {
		return base.ReturnActCmdError(error_code.ErrActBossNoFound)
	}
	if a.GetBossStat(boss.Idx) == BossStatUnlocked {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossUnLocked)
	}
	nowT := time.Now().Unix()
	boss.updateChallengeStat(nowT)
	if boss.CurrPlayerAcID != cmd.ParamAccountInfo.AccountID &&
		boss.CurrPlayerState != BossChallengeStatFighting {
		return base.ReturnActCmdError(error_code.ErrActBossCurrBossNoLocked)
	}

	logs.Trace("EndBossFight %v %d", boss, nowT)

	if hpDamage > boss.Hp {
		hpDamage = boss.Hp
	}

	if hpDamage < 0 {
		hpDamage = 0
	}

	_, realGbCount := a.OnDamage(boss, hpDamage)

	boss.MVPRank.OnPlayerSorce(cmd.ParamAccountInfo, hpDamage)
	a.TodayDamages.OnPlayerSorceAdd(cmd.ParamAccountInfo, hpDamage)
	logs.Debug("debug damange rank", a.TodayDamages)

	boss.CurrPlayerState = BossChallengeStatNormal
	boss.CurrPlayerAcID = ""
	boss.CurrPlayerName = ""
	boss.CurrPlayerAvatarID = 0
	boss.CurrPlayerStopTime = 0

	// 计算奖励
	res.ParamStrs[0] = boss.SelfLoot
	res.ParamInts[0] = 0
	res.ParamInts[1] = boss.Hp
	res.ParamItemC = boss.ItemC

	a.SetAct(cmd.ParamAccountInfo.AccountID, nowT)

	res.SetNeedSync(base.GuildActBoss)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	a.GetGuildHandler().SetNeedSave2DB()

	// logiclog statistic
	a.Statictic.JoinStic(cmd.ParamAccountInfo.AccountID)

	res.ParamInts[2] = int64(realGbCount)
	return res
}

func (a *ActivityState) SendActNotify(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	if len(cmd.ParamStrs) < 2 ||
		len(cmd.ParamInts) < 1 ||
		cmd.ParamAccountInfo == nil {
		return base.ReturnActCmdError(error_code.ErrCommonParam)
	}

	res.SetNeedSync(base.GuildActBoss)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	return res
}

func (a *ActivityState) DebugClean(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	if len(cmd.ParamInts) == 0 || cmd.ParamInts[0] == 0 {
		a.LastRefershTime = 0
	} else {
		a.BossDegree = int(cmd.ParamInts[0])
	}
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	return res
}

func (a *ActivityState) IsBossAllPassed(cmd *base.ActCommand) *base.ActCommand {
	res := new(base.ActCommand)
	res.ParamInts = make([]int64, 1) // 0 BOSS没有全死 1 全死
	if cmd.ParamInts[0] == 0 {
		for _, state := range a.Bosses {
			if state.Hp > 0 {
				res.ParamInts[0] = 0
				break
			}
		}
		res.ParamInts[0] = 1
	} else {
		if a.BigBoss.Hp > 0 {
			res.ParamInts[0] = 0
		} else {
			res.ParamInts[0] = 1
		}
	}
	return res
}
