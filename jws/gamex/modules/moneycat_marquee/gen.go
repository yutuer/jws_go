package moneycat_marquee

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*MoneyCatModule
)

func init() {
	mInstance = make(map[uint]*MoneyCatModule, 10)
	modules.RegModule(modules.Module_MoneyCat, newDataVerModule)
}

func GetModule(shard uint) *MoneyCatModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newDataVerModule(sid uint) modules.ServerModule {
	m := genDestGenFirstModule(sid)
	mInstance[sid] = m
	return m
}
