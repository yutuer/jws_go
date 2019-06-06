package logics

func (s *SyncResp) mkSurplusGachaInfo(p *Account) {
	nowTime := p.Profile.GetProfileNowTime()
	surplusInfo := p.Profile.GetHeroSurplusInfo()
	surplusInfo.TryDailyReset(nowTime)

	if nowTime < surplusInfo.EndTime {
		s.SurplusGachaEndTime = surplusInfo.EndTime
		s.SurplusDrawGachaCount = surplusInfo.DailyDrawCount[:]
		s.SurplusGachaFirstOpen = surplusInfo.DailyFirstOpen2Client
		surplusInfo.DailyFirstOpen2Client = false
	}
}
