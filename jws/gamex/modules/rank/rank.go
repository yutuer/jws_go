package rank

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/push"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/jws/gamex/models/helper"

	"vcs.taiyouxi.net/jws/gamex/modules/balance_timer"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/rank/award"

	"time"

	"runtime"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/global_count"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	rankId_RankByCorpGS = 10 + iota
	rankId_RankByCorpGSYesterday
	rankId_BossTotalLow
	rankId_BossTotalLowYesterday
	rankId_BossTotalHigh
	rankId_BossTotalHighYesterday
	rankId_BossSingleLow
	rankId_BossSingleLowYesterday
	rankId_BossSingleHigh
	rankId_BossSingleHighYesterday
	rankId_SimplePvp
	rankId_GuildAct
	rankId_GuildActWeek
	rankId_GuildActWeekLast
	rankId_GuildGS
	rankId_GuildGateEnemy
	rankId_RankByCorpTrial
	rankId_RankByCorpGS_ServerOpen
	rankId_RankByGuildGS_ServerOpen
	rankId_RankByHeroStar
	rankId_RankByHeroDiff_TU
	rankId_RankByHeroDiff_ZHAN
	rankId_RankByHeroDiff_HU
	rankId_RankByHeroDiff_SHI
	rankId_RankByDestiny
	rankId_RankByJade
	rankId_RankByEquipStarLv
	rankId_RankByCorpLv
	rankId_RankByWingStar
	rankId_RankByHersoDestiny
	rankId_RankByHeroJadeTwo
	rankId_RankByWuShuangGs
	rankId_RankByExclusiveWeapon
	rankId_RankByAstrology
	rankId_RankByCorpOfWei
	rankId_RankByCorpOfShu
	rankId_RankByCorpOfWu
	rankId_RankByCorpOfQunXiong
)

const (
	Ex_RankId_RankByDestiny     = rankId_RankByDestiny
	Ex_RankId_RankByJade        = rankId_RankByJade
	Ex_RankId_RankByEquipStarLv = rankId_RankByEquipStarLv
	Ex_RankId_RankByCorpLv      = rankId_RankByCorpLv
	Ex_RankId_RankByCorpGS      = rankId_RankByCorpGS
	Ex_RankId_RankByHeroStar    = rankId_RankByHeroStar
	Ex_RankId_RankBySwingStarLv = rankId_RankByWingStar
	Ex_RankId_HeroDestinyLv     = rankId_RankByHersoDestiny
	Ex_RankId_HeroByJadeTwo     = rankId_RankByHeroJadeTwo
	Ex_RankId_HeroByWuShuangGs  = rankId_RankByWuShuangGs
	Ex_RankId_Astrology         = rankId_RankByAstrology
	Ex_RankId_ExclusiveWeapon   = rankId_RankByExclusiveWeapon
)

func genRankModule(sid uint) *RankModule {
	m := &RankModule{
		sid:                   sid,
		RankCorpGs:            RankByCorpDelay{},
		RankSimplePvp:         RankByCorpDynamic{},
		RankCorpGsSevOpn:      RankByCorpStatic{},
		RankByCorpTrial:       RankByCorp{},
		RankByHeroStar:        RankByCorp{},
		RankGuildGs:           RankByGuildDelay{},
		RankGuildGateEnemy:    RankByGuild{},
		RankGuildSevOpn:       RankByGuildStatic{},
		RankByHeroDiff:        [gamedata.HeroDiff_Count]RankByCorp{},
		RankByDestiny:         RankByCorp{},
		RankByJade:            RankByCorp{},
		RankByEquipStarLv:     RankByCorp{},
		RankByCorpLv:          RankByCorp{},
		RankByHeroJadeTwo:     RankByCorp{},
		RankByHeroWuShuangGs:  RankByCorp{},
		RankByExclusiveWeapon: RankByCorp{},
		RankByHeroDestiny:     RankByCorp{},
		RankByWingStar:        RankByCorp{},
		RankByAstrology:       RankByCorp{},
	}
	m.sevenDayRankStartTime = game.ServerStartTime(sid)
	return m
}

