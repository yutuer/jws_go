package gates_enemy

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/gate_enemy_push"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*GatesEnemyModule
)

func init() {
	mInstance = make(map[uint]*GatesEnemyModule, 6)
	gate_enemy_push.TmpGEPush = make(map[uint]gate_enemy_push.GEPush, 6)
	modules.RegModule(modules.Module_GateEnemy, newGateEnemyModule)
}

func GetModule(shard uint) *GatesEnemyModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newGateEnemyModule(sid uint) modules.ServerModule {
	m := genGatesEnemyModule(sid)
	mInstance[sid] = m
	gate_enemy_push.TmpGEPush[sid] = m
	return m
}
