package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamChangeAvatar struct {
	Sid  uint
	Acid string
	Code int
	Info helper.ChangeAvatarInfo
}

//RetAttack ..
type RetChangeAvatar struct {
	Info helper.ChangeAvatarRetInfo
}

//MethodChangeAvatar ..
type MethodChangeAvatar struct {
	module.BaseMethod
}

func newMethodChangeAvatar(m module.Module) *MethodChangeAvatar {
	return &MethodChangeAvatar{
		module.BaseMethod{Method: MethodChangeAvatarID, Module: m},
	}
}

//NewParam ..
func (m *MethodChangeAvatar) NewParam() module.Param {
	return &ParamChangeAvatar{}
}

//NewRet ..
func (m *MethodChangeAvatar) NewRet() module.Ret {
	return &RetChangeAvatar{}
}

//Do ..
func (m *MethodChangeAvatar) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamChangeAvatar)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetChangeAvatar{
			Info: helper.ChangeAvatarRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	player := roomInfo.GetPlayer(info.AcID)
	if player == nil {
		logs.Info("[TeamBoss] Player: %v not exit in room: %v", info.AcID, info.RoomID)
		ret = RetChangeAvatar{
			Info: helper.ChangeAvatarRetInfo{
				Code: helper.RetCodePlayerNotInRoom,
			},
		}
		return
	}
	if info.Position >= len(roomInfo.PositionAcID) || info.Position < 0 {
		logs.Info("[TeamBoss] Player: %v position: %v error in room: %v", info.AcID, info.Position, info.RoomID)
		ret = RetChangeAvatar{
			Info: helper.ChangeAvatarRetInfo{
				Code: helper.RetCodeOptInvalid,
			},
		}
		return
	}
	if roomInfo.PositionAcID[info.Position] != "" && roomInfo.PositionAcID[info.Position] != info.AcID {
		logs.Info("[TeamBoss] Player: %v position: %v error in room: %v", info.AcID, info.Position, info.RoomID)
		ret = RetChangeAvatar{
			Info: helper.ChangeAvatarRetInfo{
				Code: helper.RetCodePositionOccupied,
			},
		}
		return
	}
	// change pos
	for i, item := range roomInfo.PositionAcID {
		if item == info.AcID {
			roomInfo.PositionAcID[i] = ""
		}
	}
	if info.BattleAvatar != -1 {
		roomInfo.PositionAcID[info.Position] = info.AcID
	}
	player.SimpleInfo.BattleAvatar = info.BattleAvatar
	player.SimpleInfo.Wing = info.Wing
	player.SimpleInfo.Fashion = info.Fashion
	player.SimpleInfo.MagicPet = info.MagicPet
	player.SimpleInfo.ExclusiveWeapon = info.ExclusiveWeapon
	player.SimpleInfo.Level = info.Level
	player.SimpleInfo.StarLevel = info.StarLevel
	player.BattleData = info.BattleInfo
	player.SimpleInfo.CompressGS = info.CompressGs

	roomDetail := roomInfo.genRoomDetailInfo()
	acids := roomInfo.GetExtraPlayer(info.AcID)
	for k, v := range acids {
		bm.RoomInfo(k, roomDetail, v)
	}
	ret = RetChangeAvatar{
		Info: helper.ChangeAvatarRetInfo{},
	}

	return
}