type RankModule struct {
	sid    uint
	rankdb *rankDB

	RankCorpGs             RankByCorpDelay
	RankSimplePvp          RankByCorpDynamic
	RankCorpGsSevOpn       RankByCorpStatic
	RankByCorpTrial        RankByCorp
	RankByHeroStar         RankByCorp
	RankByDestiny          RankByCorp
	RankByJade             RankByCorp
	RankByEquipStarLv      RankByCorp
	RankByCorpLv           RankByCorp
	RankByHeroDiff         [gamedata.HeroDiff_Count]RankByCorp
	RankGuildGs            RankByGuildDelay
	RankGuildGateEnemy     RankByGuild
	RankGuildSevOpn        RankByGuildStatic
	RankByWingStar         RankByCorp
	RankByHeroDestiny      RankByCorp
	RankByHeroJadeTwo      RankByCorp
	RankByHeroWuShuangGs   RankByCorp
	RankByAstrology        RankByCorp
	RankByExclusiveWeapon  RankByCorp
	RankByCorpGsOfWei      RankByCorpDelay
	RankByCorpGsOfShu      RankByCorpDelay
	RankByCorpGsOfWu       RankByCorpDelay
	RankByCorpGsOfQunXiong RankByCorpDelay
	sevenDayRankStartTime  int64
}

func (r *RankModule) BeforeStop() {
}

func (r *RankModule) Start() {
	r.rankdb = &rankDB{}

	r.buildRankCorpGs()
	r.buildRankSimplePvp()
	r.buildRankCorpGsSeverOpen()
	r.buildRankGuildGS()
	r.buildRankGuildSevOpen()
	r.buildRankGuildGateEnemy()
	r.buildRankCorpTrial()
	r.buildRankCorpHeroStar()
	r.buildRankCorpHeroDiffTU()
	r.buildRankCorpHeroDiffZHAN()
	r.buildRankCorpHeroDiffHU()
	r.buildRankCorpHeroDiffSHI()
	r.buildRankDestiny()
	r.buildRankJade()
	r.buildRankEquipStarLv()
	r.buildRankCorpLv()
	r.buildRankWingStar()
	r.buildRankHeroDesiny()
	r.buildRankWuShuangGs()
	r.buildRankJadeTwo()
	r.buildRankAstrology()
	r.buildRankExclusiveWeapon()
	r.buildRankCorpOfWei()
	r.buildRankCorpOfShu()
	r.buildRankCorpOfWu()
	r.buildRankCorpOfQunXiong()
}

func (r *RankModule) AfterStart(g *gin.Engine) {
	g.POST(game.Cfg.RankReloadUrl, r.handleHTTP)
}

func (r *RankModule) buildRankCorpGs() {
	r.RankCorpGs.Start(rankId_RankByCorpGS, TableRankCorpGs(r.sid), r.rankdb,
		"rank.RankCorpGs", func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.CurrCorpGs)
		})
	balance.GetModule(r.sid).RegBalanceNotifyChan(
		"RankCorpGs",
		r.RankCorpGs.setNeedBalanceTopN(false, func(topN [RankTopSize]CorpDataInRank,
			acid2PosScore map[string]PairPosScore) {
			logs.Trace("sevendayrank corpgs award diffday %d", GetSevenDayRankDays(r.sid))
			cfg := gamedata.GetSevOpnRankConfg()
			if !uutil.IsOverseaVer() && GetSevenDayRankDays(r.sid) == int64(cfg.GetAwardDay()-1) &&
				game.Cfg.GetHotActValidData(r.sid, uutil.Hot_Value_SevenRank) {
				logs.Trace("sevendayrank award")
				// 拷贝数据到新rank
				r.RankCorpGsSevOpn.SetTopN(topN)
				r.RankCorpGsSevOpn.setRank(acid2PosScore)
				// 发奖
				topNAcid := make([]string, 0, len(topN))
				topNScore := make([]int64, 0, len(topN))
				acid2Score := make(map[string]int64, len(acid2PosScore))
				for _, t := range topN {
					topNAcid = append(topNAcid, t.ID)
					topNScore = append(topNScore, t.Score)
				}
				for acid, v := range acid2PosScore {
					acid2Score[acid] = v.Score
				}
				award.AwardByCorpGs(r.sid, topNAcid, topNScore, acid2Score, RankByCorpDelayPowBase)
			}
		}), gamedata.GetSevOpnRankTimeBalance())
}

