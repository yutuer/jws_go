package gatemode

import (
	"fmt"
	"net"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TCPGameServer TODO 需要处理断线重连
type TCPGameServer struct {
	*client.PacketConnAgent

	//quit chan struct{}
}

func (t *TCPGameServer) GetGoneChan() <-chan struct{} {
	return t.PacketConnAgent.GetGoneChan()
}

func newTCPGameServer(name, addr string) *TCPGameServer {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logs.Error("Game Server is not reachable! %s", err.Error())
	}
	serverName := fmt.Sprintf("serverof:%s", name)
	agent := client.NewPacketConnAgent(serverName, conn)

	return &TCPGameServer{
		PacketConnAgent: agent,
		//quit:            make(chan struct{}),
	}
}

type TCPGameServerManager struct {
	wg   util.WaitGroupWrapper
	quit chan struct{}
}

func NewTCPGameServerManager() *TCPGameServerManager {
	return &TCPGameServerManager{
		quit: make(chan struct{}),
	}
}

func (t *TCPGameServerManager) NewGameServer(name, addr string) gate.GameServer {
	//TODO 检查是否是断线重连，而且后agent还没关闭
	t.wg.Add(1)
	s := newTCPGameServer(name, addr)
	go s.Start(t.quit)
	return s
}

func (t *TCPGameServerManager) RecycleGameServer(gs gate.GameServer) {
	//TODO 暂时回收当前链接，等待玩家是否会断线重连。并尝试数据回写数据库
	// 需要配合后端Game Server算法，玩家断线后Game Server实现数据落地
	logs.Trace("TCP RecycleGameServer conn.")
	t.wg.Done()
	gs.Stop()
	return
}

func (t *TCPGameServerManager) WaitAllShutdown(quit <-chan struct{}) {
	<-quit
	close(t.quit)
	t.wg.Wait()
	return
}
