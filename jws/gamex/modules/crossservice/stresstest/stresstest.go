package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/crossservice/client"
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module/worldboss"
	"vcs.taiyouxi.net/jws/crossservice/util/connect"
	//"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "rage",
			Usage:  "send message as a berserker",
			Action: rage,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c",
					Usage: "config file",
					Value: "config.json",
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "show verbose",
				},
				cli.IntFlag{
					Name:  "t",
					Usage: `after %d seconds, stop sending message`,
					Value: 0,
				},
			},
		},
	}

	app.Run(os.Args)
}

func rage(c *cli.Context) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	configFile := c.String("c")
	verbose := c.Bool("v")
	keepTime := c.Int("t")

	bs, err := ioutil.ReadFile(configFile)
	if nil != err {
		log.Printf("ReadFile Failed, %v", err)
		return
	}

	cfg := &Config{}
	if err := json.Unmarshal(bs, cfg); nil != err {
		log.Printf("Unmarshal Failed, %v", err)
		return
	}
	waitGroup := sync.WaitGroup{}
	for i := 0; i < cfg.ClientNum; i++ {
		// <-time.After(2 * time.Second)
		<-time.After(200 * time.Millisecond)
		waitGroup.Add(1)
		go func(i int) {
			defer waitGroup.Done()
			Client(cfg, verbose, 10+uint(i), keepTime)
		}(i)
	}
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	<-signalChan
	waitGroup.Wait()
	log.Printf("Stop All")
	return
}

//Client ..
func Client(cfg *Config, verbose bool, sid uint, keepTime int) {
	cs := client.NewClient(uint32(cfg.Gid), []uint32{uint32(sid)})
	//discover.SetRedisCfg(cfg.DiscoverAddr, cfg.DiscoverDB, cfg.DiscoverAuth) 	//注释已经无效的方法
	cs.AddGroupIDs([]uint32{cfg.Group})
	if err := cs.Start(); nil != err {
		log.Printf("CrossService Client Start Failed, %v", err)
		return
	}
	cs.SetConnPoolMax(800)

	st := &static{
		list: make([]*staticElem, cfg.Concurrence),
	}
	for i := 0; i < cfg.Concurrence; i++ {
		st.list[i] = &staticElem{}
	}

	stop := false
	if 0 != keepTime {
		time.AfterFunc(time.Duration(keepTime)*time.Second, func() { stop = true })
	}
	go func() {
		starttime := time.Now()
		for !stop {
			<-time.After(time.Duration(cfg.StaticInterval) * time.Second)
			sumRequestCount := int(0)
			sumRequestTimeout := int(0)
			for i := 0; i < cfg.Concurrence; i++ {
				sumRequestCount += st.list[i].requestCount
				sumRequestTimeout += st.list[i].requestTimeout
			}
			if 0 == sumRequestCount {
				continue
			}
			seconds := time.Now().Sub(starttime).Seconds()
			log.Printf("Client[%d] Static: request [%d](%0.4f/s), timeout [%d(%0.2f%%)]", sid, sumRequestCount, float64(sumRequestCount)/seconds, sumRequestTimeout, 100.0*float32(sumRequestTimeout)/float32(sumRequestCount))
			for s := 0; s < stepMax; s++ {
				sumRequestCountStep := int(0)
				sumRequestTimeoutStep := int(0)
				sumRequestCostSum := time.Duration(0)
				sumRequestCostLast := time.Duration(0)
				for i := 0; i < cfg.Concurrence; i++ {
					sumRequestCountStep += st.list[i].requestCountStep[s]
					sumRequestTimeoutStep += st.list[i].requestTimeoutStep[s]
					sumRequestCostSum += st.list[i].requestCostSum[s]
					for l := 0; l < costLastLen; l++ {
						sumRequestCostLast += st.list[i].requestCostLast[s][l]
					}
				}
				if 0 == sumRequestCountStep {
					continue
				}
				// log.Printf("Client[%d] Static: cost average [%s]", sid, sumRequestCostSum.String())
				log.Printf("Client[%d] Static: Step [%d]: request [%d], timeout [%d(%0.2f%%)], cost average [%s], cost last(%d) average [%s]",
					sid, s, sumRequestCountStep, sumRequestTimeoutStep, 100.0*float32(sumRequestTimeoutStep)/float32(sumRequestCountStep),
					(sumRequestCostSum / time.Duration(sumRequestCountStep)).String(),
					costLastLen,
					(sumRequestCostLast / time.Duration(cfg.Concurrence*costLastLen)).String(),
				)
			}
		}
	}()

	log.Printf("Start sid client: %v", sid)
	for i := 0; i < cfg.Concurrence; i++ {
		ci := i
		go func() {
			e := &exe{
				cs:      cs,
				group:   cfg.Group,
				sid:     uint32(sid),
				verbose: verbose,
			}

			<-time.After(time.Duration(float32(cfg.Interval)*rand.Float32()) * time.Millisecond)
			acid := fmt.Sprintf("%d:%d:%s", cfg.Gid, sid, uuid.NewV4())
			ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
			e.step = stepGetInfo
			for !stop {
				e.cost = 0
				e.timeout = false
				ss := e.step
				attackCount := int(60)
				select {
				case _, ok := <-ticker.C:
					if false == ok {
						return
					}
					switch e.step {
					case stepGetInfo:
						e.worldBossGetInfo(acid)
						e.step = stepJoin
					case stepJoin:
						e.worldBossJoin(acid)
						e.step = stepAttack
						attackCount = rand.Int() % 100
					case stepAttack:
						attackCount--
						if 0 >= attackCount {
							e.step = stepLeave
						}
						e.worldBossAttack(acid)
					case stepLeave:
						e.worldBossLeave(acid)
						e.step = stepGetRank
					case stepGetRank:
						e.worldBossGetRank(acid)
						e.step = stepGetFormationRank
					case stepGetFormationRank:
						e.worldBossGetFormationRank(acid)
						e.step = stepPlayerDetail
					case stepPlayerDetail:
						e.worldBossPlayerDetail(acid)
						e.step = stepGetInfo
					}
				}

				st.list[ci].requestCount++
				st.list[ci].requestCountStep[ss]++
				if e.timeout {
					st.list[ci].requestTimeout++
					st.list[ci].requestTimeoutStep[ss]++
				}
				st.list[ci].requestCostSum[ss] += e.cost
				st.list[ci].requestCostLast[ss][st.list[ci].requestCostLastIndex[ss]] = e.cost
				st.list[ci].requestCostLastIndex[ss]++
				if st.list[ci].requestCostLastIndex[ss] >= costLastLen {
					st.list[ci].requestCostLastIndex[ss] = 0
				}
			}

		}()
	}

	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	<-signalChan
	log.Printf("Stop sid client: %v", sid)
	cs.Stop()
}

