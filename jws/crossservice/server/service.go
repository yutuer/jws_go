package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"vcs.taiyouxi.net/jws/crossservice/helper"
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/metrics"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/crossservice/util/connect"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/publish"
	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type service struct {
	ServerIP   string
	ServerPort int

	cs      *connect.Server
	modules *modules

	pushConnects map[uint32]*connect.Conn
	pushLock     sync.RWMutex

	publishHandle *publish.Handle

	isClosed  bool
	processWg sync.WaitGroup
}

func newService(serviceIndex string, ip string, port int, groups []uint32) (*service, error) {
	s := &service{
		ServerIP:     ip,
		ServerPort:   port,
		pushConnects: map[uint32]*connect.Conn{},
		isClosed:     false,
	}

	//加载模块
	modules := newModules(groups)
	modules.loadModules(s.pushToClient)
	s.modules = modules
	s.modules.start()

	cs, err := connect.NewServer("tcp4", s.ServerIP, s.ServerPort)
	if nil != err {
		return nil, fmt.Errorf("start connect server failed, %v", err)
	}
	logs.Warn("connect server listen: %s", cs.LocalAddr().String())
	s.cs = cs

	s.publishHandle = publish.NewHandle(helper.ServiceName + "/" + serviceIndex)

	return s, nil
}

func (s *service) run() {
	defer logs.PanicCatcherWithInfo("CrossService Server, service run")

	go func() {
		defer logs.PanicCatcherWithInfo("CrossService Server, connect server run")
		err := s.cs.Run()
		if nil != err && connect.ErrClosed != err {
			logs.Error(fmt.Sprintf("CrossService Server, connect server end with error, %v", err))
			s.cs.Close()
		}
	}()

	for _, groupID := range s.modules.groups {
		s.publishHandle.AddElem(fmt.Sprintf("group/%d", groupID), fmt.Sprintf("%d", groupID))
	}

	go func() {
		interval := time.Second * 10
		s.publishHandle.SetTTL(interval * 10)
		ticker := time.NewTicker(interval)
		for !s.isClosed {
			s.publishHandle.Publish()
			<-ticker.C
		}
	}()

	s.regGinHandle()

	for !s.isClosed {
		msg, conn, err := s.cs.Recv()
		if nil != err {
			if connect.ErrClosed == err || true == s.isClosed {
				break
			}
			if io.EOF == err {
				continue
			}
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Err.Error() == syscall.ECONNRESET.Error() {
					continue
				}
			}
			logs.Error(fmt.Sprintf("CrossService Server, service get msg with error, %v", err))
			continue
		}
		go func() {
			defer logs.PanicCatcherWithInfo("CrossService Server, process message")
			s.processWg.Add(1)
			defer s.processWg.Done()
			cs_metrics.AddOpt(1)
			if err := s.processMsg(msg, conn); nil != err {
				logs.Error(fmt.Sprintf("CrossService Server, service process msg with error, %v", err))
			}
		}()
	}
}

func (s *service) close() {
	s.isClosed = true
	s.publishHandle.UnPublish()
	if nil != s.cs {
		s.cs.Close()
	}
	s.processWg.Wait()
	s.modules.stop()
	s.modules.close()
}

func (s *service) setIPFilter(fs []string) {
	if 0 == len(fs) {
		return
	}
	filter := connect.NewIPFilter()
	for _, str := range fs {
		filter.Add(str)
	}
	s.cs.SetIPFilter(filter)
}

func (s *service) processMsg(msg *connect.Message, conn *connect.Conn) error {
	req, err := message.DecodeMessageToRequest(msg)
	if nil != err {
		return fmt.Errorf("DecodeMessageToRequest failed, %v", err)
	}

	var rsp *message.Response
	switch req.ProtocolID {
	case message.ProtocolHelloReq:
		rsp = s.processHello(req, conn)
	case message.ProtocolSyncReq:
		rsp = s.processSync(req)
	case message.ProtocolAsyncReq:
		rsp = s.processAsync(req)
	}
	rspMsg, err := message.EncodeResponseToMessage(rsp)
	if nil != err {
		return fmt.Errorf("processMsg EncodeResponseToMessage failed, %v", err)
	}
	if err := conn.Send(rspMsg); nil != err && connect.ErrClosed != err && false == s.isClosed {
		return fmt.Errorf("processMsg Send Response failed, %v", err)
	}
	return nil
}

func (s *service) processSync(req *message.Request) *message.Response {
	tran := makeSyncTransaction(req)
	errCode, err := s.modules.dispatch(tran)
	var rsp *message.Response
	if nil != err || message.ErrCodeOK != errCode {
		rsp = message.MakeErrResponse(message.ProtocolSyncRsp, errCode, err)
	} else {
		select {
		case b, ok := <-tran.rsp:
			if false == ok {
				rsp = message.MakeErrInnerResponse(message.ProtocolSyncRsp, fmt.Errorf("Sync get Response from closed transaction"))
			} else {
				rsp = message.MakeSyncAckResponse(b)
			}
		case <-time.After(DefaultSyncTransactionTimeout):
			rsp = message.MakeErrResponse(message.ProtocolSyncRsp, message.ErrCodeTimeout, fmt.Errorf("Sync get Response from transaction timeout"))
		}
	}
	return rsp
}

