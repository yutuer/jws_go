package hour_log

import (
	"bufio"
	"time"

	"fmt"
	"os"
	"path/filepath"

	"sync"

	"github.com/gin-gonic/gin"
	gm "github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/metrics"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	首先分渠道
	每天0点到当前整点的累积值
		registerCount: 新注册数量
		deviceCount: 新注册里新设备数量
		activeCount: 活跃用户数量
		chargeAcidCount: 付费用户数量
		chargeSum: 付费总金额
	每小时统计值, 每一分钟统计一次
		pcu
		acu
*/
const (
	_ = iota
	Cmd_Login
	Cmd_Register
	Cmd_Active
	Cmd_Pay
	Cmd_Debug_log_output
)
const (
	db_counter_key = "hourlog_db"
	path           = "/opt/supervisor/log/hourlybi"
	log_profix     = "134_"
	gameId         = "134"
)

var (
	channelCCUCounter map[string]gm.Counter
	mutxChannelCCU    sync.RWMutex
)

func init() {
	channelCCUCounter = make(map[string]gm.Counter, 10)
}

type HourLog struct {
	sid      uint
	cmd_chan chan hourCmd
	waitter  util.WaitGroupWrapper

	CurDayTimeBegin int64
	LastLogHourTime int64
}

func NewHourLog(sid uint) *HourLog {
	m := new(HourLog)
	m.sid = sid
	return m
}

func (hl *HourLog) Start() {
	hl.dbLoad(hl.sid)
	hl.cmd_chan = make(chan hourCmd, 1024)
	hl.preCheck()

	hl.waitter.Wrap(func() {
		timerChan := uutil.TimerSec.After(time.Second)
		ccuRefTick := 0
		for {
			select {
			case command, ok := <-hl.cmd_chan:
				if !ok {
					logs.Warn("HourLog cmdChan already closed")
					return
				}
				func(cmd hourCmd) {
					//by YZH 这个让parent never dead, 应该如此吗？
					defer logs.PanicCatcherWithInfo("HourLog Worker Panic")
					hl.preCheck()

					hl.processCmd(&cmd)
				}(command)
			case <-timerChan:
				hl.preCheck()

				ccuRefTick++
				if ccuRefTick >= util.MinSec {
					hl.onCCU()
					ccuRefTick = 0
				}
				timerChan = uutil.TimerSec.After(time.Second)
			}
		}
	})
}

func (hl *HourLog) AfterStart(g *gin.Engine) {

}

func (hl *HourLog) BeforeStop() {

}

func (hl *HourLog) Stop() {
	close(hl.cmd_chan)

	t := time.Now().In(logiclog.GetBILogTL())
	hl.logCheck(t, true)

	hl.dbSave(hl.sid)
	hl.waitter.Wait()
}

type hourCmd struct {
	Typ     int
	Acid    string
	Channel string
	Device  string
	Money   int
}

func (hl *HourLog) preCheck() {
	t := time.Now().In(logiclog.GetBILogTL())

	hl.logCheck(t, false)
	hl.clearCheck(t)

	hl.dbSave(hl.sid)
}

func (hl *HourLog) processCmd(cmd *hourCmd) {
	switch cmd.Typ {
	case Cmd_Login:
		hl.onLogin(cmd)
	case Cmd_Register:
		hl.onRegister(cmd)
	case Cmd_Active:
		hl.onActive(cmd)
	case Cmd_Pay:
		hl.onPay(cmd)
	case Cmd_Debug_log_output:
		t := time.Now().In(logiclog.GetBILogTL())
		hl.onCCU()
		hl.logCheck(t, true)
	}
}

func (hl *HourLog) OnLogin(acid, channel, device string) {
	hl.CommandExec(hourCmd{
		Typ:     Cmd_Login,
		Acid:    acid,
		Channel: channel,
		Device:  device,
	})
}

func (hl *HourLog) OnRegister(acid, channel, device string) {
	hl.CommandExec(hourCmd{
		Typ:     Cmd_Register,
		Acid:    acid,
		Channel: channel,
		Device:  device,
	})
}

func (hl *HourLog) OnActive(acid, channel string) {
	hl.CommandExec(hourCmd{
		Typ:     Cmd_Active,
		Acid:    acid,
		Channel: channel,
	})
}

func (hl *HourLog) OnPay(acid, channel string, money int) {
	hl.CommandExec(hourCmd{
		Typ:     Cmd_Pay,
		Acid:    acid,
		Channel: channel,
		Money:   money,
	})
}

func (hl *HourLog) DebugOutput() {
	hl.CommandExec(hourCmd{
		Typ: Cmd_Debug_log_output,
	})
}

func (hl *HourLog) CommandExec(cmd hourCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	chann := hl.cmd_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("HourLog CommandExec chann full, cmd put timeout")
	}
}

