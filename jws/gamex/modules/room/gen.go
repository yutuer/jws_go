package room

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	ROOM_WARN_ENTERROOM_NUM_WRONG = errCode.FenghuoRoomNumWarn
	//主机应该等待其他玩家都Ready后才能自己Ready,或者叫开始
	ROOM_WARN_MASTER_WAITING_OTHERS_READY = errCode.FenghuoWaitOthersReady
	ROOM_WARN_CREATEROOM_NOENOUGH_MONEY   = errCode.FenghuoNoMoney
	ROOM_WARN_GETROOMINFO_NUM_WRONG       = errCode.FenghuoRoomNumWarn

	ROOM_ERR_UNKNOWN                              = 201
	ROOM_ERR_YOU_NO_PERMISSION_TO_KICK            = 202
	ROOM_ERR_GIVESYNC                             = 203
	ROOM_ERR_GIVESYNC_EXTRAREWARD_COUNTER_CONSUME = 204
	ROOM_ERR_GETROOMINFO_NOFOUND_ROOMID           = 205
	ROOM_ERR_MKREWARD_NO_ENOUGH_MONEY             = 206
)

func Get(sid uint) *module {
	sid = game.Cfg.GetShardIdByMerge(sid)
	for i := 0; i < len(moduleMap); i++ {
		if moduleMap[i].sid == sid {
			return moduleMap[i].ms
		}
	}
	return nil
}

var (
	moduleMap []struct {
		sid uint
		ms  *module
	}
)

func init() {
	modules.RegModule(modules.Module_Room, newModule)
}

func newModule(sid uint) modules.ServerModule {
	m := New(sid)
	moduleMap = append(moduleMap, struct {
		sid uint
		ms  *module
	}{
		sid,
		m,
	})
	return m
}
