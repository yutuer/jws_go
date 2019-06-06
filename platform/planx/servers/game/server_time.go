package game

import (
	"time"
	"vcs.taiyouxi.net/platform/planx/util"
)

//
// time 相关
//

// 注册 与玩家注册时间相关的时间刻画
// 下面是系统的Now时间的

// 返回注册后的秒数
func GetUnixTimeAfterRegister(regtime, nowtime int64) int64 {
	return GetUnixTimeAfterRegisterByT(regtime, nowtime)
}

// 返回注册后的天数, 注册当天返回0 以后每过一天加一
func GetDayAfterRegister(regtime, nowtime int64) int64 {
	return GetDayAfterRegisterByT(regtime, nowtime)
}

// 服务器时间 与玩家注册时间相关的时间刻画

// 返回开服后的秒数
func GetUnixTimeAfterOpenServer(shard uint) int64 {
	return GetUnixTimeAfterOpenServerByT(shard, time.Now().Unix())
}

// 返回开服后的天数, 注册当天返回0 以后每过一天加一
func GetDayAfterOpenServer(shard uint) int64 {
	return GetDayAfterOpenServerByT(shard, time.Now().Unix())
}

// 下面是提供给玩家自带的Now时间的

// 返回注册后的秒数
func GetUnixTimeAfterRegisterByT(regtime, t int64) int64 {
	r := t - regtime
	return r
}

// 返回注册后的天数, 注册当天返回0 以后每过一天加一
func GetDayAfterRegisterByT(regtime, t int64) int64 {
	return util.GetDayBeforeUnix(regtime, t)
}

// 服务器时间 与玩家注册时间相关的时间刻画

// 返回服务器开服时间
func ServerStartTime(sid uint) int64 {
	return util.GetServerStartTime(Cfg.GetShardIdByMerge(sid))
}

// 返回开服后的秒数
func GetUnixTimeAfterOpenServerByT(shard uint, t int64) int64 {
	r := t - ServerStartTime(shard)
	return r
}

// 返回开服后的天数, 注册当天返回0 以后每过一天加一
func GetDayAfterOpenServerByT(shard uint, t int64) int64 {
	return util.GetDayBeforeUnix(ServerStartTime(shard), t)
}

// 返回当前绝对时间，由开服时间计算得来
func GetNowTimeByOpenServer(shard uint) int64 {
	return GetUnixTimeAfterOpenServer(shard) + ServerStartTime(shard)
}
