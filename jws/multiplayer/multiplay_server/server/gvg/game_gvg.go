package gvg

import (
	"runtime/debug"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/timingwheel"

	"errors"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gvg_proto"
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

type GVGGame struct {
	wg                  sync.WaitGroup
	GID                 uint
	SID                 int
	GameID              string
	Level               uint32
	SceneID             string
	BossID              string
	GroupID             uint32
	BoxStatus           int
	CostID              string
	Stat                gvg_proto.GVGGameState
	Datas               gvg_proto.GVGGameDatas
	isNeedPushGameState bool
	isHasOnOver         bool
	lastPushTime        int64

	//创建过程创建的列表，表示哪些玩家可以加入这个游戏
	AcIDs []string

	cmdChannel chan GVGGameCommandMsg
	quitChan   chan bool
	tWheel     *timingwheel.TimingWheel

	// 广播用结构 用锁保护
	channelMutex sync.RWMutex
	Channel      *linkext.Channel
	acID2Players map[string]GamePlayer
	lead         string
}

func NewGVGGame(gameID string, sid uint) *GVGGame {
	return &GVGGame{
		GameID:       gameID,
		Channel:      linkext.NewChannel(),
		acID2Players: make(map[string]GamePlayer, 8),
		SID:          int(sid),
	}
}

func (g *GVGGame) SetLeadExcept(acid string) {
	for i := 0; i < len(g.Stat.Player); i++ {
		if g.Stat.Player[i].AcID != acid && !g.Stat.Player[i].IsExit() {
			g.lead = g.Stat.Player[i].AcID
		}
	}
}

func (g *GVGGame) Start(data *helper.GVGStartFightData) error {
	g.loadAccountDatas(data)
	g.loadGameDatas(data)

	logs.Trace("g.Stat %v", g.Stat)
	logs.Trace("g.Data %v", g.Datas)

	g.quitChan = make(chan bool, 1)
	g.tWheel = mulutil.GetQuickTimeWheel()
	g.cmdChannel = make(chan GVGGameCommandMsg, 64)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("GVGGame Panic, session Err %v", err)
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
				if !ok {
					logs.Error("cmd channel closed")
				}
				g.wg.Add(1)
				func() {
					defer g.wg.Done()
					g.processMsg(&command)
				}()
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

func (g *GVGGame) Stop() {
	go func() {
		g.quitChan <- true
		g.wg.Wait()
		//for _, p := range g.acID2Players {
		//	p.GetSession().Close()
		//}
		g.Close()
		logs.Info("GVGGame stop %s", g.GameID)
	}()
}

func (r *GVGGame) EnterPlayer(player GamePlayer) error {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	isInAcIDs := false
	for _, a := range r.AcIDs {
		if player.GetAcID() == a {
			isInAcIDs = true
		}
	}
	if !isInAcIDs {
		logs.Error("GVGGame.EnterPlayer player No In AcIDs %s %s",
			player.GetAcID(), r.GameID)
		//如果玩家ID不在创建游戏时声明的列表中则报错
		return errors.New("player No In AcIDs")
	}
	r.Channel.Join(player.GetSession(), player.GetAcID())
	r.acID2Players[player.GetAcID()] = player
	if r.lead == "" {
		r.lead = player.GetAcID()
	}
	return nil
}

//func (r *GVGGame) LeavePlayer(player GamePlayer) {
//	r.channelMutex.Lock()
//	defer r.channelMutex.Unlock()
//	r.Channel.Exit(player.GetAcID())
//	delete(r.acID2Players, player.GetAcID())
//}

func (r *GVGGame) broadcastMsg(data []byte) {
	r.Channel.Broadcast(data)
}

func (r *GVGGame) broadcastMsg2Other(player GamePlayer, data []byte) {
	r.Channel.BroadcastOthers(data, player.GetSession())
}

func (r *GVGGame) Close() {
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()
	r.Channel.Close()
}

func (r *GVGGame) PushCommand(msg *GVGGameCommandMsg) {
	if r.cmdChannel == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case r.cmdChannel <- *msg:
	case <-ctx.Done():
		logs.Error("GVGGame cmdChan is full")
	}
}

func (r *GVGGame) PushCommandWithRsp(msg *GVGGameCommandMsg) *GVGGameCommandResMsg {
	if r.cmdChannel == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case r.cmdChannel <- *msg:
	case <-ctx.Done():
		logs.Error("GVGGame cmdChan is full")
	}

	select {
	case ret := <-msg.ResChann:
		logs.Debug("GVGGameCommandExec success")
		return &ret
	case <-ctx.Done():
		logs.Error("GVGGame CommandExec apply <-retChan timeout")
		return nil
	}
}
