package counter

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

// GameMode Counter Type
const (
	CounterTypeNull                  = gamedata.CounterTypeNull
	CounterTypeGoldLevel             = gamedata.CounterTypeGoldLevel
	CounterTypeFineIronLevel         = gamedata.CounterTypeFineIronLevel
	CounterTypeTrial                 = gamedata.CounterTypeTrial
	CounterTypeBoss                  = gamedata.CounterTypeBoss
	CounterTypeDCLevel               = gamedata.CounterTypeDCLevel
	CounterTypeFish                  = gamedata.CounterTypeFish
	CounterTypeGeneralQuest          = gamedata.CounterTypeGeneralQuest
	CounterTypeFishHC                = gamedata.CounterTypeFishHC
	CounterTypeGVE                   = gamedata.CounterTypeGVE
	CounterTypeTeamPvp               = gamedata.CounterTypeTeamPvp
	CounterTypeTeamPvpRefresh        = gamedata.CounterTypeTeamPvpRefresh
	CounterTypeSimplePvp             = gamedata.CounterTypeSimplePvp
	CounterTypeHitHammerDailyLimit   = gamedata.CounterTypeHitHammerDailyLimit
	CounterTypeWorshipTimes          = gamedata.CounterTypeWorshipTimes
	CounterTypeFreeGuildBoss         = gamedata.CounterTypeFreeGuildBoss
	CounterTypeFreeGuildBigBoss      = gamedata.CounterTypeFreeGuildBigBoss
	CounterTypeGuildBossBuyTime      = gamedata.CounterTypeGuildBossBuyTime
	CounterTypeGuildBigBossBuyTime   = gamedata.CounterTypeGuildBigBossBuyTime
	CounterTypeEatBaozi              = gamedata.CounterTypeEatBaozi
	CounterTypeFenHuoFreeExtraReward = gamedata.CounterTypeFengHuoFreeExtraReward
	CounterTypeFenHuoFreeSeniorSweep = gamedata.CounterTypeFengHuoFreeSeniorSweep
	CounterTypeFenHuoMaxSeniorSweep  = gamedata.CounterTypeFengHuoMaxSeniorSweep
	CounterTypeExpedition            = gamedata.CounterTypeExpedition
	CounterTypeFestivalBoss          = gamedata.CounterFestivalBoss
	CounterTypeHeroDiff              = gamedata.ConnterHeroDiff
	CounterTypeWspvpChallenge        = gamedata.CounterTypeWspvpChallenge
	CounterTypeWspvpRefresh          = gamedata.CounterTypeWspvpRefresh
	CounterTypeFBInvitation          = gamedata.CounterTypeFBInvitation
	CounterTypeWBoss                 = gamedata.ConterTypeWBoss
	CounterTypeCountMax              = gamedata.CounterTypeCountMax
)

type UpdateData helper.IAccount
