package rpc

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GateRPC struct {
	quit chan struct{}
	servers.ConServer
	Cfg Config
}

func NewGateRPCServer(c Config) *GateRPC {
	scfg := servers.NewConnServerCfg{
		ListenTo:             c.RPCListen,
		NumberOfAcceptor:     1,
		NumberOfWaitingQueue: 10000,
	}
	return &GateRPC{
		ConServer: servers.NewConnServer(scfg),
		Cfg:       c,
		quit:      make(chan struct{}),
	}
}

func (g *GateRPC) Start(receiver interface{}) {
	rpc.Register(receiver)

	go g.ConServer.Start()
	var waitGroup util.WaitGroupWrapper
	waitGroup.Wrap(func() {
		for {
			select {
			case <-g.quit:
				logs.Info("Gate.rpc.rpcloop, server quit")
				return
			case conn, ok := <-g.GetWaitingConnChan():
				if ok {
					go func() {
						jsonrpc.ServeConn(conn)
						g.ReleaseConnChan(conn)
					}()
				}
			}
		}
	})
	waitGroup.Wait()
	g.ConServer.Stop()
}

func (g *GateRPC) Stop() {
	close(g.quit)
}
