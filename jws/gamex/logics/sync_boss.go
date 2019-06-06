package logics

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

//"vcs.taiyouxi.net/jws/gamex/models/gamedata"

func (s *SyncResp) mkBossInfo(p *Account) {
	if s.boss_fight_need_sync {
		playerBoss := p.Profile.GetBoss()
		s.SyncBossIDs = make([]string, 0, gamedata.MaxDegree)
		s.SyncBossCount, s.SyncBossCountRefTime = p.Profile.GetCounts().Get(
			gamedata.CounterTypeBoss, p.Account)
		s.SyncBossRewardsCount = gamedata.BossMaxReward
		s.SyncBossRewardIDs = make([]string, 0, gamedata.MaxDegree*gamedata.BossMaxReward)
		s.SyncBossRewardCounts = make([]uint32, 0, gamedata.MaxDegree*gamedata.BossMaxReward)
		s.SyncBossMaxDegree = playerBoss.MaxDegree

		for i := 0; i < len(playerBoss.Bosses); i++ {
			boss := playerBoss.Bosses[i]
			s.SyncBossIDs = append(s.SyncBossIDs, boss.BossTyp)
			for j := 0; j < gamedata.BossMaxReward; j++ {
				if j < len(boss.RewardIDs) {
					s.SyncBossRewardIDs = append(s.SyncBossRewardIDs, boss.RewardIDs[j])
					s.SyncBossRewardCounts = append(s.SyncBossRewardCounts, boss.RewardCounts[j])
				} else {
					s.SyncBossRewardIDs = append(s.SyncBossRewardIDs, "")
					s.SyncBossRewardCounts = append(s.SyncBossRewardCounts, 0)
				}
			}
		}
	}
}
