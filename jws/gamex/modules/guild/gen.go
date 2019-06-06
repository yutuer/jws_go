package guild

import (
	"sync"

	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*GuildModule
	mutx      sync.RWMutex
)

func init() {
	mInstance = make(map[uint]*GuildModule, 6)
	modules.RegModule(modules.Module_Guild, newGuildModule)
}

func GetModule(shard uint) *GuildModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGuildModule(sid uint) modules.ServerModule {
	m := genGuildModule(sid)
	mInstance[sid] = m
	return m
}