func (r *RankModule) buildRankCorpGsSeverOpen() {
	r.RankCorpGsSevOpn.Start(rankId_RankByCorpGS_ServerOpen,
		tableRankCorpGsSvrOpn(r.sid),
		r.rankdb,
		"rank.RankCorpGsSvrOpn",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.CurrCorpGs)
		})
}

/*
func (r *RankModule) buildRankGuildAct(Cfg *Config) {
	RankGuildAct.Start(rankId_GuildAct, game.Cfg.ShardId+":RankGuildAct", rankdb, "rank.RankGuildAct", func(a *guild_info.GuildInfoBase) int64 {
		return a.Base.ActivenessForXp
	})
}
*/

func (r *RankModule) buildRankGuildGateEnemy() {
	r.RankGuildGateEnemy.Start(rankId_GuildGateEnemy, TableRankGuildGateEnemy(r.sid),
		r.rankdb, "rank.RankGuildGateEnemy", func(a *guild_info.GuildSimpleInfo) int64 {
			return a.GetGEPointWeek()
		})
	balance.GetModule(r.sid).RegBalanceNotifyChan("RankGuildGateEnemy",
		r.RankGuildGateEnemy.setNeedBalanceTopN(true, func(rank int, id string) {}),
		gamedata.GetGatesEnemyWeekBalanceTime())
}

