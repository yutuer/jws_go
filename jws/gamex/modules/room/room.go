package room

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/room/info"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	cmdTypeGet = iota + 1
	cmdTypeAttachObserver
	cmdTypeDetachObserver
	cmdTypeUpdateRoom
	cmdTypeLeaveRoom
	cmdTypeChangeMaster
	cmdTypeReady
	cmdTypeCancelReady
	cmdTypeStartFight
	cmdTypeEnterRoom
	cmdTypeMakeReward
	cmdTypeGetRoomInfo
)

type FenghuoProfile struct {
	Name     string
	AcID     string
	AvatarID int
	CorpLv   uint32
	Gs       int
}

type cmd struct {
	Type    int
	Account string

	SimpleInfo FenghuoProfile
	SubLevels  *gamedata.FenghuoLevelData
	Reward     *gamedata.PriceDatas
	Avatar     *helper.Avatar2ClientByJson
	//StageData  *gamedata.FenghuoStageData

	ResChan chan *cmd

	Code       int
	ErrCode    errorcode.ErrorCode
	P1         int
	P2         int
	PB         bool
	PS         string
	PlayerChan chan<- servers.Request
	Rooms      []info.Room
	Rnd        *rand.Rand
}

type cmdSyncRoom struct {
	NewRoom []byte
	DelRoom int
}

type playerChanRegInfo struct {
	AccountID string
	IsDel     bool
	Chan      chan<- servers.Request
}

type module struct {
	sid            uint
	wg             sync.WaitGroup
	cmdChan        chan *cmd
	stopChan       chan bool
	BalanceChan    chan bool
	notifyStopChan chan bool

	roomNumArray []int
	roomNumMap   map[int]*info.Room

	roomNumAllocCurrMax int
	roomNums            []int

	newRooms [][]byte
	delRooms []int

	playerSyncRoomCmdChan chan cmdSyncRoom
	playerChanRegChan     chan playerChanRegInfo
	playerChanMap         map[string]chan<- servers.Request
}

func New(sid uint) *module {
	m := new(module)
	m.sid = sid
	return m
}

func (r *module) AfterStart(g *gin.Engine) {
}

func (r *module) BeforeStop() {
}

func (r *module) Start() {
	r.wg.Add(1)

	r.roomNumMap = make(map[int]*info.Room, 256)
	r.stopChan = make(chan bool, 1)
	r.notifyStopChan = make(chan bool, 1)
	r.cmdChan = make(chan *cmd, 256)
	r.BalanceChan = make(chan bool, 64)
	r.playerChanMap = make(map[string]chan<- servers.Request, 2048)
	r.playerChanRegChan = make(chan playerChanRegInfo, 64)
	r.playerSyncRoomCmdChan = make(chan cmdSyncRoom, 64)

	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-r.stopChan:
				logs.Trace("stopChan room")
				return
			case command, ok := <-r.cmdChan:
				if ok {
					res := r.processCmd(command)
					if command.ResChan != nil {
						command.ResChan <- res
					}
				}
			}
		}
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		timeChan := time.After(time.Second)
		for {
			select {
			case newPlayer := <-r.playerChanRegChan:
				if newPlayer.IsDel {
					delete(r.playerChanMap, newPlayer.AccountID)
				} else {
					r.playerChanMap[newPlayer.AccountID] = newPlayer.Chan
				}
			case roomSync := <-r.playerSyncRoomCmdChan:
				logs.Trace("r %v", roomSync)
				if roomSync.NewRoom != nil && len(roomSync.NewRoom) > 0 {
					r.newRooms = append(r.newRooms, roomSync.NewRoom)
				}
				if roomSync.DelRoom > 0 {
					r.delRooms = append(r.delRooms, roomSync.DelRoom)
				}
			case <-r.notifyStopChan:
				logs.Trace("notifyStopChan room")
				return
			case <-timeChan:
				timeChan = time.After(time.Second)
				r.notify()
			}
		}
	}()
}

func (r *module) Stop() {
	r.stopChan <- true
	r.notifyStopChan <- true
	logs.Flush()
	r.wg.Wait()
}

func (r *module) sendCmd(ctx context.Context, c *cmd) *cmd {
	c.ResChan = make(chan *cmd, 1)
	r.cmdChan <- c
	var t *cmd
	select {
	case t = <-c.ResChan:
	case <-ctx.Done():
		return nil
	}
	return t
}

func (r *module) sendCmdWithoutRes(ctx context.Context, c *cmd) {
	r.cmdChan <- c
}
