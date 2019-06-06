package teamboss

import (
	"sort"

	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamRoomList struct {
	Sid  uint
	Acid string
	Code int
	Info helper.RoomListInfo
}

//RetAttack ..
type RetRoomList struct {
	Info helper.RoomListRetInfo
}

//MethodRoomList ..
type MethodRoomList struct {
	module.BaseMethod
}

func newMethodRoomList(m module.Module) *MethodRoomList {
	return &MethodRoomList{
		module.BaseMethod{Method: MethodRoomListID, Module: m},
	}
}

//NewParam ..
func (m *MethodRoomList) NewParam() module.Param {
	return &ParamRoomList{}
}

//NewRet ..
func (m *MethodRoomList) NewRet() module.Ret {
	return &RetRoomList{}
}

//Do ..
func (m *MethodRoomList) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamRoomList)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	errCode = message.ErrCodeOK
	rooms := bm.Room.GetRooms(info.RoomLevel)
	roomList := make([]helper.RetRoomList, 0, len(rooms))
	for _, item := range rooms {
		if item == nil || item.RoomState == helper.TBRoomFight {
			continue
		}
		avatars := make([]helper.RoomListAvatar, len(item.Players))
		for j, jtem := range item.Players {
			avatars[j] = helper.RoomListAvatar{
				Avatar: jtem.SimpleInfo.BattleAvatar,
				StarLv: jtem.SimpleInfo.StarLevel,
			}
		}
		leadName := ""
		server := uint(0)
		for _, ktem := range item.Players {
			if ktem.SimpleInfo.AcID == item.LeadAcID {
				leadName = ktem.SimpleInfo.Name
				server = ktem.SimpleInfo.Sid
			}
		}
		if leadName == "" {
			logs.Warn("[TeamBoss] No lead in this room: %v", *item)
			continue
		}
		roomList = append(roomList, helper.RetRoomList{
			LeadName:     leadName,
			PlayerCount:  len(avatars),
			RoomStatus:   item.RoomSetting,
			Level:        info.RoomLevel,
			BattleAvatar: avatars,
			RoomID:       item.ID,
			ServerID:     server,
		})
	}

	roomList = sortRoomList(roomList)
	if len(roomList) > gamedata.BoxCfg.RoomMax {
		roomList = roomList[:gamedata.BoxCfg.RoomMax]
	}
	logs.Debug("[TeamBoss] Room list: %v", roomList)
	ret = RetRoomList{
		Info: helper.RoomListRetInfo{
			List: roomList,
		},
	}
	return
}
func sortRoomList(waitForSortRoom []helper.RetRoomList) []helper.RetRoomList {
	sort.Sort(roomList(waitForSortRoom))
	return waitForSortRoom
}

type roomList []helper.RetRoomList

func (I roomList) Len() int {
	return len(I)
}
func (I roomList) Less(i, j int) bool {
	inum, jnum := I[i].RoomStatus, I[j].RoomStatus
	if I[i].PlayerCount >= helper.RoomPlayerMaxCount {
		inum = helper.TBRoomStateFull
	}
	if I[j].PlayerCount >= helper.RoomPlayerMaxCount {
		jnum = helper.TBRoomStateFull
	}
	return inum < jnum
}
func (I roomList) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}
