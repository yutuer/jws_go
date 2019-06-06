package logics

import "vcs.taiyouxi.net/jws/gamex/models/account/events"

func addHeroLvUpHandle(a *Account) {
	acc := a.Account
	a.AddHandle(events.NewHandler().WithOnHeroLvUp(func(fromLv, toLv, toExp uint32, reason string) {
		profile := &acc.Profile
		// MaxGS可能变化
		profile.GetData().SetNeedCheckMaxGS()
		profile.GetData().SetNeedCheckCompanion(true)
	}))
}
