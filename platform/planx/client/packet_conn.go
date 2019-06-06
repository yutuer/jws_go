package client

import (
	"bufio"
	"io"
	"net"
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

type PacketConn struct {
	net.Conn
	ioread  *bufio.Reader
	iowrite *bufio.Writer
}

const defaultBufferSize = 256 * 1024
const defaultPingTime = 30 //seconds //应该小于secondsTimer的最大周期时间60*10s = 10分钟
var secondsTimer *timingwheel.TimingWheel

func init() {
	secondsTimer = timingwheel.NewTimingWheel(time.Second, 60*10)
}

func newPacketConn(connection net.Conn) PacketConn {
	conn := PacketConn{
		Conn:    connection,
		ioread:  bufio.NewReaderSize(connection, defaultBufferSize),
		iowrite: bufio.NewWriterSize(connection, defaultBufferSize),
	}
	return conn
}

func (p *PacketConn) setReadDeadline(t time.Time) error {
	return p.Conn.SetReadDeadline(t)
}

func (p *PacketConn) setWriteDeadline(t time.Time) error {
	return p.Conn.SetWriteDeadline(t)
}

func (p *PacketConn) readPacket(t time.Duration) (*Packet, error) {
	pkt, err := ReadPacket(p.ioread, func(id int) {
		rtimeout := time.Now().Add(t * time.Duration(id))
		p.setReadDeadline(rtimeout)
	})
	if err != nil {
		e, ok := err.(net.Error)
		if ok && e.Temporary() {
			if !e.Timeout() {
				logs.Warn("PacketConn.ReadPacket got Temorary error, %s, %v", p.RemoteAddr(), err)
			} else {
				logs.Debug("PacketConn.ReadPacket Timeout, Heartbeat broken, %s, %v", p.RemoteAddr(), err)
			}
		} else if err == io.EOF || util.NetIsClosed(err) {
			//logs.Debug("PacketConn.ReadPacket Closed, %s", p.RemoteAddr())
		} else {
			logs.Warn("PacketConn.ReadPacket Error, %s, %v", p.RemoteAddr(), err)
		}
		return nil, err
	}
	return pkt, err
}

func (p *PacketConn) sendPacket(pkt *Packet) error {
	if pkt == nil {
		logs.Warn("PacketConn.SendPacket got a nil packet, ignored")
		return nil
	}
	_, err := SendPacket(p.iowrite, pkt)
	p.iowrite.Flush()
	if err != nil {
		e, ok := err.(net.Error)
		if ok && e.Temporary() {
			if e.Timeout() {
				logs.Warn("PacketConn.SendPacket maybe timeout, packet might be lost, %s, %v", p.RemoteAddr(), err)
			} else {
				logs.Warn("PacketConn.SendPacket got Temporary error, packet might be lost, %s, %v", p.RemoteAddr(), err)
			}
		} else if err == io.EOF || util.NetIsClosed(err) {
			//logs.Debug("PacketConn.SendPacket client closed, %s", p.RemoteAddr())
		} else {
			logs.Warn("PacketConn.SendPacket Error, %s, %v", p.RemoteAddr(), err)
		}
	}
	return err
}

type PacketConnAgent struct {
	PacketConn
	SessionId     int64
	Name          string
	WaitSendAll   bool
	WriteDeadLine time.Duration
	IdleTimeSpan  time.Duration
	//broken         bool
	lastActionTime int64

	readingChan chan *Packet
	sendingChan chan *Packet

	sendErrCloseChan chan struct{}
	readQuitChan     chan struct{}
	gone             chan struct{}
}

func NewPacketConnAgent(name string, connection net.Conn) *PacketConnAgent {
	pca := newPacketConn(connection)
	return &PacketConnAgent{
		Name:          name,
		PacketConn:    pca,
		WriteDeadLine: time.Second * 20,

		lastActionTime: time.Now().UnixNano(),
		IdleTimeSpan:   time.Second * 5,

		readingChan: make(chan *Packet),
		sendingChan: make(chan *Packet, 16), //减少sending Block进程的可能性

		sendErrCloseChan: make(chan struct{}),
		readQuitChan:     make(chan struct{}),
		gone:             make(chan struct{}),
	}
}

func (agent *PacketConnAgent) Start(Quiting <-chan struct{}) {
	//quit := make(chan struct{})
	quit := Quiting
	var wg util.WaitGroupWrapper

	wg.Wrap(func() { agent.reading(quit) })
	agent.sending(quit)
	agent.Conn.Close()

	//sendingChan不需要给关闭，没有goroutine依赖它退出，它会被gc回收
	//close(agent.sendingChan)
	wg.Wait()
	close(agent.gone)
	logs.Debug("PacketConnAgent down")
}

//Stop is the same as Conn.Close()
//It is safe to be called many times as you like
func (agent *PacketConnAgent) Stop() {
	agent.Conn.Close()
}

func (agent *PacketConnAgent) GetReadingChan() <-chan *Packet {
	return agent.readingChan
}

func (agent *PacketConnAgent) GetGoneChan() <-chan struct{} {
	return agent.gone
}

func (agent *PacketConnAgent) IsIdle() bool {
	diff := time.Now().UnixNano() - agent.lastActionTime
	return diff > int64(agent.IdleTimeSpan)
}

// Send sendingChan <- pkt with protection
func (agent *PacketConnAgent) SendPacket(pkt *Packet) bool {
	if pkt != nil {
		select {
		case agent.sendingChan <- pkt:
		default:
			conn := &agent.PacketConn
			logs.Warn("[PacketConnAgent] [%s] SendPacket fulled!, %s", agent.Name, conn.RemoteAddr())
			return false
		}
		return true
	}
	return true
}

func (agent *PacketConnAgent) sending(Quiting <-chan struct{}) {
	conn := &agent.PacketConn
	//var timer *time.Timer
	timerdur := defaultPingTime * time.Second

	send := func(pkt *Packet) bool {
		err := conn.SetWriteDeadline(time.Now().Add(agent.WriteDeadLine))
		if err != nil {
			logs.Warn("[PacketConnAgent] [%s] WriteDeadline Error, %s, %v", agent.Name, conn.RemoteAddr(), err)
			conn.SetWriteDeadline(time.Time{})
			return false
		}

		err = agent.sendPacket(pkt)
		//DO NOT USE ELSE HERE
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Temporary() {
				if e.Timeout() {
					//logs.Error("[PacketConnAgent] [%s] Write TCP maybe timeout, packet might be lost, %s, %v", agent.Name, conn.RemoteAddr(), err)
					conn.SetWriteDeadline(time.Time{})
					return false
				} else {
					//logs.Error("[PacketConnAgent] [%s] Write TCP got Temporary error, packet might be lost, %s, %v", agent.Name, conn.RemoteAddr(), err)
				}
			} else if err == io.EOF {
				//logs.Info("[PacketConnAgent] [%s] Write TCP client closed, %s", agent.Name, conn.RemoteAddr())
				conn.SetWriteDeadline(time.Time{})
				return false
			} else {
				//logs.Error("[PacketConnAgent] [%s] Write TCP Error, %s, %v, %t", agent.Name, conn.RemoteAddr(), err, err)
				conn.SetWriteDeadline(time.Time{})
				return false
			}
		}
		conn.SetWriteDeadline(time.Time{})
		return true
	}

Loop:
	for {
		select {
		case <-agent.readQuitChan:
			break Loop
		case <-Quiting:
			break Loop
		case <-secondsTimer.After(timerdur):
			//case <-func() <-chan time.Time {
			//if timer == nil {
			//timer = time.NewTimer(timerdur)
			//} else {
			//timer.Reset(timerdur)
			//}
			//return timer.C
			//}():
			//周期性发PING呼叫客户端，可以主动发现客户端掉线问题
			if ok := send(NewPingPacket()); !ok {
				break Loop
			}
		case pkt := <-agent.sendingChan:
			agent.lastActionTime = time.Now().UnixNano()
			if ok := send(pkt); !ok {
				break Loop
			}
		}
	}
	close(agent.sendErrCloseChan) // 通知reading主loop退出
	logs.Debug("[PacketConnAgent] [%s] sending routine exit, la:%s ra:%s", agent.Name, conn.LocalAddr(), conn.RemoteAddr())
}

