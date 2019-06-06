package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const (
	layout = "2006-1-2 15:04:05"
)

func init() {

}

func TestSame(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-2-1 22:00:01", time.Local) // 周日
	t.Logf("local %v", t1)
	//t2 := time.ParseInLocation(layout, "2015-2-2 00:00:01", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-2-3 01:00:00", time.Local)
	//t4 := time.ParseInLocation(layout, "2015-2-4 23:55:01", time.Local)
	//t5 := time.ParseInLocation(layout, "2015-2-5 12:00:00", time.Local)
	t6, _ := time.ParseInLocation(layout, "2015-2-6 12:00:01", time.Local)
	t7, _ := time.ParseInLocation(layout, "2015-2-7 10:00:00", time.Local)
	t8, _ := time.ParseInLocation(layout, "2015-2-8 09:00:00", time.Local) // 周日
	t9, _ := time.ParseInLocation(layout, "2015-2-8 08:30:00", time.Local)
	t10, _ := time.ParseInLocation(layout, "2015-2-9 08:30:00", time.Local)

	Convey("IsSameDay ", t, func() {
		So(IsSameDayUnix(t8.Unix(), t9.Unix()), ShouldEqual, true)
		So(IsSameDayUnix(t1.Unix(), t9.Unix()), ShouldEqual, false)
		So(IsSameDayUnix(t1.Unix(), t7.Unix()), ShouldEqual, false)
	})

	Convey("IsSameWeek ", t, func() {
		So(IsSameWeek(t3, t6), ShouldEqual, true)
		So(IsSameWeek(t1, t3), ShouldEqual, false)
		So(IsSameWeek(t10, t3), ShouldEqual, false)
	})

}

/*
func TestWeekTime(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-2-2 00:00:00", time.Local)
	t2, _ := time.ParseInLocation(layout, "2015-2-2 00:12:00", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-2-2 00:30:00", time.Local)
	t4, _ := time.ParseInLocation(layout, "2015-2-2 01:30:00", time.Local)
	t5, _ := time.ParseInLocation(layout, "2015-2-2 12:00:00", time.Local)

	Convey("Zero Should 0", t, func() {
		So(WeekTime(t1), ShouldEqual, 0)
		So(WeekTime(t2), ShouldEqual, 0)
	})

	Convey("Should Right", t, func() {
		So(WeekTime(t3), ShouldEqual, 1)
		So(WeekTime(t4), ShouldEqual, 3)
		So(WeekTime(t5), ShouldEqual, 24)
	})
}
*/
func TestAccountTime2Point(t *testing.T) {
	p1, r1 := AccountTime2Point(3607, 4, 60)

	Convey("Account normal", t, func() {
		So(p1, ShouldEqual, 60)
		So(r1, ShouldEqual, 3)
	})

	p2, r2 := AccountTime2Point(3306, 3303, 10)

	Convey("Account no enough", t, func() {
		So(p2, ShouldEqual, 0)
		So(r2, ShouldEqual, 3)
	})

	p3, r3 := AccountTime2Point(3300, 2200, 11)
	Convey("Account just right", t, func() {
		So(p3, ShouldEqual, 100)
		So(r3, ShouldEqual, 0)
	})

}

func TestGetDayBefore(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-5-13 11:44:01", time.Local)
	t2, _ := time.ParseInLocation(layout, "2015-5-13 12:12:00", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-5-14 00:00:01", time.Local)
	t4, _ := time.ParseInLocation(layout, "2015-5-15 01:30:00", time.Local)
	t5, _ := time.ParseInLocation(layout, "2015-5-16 12:00:00", time.Local)

	Convey("TestGetDayBefore", t, func() {
		So(GetDayBefore(t1, t1), ShouldEqual, 0)
		So(GetDayBefore(t1, t2), ShouldEqual, 0)
		So(GetDayBefore(t1, t3), ShouldEqual, 1)
		So(GetDayBefore(t1, t4), ShouldEqual, 2)
		So(GetDayBefore(t1, t5), ShouldEqual, 3)
	})
}

func TestDailyBegin(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-5-13 11:44:01", time.Local)
	t.Logf("DailyBeginUnix %v", DailyBeginUnix(t1.Unix()))
}

func TestDailyTime(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-2-2 00:00:00", time.Local)
	t2, _ := time.ParseInLocation(layout, "2015-2-2 00:12:00", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-2-2 00:30:00", time.Local)
	t4, _ := time.ParseInLocation(layout, "2015-2-2 01:30:00", time.Local)
	t5, _ := time.ParseInLocation(layout, "2015-2-2 12:00:00", time.Local)

	t9, _ := time.ParseInLocation(layout, "2015-2-9 23:55:00", time.Local)

	Convey("Zero Should 0", t, func() {
		So(DailyTime(t1), ShouldEqual, 0)
		So(DailyTime(t2), ShouldEqual, 0)
	})

	Convey("Should Right", t, func() {
		So(DailyTime(t3), ShouldEqual, 1)
		So(DailyTime(t4), ShouldEqual, 3)
		So(DailyTime(t5), ShouldEqual, 24)
		So(DailyTime(t9), ShouldEqual, 47)
	})
}

func TestDailyTimeStr(t *testing.T) {
	t1, _ := time.ParseInLocation(layout, "2015-2-2 00:00:00", time.Local)
	t2, _ := time.ParseInLocation(layout, "2015-2-2 00:12:00", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-2-2 00:30:00", time.Local)
	t4, _ := time.ParseInLocation(layout, "2015-2-2 01:30:00", time.Local)
	t5, _ := time.ParseInLocation(layout, "2015-2-2 12:00:00", time.Local)

	s1 := "00:00"
	s2 := "00:12"
	s3 := "00:30"
	s4 := "01:30"
	s5 := "12:00"

	Convey("Should Right", t, func() {
		So(DailyTimeFromString(s1), ShouldEqual, DailyTime(t1))
		So(DailyTimeFromString(s2), ShouldEqual, DailyTime(t2))
		So(DailyTimeFromString(s3), ShouldEqual, DailyTime(t3))
		So(DailyTimeFromString(s4), ShouldEqual, DailyTime(t4))
		So(DailyTimeFromString(s5), ShouldEqual, DailyTime(t5))
	})
}

func TestWeek(t *testing.T) {
	tn := time.Now().Unix()
	t.Logf("Week %d", GetWeek(tn))
}
