package sPvpRander

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*simplePvpRanderModule
)

func init() {
	mInstance = make(map[uint]*simplePvpRanderModule, 6)
	modules.RegModule(modules.Module_SPvpRander, newSimplePvpModule)
}

func GetModule(shard uint) *simplePvpRanderModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newSimplePvpModule(sid uint) modules.ServerModule {
	m := genSimplePvpRanderModule(sid)
	mInstance[sid] = m
	return m
}
