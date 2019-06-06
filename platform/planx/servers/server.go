package servers

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	//"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Number_Of_Acceptor     = 1
	Number_Of_WaitingQueue = 64
)

type ConServer interface {
	Stop()
	Start()

	GetWaitingConnChan() <-chan net.Conn
	//TODO: zehong, Only useful for LimitConn Server. OPTMIZE
	ReleaseConnChan(net.Conn)
}

type SSLCertCfg struct {
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

func (s SSLCertCfg) IsSet() bool {
	if s.Cert != "" {
		return true
	}
	return false
}

type ConnServer struct {
	quit chan struct{}

	waitingConn chan net.Conn
	//quitingConn chan net.Conn

	listento string
	SslCfg   SSLCertCfg
	SslCaCfg SSLCertCfg
	listener net.Listener
}

type NewConnServerCfg struct {
	ListenTo             string
	NumberOfAcceptor     uint
	NumberOfWaitingQueue uint
	SslCfg               SSLCertCfg
	SslCaCfg             SSLCertCfg
}

func NewConnServer(cfg NewConnServerCfg) *ConnServer {
	if cfg.NumberOfAcceptor <= 0 {
		cfg.NumberOfAcceptor = Number_Of_Acceptor
	}
	if cfg.NumberOfWaitingQueue <= 0 {
		cfg.NumberOfWaitingQueue = Number_Of_WaitingQueue
	}
	return &ConnServer{
		quit: make(chan struct{}),

		waitingConn: make(chan net.Conn, cfg.NumberOfWaitingQueue),
		//quitingConn: make(chan net.Conn),
		SslCfg:   cfg.SslCfg,
		SslCaCfg: cfg.SslCaCfg,
		listento: cfg.ListenTo,
	}
}

// A listener implements a network listener (net.Listener) for TLS connections.
type xlistener struct {
	net.Listener
	config *tls.Config
}

// Accept waits for and returns the next incoming TLS connection.
// The returned connection c is a *tls.Conn.
func (l *xlistener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err != nil {
		return
	}

	if tcpconn, ok := c.(*net.TCPConn); ok {
		tcpconn.SetKeepAlive(true)
		tcpconn.SetKeepAlivePeriod(time.Second * 60)
		tcpconn.SetLinger(-1)
	}

	if l.config != nil {
		c = tls.Server(c, l.config)
	}
	return
}

func newXListener(listento string, sslCfg, sslCaCfg SSLCertCfg) (net.Listener, error) {
	logs.Info("Start Listen %s", listento)
	listener, err := net.Listen("tcp", listento)
	if err != nil {
		return nil, err
	}

	xl := &xlistener{Listener: listener}

	if sslCfg.IsSet() {
		cer, err := tls.LoadX509KeyPair(sslCfg.Cert, sslCfg.Key)
		if err != nil {
			logs.Error("getListener LoadX509KeyPair err:%s", err.Error())
			return nil, err
		}
		logs.Info("Listener build with sslcfg %v", sslCfg)
		config := &tls.Config{
			Certificates:       []tls.Certificate{cer},
			InsecureSkipVerify: true,
		}
		if sslCaCfg.IsSet() {
			certPEMBlock, err := ioutil.ReadFile(sslCaCfg.Cert)
			if err != nil {
				return nil, err
			}
			logs.Info("Listener build with sslCaCfg %v", sslCaCfg)
			capool := x509.NewCertPool()
			capool.AppendCertsFromPEM(certPEMBlock)
			//config.RootCAs = capool
			config.ClientCAs = capool
			config.ClientAuth = tls.RequireAndVerifyClientCert
		}
		logs.Info("Start Listen %s with ssl", listento)
		xl.config = config
		return xl, nil
	}

	return xl, nil
}

func (server *ConnServer) Start() {
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

func (server *ConnServer) Stop() {
	server.listener.Close()
	//close(server.quitingConn)
	close(server.waitingConn)
	close(server.quit)
}

//func (server *ConnServer) closer() {
//logs.Info("ConnServer closer started.")
//for conn := range server.quitingConn {
//conn.Close()
//}
//logs.Trace("closer done")
//}

func (server *ConnServer) acceptor() {
	logs.Info("ConnServer acceptor start %s", server.listento)
	timeoutDuration := time.Millisecond * 3000
	timer := time.NewTimer(timeoutDuration)
	for {
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
		// XXX 可能会有一点小问题，因为listener是提前close的，所以这里不应该出现写
		case server.waitingConn <- conn:
		case <-timer.C:
			logs.Info("ERR Out of service %s", conn.RemoteAddr())
			conn.Close()
		}

	}
}

func (server *ConnServer) GetWaitingConnChan() <-chan net.Conn {
	return server.waitingConn
}

//Only useful for LimitConn Server. OPTMIZE
func (server *ConnServer) ReleaseConnChan(con net.Conn) {
	//server.quitingConn <- con
}
