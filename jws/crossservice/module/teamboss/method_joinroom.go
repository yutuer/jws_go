package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamJoinRoom struct {
	Sid  uint
	Acid string
	Info helper.JoinRoomInfo
}

//RetAttack ..
type RetJoinRoom struct {
	Info helper.JoinRoomRetInfo
}

//MethodJoinRoom ..
type MethodJoinRoom struct {
	module.BaseMethod
}

func newMethodJoinRoom(m module.Module) *MethodJoinRoom {
	return &MethodJoinRoom{
		module.BaseMethod{Method: MethodJoinRoomID, Module: m},
	}
}

//NewParam ..
func (m *MethodJoinRoom) NewParam() module.Param {
	return &ParamJoinRoom{}
}

//NewRet ..
func (m *MethodJoinRoom) NewRet() module.Ret {
	return &RetJoinRoom{}
}

//Do ..
func (m *MethodJoinRoom) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamJoinRoom)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetJoinRoom{
			Info: helper.JoinRoomRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	if roomInfo.RoomState == helper.TBRoomFight {
		logs.Info("[TeamBoss] Room: %v can't entry", info.RoomID)
		ret = RetJoinRoom{
			Info: helper.JoinRoomRetInfo{
				Code: helper.RetCodeRoomInBattle,
			},
		}
		return
	}
	if roomInfo.RoomSetting == helper.TBRoomStateLock && !info.JoinInfo.IsInvited {
		logs.Info("[TeamBoss] Room: %v can't entry", info.RoomID)
		ret = RetJoinRoom{
			Info: helper.JoinRoomRetInfo{
				Code: helper.RetCodeRoomCantEntry,
			},
		}
		return
	}
	joinInfo := info.JoinInfo
	if roomInfo.GetPlayer(joinInfo.AcID) == nil && !roomInfo.IsFull() {
		roomInfo.AddPlayer(&Player{
			DetailData: joinInfo.PlayerDetailInfo,
			SimpleInfo: helper.PlayerSimpleInfo{
				AcID:         joinInfo.AcID,
				Sid:          joinInfo.Sid,
				GS:           joinInfo.GS,
				Avatar:       joinInfo.Avatar,
				Name:         joinInfo.Name,
				VIP:          joinInfo.VIP,
				BattleAvatar: -1,
			},
		})

		roomDetail := roomInfo.genRoomDetailInfo()
		ret = RetJoinRoom{
			Info: helper.JoinRoomRetInfo{
				Info: roomDetail,
			},
		}
		acids := roomInfo.GetExtraPlayer(info.JoinInfo.AcID)
		for k, v := range acids {
			bm.RoomInfo(k, roomDetail, v)
		}
	} else {
		ret = RetJoinRoom{
			Info: helper.JoinRoomRetInfo{
				Code: helper.RetCodeRoomPlayerFull,
			},
		}
		logs.Info("[TeamBoss] Room: %v can't entry for acid: %v", info.RoomID, info.JoinInfo.AcID)
	}
	return
}