func (hl *HourLog) writeLog(tu int64, ds map[string]*logData) {
	for k, v := range ds {
		logs.Debug("HourLog want writeLog %s %v", k, *v)
	}

	logs.Debug("HourLog HourLogValid %v", game.Cfg.HourLogValid)
	if !game.Cfg.HourLogValid {
		return
	}

	// 每小时log，不管服务器什么时区，都是东8区
	t := time.Unix(tu, 0).In(logiclog.GetBILogTL())
	log_file_name := fmt.Sprintf("%s%d_%d-%02d-%02d_%02d.log", log_profix, hl.sid,
		t.Year(), t.Month(), t.Day(), t.Hour())
	ppath := "."
	if uutil.PathExists(path) {
		ppath = path
	}
	fname := filepath.Join(ppath, log_file_name)
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("HourLog writeLog OpenFile err %s", err.Error())
		return
	}

	bufWriter := bufio.NewWriterSize(file, 1024)

	for channel, data := range ds {
		gid := game.Cfg.Gid
		strChan := game.Gid2Channel[gid]
		gidSid := fmt.Sprintf("%s%04s%06d", strChan, gameId, hl.sid)
		line := fmt.Sprintf("%s$$%s$$%d$$%d$$%d$$%d$$%d$$%d$$%d$$%d$$%d\n",
			gidSid, channel, data.chargeSum, data.chargeAcid,
			data.active, data.active-data.register,
			data.register, data.register, data.device,
			data.pcu, data.acu)
		logs.Debug("HourLog Write %s", line)
		bufWriter.Write([]byte(line))
	}
	bufWriter.Flush()
	file.Close()
}

func (hl *HourLog) dbLoad(shardId uint) error {
	_db := modules.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), TableBIHourModule(shardId), hl, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (hl *HourLog) dbSave(shardId uint) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb, TableBIHourModule(shardId), hl); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := modules.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := metricsModules.DoCmdBufferWrapper(db_counter_key, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}

func GetCCU() map[string]int {
	mutxChannelCCU.RLock()
	res := make(map[string]int, len(channelCCUCounter))
	for ch, c := range channelCCUCounter {
		res[ch] = int(c.Count())
	}
	mutxChannelCCU.RUnlock()
	return res
}

func AddCCU(channel string) {
	if channel == "" {
		return
	}
	mutxChannelCCU.RLock()
	c, ok := channelCCUCounter[channel]
	if ok {
		mutxChannelCCU.RUnlock()
		c.Inc(1)
		return
	}
	mutxChannelCCU.RUnlock()

	mutxChannelCCU.Lock()
	c, ok = channelCCUCounter[channel]
	if !ok {
		cn := channelCCUName(channel)
		c = metrics.NewCounter(cn)
		channelCCUCounter[channel] = c
		logs.Info("new metrics %s", cn)
	}
	c.Inc(1)
	mutxChannelCCU.Unlock()
}

func DelCCU(channel string) {
	if channel == "" {
		return
	}
	mutxChannelCCU.RLock()
	defer mutxChannelCCU.RUnlock()
	c, ok := channelCCUCounter[channel]
	if ok {
		c.Dec(1)
	}
}

func channelCCUName(channel string) string {
	return fmt.Sprintf("gamex.%s.ccu", channel)
}

func GetHourEndTS() int64 {
	t := time.Now().In(logiclog.GetBILogTL())
	nt_tl := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 0, logiclog.GetBILogTL())
	return nt_tl.Unix()
}
