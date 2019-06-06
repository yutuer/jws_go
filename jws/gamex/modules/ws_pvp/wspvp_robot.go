package ws_pvp

import (
	"fmt"
	"strings"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

type WsPvpRobot struct {
	GroupId uint
	Sid     uint
	RobotId uint
}

func GenRobotId(groupId, sid uint32, robId int) string {
	return fmt.Sprintf("%s:%d:%d:%d", WS_PVP_ROBOT_ID_PREFIX, groupId, sid, robId)
}

func InitRobot(sid uint32) {
	groupCfg := gamedata.GetWSPVPGroupCfg(sid)
	groupId := groupCfg.GetWspvpGroupID()
	names := gamedata.RandRobotNames(WS_PVP_RANK_MAX)
	robotIds := make([]string, WS_PVP_RANK_MAX)
	for i := 0; i < WS_PVP_RANK_MAX; i++ {
		robotIds[i] = GenRobotId(groupId, sid, i)
	}
	initRobotInfo(int(sid), int(groupId), robotIds, names)
	initNameRedis(sid, names, robotIds)
}

func IsRobotId(acid string) bool {
	return strings.Index(acid, WS_PVP_ROBOT_ID_PREFIX) == 0
}
