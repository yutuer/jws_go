package rank

import (
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func GetSevenDayRankStartTime(shard uint) int64 {
	return GetModule(shard).sevenDayRankStartTime
}

// 获得开服后的天数
func GetSevenDayRankDays(shard uint) int64 {
	return util.GetDayBeforeUnix(GetModule(shard).sevenDayRankStartTime,
		game.GetNowTimeByOpenServer(shard))
}

// 设置为排行榜开始第几天，1~7
func DebugSetSevenDayRankday(shard uint, day uint32) {
	GetModule(shard).sevenDayRankStartTime = util.DailyBeginUnix(game.GetNowTimeByOpenServer(shard)) -
		int64(day-1)*util.DaySec
	logs.Trace("DebugSetSevenDayRankday %d", GetModule(shard).sevenDayRankStartTime)
}
