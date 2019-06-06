package title_rank

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func GetModule(sid uint) *TitleRank {
	return mInstance[game.Cfg.GetShardIdByMerge(sid)]
}

var (
	mInstance map[uint]*TitleRank
)

func init() {
	mInstance = make(map[uint]*TitleRank, 6)
	modules.RegModule(modules.Module_TitleRank, newTitleRankModule)
}

func newTitleRankModule(sid uint) modules.ServerModule {
	m := newTitleRank(sid)
	mInstance[sid] = m
	return m
}
