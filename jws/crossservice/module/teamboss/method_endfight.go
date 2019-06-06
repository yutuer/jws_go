package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamEndFight struct {
	Sid  uint
	Acid string
	Code int
	Info helper.EndFightInfo
}

//RetAttack ..
type RetEndFight struct {
	Info helper.EndFightRetInfo
}

//MethodEndFight ..
type MethodEndFight struct {
	module.BaseMethod
}

func newMethodEndFight(m module.Module) *MethodEndFight {
	return &MethodEndFight{
		module.BaseMethod{Method: MethodEndFightID, Module: m},
	}
}

//NewParam ..
func (m *MethodEndFight) NewParam() module.Param {
	return &ParamEndFight{}
}

//NewRet ..
func (m *MethodEndFight) NewRet() module.Ret {
	return &RetEndFight{}
}

//Do ..
func (m *MethodEndFight) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamEndFight)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK

	res := bm.RewardLog.receiveReward(info.GlobalRoomID, info.AcID)
	if res == nil {
		ret = RetEndFight{
			Info: helper.EndFightRetInfo{},
		}
	} else {
		ret = RetEndFight{
			Info: helper.EndFightRetInfo{
				HasRedBox: res.HasRedBox,
				HasReward: res.hasReward,
				Level:     res.Level,
				IsCost:    res.CostID == info.AcID,
			},
		}
	}

	roomInfo := bm.Room.GetRoom(helper.Global2RoomID(info.GlobalRoomID))
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.GlobalRoomID)
		return
	}
	//if roomInfo.LostPlayer != "" {
	//	logs.Info("[TeamBoss] Remove lost player: %v in room: %v", roomInfo.ID, roomInfo.LostPlayer)
	//	roomInfo.LostPlayer = ""
	//	roomInfo.RemovePlayer(roomInfo.LostPlayer)
	//}
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
	player.SimpleInfo.InBattle = false
	isOver := true
	for _, p := range roomInfo.Players {
		if p.SimpleInfo.InBattle {
			isOver = false
			break
		}
	}
	if isOver {
		roomInfo.RoomState = helper.TBRoomIdle
	}
	acids := roomInfo.GetExtraPlayer("")
	for k, v := range acids {
		bm.RoomInfo(k, roomInfo.genRoomDetailInfo(), v)
	}
	return
}
