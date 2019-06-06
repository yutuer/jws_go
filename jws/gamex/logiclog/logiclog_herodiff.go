package logiclog

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LogicInfo_HeroDiffStart struct {
	GS      int
	StageID int
}

type LogicInfo_HeroDiffFinish struct {
	GS      int
	StageID int
	Score   int
	IsSweep bool
}

func LogHeroDiffStart(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, stageID int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroDiffStart][%s]  stageID %d info %v ", accountId, stageID, info)
	r := LogicInfo_HeroDiffStart{
		GS:      gs,
		StageID: stageID,
	}
	TypeInfo := LogicTag_HeroDiffStart
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogHeroDiffFinish(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, stageID int, score int, isSweep bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	r := LogicInfo_HeroDiffFinish{
		GS:      gs,
		StageID: stageID,
		Score:   score,
		IsSweep: isSweep,
	}
	TypeInfo := LogicTag_HeroDiffFinish
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("["+TypeInfo+"][%s][%s][%v]", accountId, info, r)
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}
