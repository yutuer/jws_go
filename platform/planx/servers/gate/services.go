package gate

import (
	"vcs.taiyouxi.net/platform/planx/client"
)

type GameServerManager interface {
	NewGameServer(name, addr string) GameServer
	RecycleGameServer(gs GameServer)
	WaitAllShutdown(quit <-chan struct{})
}

type GameServer interface {
	GetReadingChan() <-chan *client.Packet
	SendPacket(pkt *client.Packet) bool
	GetGoneChan() <-chan struct{}
	Stop()
}
