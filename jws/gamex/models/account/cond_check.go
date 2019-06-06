package account

import (
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func CondCheck(moduleId uint32, a *Account) bool {
	cond := gamedata.GetCond(moduleId)
	if cond == nil {
		logs.Error("cfg condition %d not found ", moduleId)
		return false
	}
	switch cond.GetConditionType() {
	case gamedata.FteConditionRoleOpenTypCorpLv:
		reqLevel, _ := strconv.Atoi(cond.GetConditionValue())
		level, _ := a.Profile.GetCorp().GetXpInfo()
		return level >= uint32(reqLevel)
	case gamedata.FteConditionRoleOpenTypStage:
		preStage := []string{cond.GetConditionValue()}
		return a.Profile.GetStage().IsAllPreStagePass(preStage)
	default:
		logs.Error("cfg Condition has new condition type but server not know: %d", cond.GetConditionType())
		return false
	}
}