func (s *service) processAsync(req *message.Request) *message.Response {
	tran := makeAsyncTransaction(req)
	errCode, err := s.modules.dispatch(tran)
	var rsp *message.Response
	if nil != err || message.ErrCodeOK != errCode {
		rsp = message.MakeErrResponse(message.ProtocolAsyncRsp, errCode, err)
	} else {
		rsp = message.MakeAsyncAckResponse(req)
	}
	return rsp
}

func (s *service) processHello(req *message.Request, conn *connect.Conn) *message.Response {
	hello := helper.DecodeHelloReq(req.Data[:req.DataLen])

	logs.Info("[CrossService] Get Hello from [%s] ShardIDs %+v", conn.RemoteAddr().String(), hello.ShardIDs)

	s.pushLock.Lock()
	for _, sid := range hello.ShardIDs {
		oldConn := s.pushConnects[sid]
		if nil != oldConn {
			s.publishHandle.PushDel(fmt.Sprintf("gamex/%d", sid), conn.RemoteAddr().String())
			logs.Info("[CrossService] Delete old connection [%s] as shard [%d] %+v", conn.RemoteAddr().String(), sid)
		}
		s.pushConnects[sid] = conn
		s.publishHandle.AddElem(fmt.Sprintf("gamex/%d", sid), conn.RemoteAddr().String())
	}
	s.pushLock.Unlock()

	return message.MakeHelloAckResponse()
}

func (s *service) pushToClient(sid uint32, mo string, me string, p module.Param) error {
	conn := s.getPushConn(sid)
	if nil == conn {
		return fmt.Errorf("Service get Connection to Shard %d failed", sid)
	}
	if nil != conn.Err() {
		return fmt.Errorf("Service get Connection with old error to Shard %d, error: %v", sid, conn.Err())
	}
	bs := new(bytes.Buffer)
	gob.NewEncoder(bs).Encode(p)
	req := &message.Request{
		ProtocolID: message.ProtocolPushReq,
		Module:     mo,
		Method:     me,
		Data:       bs.Bytes(),
		DataLen:    uint32(len(bs.Bytes())),
	}
	msg, err := message.EncodeRequestToMessage(req)
	if nil != err {
		return err
	}

	if err := conn.Send(msg); nil != err {
		if connect.ErrClosed == err {
			logs.Warn("[CrossService] pushToClient, Send to close connection, Shard [%d]", sid)
		} else {
			return fmt.Errorf("Push Send Response failed, %v", err)
		}
	}
	return nil
}

func (s *service) getPushConn(sid uint32) *connect.Conn {
	s.pushLock.RLock()
	defer s.pushLock.RUnlock()
	return s.pushConnects[sid]
}

func (s *service) regGinHandle() {
	http_util.POST("/SendBadPacketToShard", s.ginSendBadPacketToShard)

	http_util.GET("/ShowConnectedShard", s.ginShowConnectedShard)
}

func (s *service) ginShowConnectedShard(c *gin.Context) {
	out := "Connected Shard\n"

	list := []uint32{}
	s.pushLock.RLock()
	for sid := range s.pushConnects {
		list = append(list, sid)
	}
	s.pushLock.RUnlock()

	col := 5
	index := 0
	for _, sid := range list {
		index++
		out += fmt.Sprintf("%10d", sid)
		if 0 == index%col {
			out += "\n"
		}
	}
	out += "\n"

	c.String(200, out)
	return
}

//------debug

func (s *service) ginSendBadPacketToShard(c *gin.Context) {
	shard := c.PostForm("shard")

	shardID, err := strconv.ParseUint(shard, 10, 32)
	if nil != err {
		c.String(400, err.Error()+"\n")
		return
	}

	param := ""
	for i := 0; len(param) < 0xFFF0*2+100; i++ {
		param += fmt.Sprintf("|%d|", i)
	}

	if err := s.PushBadToClient(uint32(shardID), "module", "method", param); nil != err {
		c.String(400, err.Error()+"\n")
		return
	}
	c.String(200, shard+"\n")
	return
}

func (s *service) PushBadToClient(sid uint32, mo string, me string, p module.Param) error {
	conn := s.getPushConn(sid)
	if nil == conn {
		return fmt.Errorf("Service get Connection to Shard %d failed", sid)
	}
	if nil != conn.Err() {
		return fmt.Errorf("Service get Connection with old error to Shard %d, error: %v", sid, conn.Err())
	}
	bs := new(bytes.Buffer)
	gob.NewEncoder(bs).Encode(p)
	req := &message.Request{
		ProtocolID: message.ProtocolPushReq,
		Module:     mo,
		Method:     me,
		Data:       bs.Bytes(),
		DataLen:    uint32(len(bs.Bytes())),
	}
	msg, err := message.EncodeRequestToMessage(req)
	if nil != err {
		return err
	}

	if err := conn.SendBad(msg); nil != err {
		if connect.ErrClosed == err {
			logs.Warn("[CrossService] pushToClient, Send to close connection, Shard [%d]", sid)
		} else {
			return fmt.Errorf("Push Send Response failed, %v", err)
		}
	}
	return nil
}
