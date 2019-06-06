package logics

func (s *SyncResp) mkWspvpInfo(p *Account) {
	if s.SyncWSPVPNeed {
		p.Profile.WSPVPPersonalInfo.TryRefresh(p.GetProfileNowTime())
		p.Profile.WSPVPPersonalInfo.TryRefreshRank(p.GetWSPVPGroupId(), p.AccountID.String())
		info := p.Profile.WSPVPPersonalInfo
		s.WsRank = info.Rank
		s.NotClaimedReward = info.NotClaimedReward
		s.LastRankChangeTime = info.LastRankChangeTime
		s.HasClaimedBox = info.HasClaimedBox
		s.HasChallengeCount = info.HasChallengeCount
		s.BestRank = info.BestRank
		s.HasClaimedBestRank = info.HasClaimedBestRank
		s.DefenseFormation = CheckAndChangeFormation(p.GetDefenseFormation())
	}
}
