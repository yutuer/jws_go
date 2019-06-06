package data_hot_update

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func GetModule(sid uint) *DataHotUpdateModule {
	return mInstance[game.Cfg.GetShardIdByMerge(sid)]
}

var (
	mInstance map[uint]*DataHotUpdateModule
)

func init() {
	mInstance = make(map[uint]*DataHotUpdateModule, 6)
	modules.RegModule(modules.Module_DataHotUpdate, newDataHotUpdateModule)
}

func newDataHotUpdateModule(sid uint) modules.ServerModule {
	m := genDataHotUpdateModule(sid)
	mInstance[sid] = m
	return m
}
