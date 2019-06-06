package teamboss

import (
	"fmt"

	"sync"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// id:  level_index_time

type RoomInfo struct {
	Rooms map[uint32][]*Room
	mutex sync.RWMutex
}

func genRoomID(level uint32, index int) string {
	return fmt.Sprintf("%d_%d_%d", level, index, time.Now().Unix())
}

func (ri *RoomInfo) GetRooms(roomLevel uint32) []*Room {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	return append([]*Room{}, ri.Rooms[roomLevel][:]...)
}

func (ri *RoomInfo) GetRoom(roomID string) *Room {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	level, index, err := helper.ParseRoomID(roomID)
	if err != nil {
		logs.Error("parse room id err by %v", err)
		return nil
	}
	rooms := ri.Rooms[level]
	if index < len(rooms) {
		room := rooms[index]
		if room != nil && room.ID == roomID {
			return room
		}
	}

	return nil
}

func (ri *RoomInfo) CreateRoom(level uint32) *Room {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()
	rooms := ri.Rooms[level]
	exist := false
	var room *Room
	for i, r := range rooms {
		if r == nil {
			room = &Room{
				ID: genRoomID(level, i),
			}
			rooms[i] = room
			exist = true
			break
		}
	}
	if !exist {
		room = &Room{
			ID: genRoomID(level, len(rooms)),
		}
		rooms = append(rooms, room)
	}
	ri.Rooms[level] = rooms
	return room
}

func (ri *RoomInfo) DelRoom(roomID string) {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()
	level, index, err := helper.ParseRoomID(roomID)
	if err != nil {
		logs.Error("parse room id err by %v", err)
		return
	}
	rooms := ri.Rooms[level]
	if index < len(rooms) {
		room := rooms[index]
		if room != nil && room.ID == roomID {
			ri.Rooms[level][index] = nil
		}
	}
	return
}

type Room struct {
	ID            string
	RoomLevel     uint32
	BossID        string
	TeamTypID     uint32
	SceneID       string
	LeadAcID      string
	Players       []*Player
	RoomSetting   int // 房间设置，公开还是仅限邀请
	RoomState     int // 内部状态表明房间战斗状态
	BoxStatus     int
	AdvanceCostID string
	LostPlayer    string
	PositionAcID  [helper.RoomPlayerMaxCount]string
}

func (r *Room) GenRoomInfo() {
	r.TeamTypID = gamedata.GetTBTeamTypeID()
	r.BossID = gamedata.GetTBBossID(r.TeamTypID)
	r.SceneID = gamedata.GetTBSceneID(r.TeamTypID)
}

func (r *Room) PlayerCount() int {
	return len(r.Players)
}

func (r *Room) AddPlayer(p *Player) {
	r.Players = append(r.Players, p)
}

func (r *Room) RemovePlayer(acid string) {
	index := -1
	for i, item := range r.Players {
		if acid == item.SimpleInfo.AcID {
			index = i
		}
	}
	if index != -1 {
		r.Players = append(r.Players[:index], r.Players[index+1:]...)
		if r.AdvanceCostID == acid {
			r.AdvanceCostID = ""
			r.BoxStatus = 0
		}
		if r.LeadAcID == acid && len(r.Players) > 0 {
			r.LeadAcID = r.Players[0].SimpleInfo.AcID
			r.Players[0].SimpleInfo.Status = helper.TBPlayerStateIdle
		}
		for i, item := range r.PositionAcID {
			if item == acid {
				r.PositionAcID[i] = ""
			}
		}
	}
}

func (r *Room) GetPlayer(acid string) *Player {
	for _, item := range r.Players {
		if item.SimpleInfo.AcID == acid {
			return item
		}
	}
	return nil
}

func (r *Room) GetExtraPlayer(acid string) map[uint][]string {
	ret := make(map[uint][]string, 2)
	for _, item := range r.Players {
		if item.SimpleInfo.AcID != acid {
			ret[item.SimpleInfo.Sid] = append(ret[item.SimpleInfo.Sid], item.SimpleInfo.AcID)
		}
	}
	return ret
}

func (r *Room) genRoomDetailInfo() helper.RoomDetailInfo {
	psi := make([]helper.PlayerSimpleInfo, len(r.Players))
	for i, item := range r.Players {
		psi[i] = item.SimpleInfo
	}
	return helper.RoomDetailInfo{
		RoomID:       r.ID,
		LeadID:       r.LeadAcID,
		SimpleInfo:   psi,
		RoomStatus:   r.RoomSetting,
		BoxStatus:    r.BoxStatus,
		BossID:       r.BossID,
		TeamTypID:    r.TeamTypID,
		SceneID:      r.SceneID,
		Level:        r.RoomLevel,
		PositionAcID: r.PositionAcID,
	}
}
func (r *Room) IsFull() bool {
	return len(r.Players) >= helper.RoomPlayerMaxCount
}

type Player struct {
	BattleData []byte
	DetailData []byte
	SimpleInfo helper.PlayerSimpleInfo
}
