package gamedata

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdCommonDailyStartTime      util.TimeToBalance
	gdGuildActiveRankingTime    util.TimeToBalance
	gdPVPBalanceTime            util.TimeToBalance
	gdPVPWeekBalanceTime        util.TimeToBalance
	gdGatesEnemyWeekBalanceTime util.TimeToBalance
	gdTeamPVPBalanceTime        util.TimeToBalance
	gdWeekBalanceBeginTime      util.TimeToBalance
	gdRankWorshipBalanceTime    util.TimeToBalance
	gdGuildBagAutoRefuse        util.TimeToBalance
	gdGVGGuildGiftGet           util.TimeToBalance
	gdHeroDiffReset             util.TimeToBalance
	gdGuildWorshipReset         util.TimeToBalance
	gdWuShuangTitle             util.TimeToBalance
	gdFriendGiftRefreshTime     util.TimeToBalance
)

const (
	DailyStartTypCommon = iota
	DailyStartTypGuildActiveRanking
	DailyStartTypPVPBalance
	DailyStartTypGatesEnemyWeekBalance
	DailyStartTypTeamPVPBalance
	DailyStartTypGuildBoss
	DailyStartTypRankWorshipBalance
	DailyStartTypGuildBagAutoRefuse
	DailyStartTypGVGGuildGiftGet
	DailyStartTypHeroDiffReset
	DailyStartTypGuildWorshipReset
	DailyStartTypeWspvpRefresh
	DailyStartTypWspvpChallenge
	DailyStartTypFriendGift
	DailyStartTypCount
)

func loadRefreshTimeData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.ALLREFRESHTIME_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for _, a := range dataList.GetItems() {
		switch a.GetOption() {
		case "Common":
			gdCommonDailyStartTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdCommonDailyStartTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME CommonDailyStartTime err By " + a.GetRefreshTime()))
			}
			gdCommonDailyStartTime.WeekDay = int(a.GetDayInWeek())
			break
		case "GuildActiveRanking":
			gdGuildActiveRankingTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdGuildActiveRankingTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME GuildActiveRankingTime err By " + a.GetRefreshTime()))
			}
			gdGuildActiveRankingTime.WeekDay = int(a.GetDayInWeek())
			break
		case "PVPBalance":
			gdPVPBalanceTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdPVPBalanceTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME PVPBalanceTime err By " + a.GetRefreshTime()))
			}
			gdPVPBalanceTime.WeekDay = int(a.GetDayInWeek())
			break
		case "GEWeekRank":
			gdGatesEnemyWeekBalanceTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdGatesEnemyWeekBalanceTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME gdGatesEnemyWeekBalanceTime err By " + a.GetRefreshTime()))
			}
			gdGatesEnemyWeekBalanceTime.WeekDay = int(a.GetDayInWeek())
		case "TPVPBalance":
			gdTeamPVPBalanceTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdTeamPVPBalanceTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME TPVPBalanceTime err By " + a.GetRefreshTime()))
			}
			gdTeamPVPBalanceTime.WeekDay = int(a.GetDayInWeek())
			break
		case "RankWorship":
			gdRankWorshipBalanceTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdRankWorshipBalanceTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME RankWorshipBalanceTime err By " + a.GetRefreshTime()))
			}
			gdRankWorshipBalanceTime.WeekDay = int(a.GetDayInWeek())
			break
		case "BasicPvPRestDay":
			gdPVPWeekBalanceTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdPVPWeekBalanceTime.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME PVPWeekBalanceTime err By " + a.GetRefreshTime()))
			}
			gdPVPWeekBalanceTime.WeekDay = int(a.GetDayInWeek())
			break
		case "GuildBagAutoRefuse":
			gdGuildBagAutoRefuse.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdGuildBagAutoRefuse.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime()))
			}
			gdGuildBagAutoRefuse.WeekDay = int(a.GetDayInWeek())
		case "GVGGuildGiftGet":
			gdGVGGuildGiftGet.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdGVGGuildGiftGet.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime()))
			}
			gdGVGGuildGiftGet.WeekDay = int(a.GetDayInWeek())
		case "HeroDiffRank":
			gdHeroDiffReset.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdHeroDiffReset.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime()))
			}
			gdHeroDiffReset.WeekDay = int(a.GetDayInWeek())
		case "WorshipTime":
			gdGuildWorshipReset.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdGuildWorshipReset.DailyTime < 0 {
				panic(errors.New("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime()))
			}
			gdGuildWorshipReset.WeekDay = int(a.GetDayInWeek())
		case "WSPVP":
			gdWuShuangTitle.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdWuShuangTitle.DailyTime < 0 {
				//panic(errors.New("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime()))
			}
			gdWuShuangTitle.WeekDay = int(a.GetDayInWeek())
		case "FRIENDGIFT":
			logs.Debug("load friend gift refresh time")
			gdFriendGiftRefreshTime.DailyTime = util.DailyTimeFromString(a.GetRefreshTime())
			if gdFriendGiftRefreshTime.DailyTime < 0 {
				logs.Error("ALLREFRESHTIME GuildBagAutoRefuse err By " + a.GetRefreshTime())
			}
			gdFriendGiftRefreshTime.WeekDay = int(a.GetDayInWeek())

		default:
			logs.Error("ALLREFRESHTIME GetOption Unknown, %v", a.GetOption())
		}

	}
	logs.Trace("CommonDailyStartTime %v", gdCommonDailyStartTime)
	logs.Trace("GuildActiveRankingTime %v", gdGuildActiveRankingTime)
	logs.Trace("PVPBalanceTime %v", gdPVPBalanceTime)
	logs.Trace("gdGatesEnemyWeekBalanceTime %v", gdGatesEnemyWeekBalanceTime)
	logs.Trace("TPVPBalanceTime %v", gdTeamPVPBalanceTime)
	logs.Trace("gdGuildWorshipReset %v", gdGuildWorshipReset)
	gdWeekBalanceBeginTime = gdCommonDailyStartTime
	gdWeekBalanceBeginTime.WeekDay = 1
	logs.Trace("gdRankWorshipBalanceTime %v", gdRankWorshipBalanceTime)

}

