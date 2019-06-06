package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func addEnergyUsedHandle(a *Account) {
	acc := a.Account
	a.AddHandle(events.NewHandler().WithEnergyUsed(func(en int64) {
		logs.Info("<energy used>, %d", en)
		if hasMaxStarHero(a) {
			nowTime := acc.Profile.GetProfileNowTime()
			acc.Profile.GetHeroSurplusInfo().TryDailyReset(nowTime)
			acc.Profile.GetHeroSurplusInfo().AddUsedEn(int(en), nowTime)
		}
	}))
}

func hasMaxStarHero(a *Account) bool {
	stars := a.Profile.GetHero().HeroStarLevel
	for _, star := range stars {
		if star == 25 {
			return true
		}
	}
	return false
}
