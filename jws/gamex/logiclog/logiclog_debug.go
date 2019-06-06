package logiclog

import "vcs.taiyouxi.net/platform/planx/util/logiclog"

func LogiclogDebug(accountId string, avatar int, corpLvl uint32, channel string) {

	r := struct {
		Event string
	}{
		Event: "CheckEquipInBag",
	}

	TypeInfo := "LogiclogDebug"
	logiclog.Info(accountId, avatar, corpLvl, channel, TypeInfo, r, "")
}
