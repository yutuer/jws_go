package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamChangeRoomStatus struct {
	Sid  uint
	Acid string
	Code int
	Info helper.ChangeRoomStatusInfo
}

//RetAttack ..
type RetChangeRoomStatus struct {
	Info helper.ChangeRoomStatusRetInfo
}

//MethodChangeRoomStatus ..
type MethodChangeRoomStatus struct {
	module.BaseMethod
}

func newMethodChangeRoomStatus(m module.Module) *MethodChangeRoomStatus {
	return &MethodChangeRoomStatus{
		module.BaseMethod{Method: MethodChangeRoomStatusID, Module: m},
	}
}

//NewParam ..
func (m *MethodChangeRoomStatus) NewParam() module.Param {
	return &ParamChangeRoomStatus{}
}

//NewRet ..
func (m *MethodChangeRoomStatus) NewRet() module.Ret {
	return &RetChangeRoomStatus{}
}

//Do ..
func (m *MethodChangeRoomStatus) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamChangeRoomStatus)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetChangeRoomStatus{
			Info: helper.ChangeRoomStatusRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.AcID)
	if player == nil {
		logs.Info("[TeamBoss] Player: %v not exit in room: %v", info.AcID, info.RoomID)
		ret = RetChangeRoomStatus{
			Info: helper.ChangeRoomStatusRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}
	if info.AcID != roomInfo.LeadAcID {
		logs.Info("[TeamBoss] Player: %v permit limit for room: %v", info.AcID, info.RoomID)
		ret = RetChangeRoomStatus{
			Info: helper.ChangeRoomStatusRetInfo{
				Code: helper.RetCodeOptLimitPermission,
			},
		}
		return
	}
	roomInfo.RoomSetting = info.RoomStatus
	roomDetail := roomInfo.genRoomDetailInfo()
	acids := roomInfo.GetExtraPlayer(info.AcID)
	for k, v := range acids {
		bm.RoomInfo(k, roomDetail, v)
	}
	ret = RetChangeRoomStatus{
		Info: helper.ChangeRoomStatusRetInfo{
			RoomStatus: roomInfo.RoomSetting,
		},
	}
	return
}
