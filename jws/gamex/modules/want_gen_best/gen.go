package want_gen_best

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*WantGenBestModule
)

func init() {
	mInstance = make(map[uint]*WantGenBestModule, 6)
	modules.RegModule(modules.Module_WantGenBest, newDataVerModule)
}

func GetModule(shard uint) *WantGenBestModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newDataVerModule(sid uint) modules.ServerModule {
	m := genWantGenBestModule(sid)
	mInstance[sid] = m
	return m
}