/*
func (r *RankModule) buildRankGuildActWeek(Cfg *Config) {
	RankGuildActWeekLast.Start(rankId_GuildActWeekLast,
		game.Cfg.ShardId+":RankGuildActWeekLast",
		rankdb, "rank.RankGuildActWeekLast",
		func(a *guild_info.GuildInfoBase) int64 {
			return 0
		})
	RankGuildActWeek.setYesterdayRank(&RankGuildActWeekLast)
	RankGuildActWeek.Start(rankId_GuildActWeek,
		game.Cfg.ShardId+":RankGuildActWeek",
		rankdb, "rank.RankGuildActWeek",
		func(a *guild_info.GuildInfoBase) int64 {
			return a.Base.GetActivenessForXpWeek()
		})
	rank_balance_timer.RegBalanceNotifyChan("RankGuildActWeek",
		RankGuildActWeek.setNeedBalanceTopN(true, func(rank int, id string) {}),
		gamedata.GetGuildActiveRankingBegin())
}
*/
func (r *RankModule) buildRankSimplePvp() {
	r.RankSimplePvp.Start(
		TableRankSimplePvp(r.sid),
		r.rankdb, "rank.RankSimplePvp")

	balance.GetModule(r.sid).RegBalanceNotifyChan(
		"RankSimplePvp",
		r.RankSimplePvp.setNeedBalanceTopN(false, func(rankUids []string) {
			// 因为rankUids是拷贝的，而且这里有loaddb的操作，所以go出去
			go func() {
				for i := 0; i < len(rankUids); i++ {
					rank := i + 1
					id := rankUids[i]

					logs.Warn("Balance TopN %d --> %v", rank, id)
					if id == "" {
						return
					}

					rewards := gamedata.GetPvpDailyRewardData(rank)
					if rewards == nil {
						logs.SentryLogicCritical(id, "GetRankRewardData Nil by Rank %d", rank)
						return
					}

					SendMail(r.sid, id,
						rankId_SimplePvp,
						rank, 0,
						mail_sender.IDS_MAIL_SIMPLEPPVP_RANKREWARD_TITLE,
						[]string{fmt.Sprintf("%d", rank)},
						"SimplePvpRankMail",
						rewards.Item2Client,
						rewards.Count2Client, timail.Mail_Send_By_Rank_SimplePvp)

					// 推送和跑马灯需要角色昵称,因为数量不多,就loadsimpleinfo了
					if rank <= push.Simple_pvp_rank_limit {
						accountDBID, err := db.ParseAccount(id)
						if err != nil {
							logs.Error("RankSimplePvp Balance ParseAccount %s Err By %s", accountDBID, err.Error())
							continue
						}

						name, platformId, deviceToken := loadRankAccountInfo(id)
						if name == "" {
							logs.Trace("RankSimplePvp Balance loadRankAccountInfo %s Err", id)
							continue
						}
						// push
						t := time.Now().In(util.ServerTimeLocal)
						push.SimplePvp(rank, id,
							platformId, deviceToken,
							fmt.Sprintf("%d", t.Month()),
							fmt.Sprintf("%d", t.Day()),
							name)
						// sysnotice
						cfgSN := gamedata.SimplePvpSysNotic(uint32(rank))
						if cfgSN != nil {
							sysnotice.NewSysRollNotice(accountDBID.ServerString(), int32(cfgSN.GetServerMsgID())).
								AddParam(sysnotice.ParamType_RollName, name).Send()
						}
					}
					runtime.Gosched()
				}
			}()
			// 判断是否是周奖励的发放日
			nowWeek := util.GetWeek(time.Now().Unix())
			logs.Warn("NowWeek is: %d", nowWeek)
			if nowWeek == gamedata.GetSimplePvpConfig().GetWeekRewardResetDay() {
				for i := 0; i < len(rankUids); i++ {
					rank := i + 1
					id := rankUids[i]

					logs.Warn("Week Balance TopN %d --> %v", rank, id)
					if id == "" {
						return
					}

					rewards := gamedata.GetPvpWeekReward(rank)
					if rewards == nil {
						logs.SentryLogicCritical(id, "GetRankRewardData Nil by Rank %d", rank)
						return
					}

					SendMail(r.sid, id,
						rankId_SimplePvp,
						rank, 0,
						mail_sender.IDS_MAIL_SIMPLEPPVP_RANKREWARDWEEK_TITLE,
						[]string{fmt.Sprintf("%d", rank)},
						"SimplePvpWeekRankMail",
						rewards.Item2Client,
						rewards.Count2Client,
						timail.Mail_Send_By_WeekRank_SimpePvp)
				}

				// 周奖励发放完毕需要清除排行榜
				r.RankSimplePvp.Clean()
				// 更新战斗记录
				global_count.AddRecordCount(r.sid, uint(game.Cfg.Gid), global_count.SimplePvpRecord)
			}

			// 为title记录
			title_rank.GetModule(r.sid).SetSimplePvpRank(rankUids)
		}), gamedata.GetPVPBalanceBegin())
}

