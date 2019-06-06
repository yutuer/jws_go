package market_activity

import (
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Activity_Rank                = 66
	Activity_RankDestiny         = 67
	Activity_RankStone           = 68
	Activity_RankArmStar         = 69
	Activity_RankHeroStar        = 70
	Activity_RankPlayerLevel     = 71
	Activity_RankPlayerGs        = 72
	Activity_RankHeroDestiny     = 78
	Activity_RankHeroSwingStarLv = 79
	Activity_RankExclusiveWeapon = 80
	Activity_RankWuShuangGs      = 81
	Activity_RankHeroJadeTwo     = 83
	Activity_RankTwo             = 82
	Activity_RankAstrology       = 77
)

var tablename map[uint32]string = map[uint32]string{
	Activity_RankPlayerGs:        "SnapShootRankPlayerGs",
	Activity_RankHeroStar:        "SnapShootRankHeroStar",
	Activity_RankPlayerLevel:     "SnapShootRankPlayerLevel",
	Activity_RankDestiny:         "SnapShootRankDestiny",
	Activity_RankStone:           "SnapShootRankStone",
	Activity_RankArmStar:         "SnapShootRankArmStar",
	Activity_RankHeroDestiny:     "SnapShootRankHeroDestiny",
	Activity_RankHeroSwingStarLv: "SnapShootRankHeroSwingStarLv",
	Activity_RankHeroJadeTwo:     "SnapShootRankHeroJadeTwo",
	Activity_RankExclusiveWeapon: "SnapShootRankExclusiveWeapon",
	Activity_RankWuShuangGs:      "SnapShootRankWuShuangGs",
	Activity_RankAstrology:       "SnapShootRankAstrology",
}

const (
	Redis_ZADD_Banch = 100
	RankTopSize      = rank.RankTopSize
)

var ActivityList map[uint32][]uint32 = map[uint32][]uint32{
	Activity_Rank: {
		Activity_RankDestiny,
		Activity_RankStone,
		Activity_RankArmStar,
		Activity_RankHeroStar,
		Activity_RankPlayerLevel,
		Activity_RankPlayerGs,
	},
	Activity_RankTwo: {
		Activity_RankHeroSwingStarLv,
		Activity_RankHeroJadeTwo,
		Activity_RankExclusiveWeapon,
		Activity_RankWuShuangGs,
		Activity_RankAstrology,
		Activity_RankHeroDestiny,
	},
}

//Debug
func init() {
	//DebugTest()
}

//Debug
func DebugTest() {
	go func() {
		<-time.After(20 * time.Second)

		GetModule(10).NotifyHotDataUpdate(nil)

		<-time.After(3 * time.Second)
		acid := "0:10:6e99cf7e-9e5b-4419-8e20-25d60af32538"
		ac, _ := db.ParseAccount(acid)
		info, _ := simple_info.LoadAccountSimpleInfoProfile(ac)
		GetModule(10).GetRank(Activity_RankHeroStar, acid, info)
		//logs.Debug("[MarketActivityModule] Test GetRank Res: %v", res)

		<-time.After(3 * time.Second)
		logs.Debug("[MarketActivityModule] Test Trigger")
		cfg := gamedata.HotActivityInfo{}
		cfg.ActivityType = Activity_RankHeroStar
		cfg.ActivityId = 97
		cfg.EndTime = 10
		GetModule(10).maTimeSet.refresh(&cfg, 0)

		<-time.After(3 * time.Second)
		cfg.EndTime = 0
		GetModule(10).maTimeSet.refresh(&cfg, 0)

		<-time.After(3 * time.Second)
		GetModule(10).maTimeSet.refresh(&cfg, 0)
		GetModule(10).GetRank(Activity_RankHeroStar, acid, info)

		<-time.After(3 * time.Second)
		GetModule(10).notifyRankParentID(Activity_Rank, 9)
		cfg.EndTime = 3
		GetModule(10).maTimeSet.refresh(&cfg, 0)

		<-time.After(3 * time.Second)
		GetModule(10).notifyRankParentID(Activity_Rank, 92)
	}()

}
