package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamCreateRoom struct {
	Sid  uint
	Acid string
	Info helper.CreateRoomInfo
}

//RetAttack ..
type RetCreateRoom struct {
	Info helper.JoinRoomRetInfo
}

//MethodCreateRoom ..
type MethodCreateRoom struct {
	module.BaseMethod
}

func newMethodCreateRoom(m module.Module) *MethodCreateRoom {
	return &MethodCreateRoom{
		module.BaseMethod{Method: MethodCreateRoomID, Module: m},
	}
}

//NewParam ..
func (m *MethodCreateRoom) NewParam() module.Param {
	return &ParamCreateRoom{}
}

//NewRet ..
func (m *MethodCreateRoom) NewRet() module.Ret {
	return &RetCreateRoom{}
}

//Do ..
func (m *MethodCreateRoom) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamCreateRoom)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	roomInfo := bm.Room.CreateRoom(info.RoomLevel)
	roomInfo.RoomLevel = info.RoomLevel
	roomInfo.GenRoomInfo()

	joinInfo := info.JoinInfo
	roomInfo.LeadAcID = joinInfo.AcID
	if roomInfo != nil {
		roomInfo.AddPlayer(&Player{
			DetailData: joinInfo.PlayerDetailInfo,
			SimpleInfo: helper.PlayerSimpleInfo{
				AcID:         joinInfo.AcID,
				Sid:          joinInfo.Sid,
				GS:           joinInfo.GS,
				Avatar:       joinInfo.Avatar,
				Name:         joinInfo.Name,
				Status:       helper.TBPlayerStateIdle,
				VIP:          joinInfo.VIP,
				BattleAvatar: -1,
			},
		})
	}
	logs.Info("[TeamBoss] Create Room: %v success", *roomInfo)
	roomDetail := roomInfo.genRoomDetailInfo()
	ret = RetCreateRoom{
		Info: helper.JoinRoomRetInfo{
			Code: helper.RetCodeSuccess,
			Info: roomDetail,
		},
	}
	errCode = message.ErrCodeOK
	return
}
