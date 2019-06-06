package gate_enemy_push

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GEPush interface {
	ReadyGatesEnemyActASync(guildID string, endTime int64,
		members []helper.AccountSimpleInfo)
	StartGatesEnemyActASync(guildID string, endTime int64,
		members []helper.AccountSimpleInfo)
}

var (
	//
	TmpGEPush map[uint]GEPush
)

func GateEnemyReady(sid uint, guuid string, endTime int64,
	mems []helper.AccountSimpleInfo) {
	logs.Debug("gate_enemy_push GatesEnemyReady %v", guuid)
	if gep, ok := TmpGEPush[sid]; ok {
		gep.ReadyGatesEnemyActASync(guuid, endTime, mems)
	}
}
func GateEnemyStart(sid uint, guuid string, endTime int64,
	mems []helper.AccountSimpleInfo) {
	logs.Debug("gate_enemy_push GatesEnemyStart %v", guuid)
	if gep, ok := TmpGEPush[sid]; ok {
		gep.StartGatesEnemyActASync(guuid, endTime, mems)
	}
}
