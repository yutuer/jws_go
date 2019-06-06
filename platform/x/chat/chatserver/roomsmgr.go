package chatserver

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/serialx/hashring"
)

type Rooms struct {
	shard        string
	cityKey      string
	priorRooms   *list.List
	rooms        map[string]*Room
	hash         *hashring.HashRing
	mutex        sync.RWMutex
	adjustRoomCh chan string
	roomNum      int64 // atomic 房间总数，包括优先级房间
	priorRoomNum int64 // atomic
	tStatistic   <-chan time.Time

	roomIdSeed uint64 // 生成房间名的seed
}

func newRooms(shard, city_key string) *Rooms {
	roms := &Rooms{
		shard:        shard,
		cityKey:      city_key,
		priorRooms:   list.New(),
		rooms:        make(map[string]*Room, 256),
		adjustRoomCh: make(chan string, 1024),
		tStatistic:   time.NewTicker(5 * time.Second).C,
	}
	roomNames := make([]string, CommonCfg.CityInitRoomNum)
	for i := 0; i < CommonCfg.CityInitRoomNum; i++ {
		roomName, pNewRoom := roms._genRoom()
		roms.rooms[roomName] = pNewRoom
		roms.chgRoomNum(1)
		roomNames[i] = roomName
	}
	logs.Info("city %v  init room %v", city_key, roomNames)
	roms.hash = hashring.New(roomNames)
	//	go roms._adjustRoomsRoutine()
	return roms
}

func (rooms *Rooms) findRoom(acid string, player *Player) {
	// lock roomsmgr
	rooms.mutex.Lock()
	defer rooms.mutex.Unlock()
	// 优先返回未满CommonCfg.RoomAdjustNum人房间
	var fPriorRoom *list.Element
	var pRoom *Room
	for pRoom == nil {
		if rooms.priorRooms.Len() > 0 {
			fPriorRoom = rooms.priorRooms.Front()
			pRoom = fPriorRoom.Value.(*Room)
		} else { // hash room
			pRoom = rooms._hashRoom(acid)
		}
		// 找不到房间，目前只在清理完人数过少的房间后可能发生
		if pRoom == nil {
			roomName, pNewRoom := rooms._genRoom()
			rooms._addRoom(roomName, pNewRoom)
		}
	}

	// lock room
	pRoom.mutex.Lock()
	defer pRoom.mutex.Unlock()
	roomNum := pRoom.playerReg(player)
	// 房间人数大于CommonCfg.RoomAdjustNum，则将此房间移出优先级房间队列
	if fPriorRoom != nil && len(pRoom.sid2Players) >= CommonCfg.RoomAdjustNum {
		rooms.priorRooms.Remove(fPriorRoom)
		rooms.chgPriorRoomNum(-1)
		rooms.rooms[pRoom.Name] = pRoom
		rooms.hash = rooms.hash.AddNode(pRoom.Name)
		logs.Info("move room %v from prior to rooms ", pRoom.Name)
	}
	//	rooms.onRoomPlayerChg(pRoom.Name, roomNum)
	// 若此时房间人数第一次达上限，则从hashring中删除，并创建一个新的空房间放到优先级队列里
	if roomNum >= CommonCfg.RoomMaxNum && !pRoom.isFull {
		logs.Info("[NewRoom] room:%s up to maxnum[%d]", pRoom.Name, roomNum)
		rooms.hash = rooms.hash.RemoveNode(pRoom.Name)
		pRoom.isFull = true
	}
}

//func (rs *Rooms) onRoomPlayerChg(roomName string, roomNum int) {
//	if !(roomNum > CommonCfg.RoomMaxNum ||
//		(roomNum < CommonCfg.RoomMinNum && rs.GetRoomNum() > int64(CommonCfg.CityInitRoomNum))) {
//		return
//	}
//	rs.adjustRoomCh <- roomName
//}

func (rooms *Rooms) _adjustRoomsRoutine() {
	for {
		select {
		case roomName := <-rooms.adjustRoomCh:
			rooms._adjust(roomName)
		}
	}
}

func (rooms *Rooms) _adjust(roomName string) {
	// lock roomsmgr
	rooms.mutex.Lock()
	defer rooms.mutex.Unlock()
	pRoom, ok := rooms.rooms[roomName]
	if !ok {
		logs.Error("RoomsMgr not find room: %v", roomName)
		return
	}
	cfgInitRomNum := CommonCfg.CityInitRoomNum
	cfgRomMaxNum := CommonCfg.RoomMaxNum
	cfgRomMinNum := CommonCfg.RoomMinNum
	// lock room
	pRoom.mutex.Lock()
	defer pRoom.mutex.Unlock()
	roomPlayerNum := len(pRoom.sid2Players)
	if !(roomPlayerNum > cfgRomMaxNum || (roomPlayerNum < cfgRomMinNum && len(rooms.rooms) > cfgInitRomNum)) {
		return
	}
	logs.Info("begin adjust city:%v room:%v playernum:%v", rooms.cityKey, roomName, roomPlayerNum)
	if roomPlayerNum >= CommonCfg.RoomMaxNum {
		// new Room
		roomName, pNewRoom := rooms._genRoom()
		rooms._addRoom(roomName, pNewRoom)
		// find some player in this room and stop them
		movePlayer := _findPlayInRoom(pRoom, CommonCfg.RoomAdjustNum)
		for _, player := range movePlayer {
			player.Stop()
		}
	} else if roomPlayerNum <= cfgRomMinNum && len(rooms.rooms) > cfgInitRomNum {
		// find all player in this room
		movePlayer := _findPlayInRoom(pRoom, roomPlayerNum)
		// del room
		//解除分房间机制，孤立pRoom指针，使其等待退出
		rooms._delRoom(pRoom)
		go func(pR *Room) {
			// stop them
			for _, player := range movePlayer {
				player.Stop()
			}
			pR.WaitGroup.Wait()
			pR.Close()
		}(pRoom)
	}
}

