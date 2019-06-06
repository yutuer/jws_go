package allinone

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ChanGameServer is a reversed ChanAgent of game server, as gate agent simulation
type ChanGameServer struct {
	*game.ChanAgent
}

// GetReadingChan of ChanAgent, caller should be a gate ONLY.
func (c *ChanGameServer) GetReadingChan() <-chan *client.Packet {
	return c.Out
}

// SendPacket of ChanAgent, caller should be a gate ONLY.
func (c *ChanGameServer) SendPacket(pkt *client.Packet) bool {
	if c.IsClosed() {
		return false
	}
	select {
	case c.In <- pkt:
	case <-c.GetGoneChan():
		return false
	case <-game.GetSecondsAfterTimer(3 * time.Second):
		logs.Critical("ChanGameServer SendPacket timeout")
		return false
	}
	return true
}

//func newChanGameServer(name string) *ChanGameServer {
//agent := game.NewChanAgent(name)
//return &ChanGameServer{agent}
//}

type ChanGameServerManager struct {
	chanServer *game.ChanServer
	wg         util.WaitGroupWrapper
	quit       chan struct{}
}

func NewChanGameServerManager() *ChanGameServerManager {
	cs := game.NewChanServer()
	go func() {
		mux := logics.CreatePlayer
		cs.Start(mux)
	}()

	return &ChanGameServerManager{
		chanServer: cs,
		quit:       make(chan struct{}),
	}
}

func (t *ChanGameServerManager) NewGameServer(name, addr string) gate.GameServer {
	//TODO 检查是否是断线重连，而且后agent还没关闭
	t.wg.Add(1)

	s := &ChanGameServer{
		t.chanServer.Accept(name),
	}

	logs.Trace("New ChanGameServer conn work.")
	return s
}

func (t *ChanGameServerManager) RecycleGameServer(gs gate.GameServer) {
	//TODO 暂时回收当前链接，等待玩家是否会断线重连。并尝试数据回写数据库
	// 需要配合后端Game Server算法，玩家断线后Game Server实现数据落地
	logs.Trace("Chan RecycleGameServer.")
	t.wg.Done()
	gs.Stop()
	return
}

func (t *ChanGameServerManager) WaitAllShutdown(quit <-chan struct{}) {
	<-quit
	t.chanServer.Stop()
	close(t.quit)
	t.wg.Wait()
	return
}
