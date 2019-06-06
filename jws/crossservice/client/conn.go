package client

import (
	"fmt"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/crossservice/util/discover/publish"

	"io"

	"vcs.taiyouxi.net/jws/crossservice/helper"
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/util/connect"
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/watcher"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
)

const (
	defaultWatcherInterval = 2
	defaultConnPoolMax     = 200
)

type manageConnect struct {
	watcher     *watcher.Watcher
	catchGroups map[uint32]bool

	clientLock       sync.RWMutex
	clients          map[string]*connect.Client
	clientsWithGroup map[uint32]*connect.Client

	pushNotify chan *message.Request
	shardIDs   []uint32
	gid        uint32

	connPoolMax uint32

	handlePublish *publish.Handle

	isClosed bool
}

func newManageConnect(gid uint32, shardIDs []uint32) *manageConnect {
	publishParent := fmt.Sprintf("gamex/%d/%d/crossservice", gid, shardIDs[0])
	m := &manageConnect{
		clients:          make(map[string]*connect.Client),
		clientsWithGroup: make(map[uint32]*connect.Client),
		catchGroups:      make(map[uint32]bool),
		pushNotify:       make(chan *message.Request, DefaultPushNotifyBuff),
		shardIDs:         shardIDs,
		gid:              gid,

		connPoolMax: defaultConnPoolMax,

		handlePublish: publish.NewHandle(publishParent),

		isClosed: false,
	}
	return m
}

func (m *manageConnect) setConnPoolMax(pm uint32) {
	m.connPoolMax = pm
}

func (m *manageConnect) addService(groupID uint32) {
	m.catchGroups[groupID] = true
}

func (m *manageConnect) start() error {
	handle := watcher.NewWatcher()
	handle.SetProject(helper.ProjectName)

	filters := []watcher.Filter{}
	filters = append(filters, watcher.FilterService{Service: helper.ServiceName})
	filters = append(filters, watcher.FilterVersion{Version: version.Version})
	handle.SetFilter(filters)

	handle.OnServiceAdd = m.handleWatcherAdd
	handle.OnServiceUpdate = m.handleWatcherUpdate
	handle.OnServiceDel = m.handleWatcherDel

	handle.Start()
	return nil
}

func (m *manageConnect) stop() {
	m.isClosed = true
	m.handlePublish.UnPublish()
	m.clientLock.Lock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clientLock.Unlock()
	close(m.pushNotify)
}

func (m *manageConnect) checkInGroupCatch(gid uint32, gl []uint32) bool {
	if m.gid != gid {
		return false
	}
	for _, g := range gl {
		if m.catchGroups[g] {
			return true
		}
	}
	return false
}

func (m *manageConnect) handleWatcherAdd(s discover.Service) {
	param, err := helper.UnmarshalServiceParam(s.Extra)
	if nil != err {
		logs.Error(fmt.Sprintf("Crossservice Client, Make Connection, Parse Service Failed, %v", err))
		return
	}
	if false == m.checkInGroupCatch(param.Gid, param.GroupIDs) {
		logs.Warn("Crossservice Client, Make Connection, Ignore Service Addr[%s:%d](%s), Param%+v", s.IP, s.Port, s.Index, param)
		return
	}

	m.clientLock.Lock()
	ec, exist := m.clients[s.Index]
	m.clientLock.Unlock()
	if false == exist {
		client, err := connect.NewClient("tcp4", param.IP, param.Port, m.connPoolMax)
		if nil != err {
			logs.Error(fmt.Sprintf("Crossservice Client, Make Connection, Make Connect Client Failed, %v", err))
			return
		}
		ec = client
		logs.Info("Crossservice Client, Make Connection, Connect Success, Service Addr[%s:%d](%s), Param%+v", s.IP, s.Port, s.Index, param)
		m.clientLock.Lock()
		m.clients[s.Index] = ec
		m.clientLock.Unlock()
		go func() {
			defer logs.PanicCatcherWithInfo("CrossService Client, listen from Server")
			for {
				needRetry := m.listenFromServer(ec)
				if false == needRetry {
					logs.Warn("Crossservice Client, Connection For ListenServer exist, Service Addr[%s:%d](%s), Param%+v", s.IP, s.Port, s.Index, param)
					break
				}
			}
		}()
		go func() {
			defer logs.PanicCatcherWithInfo("CrossService Client, holding release")
			ec.HoldingRelease()
		}()
	}
	m.clientLock.Lock()
	for _, groupID := range param.GroupIDs {
		m.clientsWithGroup[groupID] = ec
	}
	m.clientLock.Unlock()

	m.handlePublish.PushAdd(fmt.Sprintf("%s:%d", s.IP, s.Port), s.Extra)
}