func (r *RankModule) buildRankGuildGS() {
	r.RankGuildGs.Start(r.sid, rankId_GuildGS, TableRankGuildGS(r.sid),
		r.rankdb, "rank.RankGuildGS",
		func(a *guild_info.GuildSimpleInfo) int64 {
			return a.GuildGSSum
		})
	balance.GetModule(r.sid).RegBalanceNotifyChan("RankGuildGS",
		r.RankGuildGs.setNeedBalanceTopN(false, rankByGuildFromCacheBalanceFunc(func(topN [RankTopSize]GuildDataInRank) {
			logs.Trace("sevendayrank guildgs award diffday %d", GetSevenDayRankDays(r.sid))
			cfg := gamedata.GetSevOpnRankConfg()
			if !uutil.IsOverseaVer() && GetSevenDayRankDays(r.sid) == int64(cfg.GetAwardDay()-1) &&
				game.Cfg.GetHotActValidData(r.sid, uutil.Hot_Value_SevenRank) {
				logs.Trace("sevendayrank guildgs award")
				// 拷贝数据到新rank
				r.RankGuildSevOpn.setTopN(topN)
				// 发奖
				topNUuid := make([]string, 0, len(topN))
				for _, t := range topN {
					topNUuid = append(topNUuid, t.UUID)
				}
				award.AwardGuild(topNUuid)
			}
			guildUuids := make([]string, 0, RankTopSize)
			for _, g := range topN {
				if g.UUID == "" {
					break
				}
				guildUuids = append(guildUuids, g.UUID)
			}
			logiclog.LogGuildRank(guildUuids)
		})), gamedata.GetSevOpnRankTimeBalance())
}

func (r *RankModule) buildRankGuildSevOpen() {
	r.RankGuildSevOpn.Start(rankId_RankByGuildGS_ServerOpen,
		tableRankGuildGsSvrOpn(r.sid),
		r.rankdb, "rank.RankGuildGsSvrOpn",
		func(a *guild_info.GuildSimpleInfo) int64 {
			return a.GuildGSSum
		})
}

func (r *RankModule) buildRankCorpTrial() {
	r.RankByCorpTrial.Start(rankId_RankByCorpTrial,
		TableRankCorpTrial(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpTrial",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.MaxTrialLv
		})
}

func (r *RankModule) buildRankCorpHeroStar() {
	r.RankByHeroStar.Start(rankId_RankByHeroStar,
		TableRankCorpHeroStar(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpHeroStar",
		func(a *helper.AccountSimpleInfo) int64 {
			var starSum uint32
			for i := 0; i < len(a.AvatarStarLvl); i++ {
				starSum += a.AvatarStarLvl[i]
			}
			return int64(starSum)
		})
}

func (r *RankModule) buildRankDestiny() {
	r.RankByDestiny.Start(rankId_RankByDestiny,
		TableRankDestiny(r.sid),
		100000,
		r.rankdb,
		"rank.RankDestiny",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.DestinyLv
		})
}

func (r *RankModule) buildRankJade() {
	r.RankByJade.Start(rankId_RankByJade,
		TableRankJade(r.sid),
		100000,
		r.rankdb,
		"rank.RankJade",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.JadeLv
		})
}

func (r *RankModule) buildRankJadeTwo() {
	r.RankByHeroJadeTwo.Start(rankId_RankByHeroJadeTwo,
		TableRankJadeTwo(r.sid),
		100000,
		r.rankdb,
		"rank.RankJadeTwo",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.JadeLv
		})
}

func (r *RankModule) buildRankWuShuangGs() {
	r.RankByHeroWuShuangGs.Start(rankId_RankByWuShuangGs,
		TableRankWuShuangGs(r.sid),
		100000,
		r.rankdb,
		"rank.RankWuShuangGs",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.WuShuangGs
		})
}

func (r *RankModule) buildRankExclusiveWeapon() {
	r.RankByExclusiveWeapon.Start(rankId_RankByExclusiveWeapon,
		TableRankExclusiveWeapon(r.sid),
		100000,
		r.rankdb,
		"rank.RankExclusiveWeapon",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.ExclusivWeapon
		})
}

func (r *RankModule) buildRankEquipStarLv() {
	r.RankByEquipStarLv.Start(rankId_RankByEquipStarLv,
		TableRankEquipStarLv(r.sid),
		100000,
		r.rankdb,
		"rank.RankEquipStarLv",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.EquipStarLv)
		})
}

