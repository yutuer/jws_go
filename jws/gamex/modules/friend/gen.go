package friend

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*friendModule
)

func init() {
	mInstance = make(map[uint]*friendModule, 6)
	modules.RegModule(modules.Module_Friend, newfriendModule)
}

func GetModule(shard uint) *friendModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newfriendModule(sid uint) modules.ServerModule {
	m := genModule(sid)
	mInstance[sid] = m
	return m
}

func genModule(sid uint) *friendModule {
	return &friendModule{
		sid: sid,
	}
}
