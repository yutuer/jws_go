package logics

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type TrialToClient struct {
	MostLevelId  int32 `codec:"ml"`
	CurLevelId   int32 `codec:"cl"`
	BonusLevelId int32 `codec:"busl"`

	SweepEndTime  int64 `codec:"swet"`
	HasSweepAward bool  `codec:"swaw"`

	AllLvlFinish bool `codec:"allfnsh"`
}

func (s *SyncResp) mkTrialAllInfo(p *Account) {
	if s.SyncTrialNeed {
		trial := p.Profile.GetPlayerTrial()
		info := TrialToClient{
			MostLevelId:   trial.MostLevelId,
			CurLevelId:    trial.CurLevelId,
			BonusLevelId:  trial.BonusLevelId,
			HasSweepAward: trial.SweepBeginLvlId > 0,
			SweepEndTime:  trial.SweepEndTime,
			AllLvlFinish:  trial.CurLevelId > gamedata.GetTrialFinalLvlId(),
		}
		s.SyncTrialInfo = encode(info)
	}
}
