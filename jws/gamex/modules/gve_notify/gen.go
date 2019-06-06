package gve_notify

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*module
)

func init() {
	mInstance = make(map[uint]*module, 6)
	modules.RegModule(modules.Module_Gve, newGveModule)
	initRand()
}

func GetModule(shard uint) *module {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGveModule(sid uint) modules.ServerModule {
	m := genModule(sid)
	mInstance[sid] = m
	return m
}
