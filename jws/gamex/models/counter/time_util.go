package counter

import (
	"vcs.taiyouxi.net/platform/planx/util"
)

// 获取当前时间之前最近的一个刷新点的Unix时刻
func getLastUpdateTimeBeforeNow(nowT, updateDailyTime int64) int64 {
	nb := util.DailyTime2UnixTime(nowT, updateDailyTime)
	if nowT >= nb {
		// 如果当前时间点在刷新点之后则就是这个刷新点
		return nb
	} else {
		// 如果在之前,那应该是昨天的那个刷新点
		return nb - util.DaySec
	}
}
