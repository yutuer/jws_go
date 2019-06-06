package logics

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (s *SyncRespNotify) mkTitleInfo(p *Account) {
	// title
	if s.SyncTitleNeed {
		mTitle := p.Profile.GetTitle()
		mTitle.UpdateTitle(p.Account, p.Profile.GetProfileNowTime())

		logs.Debug("SyncTitle after UpdateTitle, {%v}", mTitle)
		s.SyncTitleCanActivate = make([]string, 0, len(mTitle.TitleCanActivate))
		for t, _ := range mTitle.TitleCanActivate {
			s.SyncTitleCanActivate = append(s.SyncTitleCanActivate, t)
		}
		s.SyncTitles = mTitle.GetTitles()
		s.SyncTitleOn = mTitle.TitleTakeOn
		s.SyncTitleNextRefTime = mTitle.GetNextRefTime()
		s.SyncTitleHint = make([]string, 0, len(mTitle.TitleForClient))
		for t, _ := range mTitle.TitleForClient {
			s.SyncTitleHint = append(s.SyncTitleHint, t)
		}
	}
}
