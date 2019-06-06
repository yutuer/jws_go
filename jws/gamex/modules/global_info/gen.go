package global_info

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*GlobalInfoModule
)

func init() {
	mInstance = make(map[uint]*GlobalInfoModule, 6)
	modules.RegModule(modules.Module_GlobalInfo, newGlobalInfoModule)
}

func GetModule(shard uint) *GlobalInfoModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGlobalInfoModule(sid uint) modules.ServerModule {
	m := genGlobalInfoModule(sid)
	mInstance[sid] = m
	return m
}
