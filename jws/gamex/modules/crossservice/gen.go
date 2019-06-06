package crossservice

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*CrossServiceModule
)

func init() {
	mInstance = make(map[uint]*CrossServiceModule)
	modules.RegModule(modules.Module_CrossService, newCrossServiceModule)
}

//GetModule ..
func GetModule(sid uint) *CrossServiceModule {
	mergeID := game.Cfg.GetShardIdByMerge(sid)
	return mInstance[mergeID]
}

func newCrossServiceModule(sid uint) modules.ServerModule {
	mergeID := game.Cfg.GetShardIdByMerge(sid)
	if _, exist := mInstance[mergeID]; exist {
		return mInstance[mergeID]
	}

	m := genCrossServiceModule(mergeID)
	mInstance[mergeID] = m
	return m
}
