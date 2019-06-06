package player_msg

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const ()

var (
	mInstance map[uint]*PlayerMsgModule
)

func init() {
	mInstance = make(map[uint]*PlayerMsgModule, 6)
	modules.RegModule(modules.Module_PlayerMsg, newPlayerMsgModule)
}

func GetModule(shard uint) *PlayerMsgModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newPlayerMsgModule(sid uint) modules.ServerModule {
	m := genPlayerMsgModule(sid)
	mInstance[sid] = m
	return m
}
