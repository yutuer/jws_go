package market_activity

import (
	"sync"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*MarketActivityModule
	mutx      sync.RWMutex
)

func init() {
	mInstance = make(map[uint]*MarketActivityModule, 6)
	modules.RegModule(modules.Module_MarketActivity, newMarketActivityModule)
}

func GetModule(shard uint) *MarketActivityModule {
	mutx.RLock()
	mutx.RUnlock()

	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newMarketActivityModule(sid uint) modules.ServerModule {
	mutx.Lock()
	defer mutx.Unlock()

	m := genMarketActivityModule(sid)
	mInstance[sid] = m
	return m
}
