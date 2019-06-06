package hour_log

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func TableBIHourModule(sid uint) string {
	return fmt.Sprintf("%d:%d:BIHour", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

// hash
func TableBIHourRegister(sid uint) string {
	return fmt.Sprintf("%d:%d:BIHourRegister", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

// hyperloglog
func TableBIHourDevice(sid uint, channel string) string {
	return fmt.Sprintf("%d:%d:BIHourDevice-%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid), channel)
}

// hyperloglog
func TableBIHourActive(sid uint, channel string) string {
	return fmt.Sprintf("%d:%d:BIHourActive-%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid), channel)
}

// hyperloglog
func TableBIHourChargeAcid(sid uint, channel string) string {
	return fmt.Sprintf("%d:%d:BIHourChargeAcid-%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid), channel)
}

// hash
func TableBIHourChargeSum(sid uint) string {
	return fmt.Sprintf("%d:%d:BIHourChargeSum", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

// set
func TableBIHourChannels(sid uint) string {
	return fmt.Sprintf("%d:%d:BIHourChannels", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

// sorted set
func TableBIHourCCU(sid uint) string {
	return fmt.Sprintf("%d:%d:BIHourCCU", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
