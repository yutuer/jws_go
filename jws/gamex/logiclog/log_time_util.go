package logiclog

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
)

/*
	因为bilog全部要求用北京时间，所有做此文件
*/

var (
	timeLocal         *time.Location
	timeUnixBeginUnix int64
)

func init() {
	timeLocal, _ = time.LoadLocation("Asia/Shanghai")
	_TimeUnixBegin, _ := time.ParseInLocation("2006/1/2", "2000/1/1", timeLocal)
	timeUnixBeginUnix = _TimeUnixBegin.Unix()
}

func GetBILogTL() *time.Location {
	return timeLocal
}

func BIDailyBeginUnix(u int64) int64 {
	return u - (u-timeUnixBeginUnix)%util.DaySec
}
