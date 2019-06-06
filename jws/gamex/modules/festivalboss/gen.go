package festivalboss

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*FestivalBossModule
)

func init() {
	mInstance = make(map[uint]*FestivalBossModule, 10)
	modules.RegModule(modules.Module_FestivalBoss, newDataVerModule)
}

func GetModule(shard uint) *FestivalBossModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newDataVerModule(sid uint) modules.ServerModule {
	m := genFestivalBossModule(sid)
	mInstance[sid] = m
	return m
}
