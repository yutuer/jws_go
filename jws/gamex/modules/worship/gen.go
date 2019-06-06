package worship

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func Get(sid uint) *module {
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
		ms  *module
	}
)

func init() {
	modules.RegModule(modules.Module_Worship, newModule)
}

func newModule(sid uint) modules.ServerModule {
	m := New(sid)
	moduleMap = append(moduleMap, struct {
		sid uint
		ms  *module
	}{
		sid,
		m,
	})
	return m
}
