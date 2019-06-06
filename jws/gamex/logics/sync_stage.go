package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func (s *SyncResp) mkStageInfo(p *Account) {
	profile := &p.Profile
	if s.stage_update_all {
		all := profile.GetStage().GetAll(gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()))
		s.SyncStage = make([][]byte, 0, len(all))
		for i := 0; i < len(all); i++ {
			if all[i].MaxStar <= 0 {
				continue
			}

			s.SyncStage = append(s.SyncStage,
				encode(StageInfo2Client{
					all[i].Id,
					all[i].T_count,
					all[i].T_refresh,
					all[i].Sum_count,
					all[i].MaxStar}))
		}
		s.SyncLastStage = p.Profile.GetStage().GetLastStageId()
	}

	if s.last_stage_update {
		s.SyncLastStage = p.Profile.GetStage().GetLastStageId()
	}

	if s.stage_update != "" {
		s.SyncStage = make([][]byte, 0, 1)

		stage_data := p.Profile.GetStage().GetStageInfo(
			gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()), s.stage_update, p.GetRand())

		s.SyncStage = append(s.SyncStage,
			encode(StageInfo2Client{
				stage_data.Id,
				stage_data.T_count,
				stage_data.T_refresh,
				stage_data.Sum_count,
				stage_data.MaxStar}))

		s.SyncLastStage = p.Profile.GetStage().GetLastStageId()
	}

	if s.chapter_update_all {
		chapters := profile.GetStage().Chapters
		s.SyncChapter = make([][]byte, 0, len(chapters))
		for i := 0; i < len(chapters); i++ {
			ch := chapters[i]
			ch2c := account.Chapter2Client{ch.ChapterId, ch.Star, ch.Has_awardId}
			s.SyncChapter = append(s.SyncChapter, encode(ch2c))
		}
	}

	if s.chapter_update != "" {
		ch := p.Profile.GetStage().GetChapterInfo(s.chapter_update)
		if ch != nil {
			ch2c := account.Chapter2Client{ch.ChapterId, ch.Star, ch.Has_awardId}
			s.SyncChapter = append(s.SyncChapter, encode(ch2c))
		}
	}
}
