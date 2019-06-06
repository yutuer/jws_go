package rank

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*RankModule
)

func init() {
	mInstance = make(map[uint]*RankModule, 6)
	modules.RegModule(modules.Module_Rank, newRankModule)
}

func GetModule(shard uint) *RankModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newRankModule(sid uint) modules.ServerModule {
	m := genRankModule(sid)
	mInstance[sid] = m
	return m
}
