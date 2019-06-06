package city_fish

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*CityFish
)

func init() {
	mInstance = make(map[uint]*CityFish, 6)
	modules.RegModule(modules.Module_CityFish, newFishModule)
}

func GetModule(shard uint) *CityFish {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newFishModule(sid uint) modules.ServerModule {
	m := genFishModule(sid)
	mInstance[sid] = m
	return m
}
