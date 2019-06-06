package balance

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func GetModule(sid uint) *BalanceModule {
	return mInstance[game.Cfg.GetShardIdByMerge(sid)]
}

var (
	mInstance map[uint]*BalanceModule
)

func init() {
	mInstance = make(map[uint]*BalanceModule, 6)
	modules.RegModule(modules.Module_Balance, newBalanceModule)
}

func newBalanceModule(sid uint) modules.ServerModule {
	m := newBalance(sid)
	mInstance[sid] = m
	return m
}
