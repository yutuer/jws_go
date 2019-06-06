package global_mail

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*GlobalMailModule
)

func init() {
	mInstance = make(map[uint]*GlobalMailModule, 6)
	modules.RegModule(modules.Module_GlobalMail, newGlobalMailModule)
}

func GetModule(shard uint) *GlobalMailModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGlobalMailModule(sid uint) modules.ServerModule {
	m := genGlobalMailModule(sid)
	mInstance[sid] = m
	return m
}