func IsSameDayCommon(t1, t2 int64) bool {
	return util.IsSameUnixByStartTime(t1, t2, gdCommonDailyStartTime)
}

func IsSameDayGuildWorship(t1, t2 int64) bool {
	return util.IsSameUnixByStartTime(t1, t2, gdGuildWorshipReset)
}

func IsSameDayFriendGift(t1, t2 int64) bool {
	logs.Debug("gdFriendGiftRefreshTime: %d", gdFriendGiftRefreshTime)
	return util.IsSameUnixByStartTime(t1, t2, gdFriendGiftRefreshTime)
}

func GetCommonDayDiff(t_begin, t_end int64) int64 {
	return util.DailyBeginUnixByStartTime(t_end, gdCommonDailyStartTime)/util.DaySec -
		util.DailyBeginUnixByStartTime(t_begin, gdCommonDailyStartTime)/util.DaySec
}

func GetGVGGuildDayDiff(t_begin, t_end int64) int64 {
	return util.DailyBeginUnixByStartTime(t_end, gdGVGGuildGiftGet)/util.DaySec -
		util.DailyBeginUnixByStartTime(t_begin, gdGVGGuildGiftGet)/util.DaySec
}

func GetCommonDayDiffC(t_begin, t_end int64) int64 {
	return GetCommonDayDiff(t_begin, t_end) + 1
}

func GetCommonDayBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdCommonDailyStartTime)
}

func GetFriendGiftBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdFriendGiftRefreshTime)
}

func GetPVPBalanceBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdPVPBalanceTime)
}

func GetCommonDayBegin() util.TimeToBalance {
	return gdCommonDailyStartTime
}

func GetGuildActiveRankingBegin() util.TimeToBalance {
	return gdGuildActiveRankingTime
}

func GetPVPBalanceBegin() util.TimeToBalance {
	return gdPVPBalanceTime
}

func GetGatesEnemyWeekBalanceTime() util.TimeToBalance {
	return gdGatesEnemyWeekBalanceTime
}

func GetTeamPVPBalanceBegin() util.TimeToBalance {
	return gdTeamPVPBalanceTime
}

func GetWeekBalanceBegin() util.TimeToBalance {
	return gdWeekBalanceBeginTime
}

func GetRankWorshipBalance() util.TimeToBalance {
	return gdRankWorshipBalanceTime
}

func GetPVPWeekBalanceBegin() util.TimeToBalance {
	return gdPVPWeekBalanceTime
}

func GetGuildWorshipTime() int64 {
	return gdGuildWorshipReset.DailyTime
}

func GetGuildBagRefuseBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdGuildBagAutoRefuse)
}

func GetHeroDiffResetBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdHeroDiffReset)
}

func GetGVGGuildGiftBeginSec(t int64) int64 {
	return util.DailyBeginUnixByStartTime(t, gdGVGGuildGiftGet)
}

func GetBeginTimeByTyp(typ int) util.TimeToBalance {
	switch typ {
	case DailyStartTypCommon:
		return gdCommonDailyStartTime
	case DailyStartTypGuildActiveRanking:
		return gdGuildActiveRankingTime
	case DailyStartTypPVPBalance:
		return gdPVPBalanceTime
	case DailyStartTypGatesEnemyWeekBalance:
		return gdGatesEnemyWeekBalanceTime
	case DailyStartTypTeamPVPBalance:
		return gdTeamPVPBalanceTime
	case DailyStartTypGuildBoss:
		return gdGuildGBRestartTime
	case DailyStartTypRankWorshipBalance:
		return gdRankWorshipBalanceTime
	case DailyStartTypGuildBagAutoRefuse:
		return gdGuildBagAutoRefuse
	case DailyStartTypGVGGuildGiftGet:
		return gdGVGGuildGiftGet
	case DailyStartTypHeroDiffReset:
		return gdHeroDiffReset
	case DailyStartTypGuildWorshipReset:
		return gdGuildWorshipReset
	case DailyStartTypeWspvpRefresh:
		return gdWuShuangTitle
	case DailyStartTypFriendGift:
		return gdFriendGiftRefreshTime
	default:
		return util.NewTimeToBalanceNil()
	}
}

func GetIntervalDayByCommon(startTime, endTime int64) int {
	return util.GetIntervalDay(startTime, endTime, gdCommonDailyStartTime)
}
