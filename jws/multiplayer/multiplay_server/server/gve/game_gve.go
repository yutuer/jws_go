package gve

import (
	"runtime/debug"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/timingwheel"

	"errors"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	mulutil "vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/funny/linkext"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GamePlayer interface {
	GetAcID() string
	GetSession() *link.Session
}

type GVEGame struct {
	wg                  sync.WaitGroup
	GameID              string
	Stat                gve_proto.GVEGameState
	Datas               gve_proto.GVEGameDatas
	isNeedPushGameState bool
	isHasOnOver         bool
	lastPushTime        int64

	//创建过程创建的列表，表示哪些玩家可以加入这个游戏
	AcIDs []string

	cmdChannel chan GVEGameCommandMsg
	quitChan   chan bool
	tWheel     *timingwheel.TimingWheel

	// 广播用结构 用锁保护
	channelMutex sync.RWMutex
	Channel      *linkext.Channel
	acID2Players map[string]GamePlayer
}

func NewGVEGame(gameID string) *GVEGame {
	return &GVEGame{
		GameID:       gameID,
		Channel:      linkext.NewChannel(),
		acID2Players: make(map[string]GamePlayer, 8),
	}
}

func (g *GVEGame) Start() error {
	err := g.loadAccountDatas()
	if err != nil {
		return err
	}
	g.loadGameDatas()

	logs.Trace("g.Stat %v", g.Stat)
	logs.Trace("g.Data %v", g.Datas)

	g.quitChan = make(chan bool, 1)
	g.tWheel = mulutil.GetQuickTimeWheel()
	g.cmdChannel = make(chan GVEGameCommandMsg, 64)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("GVEGame Panic, session Err %v", err)
				debug.PrintStack()
			}
		}()
		g.wg.Add(1)
		defer g.wg.Done()
		timerChan := g.tWheel.After(time.Second)
		for {
			//nowT := time.Now().Unix()
			select {
			case command, ok := <-g.cmdChannel:
				if ok {
					g.wg.Add(1)
					func() {
						defer g.wg.Done()
						logs.Trace("command %v", command)
						g.processMsg(&command)
					}()
				}
			case <-timerChan:
				timerChan = g.tWheel.After(time.Second)
				g.wg.Add(1)
				func() {
					defer g.wg.Done()
					g.checkLoss()
					g.loop()
				}()
			case <-g.quitChan:
				g.PushGameState()
				g.onExitGame()
				return
			}
		}
	}()

	return nil
}

func (g *GVEGame) Stop() {
	go func() {
		g.quitChan <- true
		g.wg.Wait()
		g.Close()
		logs.Info("GVEGame stop %s", g.GameID)
	}()
}

func (r *GVEGame) EnterPlayer(player GamePlayer) error {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	isInAcIDs := false
	for _, a := range r.AcIDs {
		if player.GetAcID() == a {
			isInAcIDs = true
		}
	}
	if !isInAcIDs {
		logs.Error("GVEGame.EnterPlayer player No In AcIDs %s %s",
			player.GetAcID(), r.GameID)
		//如果玩家ID不在创建游戏时声明的列表中则报错
		return errors.New("player No In AcIDs")
	}
	r.Channel.Join(player.GetSession(), player.GetAcID())
	r.acID2Players[player.GetAcID()] = player
	return nil
}

//func (r *GVEGame) LeavePlayer(player GamePlayer) {
//	r.channelMutex.Lock()
//	defer r.channelMutex.Unlock()
//	r.Channel.Exit(player.GetAcID())
//	delete(r.acID2Players, player.GetAcID())
//}

func (r *GVEGame) broadcastMsg(data []byte) {
	r.Channel.Broadcast(data)
}

func (r *GVEGame) broadcastMsg2Other(player GamePlayer, data []byte) {
	r.Channel.BroadcastOthers(data, player.GetSession())
}

func (r *GVEGame) Close() {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	r.Channel.Close()
}

func (r *GVEGame) PushCommand(msg *GVEGameCommandMsg) {
	if r.cmdChannel == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case r.cmdChannel <- *msg:
	case <-ctx.Done():
		logs.Error("TBGame cmdChan is full")
	}
}

func (r *GVEGame) PushCommandWithRsp(msg *GVEGameCommandMsg) *GVEGameCommandResMsg {
	if r.cmdChannel == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case r.cmdChannel <- *msg:
	case <-ctx.Done():
		logs.Error("TBGame cmdChan is full")
	}

	select {
	case ret := <-msg.ResChann:
		logs.Debug("TBGameCommandExec success")
		return &ret
	case <-ctx.Done():
		logs.Error("TBGame CommandExec apply <-retChan timeout")
		return nil
	}
}
