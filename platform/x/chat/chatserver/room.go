package chatserver

import (
	"sync"
	"sync/atomic"

	// "vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/funny/linkext"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Room struct {
	sync.WaitGroup
	Name        string
	Channel     *linkext.Channel
	sid2Players map[string]*Player
	mutex       sync.RWMutex
	RoomsMgr    *Rooms
	playerNum   int32 // for statistic
	isFull      bool  // 记录曾经满的状态，用来从hashring中移除后，再加回来
}

func NewRoom(name string, rooms *Rooms) *Room {
	return &Room{
		Name:        name,
		Channel:     linkext.NewChannel(),
		sid2Players: make(map[string]*Player, CommonCfg.RoomMaxNum),
		RoomsMgr:    rooms,
	}
}

func (r *Room) OnPlayerUnReg(player *Player) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.playerUnReg(player)
	// 若此时房间人数之前达上限，此时第一次小于上限，则重新加回hashring
	if r.isFull {
		r.RoomsMgr.mutex.Lock()
		defer r.RoomsMgr.mutex.Unlock()

		_, ok := r.RoomsMgr.rooms[r.Name]
		if ok {
			r.RoomsMgr.hash = r.RoomsMgr.hash.AddNode(r.Name)
		}
		r.isFull = false
	}
}

// must in lock
func (r *Room) playerReg(player *Player) int {
	accountId := string(player.roleInfo.AccountId())
	oldPlayer, ok := r.sid2Players[accountId]
	if ok {
		delete(r.sid2Players, accountId)
		go func() {
			oldPlayer.session.Close()
		}()
		logs.Warn("playerReg again kick old, acid %s name %s", player.roleInfo.AccountId(), player.roleInfo.Name())
	}

	//确保调用外层有锁保护，下面3句作为一个原子操作
	r.sid2Players[accountId] = player
	player.room = r
	r.WaitGroup.Add(1)

	r.Channel.Join(player.session, accountId)
	r.chgPlayerNum(1)
	logs.Info("city %v room [enter]  %v num:%v %s %s",
		r.RoomsMgr.cityKey, r.Name, len(r.sid2Players),
		accountId, player.Name)
	return len(r.sid2Players)
}

// must in lock
func (r *Room) playerUnReg(player *Player) int {
	accountId := string(player.roleInfo.AccountId())
	if _, ok := r.sid2Players[accountId]; ok {
		r.Channel.Exit(accountId)

		//确保调用外层有锁保护，下面3句作为一个原子操作
		delete(r.sid2Players, accountId)
		player.room = nil
		r.WaitGroup.Done()
		r.chgPlayerNum(-1)
		logs.Info("city %v room [leave]  %v num:%v %s %s",
			r.RoomsMgr.cityKey, r.Name, len(r.sid2Players),
			accountId, player.Name)
	}
	return len(r.sid2Players)
}

func (r *Room) broadcast2OtherMsg(player *Player, msg []byte) {
	if r == nil {
		return
	}
	r.Channel.BroadcastOthers(msg, player.session)
}

func (r *Room) broadcastMsg(typ, msg string, acids []string) {
	if acids == nil || len(acids) <= 0 {
		r.Channel.Broadcast(Gen_SysNotice(typ, msg))
	} else {
		r.mutex.RLock()
		msids := make(map[uint64]struct{}, 16)
		for _, acid := range acids {
			if p, ok := r.sid2Players[acid]; ok {
				msids[p.session.Id()] = struct{}{}
			}
		}
		r.mutex.RUnlock()
		r.Channel.BroadcastSomeOthers(Gen_SysNotice(typ, msg), msids)
	}
	logs.Warn("room %v broadcast %s, %s, %v", r.Name, typ, msg, acids)
}

func (r *Room) Exit(player *Player) {
	r.Channel.Exit(player.AccountID())
}

func (r *Room) Close() {
	r.Channel.Close()
}

func (r *Room) getPlayerNum() int {
	return int(atomic.LoadInt32(&r.playerNum))
}

func (r *Room) chgPlayerNum(chg int32) {
	atomic.AddInt32(&r.playerNum, chg)
}
