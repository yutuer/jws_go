package teamboss

import (
	"time"

	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamLeaveRoom struct {
	Sid  uint
	Acid string
	Info helper.LeaveRoomInfo
}

//RetAttack ..
type RetLeaveRoom struct {
	Info helper.LeaveRoomRetInfo
}

//MethodLeaveRoom ..
type MethodLeaveRoom struct {
	module.BaseMethod
}

func newMethodLeaveRoom(m module.Module) *MethodLeaveRoom {
	return &MethodLeaveRoom{
		module.BaseMethod{Method: MethodLeaveRoomID, Module: m},
	}
}

//NewParam ..
func (m *MethodLeaveRoom) NewParam() module.Param {
	return &ParamLeaveRoom{}
}

//NewRet ..
func (m *MethodLeaveRoom) NewRet() module.Ret {
	return &RetLeaveRoom{}
}

//Do ..
func (m *MethodLeaveRoom) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamLeaveRoom)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		ret = RetLeaveRoom{
			Info: helper.LeaveRoomRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.TgtAcID)
	if player == nil || roomInfo.GetPlayer(info.OptAcID) == nil {
		ret = RetLeaveRoom{
			Info: helper.LeaveRoomRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}
	isKick := info.OptAcID != info.TgtAcID
	if isKick && info.OptAcID != roomInfo.LeadAcID {
		ret = RetLeaveRoom{
			Info: helper.LeaveRoomRetInfo{
				Code: helper.RetCodeOptLimitPermission,
			},
		}
		return
	}
	if isKick && roomInfo.RoomState == helper.TBRoomFight {
		ret = RetLeaveRoom{
			Info: helper.LeaveRoomRetInfo{
				Code: helper.RetCodeKickFightingRoom,
			},
		}
		return
	}

	//if roomInfo.RoomState == helper.TBRoomFight {
	//	// a room only include 2 players
	//	if roomInfo.LostPlayer != info.TgtAcID && roomInfo.LostPlayer != "" {
	//		bm.Room.DelRoom(info.RoomID)
	//	} else {
	//		roomInfo.LostPlayer = info.TgtAcID
	//	}
	//	ret = RetLeaveRoom{
	//		Info: helper.LeaveRoomRetInfo{
	//			Param: lrp,
	//		},
	//	}
	//	return
	//}
	refresh := true
	logs.Debug("[TeamBoss] Before leave, roomInfo: %v", *roomInfo)
	roomInfo.RemovePlayer(info.TgtAcID)
	if roomInfo.PlayerCount() <= 0 {
		bm.Room.DelRoom(info.RoomID)
		refresh = false
	}
	lrp := helper.LeaveRoomParam{
		RoomID:    roomInfo.ID,
		TgtAcID:   info.TgtAcID,
		LeaveTime: time.Now().Unix(),
		IsRefresh: refresh,
	}
	ret = RetLeaveRoom{
		Info: helper.LeaveRoomRetInfo{
			Param: lrp,
		},
	}
	if isKick {
		bm.Kick(uint32(player.SimpleInfo.Sid), lrp, []string{player.SimpleInfo.AcID})
	}
	roomInfo = bm.Room.GetRoom(info.RoomID)
	if roomInfo != nil {
		roomDetail := roomInfo.genRoomDetailInfo()
		acids := roomInfo.GetExtraPlayer(info.TgtAcID)
		for k, v := range acids {
			bm.RoomInfo(k, roomDetail, v)
		}
		logs.Debug("[TeamBoss] After leave, roomInfo: %v", *roomInfo)
	}

	return
}
