package util

import (
	"errors"
	"fmt"
	"time"

	"strconv"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
#Time Util
---------------------
##Day Time
##Week Time
*/

// 注意时区的影响
// TBD WeekTime支持其他时区

const (
	MinSec       = 60
	HourSec      = 60 * MinSec
	DaySec       = 24 * 3600
	WeekDayCount = 7
	WeekSec      = WeekDayCount * 24 * 3600
	YearSec      = 365 * DaySec
	TimeSlotMin  = 30
	TimeSlotSec  = TimeSlotMin * MinSec
)

// 处理时区产生的问题
var TimeUnixBegin time.Time
var TimeUnixBeginUnix int64
var FirstMonday time.Time
var FirstMondayUnix int64
var ServerTimeLocal *time.Location
var serverStartTime map[uint]int64

func init() {
	SetTimeLocal("Asia/Shanghai")
}

func SetTimeLocal(local_str string) {
	l, err := time.LoadLocation(local_str)
	if err != nil {
		panic(err)
	}
	ServerTimeLocal = l
	initTimeConst(l)
}
func initTimeConst(local *time.Location) {
	var err error
	TimeUnixBegin, err = time.ParseInLocation("2006/1/2", "2000/1/1", local)
	if err != nil {
		panic(fmt.Errorf("TimeUnixBegin Parse err %s", err.Error()))
	}
	//logs.Trace("TimeUnixBegin %v", TimeUnixBegin)
	TimeUnixBeginUnix = TimeUnixBegin.Unix()

	FirstMonday, err = time.ParseInLocation("2006/1/2", "2000/1/5", local)
	if err != nil {
		panic(fmt.Errorf("FirstMonday Parse err %s", err.Error()))
	}
	//logs.Trace("FirstMonday %v", FirstMonday)
	FirstMondayUnix = FirstMonday.Unix()
}

func SetServerStartTimes(shard2ServerStartTime map[uint]string) {
	if len(shard2ServerStartTime) <= 0 {
		// 这就不让起服了
		//		panic(errors.New("GetServerStartTime Cfg Err"))
		return
	}
	serverStartTime = make(map[uint]int64, len(shard2ServerStartTime))
	for shard, sst := range shard2ServerStartTime {
		if sst == "" {
			// 这就不让起服了
			panic(errors.New("GetServerStartTime Cfg Err"))
		}
		startTime := TimeFromString(sst)
		if startTime < 0 {
			// 这就不让起服了
			panic(errors.New("GetServerStartTime Cfg Err"))
		}
		serverStartTime[shard] = startTime
		logs.Warn("set shard %d start time %d", shard, startTime)
	}
}

// 不要用这个接口，用game下的，有合服shard转换
func GetServerStartTime(sid uint) int64 {
	return serverStartTime[sid]
}
func SetServerStartTime(sid uint, t int64) {
	serverStartTime[sid] = t
}

func IsSameDayUnix(t1, t2 int64) bool {
	return DailyBeginUnix(t1) == DailyBeginUnix(t2)
}

/*
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return (y1 == y2) && (m1 == m2) && (d1 == d2)
}

func IsNowDay(t time.Time) bool {
	return IsSameDay(t, time.Now())
}
*/

func DailyBeginUnix(u int64) int64 {
	return u - (u-TimeUnixBeginUnix)%DaySec
}

// Week 以周一零点为开始
func IsSameWeek(t1, t2 time.Time) bool {
	return IsSameWeekUnix(t1.Unix(), t2.Unix())
}

func IsSameWeekUnix(u1, u2 int64) bool {
	w1 := (u1 - FirstMondayUnix) / WeekSec
	w2 := (u2 - FirstMondayUnix) / WeekSec
	return w1 == w2
}

func IsNowWeek(t time.Time) bool {
	return IsSameWeek(t, time.Now())
}

func IsNowWeekUnix(u int64) bool {
	return IsSameWeekUnix(u, time.Now().Unix())
}

func GetDayBefore(t_begin time.Time, t_end time.Time) int64 {
	// 注意时区的影响
	t_b := (t_begin.Unix() - TimeUnixBeginUnix) / DaySec
	t_e := (t_end.Unix() - TimeUnixBeginUnix) / DaySec

	return t_e - t_b

}

func GetDayBeforeUnix(t_begin, t_end int64) int64 {
	// 注意时区的影响
	t_b := (t_begin - TimeUnixBeginUnix) / DaySec
	t_e := (t_end - TimeUnixBeginUnix) / DaySec

	return t_e - t_b

}

func IsTimeBetween(begin, end, t time.Time) bool {
	if !begin.IsZero() && !t.After(begin) {
		return false
	}

	if !end.IsZero() && t.After(end) {
		return false
	}

	return true
}

func IsTimeBetweenUnix(begin, end, t int64) bool {
	return (begin <= 0 || t > begin) && (end <= 0 || t <= end)
}

/* WeekTime
   从每周一零点开始每个半点和整点的计数
   如 0 - 周一 0:00
      1 -      0:30
	  2 -      1:00
	  2 -      1:11
	  3 -      1:30
*/
func WeekTime(u int64, offset_tm string) int64 {
	return WeekTimeUnix(u, offset_tm)
}

func WeekTimeUnix(u int64, offset_tm string) int64 {
	ws := (u - FirstMondayUnix) % WeekSec
	t0 := ws / TimeSlotSec
	tb := DailyTimeFromString(offset_tm)
	if t0 >= tb {
		return t0 - tb
	} else {
		return t0 + (48 - tb)
	}
}

// 将周几和时间转换成一周内的时间段，效率一般，适合读配置时预先转换
func WeeklyTimeFromString(weekDay int, offset_tm string, tm string) int64 {
	t := fmt.Sprintf("2015/8/%v %v", 2+weekDay, tm)
	t_in_time, err := time.ParseInLocation("2006/1/2 15:04", t, ServerTimeLocal)
	if err != nil {
		logs.Error("WeeklyTimeFromString Parse %s err %s", t, err.Error())
		return -1
	}
	return WeekTime(t_in_time.Unix(), offset_tm)
}

func WeeklyBeginUnix(offset_tm string, u int64) int64 {
	return FirstMondayUnix + ((u-FirstMondayUnix)/WeekSec)*WeekSec + DailyTimeFromString(offset_tm)*TimeSlotSec
}

func WeeklySlot2Unix(slot int64, offset_tm string, u int64) int64 {
	return WeeklyBeginUnix(offset_tm, u) + slot*TimeSlotSec
}

func IsNowMouth(t time.Time) bool {
	y1, m1, _ := time.Now().Date()
	y2, m2, _ := t.Date()
	return (y1 == y2) && (m1 == m2)
}

/*
   DailyTime
   从每天零点开始每个半点和整点的计数
   如 0 -      0:00
      1 -      0:30
	  2 -      1:00
	  2 -      1:11
	  3 -      1:30
*/

func DailyTime(t time.Time) int64 {
	return DailyTimeUnix(t.Unix())
}

func DailyTimeUnix(u int64) int64 {
	nb := DailyBeginUnix(u)
	return (u - nb) / TimeSlotSec
}

//DailyRealTimeUnix 当天零点开始的秒数
func DailyRealTimeUnix(u int64) int64 {
	nb := DailyBeginUnix(u)
	return u - nb
}

func DailyTime2UnixTime(now_time int64, daily_time int64) int64 {
	nb := DailyBeginUnix(now_time)
	return nb + (TimeSlotSec * daily_time)
}

// 从字符串转成DailyTime，效率一般，适合读配置时预先转换
// "00:00" -> 0
// "00:30" -> 1
// "01:00" -> 2
func DailyTimeFromString(s string) int64 {
	t := "2015/1/1 "
	t_in_time, err := time.ParseInLocation("2006/1/2 15:04", t+s, ServerTimeLocal)
	if err != nil {
		logs.Error("DailyTimeFromString Parse %s err %s", s, err.Error())
		return -1
	}
	return DailyTime(t_in_time)
}

//DailyRealTimeFromString 从字符串转成 DailyRealTime, 效率一般，适合读配置时预先转换
// "05:00" -> 18000
func DailyRealTimeFromString(s string) int64 {
	pre := "2015/1/1 "
	t, err := time.ParseInLocation("2006/1/2 15:04", pre+s, ServerTimeLocal)
	if err != nil {
		logs.Error("DailyTimeFromString Parse %s err %s", s, err.Error())
		return -1
	}
	tUnix := t.Unix()
	return tUnix - DailyBeginUnix(tUnix)
}

func TimeFromString(s string) int64 {
	t_in_time, err := time.ParseInLocation("2006/1/2 15:04", s, ServerTimeLocal)
	if err != nil {
		logs.Error("DailyTimeFromString Parse %s err %s", s, err.Error())
		return -1
	}
	return t_in_time.Unix()
}

// 当前时间距离某个偏移时间的整天数
func DayOffsetFromString(s string, u int64) int64 {
	t := "2015/1/1 "
	t_in_time, err := time.ParseInLocation("2006/1/2 15:04", t+s, ServerTimeLocal)
	if err != nil {
		logs.Error("DayOffsetFromString Parse %s err %s", s, err.Error())
		return -1
	}
	return (u - t_in_time.Unix()) / DaySec
}

/*
 计算时先计算这一段时间中可以增加多少,
 将结余的时间返回
*/
func AccountTime2Point(now_unix_sec, last_unix_sex, one_need int64) (int64, int64) {
	/*
	                 + s +
	   ----+---------+---+---+--------
	      last     this now  next
	   ----+---------+---+---+--------
	       +    add  +   + r +

	                 + one   +
	*/
	time_in_v := now_unix_sec - last_unix_sex
	s_v := time_in_v % one_need
	add_v := time_in_v - s_v

	return add_v / one_need, s_v
}

/*
	纳秒从2015年8月1日开始, 目前用在Log中标记时间(数值比较小)
*/
func tiNowUnixNanoStartTime() int64 {
	return time.Date(2015, 8, 1, 0, 0, 0, 0, time.UTC).UnixNano() // 这个直接用UTC
}

var TiNowUnixNanoStart = tiNowUnixNanoStartTime()

func TiNowUnixNano() int64 {
	return time.Now().UnixNano() - TiNowUnixNanoStart
}

/*
	转换weekday，time里周日为0，cfg一般为7
*/
func TimeWeekDayTranslate(weekday time.Weekday) int {
	ret := int(weekday)
	if ret == 0 {
		ret = 7
	}
	return ret
}

/*
	转换weekday，time里周日为0，cfg一般为7
*/
func TimeWeekDayTranslateFromCfg(weekdayCfg int) int {
	ret := weekdayCfg
	if weekdayCfg == 7 {
		ret = 0
	}
	return ret
}

// 获取当前时间是周几 0 - 6
func GetWeek(t int64) int {
	tt := time.Unix(t, 0)
	return int(tt.In(ServerTimeLocal).Weekday())
}

// 返回当前时间所在天的某个整点时间
func GetCurDayTimeAtHour(t int64, hour int) (ret int64) {
	tmpTime := time.Unix(t, 0).In(ServerTimeLocal)
	ret = time.Date(tmpTime.Year(), tmpTime.Month(), tmpTime.Day(), hour,
		0, 0, 0, ServerTimeLocal).Unix()
	return
}

func GetWeekTime(now_t int64, weekDay int, hourMin string) int64 {
	t := time.Unix(now_t, 0)
	t = t.In(ServerTimeLocal)
	_t, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", t.Year(), t.Month(), t.Day(),
			hourMin), ServerTimeLocal)
	if err != nil {
		logs.Error("GetNextWeekTime time.ParseInLocation err %v", err)
		return 0
	}
	t = _t
	if weekDay > int(time.Saturday) {
		weekDay = int(time.Sunday)
	}
	ts := t.Unix() + int64(weekDay-int(t.Weekday()))*int64(DaySec)
	return ts
}

func GetNextWeekTime(now_t int64, weekDay int, hourMin string) int64 {
	ts := GetWeekTime(now_t, weekDay, hourMin)
	if ts <= now_t {
		return ts + int64(WeekSec)
	}
	return ts
}

func Clock(t time.Time) (int, int, int) {
	return t.In(ServerTimeLocal).Clock()
}

//输入参数：Unix时间戳字符串，输出当地时间time结构
//警告：如果返回错误，不要使用返回值。
func UnixStringToTimeByLoacl(unixTimeString string) (time.Time, error) {
	//将输入的字符串转换为int64
	unixtimeint64, err := strconv.ParseInt(unixTimeString, 10, 64)
	if err != nil {
		logs.Error("time_util:UnixStringToTimeBylocal err :%v", err)
		return time.Time{}, err
	}
	logs.Trace("当前时区为：%v", ServerTimeLocal)
	//将int64的时间戳转换成时间结构,并直接获取在当地时区下的时间结构
	unixtime := time.Unix(unixtimeint64, 0).In(ServerTimeLocal)

	return unixtime, nil
}
