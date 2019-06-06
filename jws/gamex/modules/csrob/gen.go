package csrob

import (
	"sync"

	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*CSRobModule
	mutex     sync.RWMutex
)

func init() {
	mInstance = make(map[uint]*CSRobModule)
	modules.RegModule(modules.Module_CSRob, newCSRobModule)
}

//GetModule ..
func GetModule(shard uint) *CSRobModule {
	mutex.RLock()
	defer mutex.RUnlock()

	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newCSRobModule(sid uint) modules.ServerModule {
	mutex.Lock()
	defer mutex.Unlock()

	m := genCSRobModule(sid)
	mInstance[sid] = m
	return m
}