// must in lock if not init
func (rooms *Rooms) _genRoom() (string, *Room) {
	roomName := rooms._genRoomName(rooms.cityKey)
	pNewRoom := NewRoom(roomName, rooms)
	return roomName, pNewRoom
}

// must in lock
func (rooms *Rooms) _hashRoom(acid string) *Room {
	roomName, _ := rooms.hash.GetNode(acid)
	return rooms.rooms[roomName]
}

// must in lock
func (rooms *Rooms) _addRoom(roomName string, pNewRoom *Room) {
	rooms.priorRooms.PushBack(pNewRoom)
	rooms.chgRoomNum(1)
	rooms.chgPriorRoomNum(1)
	logs.Info("add room %v to prior", roomName)
}

// must in lock
func (rooms *Rooms) _delRoom(pRoom *Room) {
	rooms.hash = rooms.hash.RemoveNode(pRoom.Name)
	if _, ok := rooms.rooms[pRoom.Name]; ok {
		delete(rooms.rooms, pRoom.Name)
	} else {
		for it := rooms.priorRooms.Front(); it != nil; it = it.Next() {
			pRoom := it.Value.(*Room)
			if pRoom.Name == pRoom.Name {

				rooms.priorRooms.Remove(it)
				rooms.chgPriorRoomNum(-1)
				break
			}
		}
	}
	rooms.chgRoomNum(-1)
	logs.Info("delete room %v", pRoom.Name)
}

func (rooms *Rooms) broadcast(typ string, msg string, acids []string) {
	room := []*Room{}
	rooms.mutex.RLock()
	for _, v := range rooms.rooms {
		room = append(room, v)
	}
	for it := rooms.priorRooms.Front(); it != nil; it = it.Next() {
		room = append(room, it.Value.(*Room))
	}
	rooms.mutex.RUnlock()
	for _, r := range room {
		switch typ {
		case Typ_SysNotice, Typ_FishInfo, Typ_Gve, Typ_FengHuo:
			r.broadcastMsg(typ, msg, nil)
		case Typ_GuildRoom:
			r.broadcastMsg(typ, msg, acids)

		case Typ_FilterRoom:
			// 清理人数过少的房间
			rooms.mutex.Lock()
			r.mutex.Lock()
			roomPlayerNum := r.getPlayerNum()
			logs.Debug("[TryDelRoom] room:%s num:%d ", r.Name, roomPlayerNum)
			// 保证有可用的房间（还有空位的房间），避免清理过后只剩满的房间了
			if roomPlayerNum <= CommonCfg.RoomMinNum && rooms.GetRoomNum() > int64(CommonCfg.CityInitRoomNum) {
				logs.Info("[DelRoom] room:%s num:%d reassign", r.Name, roomPlayerNum)
				movePlayer := _findPlayInRoom(r, roomPlayerNum)
				rooms._delRoom(r)
				go func(pR *Room) {
					// stop them
					for _, player := range movePlayer {
						player.Stop()
					}
					pR.WaitGroup.Wait()
					pR.Close()
				}(r)
			}
			r.mutex.Unlock()
			rooms.mutex.Unlock()
		default:
			logs.Error("rooms broadcast typ[%s] not define", typ)
		}

	}
}

func (rooms *Rooms) getRoomPlayerNum() int64 {
	room := []*Room{}
	rooms.mutex.RLock()
	for _, v := range rooms.rooms {
		room = append(room, v)
	}
	for it := rooms.priorRooms.Front(); it != nil; it = it.Next() {
		room = append(room, it.Value.(*Room))
	}
	rooms.mutex.RUnlock()
	var res int64
	for _, r := range room {
		res += int64(r.getPlayerNum())
	}
	return res
}
func (rooms *Rooms) _genRoomName(city_key string) string {
	return fmt.Sprintf("%s_r%d", city_key, atomic.AddUint64(&rooms.roomIdSeed, 1))
}

// must in room lock
func _findPlayInRoom(r *Room, playerNum int) map[string]*Player {
	if playerNum <= 0 {
		return nil
	}
	ret := make(map[string]*Player, playerNum)
	// temp start for change player
	//	for _, player := range r.sid2Players {
	//		if strings.Index(player.Name, "robot") < 0 {
	//			ret[player.Name] = player
	//			playerNum--
	//			logs.Warn("change room %v", player.Name)
	//			break
	//		}
	//	}
	// temp end
	if sumCloseByAdjustRoom != nil {
		sumCloseByAdjustRoom.Inc(int64(playerNum))
	}
	for _, player := range r.sid2Players {
		if playerNum <= 0 {
			break
		}
		ret[player.Name] = player
		playerNum--
	}
	return ret
}

func (rooms *Rooms) chgRoomNum(chg int64) {
	atomic.AddInt64(&rooms.roomNum, chg)
}

func (rooms *Rooms) GetRoomNum() int64 {
	return atomic.LoadInt64(&rooms.roomNum)
}

func (rooms *Rooms) chgPriorRoomNum(chg int64) {
	atomic.AddInt64(&rooms.priorRoomNum, chg)
}

func (rooms *Rooms) GetPriorRoomNum() int64 {
	return atomic.LoadInt64(&rooms.priorRoomNum)
}
