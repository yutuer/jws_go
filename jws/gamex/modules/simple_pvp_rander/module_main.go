package sPvpRander

import (
	//"fmt"
	"sync"
	"time"

	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	commandTypNull = iota
	commandTypUpdateRander
	commandTypRand
)

type command struct {
	typ    int
	uid    string
	count  int
	rank   int
	rander simplePvpRander
	res    chan []string
}

//RandSimplePvpEnemy simplePVP随机一个对手
// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func RandSimplePvpEnemy(sid uint, selfID string, count, rank int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	resChan := make(chan []string, 1)
	select {
	case GetModule(sid).randerIoChan <- command{
		typ:   commandTypRand,
		uid:   selfID,
		count: count,
		rank:  rank,
		res:   resChan,
	}:
	case <-ctx.Done():
		logs.Error("RandSimplePvpEnemy cmd put timeout")
	}

	select {
	case res := <-resChan:
		return res
	case <-ctx.Done():
		logs.Error("RandSimplePvpEnemy <-res_chan timeout")
		return nil
	}
}

const (
	commandIOChanSize = 1024
	randerMakeIn      = 30
)

func genSimplePvpRanderModule(sid uint) *simplePvpRanderModule {
	return &simplePvpRanderModule{
		sid: sid,
	}
}

type simplePvpRanderModule struct {
	sid          uint
	waitter      sync.WaitGroup
	quitWaitter  sync.WaitGroup
	randerIoChan chan command
	rander       simplePvpRander
	quitChan     chan bool
}

func (r *simplePvpRanderModule) MakeRanderPool() {
	newRander := simplePvpRander{}
	err := newRander.Make(r.sid)
	if err != nil {
		logs.Error("Make simplePvpRander Err By %s", err.Error())
	} else {
		r.randerIoChan <- command{
			typ:    commandTypUpdateRander,
			rander: newRander,
		}
	}
}

func (r *simplePvpRanderModule) AfterStart(g *gin.Engine) {
	r.MakeRanderPool()
}

func (r *simplePvpRanderModule) BeforeStop() {
}

func (r *simplePvpRanderModule) Start() {
	r.waitter.Add(1)
	defer r.waitter.Done()

	r.randerIoChan = make(chan command, commandIOChanSize)
	timerChan := uutil.TimerMS.After(30 * time.Second)
	r.quitChan = make(chan bool, 1)

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			cmd, ok := <-r.randerIoChan
			logs.Trace("randerIoChan command %v", cmd)

			if !ok {
				logs.Warn("randerIoChan command close")
				return
			}

			switch cmd.typ {
			case commandTypUpdateRander:
				r.rander = cmd.rander
			case commandTypRand:
				ids, err := r.rander.randEnemy(
					cmd.uid,
					cmd.rank,
					cmd.count)
				if err != nil {
					cmd.res <- nil
				} else {
					cmd.res <- ids
				}
			}

		}
	}()

	r.quitWaitter.Add(1)
	go func() {
		var last int64
		defer r.quitWaitter.Done()
		for {
			select {
			case <-timerChan:
				timerChan = uutil.TimerMS.After(30 * time.Second)
				func() {
					defer func() {
						if err := recover(); err != nil {
							logs.Error("rander_mk panic, Err %v", err)
						}
					}()
					nowT := time.Now().Unix()
					if nowT-last > randerMakeIn {
						last = nowT
						logs.Trace("MakeRanderPool")
						r.MakeRanderPool()
					}
				}()
			case <-r.quitChan:
				logs.Warn("Stop MakeRanderPool")
				return
			}
		}
	}()
}

func (r *simplePvpRanderModule) Stop() {
	r.quitChan <- true
	r.quitWaitter.Wait()
	close(r.randerIoChan)
	r.waitter.Wait()
}
