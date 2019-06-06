package team_pvp

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func GetModule(sid uint) *TeamPvp {
	return mInstance[game.Cfg.GetShardIdByMerge(sid)]
}

var (
	mInstance map[uint]*TeamPvp
)

func init() {
	mInstance = make(map[uint]*TeamPvp, 6)
	modules.RegModule(modules.Module_TeamPvp, newTeamPvpModule)
}

func newTeamPvpModule(sid uint) modules.ServerModule {
	m := newTeamPvp(sid)
	mInstance[sid] = m
	return m
}
