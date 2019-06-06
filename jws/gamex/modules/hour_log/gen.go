package hour_log

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func Get(sid uint) *HourLog {
	sid = game.Cfg.GetShardIdByMerge(sid)
	for i := 0; i < len(moduleMap); i++ {
		if moduleMap[i].sid == sid {
			return moduleMap[i].ms
		}
	}
	return nil
}

var (
	moduleMap []struct {
		sid uint
		ms  *HourLog
	}
)

func init() {
	modules.RegModule(modules.Module_HourLog, newModule)
}

func newModule(sid uint) modules.ServerModule {
	m := NewHourLog(sid)
	moduleMap = append(moduleMap, struct {
		sid uint
		ms  *HourLog
	}{
		sid,
		m,
	})
	return m
}
