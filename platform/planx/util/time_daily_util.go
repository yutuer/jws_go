package util

import (
	"time"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type TimeToBalance struct {
	WeekDay   int   `json:"w"`
	DailyTime int64 `json:"d"`
}

func (t *TimeToBalance) IsNil() bool {
	return t.WeekDay < 0 && t.DailyTime < 0
}

func NewTimeToBalanceNil() TimeToBalance {
	return TimeToBalance{
		WeekDay:   -1,
		DailyTime: -1,
	}
}

func IsSameByStartTime(t1, t2 time.Time, dailyTime TimeToBalance) bool {
	return IsSameUnixByStartTime(t1.Unix(), t2.Unix(), dailyTime)
}

func IsSameUnixByStartTime(t1, t2 int64, dailyTime TimeToBalance) bool {
	return DailyBeginUnixByStartTime(t1, dailyTime) == DailyBeginUnixByStartTime(t2, dailyTime)
}

func DailyBeginUnixByStartTime(u int64, dailyTime TimeToBalance) int64 {
	if dailyTime.WeekDay == 0 {
		// 按天刷新
		return u - (u-(TimeUnixBeginUnix+(dailyTime.DailyTime*TimeSlotSec)))%DaySec
	} else {
		// 按周刷新
		//logs.Error("DailyBeginUnixByStartTime %v %v", u, dailyTime)
		week := TimeWeekDayTranslate(time.Weekday(GetWeekByStartTime(u, dailyTime.DailyTime)))
		//logs.Error("week %v", week)
		todayBegin := u - (u-(TimeUnixBeginUnix+(dailyTime.DailyTime*TimeSlotSec)))%DaySec
		//logs.Error("week %v", todayBegin)
		dayCount := week - dailyTime.WeekDay
		//logs.Error("week %v", dayCount)
		if week < dailyTime.WeekDay {
			dayCount = WeekDayCount + dayCount
			//logs.Error("week %v", dayCount)
		}
		//logs.Error("week %v", todayBegin-(DaySec*int64(dayCount)))
		return todayBegin - (DaySec * int64(dayCount))
	}
}

func GetWeekByStartTime(u int64, dailyTime int64) int {
	dayTimeUnix := DailyTime2UnixTime(u, dailyTime)
	week := GetWeek(u)
	if u >= dayTimeUnix {
		return week
	} else {
		week = week - 1
		if week < 0 {
			week = week + WeekDayCount
		}
		return week
	}

}

// t 当天的起始时间
func GetNextDailyTime(t, now_t int64) int64 {
	if t <= now_t {
		return t + DaySec
	}
	return t
}

func GetNextDailyWeekTime(t, now_t int64) int64 {
	if t <= now_t {
		return t + WeekSec
	}
	return t
}

// 获取间隔天数, 不包括起始和最终天数， 比如（第一天，第三天) = 1
func GetIntervalDay(start, end int64, dailyStartTime TimeToBalance) int {
	startDay := GetNextDailyTime(DailyBeginUnixByStartTime(start, dailyStartTime), start)
	endDay := DailyBeginUnixByStartTime(end, dailyStartTime)
	retInterval := int((endDay - startDay) / DaySec)
	if retInterval < 0 {
		retInterval = 0
	}
	return retInterval
}
