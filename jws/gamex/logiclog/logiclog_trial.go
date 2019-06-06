package logiclog

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LogicInfo_TrialLvl struct {
	LevelId       int32
	IsWin         int
	Gs            int
	CostTime      int64
	SkillGenerals [helper.DestinyGeneralSkillMax]int
}
type LogicInfo_TrialReset struct {
	MostLvl int32
}

type LogicInfo_TrialSweep struct {
	SweepEvent string
}

func LogTrialLvlFinish(accountId string, avatar int, corpLvl uint32, channel string,
	lvlId int32, isWin bool, costTime int64,
	gs int, skillGeneral [helper.DestinyGeneralSkillMax]int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[TrialLvlFinish][%s]  lvl %d isWin %v  %s", accountId, lvlId, isWin, info)

	iwin := 0
	if !isWin {
		iwin = 1
	}

	r := LogicInfo_TrialLvl{
		LevelId:       lvlId,
		IsWin:         iwin,
		Gs:            gs,
		CostTime:      costTime,
		SkillGenerals: skillGeneral,
	}
	TypeInfo := LogicTag_TrialLvlFinish
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogTrialReset(accountId string, avatar int, corpLvl uint32, channel string,
	mostLvl int32, fgs GetLastSetCurLogType,
	format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[TrialReset][%s]    %s", accountId, info)

	r := LogicInfo_TrialReset{
		MostLvl: mostLvl,
	}

	TypeInfo := LogicTag_TrialReset
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogTrialSweep(accountId string, avatar int, corpLvl uint32, channel string,
	sweepEvent string, fgs GetLastSetCurLogType,
	format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[TrialSweep][%s]    %s", accountId, info)

	r := LogicInfo_TrialSweep{
		SweepEvent: sweepEvent,
	}

	TypeInfo := LogicTag_TrialSweep
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}