//Config ..
type Config struct {
	Gid            uint
	Group          uint32
	ServerAddr     string
	Interval       int
	Concurrence    int
	DiscoverAddr   string
	DiscoverDB     int
	DiscoverAuth   string
	ClientNum      int
	StaticInterval int
}

type exe struct {
	cs      *client.Client
	sid     uint32
	group   uint32
	cost    time.Duration
	timeout bool
	step    int

	verbose bool
}

type static struct {
	list []*staticElem
}

type staticElem struct {
	requestCount         int
	requestCountStep     [stepMax]int
	requestCostSum       [stepMax]time.Duration
	requestCostLast      [stepMax][costLastLen]time.Duration
	requestCostLastIndex [stepMax]int
	requestTimeout       int
	requestTimeoutStep   [stepMax]int
}

const (
	costLastLen = 5
)

const (
	stepGetInfo = iota
	stepJoin
	stepAttack
	stepLeave
	stepGetRank
	stepGetFormationRank
	stepPlayerDetail
	stepMax
)

func (e *exe) worldBossGetInfo(acid string) {
	param := &worldboss.ParamGetInfo{
		Sid:  e.sid,
		Acid: acid,
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodGetInfoID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossGetInfo CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossGetInfo Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetGetInfo))
	}
}

func (e *exe) worldBossJoin(acid string) {
	param := &worldboss.ParamJoin{
		Sid:  e.sid,
		Acid: acid,
		Player: worldboss.PlayerInfo{
			Acid: acid,
			Sid:  e.sid,
			Name: acid,
		},
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodJoinID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossJoin CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossJoin Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetJoin))
	}
}

var bossLevel uint32 = 1

func (e *exe) worldBossAttack(acid string) {
	param := &worldboss.ParamAttack{
		Sid:  e.sid,
		Acid: acid,
		Attack: worldboss.AttackInfo{
			Damage: uint64(rand.Uint32() % 10240),
			Level:  bossLevel,
		},
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodAttackID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossAttack CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossAttack Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetAttack))
	}
	if nil != ret {
		bossLevel = ret.(*worldboss.RetAttack).Boss.Level
	}
}

func (e *exe) worldBossLeave(acid string) {
	param := &worldboss.ParamLeave{
		Sid:  e.sid,
		Acid: acid,
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodLeaveID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossLeave CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossLeave Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetLeave))
	}
}

func (e *exe) worldBossGetRank(acid string) {
	param := &worldboss.ParamGetRank{
		Sid:  e.sid,
		Acid: acid,
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodGetRankID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossGetRank CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossGetRank Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetGetRank))
	}
}

func (e *exe) worldBossGetFormationRank(acid string) {
	param := &worldboss.ParamGetFormationRank{
		Sid:  e.sid,
		Acid: acid,
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodGetFormationRankID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossGetFormationRank CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossGetFormationRank Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetGetFormationRank))
	}
}

func (e *exe) worldBossPlayerDetail(acid string) {
	param := &worldboss.ParamPlayerDetail{
		Sid:  e.sid,
		Acid: acid,
	}
	start := time.Now()
	ret, ec, err := e.cs.CallSync(e.group, worldboss.ModuleID, worldboss.MethodPlayerDetailID, acid, param)
	if nil != err && message.ErrCodeClosed != ec && message.ErrCodeTimeout != ec && connect.ErrTimeout != err {
		log.Printf("worldBossPlayerDetail CallSync failed %d, %v", ec, err)
		return
	}
	if message.ErrCodeTimeout == ec || connect.ErrTimeout == err {
		e.timeout = true
	}
	e.cost = time.Now().Sub(start)
	if e.verbose {
		log.Printf("worldBossPlayerDetail Cost %f, Ret %+v", e.cost.Seconds(), ret.(*worldboss.RetPlayerDetail))
	}
}
