package modules

import (
	"fmt"
	"strings"
	"sync"
	"time"

	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

const (
	metrics_prefix = "gamex.modules."
)

var (
	db_counters = make(map[string]gm.Counter, 16)
	mutex       sync.RWMutex

	db_gauge    = make(map[string]gm.Gauge, 16)
	gague_mutex sync.RWMutex
)

func DoWraper(typ string, db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	addCounter(typ)
	if db.IsNil() {
		return nil, fmt.Errorf("ModuleDBType %s, db is nil by DoWraper", typ)
	}
	defer TimeTracker(typ, time.Now())
	return db.Do(commandName, args...)
}

func DoCmdBufferWrapper(typ string, db redispool.RedisPoolConn, cb redis.CmdBuffer, transaction bool) (reply interface{}, err error) {
	addCounter(typ)
	if db.IsNil() {
		return nil, fmt.Errorf("ModuleDBType %d, db is nil by DoCmdBufferWrapper", typ)
	}
	defer TimeTracker(typ, time.Now())
	return db.DoCmdBuffer(cb, transaction)
}

func addCounter(typ string) {
	mutex.RLock()
	counter, ok := db_counters[typ]
	mutex.RUnlock()
	if !ok {
		mutex.Lock()
		if _, ok := db_counters[typ]; !ok {
			logs.Info("new counter %s", getPrefix()+"."+metrics_prefix+typ)
			counter = metrics.NewCustomCounter(getPrefix() + "." + metrics_prefix + typ)
			db_counters[typ] = counter
		}
		mutex.Unlock()
	}
	mutex.RLock()
	defer mutex.RUnlock()
	counter = db_counters[typ]
	counter.Inc(1)
}

func TimeTracker(typ string, t time.Time) {
	UpdateGaugebyTime(typ, time.Now().Sub(t).Nanoseconds())
}

func getPrefix() string {
	return "r" + strings.Replace(strings.Split(game.Cfg.Redis, ":")[0], ".", "_", -1)
}

func UpdateGaugebyTime(typ string, value int64) {
	updateGauge(fmt.Sprintf(getPrefix()+".gamex.modules_time.%s", typ), value)
}

func updateGauge(name string, value int64) {
	gague_mutex.RLock()
	gauge, ok := db_gauge[name]
	gague_mutex.RUnlock()
	if !ok {
		gague_mutex.Lock()
		if _, ok := db_gauge[name]; !ok {
			logs.Info("new gauge %s", name)
			gauge = metrics.NewCustomGauge(name)
			db_gauge[name] = gauge
		}
		gague_mutex.Unlock()
	}
	gague_mutex.RLock()
	defer gague_mutex.RUnlock()
	gauge = db_gauge[name]

	gauge.Update(value)
}

func UpdateGaugeByGate(value int64) {
	updateGauge(getPrefix()+"."+metrics_prefix+"gatecount", value)
}
