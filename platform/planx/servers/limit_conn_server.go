package servers

import (
	"net"
	"os"
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LimitConnServer struct {
	ConnServer

	spaces  chan struct{}
	maxconn uint
}

type NewLimitConnServerCfg struct {
	NewConnServerCfg
	MaxConn uint
}

func NewLimitConnServer(cfg NewLimitConnServerCfg) *LimitConnServer {
	if cfg.MaxConn <= 0 {
		cfg.MaxConn = 3
	}

	s := &LimitConnServer{
		ConnServer: *NewConnServer(cfg.NewConnServerCfg),

		maxconn: cfg.MaxConn,
		spaces:  make(chan struct{}, cfg.MaxConn),
	}
	for i := uint(0); i < cfg.MaxConn; i++ {
		s.spaces <- struct{}{}
	}
	return s
}

func (server *LimitConnServer) Start() {
	listener, err := newXListener(server.listento, server.SslCfg, server.SslCaCfg)
	if err != nil {
		logs.Critical(err.Error())
		<-time.After(time.Second * 1)
		os.Exit(1)
	} else {
		server.listener = listener
	}

	var waitGroup util.WaitGroupWrapper
	acceptors := Number_Of_Acceptor
	if acceptors < 1 {
		acceptors = 1
	}
	for i := 0; i < acceptors; i++ {
		waitGroup.Wrap(func() { server.acceptor() })
	}
	//waitGroup.Wrap(func() { server.closer() })
	waitGroup.Wait()
}

//func (server *LimitConnServer) closer() {
//logs.Info("LimitConnServer closer started.")
//for conn := range server.quitingConn {
//conn.Close()
//server.spaces <- true
//}
//logs.Trace("closer done")

//}

func (server *LimitConnServer) acceptor() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("[LimitConnServer] acceptor recover error %v", err)
		}
	}()
	logs.Info("LimitConnServer acceptor started.")
	timeoutDuration := time.Millisecond * 5000
	timer := time.NewTimer(timeoutDuration)
	for {
		//server.listener.SetDeadline()
		conn, err := server.listener.Accept()

		//PS: learn from Serve func in net/http in go source code
		var tempDelay time.Duration // how long to sleep on accept failure
		if err != nil {
			if util.NetIsClosed(err) {
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logs.Warn("Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			tempDelay = 0
			logs.Error(err.Error())
			logs.Info("Accepter quit")
			return
		}

		for len(timer.C) > 0 {
			<-timer.C
		}
		timer.Reset(timeoutDuration)
		select {
		case _, ok := <-server.spaces:
			if ok {
				logs.Info("limit conn server accept %s", conn.RemoteAddr())
				server.waitingConn <- conn
			}
		case <-timer.C:
			logs.Info("ERR Out of service %s", conn.RemoteAddr())
			conn.Close()
		}
	}
}

func (server *LimitConnServer) GetMaxConn() uint {
	return server.maxconn
}

//Only useful for LimitConn Server. OPTMIZE
func (server *LimitConnServer) ReleaseConnChan(con net.Conn) {
	//server.quitingConn <- con
	server.spaces <- struct{}{}
}
