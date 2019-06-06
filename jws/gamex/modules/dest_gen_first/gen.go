package dest_gen_first

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*DestGenFirstModule
)

func init() {
	mInstance = make(map[uint]*DestGenFirstModule, 6)
	modules.RegModule(modules.Module_DestingGeneralFirst, newDataVerModule)
}

func GetModule(shard uint) *DestGenFirstModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newDataVerModule(sid uint) modules.ServerModule {
	m := genDestGenFirstModule(sid)
	mInstance[sid] = m
	return m
}
