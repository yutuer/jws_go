package global_count

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*GlobalCountModule
)

func init() {
	mInstance = make(map[uint]*GlobalCountModule, 6)
	modules.RegModule(modules.Module_GlobalCount, newGlobalCountModule)
}

func GetModule(shard uint) *GlobalCountModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGlobalCountModule(sid uint) modules.ServerModule {
	m := genGlobalCountModule(sid)
	mInstance[sid] = m
	return m
}
