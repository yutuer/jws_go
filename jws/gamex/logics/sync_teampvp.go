package logics

func (s *SyncRespNotify) mkTeamPvpInfo(p *Account) {
	// team pvp
	if s.SyncTeamPvpNeed {
		tpvp := p.Profile.TeamPvp
		s.SyncTeamPvpRank = tpvp.Rank
		s.SyncTeamPvpAvatars = tpvp.FightAvatars
		s.SyncTeamPvpCountToday = tpvp.PvpCountToday
		s.SyncTeamPvpOpenedChests = make([]uint32, 0, len(tpvp.OpenedChestIDs))
		for _, id := range tpvp.OpenedChestIDs {
			s.SyncTeamPvpOpenedChests = append(s.SyncTeamPvpOpenedChests, id)
		}
	}
}
