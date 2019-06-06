package game

import (
	"sync/atomic"
	"time"

	"github.com/siddontang/go/timingwheel"
	"github.com/ugorji/go/codec"

	"vcs.taiyouxi.net/platform/planx/client"
	//"vcs.taiyouxi.net/platform/planx/servers"
	"strconv"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var secondsTimer *timingwheel.TimingWheel

func init() {
	secondsTimer = timingwheel.NewTimingWheel(time.Second, 60*10)
}

type ChanAgent struct {
	Name string

	In  chan *client.Packet
	Out chan *client.Packet

	closeFlag int32

	done      chan struct{}
	waitGroup util.WaitGroupWrapper
}

func NewChanAgent(name string) *ChanAgent {
	agent := &ChanAgent{
		Name: name,

		In:  make(chan *client.Packet, 64),
		Out: make(chan *client.Packet, 64),

		done: make(chan struct{}),
	}
	return agent
}

func (agent *ChanAgent) IsClosed() bool { return atomic.LoadInt32(&agent.closeFlag) != 0 }

// GetReadingChan of ChanAgent, caller should be a game ONLY.
func (agent *ChanAgent) GetReadingChan() <-chan *client.Packet {
	return agent.In
}

func GetSecondsAfterTimer(dur time.Duration) <-chan struct{} {
	return secondsTimer.After(dur)
}

// SendPacket of ChanAgent, caller should be a game ONLY.
func (agent *ChanAgent) SendPacket(pkt *client.Packet) bool {
	if agent.IsClosed() {
		return false
	}
	select {
	case agent.Out <- pkt:
	case <-agent.GetGoneChan():
		return false
	case <-GetSecondsAfterTimer(10 * time.Second):
		logs.Critical("ChanAgent SendPacket timeout")
		return false
	}

	return true
}

func (agent *ChanAgent) GetGoneChan() <-chan struct{} {
	return agent.done
}

//AllinOne模式特有：
//因为ChanAgent是GameServer的代理，模拟tcp模式下的使用
//因此同一个实例的ChanAgent在Gate中被当成相反的读写模式
// GameServer(Write) --> ChanAgent.Out --> Gate(GetReadingChan)
// GameServer(Read) <-- ChanAgent.In <-- Gate(SendPacket)
//因此不得不使用独立的chan作为ChanAgent模拟断线的形式。
func (agent *ChanAgent) close() {
	close(agent.done)
}

func (agent *ChanAgent) Stop() {
	if atomic.CompareAndSwapInt32(&agent.closeFlag, 0, 1) {
		close(agent.In)
		close(agent.Out)
	}
}

type ChanServer struct {
	quit             chan struct{}
	prepareHandshake PreparePlayer
	waitGroup        util.WaitGroupWrapper
}

func NewChanServer() *ChanServer {
	gamesrv := &ChanServer{
		quit: make(chan struct{}),
	}
	return gamesrv
}

func (c *ChanServer) Accept(name string) *ChanAgent {
	ca := NewChanAgent(name)
	c.waitGroup.Wrap(func() {
		c.handleConnection(ca)
		logs.Trace("ChanServer Agent handleConnection quit.")
		//ca.Stop()
	})

	//XXX 不需要释放ca，因为这里的ca在外层会转变成另外一个角色，会被Stop
	//所以不需要在handleConnection函数里面处理ca
	return ca
}

func (c *ChanServer) handleConnection(agent *ChanAgent) {
	defer logs.PanicCatcher()

	incoming := agent.GetReadingChan()
	var accountid, ip, gziplimit string
	var gzipLimit uint64
	gzipLimit = 0
	select {
	case <-c.quit:
		return
	case pkt := <-incoming:
		if pkt.GetContentType() != client.PacketIDGateSession {
			logs.Error("[GameCHANSerever] should get PacketIDGateSession firstly.")
			return
		}
		//accountid = "0:0:1001"
		var mh codec.MsgpackHandle
		mh.RawToString = true

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
		logs.Trace("[GameCHANSerever] handleConnection with account %s", accountid)
	}

	playerRes := c.prepareHandshake(accountid, ip)
	defer func() {
		playerRes.OnExit()
		agent.close()
		logs.Trace("[GameCHANSerever] handleConnection account %s, exited in game part", accountid)
	}()

	playerProcessor(c.quit, playerRes, incoming, agent.SendPacket, gzipLimit)
}

func (c *ChanServer) Stop() {
	close(c.quit)
}

func (c *ChanServer) Start(pp PreparePlayer) {
	if pp != nil {
		c.prepareHandshake = pp
	} else {
		panic("[GameCHANSerever] Should start with PreparePlayer function")
	}
	<-c.quit
	c.waitGroup.Wait()
	logs.Info("[GameCHANSerever] Game Server Chan Mode exit.")
}
