package util

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSameByStartTime(t *testing.T) {
	t0, _ := time.ParseInLocation(layout, "2015-2-1 21:00:01", time.Local) // 周日
	t1, _ := time.ParseInLocation(layout, "2015-2-1 22:00:01", time.Local) // 周日
	t.Logf("local %v", t1)
	t2, _ := time.ParseInLocation(layout, "2015-2-2 00:00:01", time.Local)
	t3, _ := time.ParseInLocation(layout, "2015-2-2 01:00:00", time.Local)
	t4, _ := time.ParseInLocation(layout, "2015-2-2 15:55:01", time.Local)
	t5, _ := time.ParseInLocation(layout, "2015-2-3 04:00:00", time.Local)
	t6, _ := time.ParseInLocation(layout, "2015-2-6 12:00:01", time.Local)
	//t7, _ := time.ParseInLocation(layout, "2015-2-7 16:00:00", time.Local)
	t8, _ := time.ParseInLocation(layout, "2015-2-8 21:00:00", time.Local) // 周日
	t9, _ := time.ParseInLocation(layout, "2015-2-9 08:30:00", time.Local)

	tb1 := TimeToBalance{
		WeekDay:   0,
		DailyTime: DailyTimeFromString("05:00"),
	}

	tb2 := TimeToBalance{
		WeekDay:   7,
		DailyTime: DailyTimeFromString("22:00"),
	}

	Convey("IsSameUnixByStartTime ", t, func() {
		So(IsSameUnixByStartTime(t2.Unix(), t4.Unix(), tb1), ShouldEqual, false)
		So(IsSameUnixByStartTime(t3.Unix(), t1.Unix(), tb1), ShouldEqual, true)
		So(IsSameUnixByStartTime(t4.Unix(), t5.Unix(), tb1), ShouldEqual, true)
	})

	Convey("IsSameUnixByStartTimeWeek ", t, func() {
		So(IsSameUnixByStartTime(t6.Unix(), t8.Unix(), tb2), ShouldEqual, true)
		So(IsSameUnixByStartTime(t8.Unix(), t9.Unix(), tb2), ShouldEqual, false)
		So(IsSameUnixByStartTime(t0.Unix(), t1.Unix(), tb2), ShouldEqual, false)
	})

}
