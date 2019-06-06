package game

import (
	"net"

	"github.com/ugorji/go/codec"

	"strconv"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type TCPServer struct {
	*servers.ConnServer
	quit             chan struct{}
	prepareHandshake PreparePlayer
}

func NewTCPServer(listento string) *TCPServer {
	scfg := servers.NewConnServerCfg{
		ListenTo:         listento,
		NumberOfAcceptor: 1,
	}
	gamesrv := &TCPServer{
		ConnServer: servers.NewConnServer(scfg),
		quit:       make(chan struct{}),
	}
	return gamesrv
}

func (c *TCPServer) handleConnection(con net.Conn) {
	defer logs.PanicCatcher()
	defer c.ReleaseConnChan(con)
	defer con.Close()
	//defer func() {
	//if v := recover(); v != nil {
	//logs.Error("[GameTCPServer] handleConnection recover error %v", v)
	//trace := make([]byte, 1024)
	//count := runtime.Stack(trace, true)
	//logs.Error("[GameTCPServer] Stack of %d bytes: %s\n", count, trace)
	//}
	//}()

	agent := client.NewPacketConnAgent(con.RemoteAddr().String(), con)
	go agent.Start(c.quit)
	logs.Info("[GameTCPServer] handleConnection %s", con.RemoteAddr())
	agentReadingChan := agent.GetReadingChan()

	var accountid, ip, gziplimit string
	var gzipLimit uint64
	select {
	case <-c.quit:
		break
	case pkt := <-agentReadingChan:
		if pkt.GetContentType() != client.PacketIDGateSession {
			logs.Error("[GameCHANSerever] should get PacketIDGateSession firstly.")
			break
		}
		var mh codec.MsgpackHandle
		mh.RawToString = true
		//accountid = "0:0:1001"
		dec := codec.NewDecoderBytes(pkt.GetBytes(), &mh)
		dec.Decode(&accountid)
		dec.Decode(&ip)
		dec.Decode(&gziplimit)
		gl, err2 := strconv.ParseUint(gziplimit, 10, 64)
		if err2 != nil {
			logs.Debug("playerProcessor allow gzip failed with %s", gziplimit)
			gzipLimit = 0
		} else {
			gzipLimit = gl
			logs.Debug("playerProcessor allow gzip with %d", gl)
		}
		logs.Trace("[GameTCPServer] handleConnection with account %s", accountid)

		playerRes := c.prepareHandshake(accountid, ip)
		defer playerRes.OnExit()
		playerProcessor(c.quit, playerRes, agentReadingChan, agent.SendPacket, gzipLimit)
	}

	logs.Info("[GameTCPServer] handleConnection %s exit", con.RemoteAddr())
}

func (c *TCPServer) loop() {
	conChan := c.GetWaitingConnChan()
	for {
		select {
		case <-c.quit:
			logs.Info("[GameTCPServer] dispatcher, server quit")
			return
		case conn, ok := <-conChan:
			if ok {
				myc := conn
				go c.handleConnection(myc)
			}
		}
	}
}

func (c *TCPServer) Stop() {
	close(c.quit)
}

func (c *TCPServer) Start(pp PreparePlayer) {
	if pp != nil {
		c.prepareHandshake = pp
	} else {
		panic("[GameTCPServer] Should start with PreparePlayer function")
	}
	go c.ConnServer.Start()
	var waitGroup util.WaitGroupWrapper
	waitGroup.Wrap(func() { c.loop() })
	waitGroup.Wait()
	c.ConnServer.Stop()
}
