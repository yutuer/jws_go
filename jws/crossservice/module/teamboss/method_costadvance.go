package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamCostAdvance struct {
	Sid  uint
	Acid string
	Code int
	Info helper.CostAdvanceInfo
}

//RetAttack ..
type RetCostAdvance struct {
	Info helper.CostAdvanceRetInfo
}

//MethodCostAdvance ..
type MethodCostAdvance struct {
	module.BaseMethod
}

func newMethodCostAdvance(m module.Module) *MethodCostAdvance {
	return &MethodCostAdvance{
		module.BaseMethod{Method: MethodCostAdvanceID, Module: m},
	}
}

//NewParam ..
func (m *MethodCostAdvance) NewParam() module.Param {
	return &ParamCostAdvance{}
}

//NewRet ..
func (m *MethodCostAdvance) NewRet() module.Ret {
	return &RetCostAdvance{}
}

//Do ..
func (m *MethodCostAdvance) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamCostAdvance)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetCostAdvance{
			Info: helper.CostAdvanceRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.AcID)
	if player == nil {
		logs.Info("[TeamBoss] Player: %v not exit in room: %v", info.AcID, info.RoomID)
		ret = RetCostAdvance{
			Info: helper.CostAdvanceRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}
	if roomInfo.AdvanceCostID != "" && roomInfo.AdvanceCostID != info.AcID {
		logs.Info("[TeamBoss] Player: %v no permission for room: %v, status: %v", info.AcID, info.RoomID)
		ret = RetCostAdvance{
			Info: helper.CostAdvanceRetInfo{
				Code: helper.RetCodeAlreadyTickRedBox,
			},
		}
		return
	}
	roomInfo.BoxStatus = info.BoxStatus
	if info.BoxStatus == 1 {
		roomInfo.AdvanceCostID = info.AcID
	} else {
		roomInfo.AdvanceCostID = ""
	}
	roomDetail := roomInfo.genRoomDetailInfo()
	acids := roomInfo.GetExtraPlayer(info.AcID)
	for k, v := range acids {
		bm.RoomInfo(k, roomDetail, v)
	}
	ret = RetCostAdvance{
		Info: helper.CostAdvanceRetInfo{},
	}
	return
}
