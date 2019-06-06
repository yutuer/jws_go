package logics

type ExpeditionToClient struct {
	ExpeditionIds     []string
	ExpeditionNames   []string
	ExpeditionState   int32
	ExpeditionAward   int32
	ExpeditionNum     int32
	ExpeditionREstNum int32
}

func (s *SyncResp) mkExpeditionInfo(p *Account) {
	acid := p.AccountID.String()
	now_t := p.Profile.GetProfileNowTime()

	// TODO 整理下面两块逻辑
	if p.Profile.GetExpeditionInfo().IsActive {
		p.Profile.GetExpeditionInfo().LoadEnemyToday(acid,
			int64(p.Profile.GetData().CorpCurrGS_HistoryMax), now_t)
	}

	if p.expeditionFirstActivate() {
		p.Profile.GetExpeditionInfo().LoadEnemyToday(acid,
			int64(p.Profile.GetData().CorpCurrGS_HistoryMax), now_t)
	}

	//远征
	pg := p.Profile.GetExpeditionInfo()
	if s.SyncExpeditionInfoNeed {
		s.ExpeditionState = int64(pg.ExpeditionState)
		s.ExpeditionAvard = int64(pg.ExpeditionAward)
		s.ExpeditionNum = int64(pg.ExpeditionNum)
		s.ExpeditionStep = pg.ExpeditionStep
	}
}