func (agent *PacketConnAgent) reading(Quiting <-chan struct{}) {

	conn := &agent.PacketConn
Loop:
	for {
		select {
		case <-Quiting:
			break Loop
		case <-agent.sendErrCloseChan:
			break Loop
		default:
		}

		//err := conn.SetReadDeadline(time.Now().Add(time.Second * 15))
		//if err != nil {
		//logs.Error("[PacketConnAgent] [%s] SetReadDeadline Error, %s, %v", agent.Name, conn.RemoteAddr(), err)
		//break Loop
		//}

		pkt, err := conn.readPacket(time.Second)
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Temporary() {
				continue Loop
			} else {
				if err == msg_size_err {
					logs.Error("[PacketConnAgent] [%s] Msg Size Error, %s, %v", agent.Name, conn.RemoteAddr(), err.Error())
				}
				break Loop
			}
		}
		conn.SetReadDeadline(time.Time{})

		select {
		case agent.readingChan <- pkt:
			agent.lastActionTime = time.Now().UnixNano()
		case <-time.After(10 * time.Second):
			logs.Error("[PacketConnAgent] [%s] reading ignore a packet, due to chan timeout!, la:%s ra:%s", agent.Name, conn.LocalAddr(), conn.RemoteAddr())
		}
	}
	close(agent.readingChan)
	close(agent.readQuitChan)
	logs.Debug("[PacketConnAgent] [%s] reading routine exit, la:%s ra:%s", agent.Name, conn.LocalAddr(), conn.RemoteAddr())
}
