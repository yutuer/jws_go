package logics

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
)

//mkFirstPassRewardInfo  SyncResp GetModule
func (s *SyncRespNotify) mkFirstPassRewardInfo(p *Account) {
	// team pvp
	if s.SyncFirstPassReward || s.SyncTeamPvpNeed {
		fpr := p.Profile.GetFirstPassRank()
		bs := time.Now().UnixNano()
		_, score := rank.GetModule(p.AccountID.ShardId).RankSimplePvp.GetPos(p.AccountID.String())
		metric_send(p.AccountID, "RankSPvpGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
		fpr.OnRank(gamedata.FirstPassRankTypSimplePvp, int(score/rank.SimplePvpScorePow))

		s.SyncTeamPvFirstPassReward =
			fpr.RewardStat[gamedata.FirstPassRankTypTeamPvp]
		s.SyncSimplePvpFirstPassReward =
			fpr.RewardStat[gamedata.FirstPassRankTypSimplePvp]
		s.SyncSimplePvpMaxRanks = fpr.MaxRank[gamedata.FirstPassRankTypSimplePvp]
		s.SyncTeamPvpMaxRank = fpr.MaxRank[gamedata.FirstPassRankTypTeamPvp]
	}
}

func (s *SyncRespNotify) OnChangeFirstPassRewardInfo() {
	s.SyncFirstPassReward = true
	s.SyncTeamPvpNeed = true
}
