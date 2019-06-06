package logiclog

import "vcs.taiyouxi.net/platform/planx/util/eslogger"

/*
	log分级说明：
		Error: 游戏事件log，特别是英雄和运营需要的
		Info: 辅助事件log，目前就：回城，进关卡，进boss战，用来辅助计算关卡内时间
		Debug: ClientTimeEvent log
		Trace: MN log
*/
var PLogicLogger *eslogger.ESLogger

func init() {
	PLogicLogger = &eslogger.ESLogger{}
}

// bi event log
func ErrorForGuild(accountID, channel string, guildID string, typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Error(PLogicLogger, accountID, 0, 0, channel, guildID, typeInfo, info, "", format, v...)
}

func Error(accountID string, avatar int, corpLvl uint32, channel string,
	typeInfo string, info interface{}, lastEvent string, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Error(PLogicLogger, accountID, avatar, corpLvl, channel, "", typeInfo, info, lastEvent, format, v...)
}

// client event log
func Info(accountID string, avatar int, corpLvl uint32, channel string,
	typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Info(PLogicLogger, accountID, avatar, corpLvl, channel, typeInfo, info, format, v...)
}

// ClientTimeEvent log
func Debug(accountID string, avatar int, corpLvl uint32, channel string,
	typeInfo string, info interface{}, last string, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Debug(PLogicLogger, accountID, avatar, corpLvl, channel, typeInfo, info, last, format, v...)
}

// mn
func Trace(accountID, typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Trace(PLogicLogger, accountID, typeInfo, info, format, v...)
}

// multiplayer
func MultiInfo(typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Info(PLogicLogger, "", 0, 0, "", typeInfo, info, format, v...)
}

func MultiDebug(typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Debug(PLogicLogger, "", 0, 0, "", typeInfo, info, "", format, v...)
}

func MultiTrace(typeInfo string, info interface{}, format string, v ...interface{}) {
	if PLogicLogger == nil {
		return
	}
	eslogger.Trace(PLogicLogger, "", typeInfo, info, format, v...)
}
