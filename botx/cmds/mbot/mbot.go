package mbot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	//"time"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/botx/bot"
	"vcs.taiyouxi.net/botx/cmds"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	//"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

func init() {

	logs.Trace("MBot cmd loaded")
	cmds.Register(&cli.Command{
		Name:        "mbot",
		Usage:       "mbot run as web server",
		Description: "run many bots in the same time",
		Action:      BotStart,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "rpc, r",
				Value: "127.0.0.1:8668",
				Usage: "gate server rpc ip:port",
			},
			cli.StringFlag{
				Name:  "server",
				Value: "127.0.0.1:8667",
				Usage: "gate server ip:port",
			},
			cli.BoolFlag{
				Name:  "onlymsg",
				Usage: "Ignore all session related msg",
			},
			cli.Float64Flag{
				Name:  "speed, s",
				Value: 1.0,
				Usage: "2.0 2xtimes faster",
			},
		},
	})
}

//单个机器人模式下，数据是流式读取, 解析成LogEntry然后发送给机器人。
//在单机模拟很多文件模式的情况下，需要一次性读取然后关闭FD，然后理由io.Reader进行重复利用。
//这样就能够大量节省实验机器的FD消耗。

type BotMaker struct {
	sync.WaitGroup
	Identify string
	rawBytes []byte

	nLiveBots uint32
	nCCU      uint32
}

func (bm *BotMaker) Up() {
	atomic.AddUint32(&bm.nLiveBots, 1)
	bm.Add(1)
}

func (bm *BotMaker) Down() {
	atomic.AddUint32(&bm.nLiveBots, ^uint32(0))
	bm.Done()
}

func (bm *BotMaker) CCUAdd() {
	atomic.AddUint32(&bm.nCCU, 1)
	bm.Add(1)
}

func (bm *BotMaker) CCUSub() {
	atomic.AddUint32(&bm.nCCU, ^uint32(0))
	bm.Done()
}

func (bm *BotMaker) GetNumberLiveBots() uint32 {
	return bm.nLiveBots
}

func (bm *BotMaker) GetCCU() uint32 {
	return bm.nCCU
}

type BotMakeInfo struct {
	Identify string
	Account  db.Account
	Server   string
	Rpc      string
	Speed    float64
}

func (bm *BotMaker) RunABot(bmi BotMakeInfo, be bot.BotRpsEvent) *bot.PlayerBot {
	speed := bmi.Speed

	mybot := bot.NewPlayerBot(
		bmi.Account.String(),
		bmi.Server,
		bmi.Rpc,
		speed,
	)
	if mybot == nil {
		logs.Error("mybot create failed!")
		return nil
	}

	r := bytes.NewReader(bm.rawBytes)
	go mybot.Run(r, bm, be)
	return mybot

}

const BOT_MAKE_INFO_CHAN_BUF_LEN = 1024

type BotFactory struct {
	mu     sync.RWMutex
	Makers map[string]*BotMaker

	speedParam float64

	mulog      sync.RWMutex
	replaylogs []string

	nextAccount db.Account
	server      string
	rpc         string

	makerChan chan BotMakeInfo
}

func NewBotFactory(baseAccount db.Account, server, rpc string) *BotFactory {
	bf := &BotFactory{
		Makers:      make(map[string]*BotMaker),
		makerChan:   make(chan BotMakeInfo, BOT_MAKE_INFO_CHAN_BUF_LEN),
		nextAccount: baseAccount,
		server:      server,
		rpc:         rpc,
		replaylogs:  make([]string, 0, 100),
	}
	go bf.gen()
	return bf
}

//GetLogs处理的并不严谨，这里都在架设使用者不会修改里面的值。
//但是作为一个测试用框架，这里暂时不会去修正
func (bf *BotFactory) GetLogs() []string {
	return bf.replaylogs
}
func (bf *BotFactory) AddLogFile(id string) {
	bf.mulog.Lock()
	defer bf.mulog.Unlock()
	bf.replaylogs = append(bf.replaylogs, id)
}

