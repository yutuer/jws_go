package guild_boss

import (
	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/info"
)

const RankToClient = 5

func (a *ActivityState) mkBoss2Client(nowT int64,
	i int,
	boss2Client *info.ActBossData2Client,
	d *BossState, typ int64) {
	d.updateChallengeStat(nowT)
	boss2Client.BossId = d.ID
	boss2Client.BossType = typ
	boss2Client.BossGroup = d.GroupId
	boss2Client.BossState = a.GetBossStat(i)
	boss2Client.BossEndTime = d.CurrPlayerStopTime
	boss2Client.BossHp = d.Hp
	boss2Client.BossTotalHp = d.TotalHp
	boss2Client.BossIsLock = d.CurrPlayerState == BossChallengeStatLocked ||
		d.CurrPlayerState == BossChallengeStatFighting
	boss2Client.BossPlayerState = int64(d.CurrPlayerState)
	boss2Client.PlayerID = d.CurrPlayerAcID
	boss2Client.PlayerName = d.CurrPlayerName
	boss2Client.PlayerAvatarID = d.CurrPlayerAvatarID
	boss2Client.RewardCount = int64(d.PartRewardDroped)
	boss2Client.RankPlayerIDs = make([]string, 0, RankToClient)
	boss2Client.RankPlayerNames = make([]string, 0, RankToClient)
	boss2Client.RankPlayerScore = make([]int64, 0, RankToClient)
	rank := d.MVPRank.Players[:]
	for j := 0; j < RankToClient && j < len(rank); j++ {
		boss2Client.RankPlayerIDs = append(boss2Client.RankPlayerIDs, rank[j].AccountID)
		boss2Client.RankPlayerNames = append(boss2Client.RankPlayerNames, rank[j].Name)
		boss2Client.RankPlayerScore = append(boss2Client.RankPlayerScore, rank[j].Score)
	}
}

func (a *ActivityState) Refersh(guildUuid, name string, nowT int64) {
	if a.IsNeedRefersh(nowT, gamedata.GetGuildBossRestartTime()) {
		a.Clean(guildUuid, name)
		a.SetHasRefersh(nowT)
		a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	}
}

func (a *ActivityState) ToClient(nowT int64) *info.ActBoss2Client {
	res := new(info.ActBoss2Client)
	res.CurrBossLevel = int64(a.BossDegree)
	res.CurrPlayerNum = int64(a.GetActNum(
		nowT,
		gamedata.GetGuildBossRestartTime()))

	res.BossStats = make([][]byte, 0, len(a.Bosses)+1)
	for i := 0; i < len(a.Bosses); i++ {
		boss2Client := info.ActBossData2Client{}
		a.mkBoss2Client(nowT, i, &boss2Client, &a.Bosses[i], BossTypNormal)
		res.BossStats = append(res.BossStats, codec.Encode(boss2Client))
	}
	bigBoss := info.ActBossData2Client{}
	a.mkBoss2Client(nowT, len(a.Bosses), &bigBoss, &a.BigBoss, BossTypBig)
	res.BossStats = append(res.BossStats, codec.Encode(bigBoss))

	res.DamagePlayerNames = make([]string, 0)
	res.DamagePlayerScore = make([]int64, 0)
	for _, player := range a.TodayDamages.Players {
		if player.AccountID != "" {
			res.DamagePlayerNames = append(res.DamagePlayerNames, player.Name)
			res.DamagePlayerScore = append(res.DamagePlayerScore, player.Score)
		}
	}
	return res
}