func (m *manageConnect) handleWatcherDel(s discover.Service) {
	param, err := helper.UnmarshalServiceParam(s.Extra)
	if nil != err {
		return
	}
	if false == m.checkInGroupCatch(param.Gid, param.GroupIDs) {
		logs.Warn("Crossservice Client, Remove Connection, Ignore Service Addr[%s:%d](%s), Param%+v", s.IP, s.Port, s.Index, param)
		return
	}
	m.clientLock.Lock()
	for _, groupID := range param.GroupIDs {
		delete(m.clientsWithGroup, groupID)
	}
	ec := m.clients[s.Index]
	delete(m.clients, s.Index)
	m.clientLock.Unlock()

	if nil != ec {
		ec.Close()
	}
	m.handlePublish.PushDel(fmt.Sprintf("%s:%d", s.IP, s.Port), s.Extra)
}

func (m *manageConnect) handleWatcherUpdate(s discover.Service) {
	m.handleWatcherDel(s)
	m.handleWatcherAdd(s)
}

func (m *manageConnect) getConn(groupID uint32) (*connect.Conn, error) {
	m.clientLock.RLock()
	client, exist := m.clientsWithGroup[groupID]
	m.clientLock.RUnlock()

	if false == exist {
		return nil, fmt.Errorf("group %d is not holding", groupID)
	}
	return client.GetConn()
}

func (m *manageConnect) listenFromServer(client *connect.Client) bool {
	conn, err := client.GetConn()
	if nil != err {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, get Conn failed, %v", err))
		return false
	}
	defer conn.Release()

	param := helper.EncodeHelloReq(&helper.HelloReq{ShardIDs: m.shardIDs})
	req := &message.Request{
		ProtocolID: message.ProtocolHelloReq,
		Data:       param,
		DataLen:    uint32(len(param)),
	}
	reqMsg, err := message.EncodeRequestToMessage(req)
	if nil != err {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Encode Request To Message failed, %v", err))
		return true
	}
	if err := conn.Send(reqMsg); nil != err {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Send Message failed, %v", err))
		return true
	}

	timeout := time.Now().Add(DefaultRecvTimeout)
	rspMsg, err := conn.RecvWithTimeout(timeout)
	if nil != err {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Recv Hello Rsp Message failed, %v", err))
		return true
	}
	rsp, err := message.DecodeMessageToResponse(rspMsg)
	if nil != err {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Decode Message To Request failed, %v", err))
		return true
	}
	if message.ProtocolHelloRsp != rsp.ProtocolID {
		logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Recv Hello Rsp Message un-expect, %+v", rsp))
		return true
	}

	needRetry := true
	for {
		msg, err := conn.Recv()
		if nil != err {
			if m.isClosed {
				break
			}
			if io.EOF == err {
				logs.Warn("Cross Service Client, Listen From Server, Recv EOF")
				needRetry = false
				break
			}
			logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Recv Message failed, %v", err))
			break
		}
		req, err := message.DecodeMessageToRequest(msg)
		if nil != err {
			logs.Error(fmt.Sprintf("Cross Service Client, Listen From Server, Decode Message To Request failed, %v", err))
			continue
		}
		m.pushNotify <- req
	}

	return needRetry
}

func (m *manageConnect) pull() <-chan *message.Request {
	return m.pushNotify
}