func (r *RankModule) buildRankCorpLv() {
	r.RankByCorpLv.Start(rankId_RankByCorpLv,
		TableRankCorpLv(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpLv",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.CorpLv)
		})
}

func (r *RankModule) buildRankWingStar() {
	r.RankByWingStar.Start(rankId_RankByWingStar,
		TableRankSwingStarLv(r.sid),
		100000,
		r.rankdb,
		"rank.SwingStarLv",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.SwingStarLv)
		})
}

func (r *RankModule) buildRankHeroDesiny() {
	r.RankByHeroDestiny.Start(rankId_RankByHersoDestiny,
		TableRankHeroDestinyLv(r.sid),
		100000,
		r.rankdb,
		"rank.HeroDestinyLv",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.HeroDestinyLv)
		})
}

func (r *RankModule) buildRankAstrology() {
	r.RankByAstrology.Start(rankId_RankByAstrology,
		TableRankAstrology(r.sid),
		100000,
		r.rankdb,
		"rank.Astrology",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.Astrology)
		})
}

func (r *RankModule) buildRankCorpHeroDiffTU() {
	r.RankByHeroDiff[gamedata.HeroDiff_TU].Start(rankId_RankByHeroDiff_TU,
		TableRankCorpHeroDiffTU(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpHeroDiff.TU",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.HeroDiffScore[gamedata.HeroDiff_TU])
		})
}
func (r *RankModule) buildRankCorpHeroDiffZHAN() {
	r.RankByHeroDiff[gamedata.HeroDiff_ZHAN].Start(rankId_RankByHeroDiff_ZHAN,
		TableRankCorpHeroDiffZHAN(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpHeroDiff.ZHAN",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.HeroDiffScore[gamedata.HeroDiff_ZHAN])
		})
}
func (r *RankModule) buildRankCorpHeroDiffHU() {
	r.RankByHeroDiff[gamedata.HeroDiff_HU].Start(rankId_RankByHeroDiff_HU,
		TableRankCorpHeroDiffHU(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpHeroDiff.HU",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.HeroDiffScore[gamedata.HeroDiff_HU])
		})
}
func (r *RankModule) buildRankCorpHeroDiffSHI() {
	r.RankByHeroDiff[gamedata.HeroDiff_SHI].Start(rankId_RankByHeroDiff_SHI,
		TableRankCorpHeroDiffSHI(r.sid),
		100000,
		r.rankdb,
		"rank.RankCorpHeroDiff.SHI",
		func(a *helper.AccountSimpleInfo) int64 {
			return int64(a.HeroDiffScore[gamedata.HeroDiff_SHI])
		})
}
func (r *RankModule) buildRankCorpOfWei() {
	r.RankByCorpGsOfWei.Start(rankId_RankByCorpOfWei,
		TableRankCorpOfWei(r.sid),
		r.rankdb,
		"rank.RankCorpOfWei",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.TopGsByCountry[helper.Country_Wei]
		})
}
func (r *RankModule) buildRankCorpOfShu() {
	r.RankByCorpGsOfShu.Start(rankId_RankByCorpOfShu,
		TableRankCorpOfShu(r.sid),
		r.rankdb,
		"rank.RankCorpOfShu",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.TopGsByCountry[helper.Country_Shu]
		})
}
func (r *RankModule) buildRankCorpOfWu() {
	r.RankByCorpGsOfWu.Start(rankId_RankByCorpOfWu,
		TableRankCorpOfWu(r.sid),
		r.rankdb,
		"rank.RankCorpOfWu",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.TopGsByCountry[helper.Country_Wu]
		})
}
func (r *RankModule) buildRankCorpOfQunXiong() {
	r.RankByCorpGsOfQunXiong.Start(rankId_RankByCorpOfQunXiong,
		TableRankCorpOfQunXiong(r.sid),
		r.rankdb,
		"rank.RankCorpOfQunXiong",
		func(a *helper.AccountSimpleInfo) int64 {
			return a.TopGsByCountry[helper.Country_Qun]
		})
}

