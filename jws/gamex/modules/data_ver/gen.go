package data_ver

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*DataVerModule
)

func init() {
	mInstance = make(map[uint]*DataVerModule, 6)
	modules.RegModule(modules.Module_DataVer, newDataVerModule)
}

func GetModule(shard uint) *DataVerModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newDataVerModule(sid uint) modules.ServerModule {
	m := genDataVerModule(sid)
	mInstance[sid] = m
	return m
}
