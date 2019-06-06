package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamReadyFight struct {
	Sid  uint
	Acid string
	Code int
	Info helper.ReadyFightInfo
}

//RetAttack ..
type RetReadyFight struct {
	Info helper.ReadyFightRetInfo
}

//MethodReadyFight ..
type MethodReadyFight struct {
	module.BaseMethod
}

func newMethodReadyFight(m module.Module) *MethodReadyFight {
	return &MethodReadyFight{
		module.BaseMethod{Method: MethodReadyFightID, Module: m},
	}
}

//NewParam ..
func (m *MethodReadyFight) NewParam() module.Param {
	return &ParamReadyFight{}
}

//NewRet ..
func (m *MethodReadyFight) NewRet() module.Ret {
	return &RetReadyFight{}
}

//Do ..
func (m *MethodReadyFight) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamReadyFight)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		ret = RetReadyFight{
			Info: helper.ReadyFightRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.AcID)
	if player == nil {
		ret = RetReadyFight{
			Info: helper.ReadyFightRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}
	if player.SimpleInfo.AcID == roomInfo.LeadAcID {
		ret = RetReadyFight{
			Info: helper.ReadyFightRetInfo{
				Code: helper.RetCodeReadyFailed,
			},
		}
		return
	}
	player.SimpleInfo.Status = info.Status
	roomDetail := roomInfo.genRoomDetailInfo()
	acids := roomInfo.GetExtraPlayer(info.AcID)
	for k, v := range acids {
		bm.RoomInfo(k, roomDetail, v)
	}
	ret = RetReadyFight{
		Info: helper.ReadyFightRetInfo{
			Status: player.SimpleInfo.Status,
		},
	}

	return
}
