package connect

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/crossservice/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Server ..
type Server struct {
	listener   *net.TCPListener
	serverAddr *net.TCPAddr

	filter *IPFilter

	requestQueue chan *request

	cacheConn     map[*origConn]bool
	cacheConnLock sync.Mutex

	wg       sync.WaitGroup
	isClosed bool
}

//NewServer ..
func NewServer(nt string, ip string, port int) (*Server, error) {
	server := &Server{
		requestQueue: make(chan *request, DefaultServerRequestQueueLength),

		isClosed: false,
	}

	serverAddr, err := net.ResolveTCPAddr(nt, fmt.Sprintf("%s:%d", ip, port))
	if nil != err {
		return nil, err
	}

	listener, err := net.ListenTCP(nt, serverAddr)
	if nil != err {
		return nil, err
	}
	server.serverAddr = listener.Addr().(*net.TCPAddr)
	server.listener = listener

	server.cacheConn = make(map[*origConn]bool)

	return server, nil
}

//LocalAddr ..
func (server *Server) LocalAddr() *net.TCPAddr {
	return server.serverAddr
}

//SetIPFilter ..
func (server *Server) SetIPFilter(f *IPFilter) {
	server.filter = f
}

//Run ..
func (server *Server) Run() error {
	if nil == server.listener {
		return ErrInvalid
	}
	return server.run()
}

//Recv ..
func (server *Server) Recv() (*Message, *Conn, error) {
	req, ok := <-server.requestQueue
	if false == ok {
		return nil, nil, ErrClosed
	}
	if nil != req.err {
		return nil, nil, req.err
	}

	msg := NewMessage(req.pac.data, req.pac.length)
	c := newConn(req.oc)
	return msg, c, nil
}

//Close ..
func (server *Server) Close() bool {
	if server.isClosed {
		return false
	}

	server.isClosed = true

	server.listener.Close()

	server.cacheConnLock.Lock()
	for c := range server.cacheConn {
		c.close()
	}
	server.cacheConnLock.Unlock()

	server.wg.Wait()

	close(server.requestQueue)
	return true
}

func (server *Server) run() error {
	server.wg.Add(1)
	defer server.wg.Done()

	retErr := error(nil)
	for !server.isClosed {
		tc, err := server.listener.AcceptTCP()
		if nil != err {
			if server.isClosed {
				retErr = ErrClosed
			} else {
				retErr = err
			}
			break
		}
		if nil != server.filter {
			remoteIP := tc.RemoteAddr().(*net.TCPAddr).IP.String()
			if false == server.filter.check(remoteIP) {
				logs.Warn("Connection from [%s] was ignore by ip filter", remoteIP)
				tc.Close()
				continue
			}
		}
		oc := newOrigConn(tc)
		go server.holdOrigConn(oc)
	}

	return retErr
}

func (server *Server) holdOrigConn(oc *origConn) {
	server.wg.Add(1)
	defer server.wg.Done()

	server.cacheConnLock.Lock()
	server.cacheConn[oc] = true
	cs_metrics.UpdateConn(int64(len(server.cacheConn)))
	server.cacheConnLock.Unlock()
	defer func() {
		server.cacheConnLock.Lock()
		delete(server.cacheConn, oc)
		cs_metrics.UpdateConn(int64(len(server.cacheConn)))
		server.cacheConnLock.Unlock()
	}()

	for !server.isClosed {
		pac, err := oc.readPacket()
		var req *request
		if server.isClosed {
			req = &request{pac, oc, ErrClosed}
		} else {
			req = &request{pac, oc, err}
		}

		select {
		case server.requestQueue <- req:
		case <-time.After(DefaultRequestInTimeout):
		}

		if io.EOF == err {
			break
		}
	}
}

type request struct {
	pac *packet
	oc  *origConn
	err error
}