func (r *RankModule) Stop() {
	r.RankCorpGs.Stop()
	r.RankSimplePvp.Stop()
	//RankGuildAct.Stop()
	//RankGuildActWeek.Stop()
	//RankGuildActWeekLast.Stop()
	r.RankGuildGs.Stop()
	r.RankGuildGateEnemy.Stop()
	r.RankByCorpTrial.Stop()
	r.RankByHeroStar.Stop()
	r.RankByEquipStarLv.Stop()
	r.RankByCorpLv.Stop()
	r.RankByJade.Stop()
	r.RankByDestiny.Stop()
	r.RankByWingStar.Stop()
	r.RankByHeroDestiny.Stop()
	r.RankByHeroJadeTwo.Stop()
	r.RankByHeroWuShuangGs.Stop()
	r.RankByAstrology.Stop()
	r.RankByExclusiveWeapon.Stop()
	r.RankByCorpGsOfShu.Stop()
	r.RankByCorpGsOfWei.Stop()
	r.RankByCorpGsOfWu.Stop()
	r.RankByCorpGsOfQunXiong.Stop()
	for i := 0; i < len(r.RankByHeroDiff); i++ {
		r.RankByHeroDiff[i].Stop()
	}
}

func (r *RankModule) GetRankContent(ID int) map[string]float64 {
	tableName := ""
	switch ID {
	case Ex_RankId_RankByCorpLv:
		tableName = TableRankCorpLv(r.sid)
	case Ex_RankId_RankByDestiny:
		tableName = TableRankDestiny(r.sid)
	case Ex_RankId_RankByJade:
		tableName = TableRankJade(r.sid)
	case Ex_RankId_RankByEquipStarLv:
		tableName = TableRankEquipStarLv(r.sid)
	case Ex_RankId_RankByCorpGS:
		tableName = TableRankCorpGs(r.sid)
	case Ex_RankId_RankByHeroStar:
		tableName = TableRankCorpHeroStar(r.sid)
	case Ex_RankId_RankBySwingStarLv:
		tableName = TableRankSwingStarLv(r.sid)
	case Ex_RankId_HeroDestinyLv:
		tableName = TableRankHeroDestinyLv(r.sid)
	case Ex_RankId_HeroByJadeTwo:
		tableName = TableRankJadeTwo(r.sid)
	case Ex_RankId_HeroByWuShuangGs:
		tableName = TableRankWuShuangGs(r.sid)
	case Ex_RankId_Astrology:
		tableName = TableRankAstrology(r.sid)
	case Ex_RankId_ExclusiveWeapon:
		tableName = TableRankExclusiveWeapon(r.sid)
	default:
		logs.Error("no rank db for id: %d", ID)
		return nil
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("get db nil")
		return nil
	}

	reply, err := redis.Values(conn.Do("ZRANGE", tableName, "0", "-1", "WITHSCORES"))
	if err != nil {
		logs.Error("zrange err by %v", err)
		return nil
	}
	if len(reply)%2 != 0 {
		logs.Error("zrange ret format err by %v", err)
		return nil
	}
	ret := make(map[string]float64, 0)
	for i := 0; i < len(reply); i += 2 {
		id, err := redis.String(reply[i], nil)
		if err != nil {
			logs.Error("zrange item format err for value convert string: % by err %v", reply[i], err)
			return nil
		}
		score, err := redis.Float64(reply[i+1], nil)
		if err != nil {
			logs.Error("zrange item format err for value convert float64: % by err %v", reply[i+1], err)
			return nil
		}
		ret[id] = score
	}
	return ret
}
func metricsSend(value string, rankName string) {
	name := fmt.Sprintf("sync.%d.%s.%s", game.Cfg.Gid, rankName, "time")
	metrics.SimpleSend(name, value)
}
