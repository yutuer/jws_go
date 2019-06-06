package redeemCodeModule

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*redeemCodeModules
)

func init() {
	mInstance = make(map[uint]*redeemCodeModules, 6)
	modules.RegModule(modules.Module_RedeemCode, newRedeemModule)
}

func GetModule(shard uint) *redeemCodeModules {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newRedeemModule(sid uint) modules.ServerModule {
	m := genredeemCodeModules(sid)
	mInstance[sid] = m
	return m
}
