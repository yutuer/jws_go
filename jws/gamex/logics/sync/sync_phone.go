package syncData

import "vcs.taiyouxi.net/jws/gamex/models/account"

type SyncPhone struct {
	IsHasGetPhoneReward int64 `codec:"phoneGet_"`
	PhoneSMSCoolDown    int64 `codec:"phoneCD_"`

	phoneNeedSync bool
}

func (s *SyncPhone) OnChangePhone() {
	s.phoneNeedSync = true
}

func (s *SyncPhone) MkPhoneData(p *account.Account) {
	if s.phoneNeedSync {
		phone := p.Profile.GetPhone()
		nowT := p.Profile.GetProfileNowTime()
		if phone.IsHasBindPhone() {
			s.IsHasGetPhoneReward = 2
		} else {
			s.IsHasGetPhoneReward = 1
		}
		s.PhoneSMSCoolDown = nowT - phone.LastGetCodeTime
	}
}
