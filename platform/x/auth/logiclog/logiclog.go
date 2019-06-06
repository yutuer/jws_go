package logiclog

import (
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const BITag = "[BI]"
const DebugTag = "[Debug]"

type loginInfo struct {
	DeviceID  string
	ChannelId string
	Device    string
	Team      string
	IsReg     bool
	Gameid    uint
	Shardid   uint
	CfgTeam   string
}

func LogLogin(uid, DeviceID, channelId, device string, isReg bool) {
	r := loginInfo{
		DeviceID:  DeviceID,
		ChannelId: channelId,
		Device:    device,
		IsReg:     isReg,
	}
	logs.Trace("LogLogin %s %v", BITag, r)
	logiclog.Error(uid, 0, 0, channelId, "Login", r, "", BITag)
}

func LogLoginTeam(uid, team, channelId, cfgTeam string, gameid, sharid uint) {
	r := loginInfo{
		Team:    team,
		Gameid:  gameid,
		Shardid: sharid,
		CfgTeam: cfgTeam,
	}
	logs.Trace("LoginTeam %s %s %v", BITag, uid, r)
	logiclog.Error(uid, 0, 0, channelId, "LoginTeam", r, "", BITag)
}

func LogDebug(info interface{}) {
	logs.Trace("LogDebug %s %v", DebugTag, info)
	logiclog.Error("", 0, 0, "", "Debug", info, "", DebugTag)
}
