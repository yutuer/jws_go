package server

import (
	"fmt"
	"sync"
	"time"

	"bytes"
	"encoding/gob"

	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//modules ..
type modules struct {
	ms     map[uint32]map[string]module.Module
	groups []uint32

	poolChannel map[uint32]map[string][]*dispatchChannel
	poolLock    sync.RWMutex

	isClosed bool
}

//newModules ..
func newModules(groups []uint32) *modules {
	m := &modules{
		ms:     make(map[uint32]map[string]module.Module),
		groups: groups,

		poolChannel: make(map[uint32]map[string][]*dispatchChannel),
	}
	return m
}

func (m *modules) close() {
	m.isClosed = true
	// m.poolLock.Lock()
	// defer m.poolLock.Unlock()
	for _, gc := range m.poolChannel {
		for _, mc := range gc {
			for _, dc := range mc {
				if nil != dc {
					close(dc.ch)
				}
			}
		}
	}
}

func (m *modules) loadModules(f module.FuncPush) {
	for _, g := range m.groups {
		m.ms[g] = make(map[string]module.Module)
		m.poolChannel[g] = make(map[string][]*dispatchChannel)
		for _, mg := range module.LoadModulesList {
			mo := mg.NewModule(g)
			mo.SetFuncPush(f)
			m.ms[g][mo.ModuleID()] = mo
			m.poolChannel[g][mo.ModuleID()] = make([]*dispatchChannel, mo.HashMask())
		}
	}
}

func (m *modules) start() {
	for _, gm := range m.ms {
		for _, mo := range gm {
			mo.Start()
		}
	}
	for _, gm := range m.ms {
		for _, mo := range gm {
			mo.AfterStart()
		}
	}
}

func (m *modules) stop() {
	for _, gm := range m.ms {
		for _, mo := range gm {
			mo.BeforeStop()
		}
	}
	for _, gm := range m.ms {
		for _, mo := range gm {
			mo.Stop()
		}
	}
}

func (m *modules) getModule(group uint32, moName string) module.Module {
	if mom, exist := m.ms[group]; exist {
		if mo, exist := mom[moName]; exist {
			return mo
		}
	}
	return nil
}

func (m *modules) dispatch(t *transaction) (int, error) {
	mo := m.getModule(t.req.GroupID, t.req.Module)
	if nil == mo {
		return message.ErrCodeInner, fmt.Errorf("modules dispatch, have no %s module", t.req.Module)
	}
	me := mo.GetMethod(t.req.Method)
	if nil == me {
		return message.ErrCodeInner, fmt.Errorf("modules dispatch, have no %s.%s method", t.req.Module, t.req.Method)
	}

	t.method = me
	dc := m.getChannel(t.req.GroupID, mo.ModuleID(), mo.Hash(t.req.HashSource)%mo.HashMask())
	if nil == dc {
		return message.ErrCodeInner, fmt.Errorf("modules dispatch, have no channel of %s:%s", t.req.Module, t.req.HashSource)
	}
	select {
	case dc.ch <- t:
	case <-time.After(DefaultSyncTransactionTimeout):
		return message.ErrCodeTimeout, fmt.Errorf("Sync get Response from transaction timeout")
	}
	return message.ErrCodeOK, nil
}

func (m *modules) getChannel(group uint32, moName string, hash uint32) *dispatchChannel {
	dc := m.checkoutChannel(group, moName, hash)
	if nil == dc {
		dc = m.createChannel(group, moName, hash)
		if nil == dc {
			return nil
		}
		go func() {
			defer func() {
				logs.Warn("Modules RemoveChannel %d:%s:%d", dc.group, dc.module, dc.hash)
				m.removeChannel(dc)
			}()
			dc.run()
		}()
	}
	return dc
}

func (m *modules) checkoutChannel(group uint32, moName string, hash uint32) *dispatchChannel {
	return m.poolChannel[group][moName][hash]
}

func (m *modules) createChannel(group uint32, moName string, hash uint32) *dispatchChannel {
	if m.isClosed {
		return nil
	}
	dc := &dispatchChannel{
		ch:     make(chan *transaction, DefaultServiceProcessChannelCapacity),
		hash:   hash,
		group:  group,
		module: moName,
	}
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	if nil != m.poolChannel[group][moName][hash] {
		return m.poolChannel[group][moName][hash]
	}
	m.poolChannel[group][moName][hash] = dc
	return dc
}

func (m *modules) removeChannel(dc *dispatchChannel) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	m.poolChannel[dc.group][dc.module][dc.hash] = nil
}

type dispatchChannel struct {
	ch     chan *transaction
	module string
	group  uint32
	hash   uint32
}

func (c *dispatchChannel) run() {
	defer logs.PanicCatcherWithInfo("CrossService Server, Dispatch Channel run")

	for {
		// logs.Debug("CrossService Server, Dispatch Channel %d:%s:%d do one", c.group, c.module, c.hash)
		tran, ok := <-c.ch
		if false == ok {
			break
		}
		if nil == tran || nil == tran.method {
			continue
		}
		t := module.Transaction{
			GroupID:    tran.req.GroupID,
			HashSource: tran.req.HashSource,
		}
		param := tran.method.NewParam()
		gob.NewDecoder(bytes.NewBuffer(tran.req.Data[:tran.req.DataLen])).Decode(param)
		ret := tran.method.NewRet()
		errCode := uint32(message.ErrCodeInner)
		func() {
			defer logs.PanicCatcherWithInfo("CrossService Server, Dispatch Channel Panic When Do Method")
			errCode, ret = tran.method.Do(t, param)
		}()
		// logs.Debug("CrossService Server, Dispatch Channel %d:%s:%d do one step 2", c.group, c.module, c.hash)
		if tran.sync {
			bs := new(bytes.Buffer)
			gob.NewEncoder(bs).Encode(ret)
			tran.rsp <- message.MakeTmpResponse(errCode, bs.Bytes())
		}
		// logs.Debug("CrossService Server, Dispatch Channel %d:%s:%d do one step 3", c.group, c.module, c.hash)
	}
}
