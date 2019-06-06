package linkext

import (
	"net"
	"sync"

	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type serverAcceptHandle func(s *link.Session, err error) error

type ServerExt struct {
	*link.Server

	acceptHandle serverAcceptHandle

	stopWait sync.WaitGroup
}

func NewServerExt(listener net.Listener, codecType link.CodecType) *ServerExt {
	server := &ServerExt{
		Server: link.NewServer(listener, codecType),
	}
	return server
}

func (server *ServerExt) AcceptHandle(handle serverAcceptHandle) {
	server.acceptHandle = handle
}

func (server *ServerExt) StartServer() {
	server.stopWait.Add(1)
	go func() {
		defer server.stopWait.Done()
		for {
			session, err := server.Accept()
			if server.acceptHandle != nil {
				err = server.acceptHandle(session, err)
				if err == closeErr {
					logs.Info("Receive EOF, close")
					return
				}
				if err != nil {
					logs.Error("Accept Err by %s", err.Error())
					return
				}
			}

		}
	}()
}

func (server *ServerExt) Wait() {
	server.stopWait.Wait()
}
