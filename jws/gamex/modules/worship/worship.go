package worship

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/balance_timer"
)

const (
	cmdTypeGetWorshipData = iota + 1
	cmdTypeWorship
	cmdTypeGetWorship
	cmdTypeRefershTop
)

type cmd struct {
	Type    int
	Account string
	Top     []worshipAccData

	ResChan chan *cmd
}

type module struct {
	top TopAccountWorship

	sid         uint
	wg          sync.WaitGroup
	cmdChan     chan *cmd
	stopChan    chan bool
	BalanceChan chan bool
}

func New(sid uint) *module {
	m := new(module)
	m.top.init(sid)
	m.sid = sid
	return m
}

func (r *module) AfterStart(g *gin.Engine) {
	balance.GetModule(r.sid).RegBalanceNotifyChan(
		"worship",
		r.BalanceChan, gamedata.GetRankWorshipBalance())
}

func (r *module) BeforeStop() {
}

func (r *module) Start() {
	r.wg.Add(1)
	timeChan := time.After(60 * time.Second)
	r.stopChan = make(chan bool, 1)
	r.cmdChan = make(chan *cmd, 1024)
	r.BalanceChan = make(chan bool, 64)

	err := r.top.loadDB()
	if err != nil {
		panic(err)
	}

	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-timeChan:
				timeChan = time.After(60 * time.Second)
				r.top.saveDB()
			case <-r.stopChan:
				r.top.saveDB()
				return
			case command := <-r.cmdChan:
				res := r.processCmd(command)
				if command.ResChan != nil {
					command.ResChan <- res
				}
			case <-r.BalanceChan:
				r.top.clean()
			}
		}
	}()
}

func (r *module) Stop() {
	r.stopChan <- true
	r.wg.Wait()
}

func (r *module) sendCmd(ctx context.Context, c *cmd) *cmd {
	c.ResChan = make(chan *cmd, 1)
	r.cmdChan <- c
	var t *cmd
	select {
	case t = <-c.ResChan:
	case <-ctx.Done():
		return nil
	}
	return t
}
