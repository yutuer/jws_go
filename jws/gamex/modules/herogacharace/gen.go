package herogacharace

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func Get(sid uint) *HeroGachaRace {
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
		ms  *HeroGachaRace
	}
)

func init() {
	modules.RegModule(modules.Module_HeroGachaRace, newModule)
}

func newModule(sid uint) modules.ServerModule {
	m := NewHeroGachaRace(sid)
	moduleMap = append(moduleMap, struct {
		sid uint
		ms  *HeroGachaRace
	}{
		sid,
		m,
	})
	return m
}
