package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamGetPlayerDetail struct {
	Sid  uint
	Acid string
	Info helper.GetPlayerDetailInfo
}

//RetAttack ..
type RetGetPlayerDetail struct {
	Info helper.GetPlayerDetailRetInfo
}

//MethodGetPlayerDetail ..
type MethodGetPlayerDetail struct {
	module.BaseMethod
}

func newMethodGetPlayerDetail(m module.Module) *MethodGetPlayerDetail {
	return &MethodGetPlayerDetail{
		module.BaseMethod{Method: MethodGetPlayerDetailID, Module: m},
	}
}

//NewParam ..
func (m *MethodGetPlayerDetail) NewParam() module.Param {
	return &ParamGetPlayerDetail{}
}

//NewRet ..
func (m *MethodGetPlayerDetail) NewRet() module.Ret {
	return &RetGetPlayerDetail{}
}

//Do ..
func (m *MethodGetPlayerDetail) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamGetPlayerDetail)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetGetPlayerDetail{
			Info: helper.GetPlayerDetailRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.AcID)
	if player == nil {
		logs.Info("[TeamBoss] Player: %v not exit in room: %v", info.AcID, info.RoomID)
		ret = RetGetPlayerDetail{
			Info: helper.GetPlayerDetailRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}

	tgtPlayer := roomInfo.GetPlayer(info.TgtID)
	if tgtPlayer == nil {
		logs.Info("[TeamBoss] Player: %v not exit in room: %v", info.TgtID, info.RoomID)
		ret = RetGetPlayerDetail{
			Info: helper.GetPlayerDetailRetInfo{
				Code: helper.RetCodeOptInvalid,
			},
		}
		return
	}
	ret = RetGetPlayerDetail{
		Info: helper.GetPlayerDetailRetInfo{
			Detail: tgtPlayer.DetailData,
		},
	}
	return
}
