package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/logiclog"
	"vcs.taiyouxi.net/platform/x/auth/models"
)

const (
	Ch_YYB_New = "6000"
	Ch_YYB_Old = "32"
)

var (
	channel_map = make(map[string]string, 8)
)

func init() {
	channel_map[Ch_YYB_New] = Ch_YYB_Old
}

/*
	韩国google和onestore互相冲突情况处理
	1、5005的还继续在5005里玩
	2、5006的先判断在5006里是否有账号，如果有继续在5006里玩；如果没有就去5005
*/
func ko_spec_channel(sdkFlag, sdkUid, _channelId string) string {
	channelId := convertChannel(_channelId)
	deviceID := fmt.Sprintf("%s@%s@%s", sdkUid, channelId, sdkFlag)
	_, err, _ := models.TryToCheckDeviceInfo(deviceID)
	if channelId == util.Android_Enjoy_Korea_OneStore_Channel {
		if err == models.XErrUserNotExist {
			return util.Android_Enjoy_Korea_GP_Channel
		}
	}
	return channelId
}

func login(c *gin.Context, sdkFlag, sdkUid, channelId, device string) (res bool, authToken string, display, uidstr string) {
	channelId = convertChannel(channelId)
	deviceID := fmt.Sprintf("%s@%s@%s", sdkUid, channelId, sdkFlag)
	di, err, isGM := models.TryToCheckDeviceInfo(deviceID)
	isReg := false
	if err == models.XErrUserNotExist {
		//可以创建新帐号
		di, err = models.AddDevice(deviceID, "", channelId, device)
		uid := db.InvalidUserID
		if di != nil {
			uid = di.UserId
		}
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("AddDevice error: uid(%d), error(%s)", uid, err),
				errorctl.ClientErrorMaybeDBProblem,
			)
			return false, "", "", ""
		}
		isReg = true
	} else if err != nil {
		errorctl.CtrlErrorReturn(c, "[Auth.Device]",
			err,
			errorctl.ClientErrorMaybeDBProblem,
		)
		return false, "", "", ""
	}

	uid := di.UserId

	logiclog.LogLogin(uid.String(), deviceID, channelId, device, isReg)
	if !isGM {
		if err, ban_time, ban_reason := models.UserBanInfo(uid); err != nil {
			errorctl.CtrlBanReturn(c, "[Auth.Device]",
				err,
				errorctl.ClientErrorBanByGM, ban_time, ban_reason)
			return false, "", "", ""
		}
		//下面逻辑err == nil
		authToken, err = models.AuthToken(uid, deviceID)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				err,
				errorctl.ClientErrorMaybeDBProblem,
			)
			return false, "", "", ""
		}
	} else {
		authToken = uuid.NewV4().String()
	}

	//Auth token 发送到登录校验服务器，并设置有效时间
	notifyLoginServer(authToken, uid)
	return true, authToken, di.Display, uid.String()
}

func _getUserShardInfo(uid string) (userShard []models.ShardHasRole, lastShard string) {
	_userShard, err := models.GetUserShardHasRole(uid)
	if err != nil {
		logs.Warn("[quick] login GetUserShard err: %v", err)
	} else {
		userShard = _userShard
	}

	_lastShard, err := models.GetUserLastShard(uid)
	if err != nil {
		logs.Warn("[quick] login GetUserLastShard err: %v", err)
	} else {
		lastShard = _lastShard
	}
	return
}

func convertChannel(channel string) string {
	old, ok := channel_map[channel]
	if ok {
		return old
	}
	return channel
}

func GetServerTeam(accountId []byte, gid uint, sid uint) (string, string) {
	teamAB_parent := fmt.Sprintf("%s/%d/%d/%s", config.Cfg.EtcdRoot, gid, sid, etcd.KeyTeamAB)
	teamAB, err := etcd.Get(teamAB_parent)
	if err != nil {
		return "a", teamAB
	}
	switch teamAB {
	case "A":
		return "a", teamAB
	case "B":
		return "b", teamAB
	case "AB":
		lastLetterByte := accountId[len(accountId)-1]
		if isOdd(int(lastLetterByte)) {
			return "a", teamAB
		} else {
			return "b", teamAB
		}
	}
	logs.Debug("teamAB_parent %s team %s", teamAB_parent, "team is wrong")
	return "", ""

}
func isOdd(num int) bool {
	return num%2 != 0
}
