package logics

import "vcs.taiyouxi.net/jws/gamex/models/account/events"

func addTrialHandle(a *Account) {
	acc := a.Account
	acc.AddHandle(
		events.NewHandler().WithFirstPassStage(func(stageID string) {
			if a.trialFirstActivate() {
				a.Tmp.TrialFirst = true
			}

			now_t := a.Profile.GetProfileNowTime()
			if a.expeditionFirstActivate() {
				a.Tmp.ExpeditionFirst = true
				a.Profile.GetExpeditionInfo().LoadEnemyToday(a.AccountID.String(),
					int64(a.Profile.GetData().CorpCurrGS_HistoryMax), now_t)
			}
		}))
}
