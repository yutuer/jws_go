package fenghuo

import (
	"runtime/debug"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/timingwheel"

	"errors"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuomsg"
	"vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/platform/planx/funny/linkext"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const fenghuoMaxPlayer = gamedata.FegnhuoRoomMaxPlayer
const fenghuoMaxSublevles = gamedata.FenghuoStageMaxNum
const fenghuoMaxEnemies = 32

type acid2player struct {
	IDX    int
	Player *FenghuoPlayer
}

type FenghuoGame struct {
	wg     sync.WaitGroup
	GameID string

	GameState     int
	LastGameState int

	//Max is fenghuoMaxPlayer
	PlayerOnline [fenghuoMaxPlayer]bool
	PlayerHPs    [fenghuoMaxPlayer]int
	Players      [fenghuoMaxPlayer]FenghuoPlayer
	//Slice of enemyHps
	EnemyHps   []int
	NumEnemies int
	enemyHps   [fenghuoMaxEnemies]int
	//这个值[1,8]范围有效
	SubLevel          int
	SubLevelStatus    [fenghuoMaxSublevles][fenghuoMaxPlayer]int
	SubLevelDoneCount int

	//isNeedPushGameState bool
	//isHasOnOver         bool
	//lastPushTime        int64

	//创建过程创建的列表，表示哪些玩家可以加入这个游戏
	AcIDs   [fenghuoMaxPlayer]string
	Avatars [fenghuoMaxPlayer]*helper.Avatar2ClientByJson

	//cmdChannel chan GVEGameCommandMsg
	quitChan chan bool
	tWheel   *timingwheel.TimingWheel

	// 广播用结构 用锁保护
	channelMutex sync.RWMutex
	Channel      *linkext.Channel
	acID2Players map[string]acid2player
}

func NewFenghuoGame(gameID string) *FenghuoGame {
	g := &FenghuoGame{
		GameID:        gameID,
		Channel:       linkext.NewChannel(),
		acID2Players:  make(map[string]acid2player, fenghuoMaxPlayer),
		GameState:     fenghuomsg.FenghuoGameStatusWaitingInit,
		LastGameState: fenghuomsg.FenghuoGameStatusWaitingInit,
	}
	for i := range g.PlayerHPs {
		g.PlayerHPs[i] = 9999
	}
	return g
}

func (g *FenghuoGame) Start() error {
	g.wg.Add(1)
	defer g.wg.Done()

	g.quitChan = make(chan bool, 1)
	g.tWheel = util.GetQuickTimeWheel()
	//g.cmdChannel = make(chan GVEGameCommandMsg, 64)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("GVEGame Panic, session Err %v", err)
				debug.PrintStack()
			}
		}()
		timerChan := g.tWheel.After(time.Second)
		for {
			select {
			//case command, _ := <-g.cmdChannel:
			//	g.wg.Add(1)
			//	func() {
			//		defer g.wg.Done()
			//		logs.Trace("command %v", command)
			//		g.processMsg(&command)
			//	}()
			case <-timerChan:
				timerChan = g.tWheel.After(time.Second)
				g.wg.Add(1)
				func() {
					defer g.wg.Done()
					g.checkLoss()
					g.loop()
				}()
			case <-g.quitChan:
				//g.PushGameState()
				//g.onExitGame()
				return
			}
		}
	}()

	return nil
}

func (g *FenghuoGame) AllSubLevelDone() bool {
	return g.GameState >= fenghuomsg.FenghuoGameStatusGameOver
}

func (g *FenghuoGame) ForceOver() {
	g.GameState = fenghuomsg.FenghuoGameStatusForceOver
}

func (g *FenghuoGame) GameOver() {
	g.channelMutex.Lock()
	defer g.channelMutex.Unlock()
	logs.Info("FenghuoGame Destroy Room ID %s", g.GameID)
	for i := range g.Players {
		g.Players[i].GetSession().Close()
		acID := g.Players[i].GetAcID()
		g.Channel.Exit(acID)
		delete(g.acID2Players, acID)
	}
	g.Stop()
}

func (g *FenghuoGame) Stop() {
	g.quitChan <- true
	g.wg.Wait()
	g.Close()
}

func (r *FenghuoGame) EnterPlayer(player *FenghuoPlayer) (int, error) {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	playeridx := 0
	isInAcIDs := false
	for i, a := range r.AcIDs {
		if player.GetAcID() == a {
			isInAcIDs = true
			playeridx = i
			break
		}
	}
	if !isInAcIDs {
		//如果玩家ID不在创建游戏时声明的列表中则报错
		return 0, errors.New("player No In AcIDs")
	}

	r.Channel.Join(player.GetSession(), player.GetAcID())
	r.acID2Players[player.GetAcID()] = acid2player{IDX: playeridx, Player: player}
	r.Players[playeridx] = *player
	r.PlayerOnline[playeridx] = true
	return playeridx, nil
}

func (r *FenghuoGame) IsPlayerOnline(idx int) bool {
	if idx < 0 && idx >= fenghuoMaxPlayer {
		return false
	}
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()

	return r.PlayerOnline[idx]
}

//func (r *FenghuoGame) LeavePlayer(player FenghuoPlayer) {
//	r.channelMutex.Lock()
//	defer r.channelMutex.Unlock()
//	r.Channel.Exit(player.GetAcID())
//	player.GetSession().Close()
//	delete(r.acID2Players, player.GetAcID())
//}

func (r *FenghuoGame) broadcastMsg(data []byte) {
	r.Channel.Broadcast(data)
}

func (r *FenghuoGame) broadcastMsg2Other(player FenghuoPlayer, data []byte) {
	r.Channel.BroadcastOthers(data, player.GetSession())
}

func (r *FenghuoGame) Close() {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	r.Channel.Close()
}
