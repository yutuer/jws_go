package chatserver

import (
	// "fmt"
	"sync"
	// "time"

	gm "github.com/rcrowley/go-metrics"
	// "vcs.taiyouxi.net/platform/planx/funny/link"
	// "vcs.taiyouxi.net/platform/planx/metrics"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	city_mgr             *CityMgr
	sumCloseByAdjustRoom gm.Counter
)

type CityMgr struct {
	city2Rooms map[string]*Rooms
	mutex      sync.RWMutex
}

func init() {
	city_mgr = &CityMgr{
		city2Rooms: make(map[string]*Rooms, 128),
	}
	go city_mgr.statistic()
}

func enterRoomByCity(shard, city string, acid string, player *Player) {
	city_mgr._enterRoom(GetGidSidMerge(shard), city, acid, player)
}

func (pCity *CityMgr) _enterRoom(shard, city string, acid string, player *Player) {
	city_key := shard + "_" + city
	pCity.mutex.RLock()
	rooms, ok := pCity.city2Rooms[city_key]
	if ok {
		pCity.mutex.RUnlock()
		rooms.findRoom(acid, player)
		return
	} else {
		pCity.mutex.RUnlock()
		pCity.mutex.Lock()
		var pRooms *Rooms
		if _pRooms, ok := pCity.city2Rooms[city_key]; !ok {
			pRooms = newRooms(shard, city_key)
			pCity.city2Rooms[city_key] = pRooms
		} else {
			pRooms = _pRooms
		}
		pCity.mutex.Unlock()
		pRooms.findRoom(acid, player)
		return
	}
}

func (pCity *CityMgr) broadCast(typ, shard, msg string, acids []string) {
	rooms := []*Rooms{}
	pCity.mutex.RLock()
	for _, v := range pCity.city2Rooms {
		rooms = append(rooms, v)
	}
	pCity.mutex.RUnlock()
	for _, room := range rooms {
		if room.shard == shard {
			room.broadcast(typ, msg, acids)
		}
	}
}

func (pCity *CityMgr) GetRoomAndPlayerNum() (roomNum, priorRoomNum, playerNum int64) {
	rooms := []*Rooms{}
	pCity.mutex.RLock()
	for _, v := range pCity.city2Rooms {
		rooms = append(rooms, v)
	}
	pCity.mutex.RUnlock()
	for _, room := range rooms {
		roomNum += room.GetRoomNum()
		priorRoomNum += room.GetPriorRoomNum()
		playerNum += room.getRoomPlayerNum()
	}
	return
}

func (pCity *CityMgr) statistic() {
	//TODO YZH
	defer func() {
		if err := recover(); err != nil {
			logs.Error("[citymgr Statistic] recover error %v", err)
		}
	}()
	NewGauge := func(name string) gm.Gauge {
		return metrics.NewGauge(fmt.Sprintf("%s.%s.%s", "chattown", metrics.GetIPToken(), name))
	}
	//NewCounter := func(name string) gm.Counter {
	//    return metrics.NewCounter(fmt.Sprintf("%s.%s.%s", "chattown", metrics.GetIPToken(), name))
	//}
	roomNum := NewGauge("RoomNum")
	priorRoomNum := NewGauge("PriorRoomNum")
	playerNum := NewGauge("PlayerNum")
	//sumCloseByAdjustRoom = NewCounter("SumCloseByAdjustRoom")
	//fullBuffCloseSessionCounter := NewCounter("FullBuffCloseNum")
	//sumRecCounter := NewCounter("SumRecCounter")
	//sumSendCounter := NewCounter("SumSendCounter")
	//sumWantSendCounter := NewCounter("SumWantSendCounter")
	//sumConnCounter := NewCounter("SumConnCounter")
	//link.RegisterCounter(&mono_counter{mc: fullBuffCloseSessionCounter},
	//    &mono_counter{mc: sumRecCounter},
	//    &mono_counter{mc: sumSendCounter},
	//    &mono_counter{mc: sumWantSendCounter},
	//    &mono_counter{mc: sumConnCounter})

	for {
		select {
		case <-timingWheel.After(time.Second * 5):
			rn, prn, pn := pCity.GetRoomAndPlayerNum()
			roomNum.Update(rn)
			priorRoomNum.Update(prn)
			playerNum.Update(pn)
		}
	}
}

type mono_counter struct {
	mc gm.Counter
}

func (c *mono_counter) Inc() {
	c.mc.Inc(1)
}
