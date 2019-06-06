package gvg

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*gvgModule
)

func init() {
	mInstance = make(map[uint]*gvgModule, 6)
	modules.RegModule(modules.Module_GvG, newGvgModule)
}

func GetModule(shard uint) *gvgModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGvgModule(sid uint) modules.ServerModule {
	m := genModule(sid)
	mInstance[sid] = m
	return m
}

func genModule(sid uint) *gvgModule {
	return &gvgModule{
		sid: sid,
	}
}
