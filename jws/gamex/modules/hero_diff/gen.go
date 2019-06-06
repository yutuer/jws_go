package hero_diff

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*heroDiffModule
)

func init() {
	mInstance = make(map[uint]*heroDiffModule, 6)
	modules.RegModule(modules.Module_HeroDiff, newHeroDiffModule)
}

func GetModule(shard uint) *heroDiffModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newHeroDiffModule(sid uint) modules.ServerModule {
	m := genModule(sid)
	mInstance[sid] = m
	return m
}

func genModule(sid uint) *heroDiffModule {
	return &heroDiffModule{
		sid: sid,
	}
}