func (bf *BotFactory) SimpleGenerator(id string, number int) {
	rand.Seed(time.Now().UnixNano())
	logs.Info("SimpleGenerator run with number %d.", number)
	for i := 0; i < number; i++ {
		sp := speed
		if speed <= 0 {
			sp = float64(rand.Int63n(5) + 5)
		}
		bmi := BotMakeInfo{
			Identify: id,
			Server:   bf.server,
			Rpc:      bf.rpc,
			Speed:    sp,
		}
		bf.makerChan <- bmi
		slp := rand.Intn(50)
		time.Sleep(time.Duration(slp) * time.Millisecond)

	}
	logs.Info("SimpleGenerator has finished.")
}

//RandomGenerator sleep: 随机分批生成的每次休息间隔
func (bf *BotFactory) RandomGenerator(idlist []string, number, once int, sleep int) {
	nlogfiles := len(idlist)
	halfnum := once
	rand.Seed(time.Now().UnixNano())
	if sleep <= 0 {
		sleep = 1
	}
	for {
		ididx := rand.Intn(nlogfiles)
		nbots := rand.Intn(halfnum)
		if nbots == 0 {
			continue
		}

		if nbots > number {
			nbots = number
		}
		bf.SimpleGenerator(idlist[ididx], nbots)
		logs.Trace("RandomGenerator:%s: %d - %d", idlist[ididx], number, nbots)
		number -= nbots
		if number <= 0 {
			break
		}
		time.Sleep(time.Second * time.Duration(sleep))
	}
	logs.Info("RandomGenerator has finished.")
}

func (bf *BotFactory) gen() {
	for bmi := range bf.makerChan {
		id := bmi.Identify

		bmi.Account = bf.nextAccount
		bf.nextAccount.UserId = db.NewUserID()

		mker, err := bf.getMaker(id)
		if err != nil {
			logs.Error("bot %s start error, %s", bmi.Account.String(), err.Error())
			continue
		}

		mker.RunABot(bmi, nil)
	}
}

func (bf *BotFactory) getMaker(id string) (*BotMaker, error) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if m, ok := bf.Makers[id]; ok {
		return m, nil
	} else {
		rawBytes, err := ioutil.ReadFile(id)
		if err != nil {
			return nil, err
		}
		bm := &BotMaker{
			Identify: id,
			rawBytes: rawBytes}
		bf.Makers[id] = bm
		return bm, nil
	}
}

func (bf *BotFactory) GetNumberLiveBots() uint32 {
	var sum uint32
	for _, bm := range bf.Makers {
		sum += bm.GetNumberLiveBots()
	}
	return sum
}

func (bf *BotFactory) GetCCU() uint32 {
	var sum uint32
	for _, bm := range bf.Makers {
		sum += bm.GetCCU()
	}
	return sum
}

//func (bf *BotFactory) Raise(
//id_gen func() string,
//number uint32,
//strategy_gen func(uint32) func() (int, time.Duration),
//) {
//strategy := strategy_gen(number)
//for {
//num, sleep := strategy()
//for i := 0; i < num; i++ {
//id := id_gen()
//mker := bf.GetMaker(id)
////mker.RunABot()
//}
//time.Sleep(sleep)

//}

//}
var speed float64

func BotStart(c *cli.Context) {
	defer logs.Flush()
	baseAccount, err := db.ParseAccount("0:0:87d5f092-7bb7-44a5-9869-b42fd9bf5858")
	if err != nil {
		logs.Error("Account flag can not be parsed!, %s", err.Error())
		bot.OSExit(-1)
	}

	if !c.Args().Present() {
		logs.Error("At least one argument is needed for replay file folder!")
		bot.OSExit(-1)
	}
	bot.OnlyMsg = c.Bool("onlymsg")
	speed = c.Float64("speed")
	bot_factory := NewBotFactory(baseAccount, c.String("server"), c.String("rpc"))
	bot_factory.speedParam = c.Float64("speed")

	botpath := fmt.Sprintf("./%s/*.log", c.Args().First())
	botfiles, err := filepath.Glob(botpath)
	if err != nil {
		logs.Error("filepath Glob Error: %s", err.Error())
		bot.OSExit(-4)
	}
	if len(botfiles) > 0 {
		for _, f := range botfiles {
			logs.Info("botfile added: %s", f)
			bot_factory.AddLogFile(f)
		}
		api(bot_factory)
	} else {
		logs.Error("No enough log files found.")
	}
}
