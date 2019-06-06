package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
)

func addTitleHandle(a *Account) {
	acc := a.Account
	acc.AddHandle(
		events.NewHandler().WithTitleOnChg(func(oldTitle, newTitle string) {
			simpleInfo := a.GetSimpleInfo()
			lv, _ := a.Profile.GetCorp().GetXpInfo()
			if lv >= FirstIntoCorpLevel {
				rank.GetModule(a.AccountID.ShardId).RankCorpGs.Add(&simpleInfo,
					int64(simpleInfo.CurrCorpGs), int64(simpleInfo.CurrCorpGs))
			}
			guild.GetModule(a.AccountID.ShardId).UpdateAccountInfo(simpleInfo)
		}))
}
