package redispool

import (
	"fmt"
	"time"

	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Config struct {
	HostAddr       string
	Capacity       int
	MaxCapacity    int
	IdleTimeout    time.Duration
	DBSelected     int
	DBPwd          string
	ConnectTimeout time.Duration //建立链接需要的时间
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

const (
	Default_RedisPool_Capacity = 100
)

//优化Pool，因为Redigo的操作不说thread-safe的
//参考:https://github.com/garyburd/redigo/wiki/FAQ
// https://github.com/garyburd/redigo/wiki/Vitess-Example

// TODO 基于pool.WaitCout() WaitTime() 发展趋势统计调整Capacity
func NewSimpleRedisPool(pooName, redisServer string, dbSeleccted int,
	dbPwd string, devmode bool, capacity int, use_new_pool bool) IPool {
	logs.Info("Gamex DB Setup with %s,  db:%d", redisServer, dbSeleccted)
	cfg := Config{
		HostAddr:       redisServer,
		Capacity:       capacity,
		MaxCapacity:    3000,
		IdleTimeout:    5 * time.Minute,
		DBSelected:     dbSeleccted,
		DBPwd:          dbPwd,
		ConnectTimeout: 2000 * time.Millisecond,
		ReadTimeout:    0,
		WriteTimeout:   0,
	}

	pool := NewRedisPool(cfg, pooName, use_new_pool)
	go poolStatistic(pooName, pool)
	return pool
}

func poolStatistic(poolName string, pool IPool) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("[RedisPoolStatistic] recover error %v", err)
		}
	}()

	NewGauge := func(name string) gm.Gauge {
		return metrics.NewGauge(fmt.Sprintf("%s.RedisPool.%s", poolName, name))
	}
	poolCapacity := NewGauge("Capacity")
	poolAvailable := NewGauge("Available")
	poolMaxCap := NewGauge("MaxCap")
	poolWaitCount := NewGauge("WaitCount")
	poolWaitTime := NewGauge("WaitTime")       // nanosecond
	poolIdleTimeout := NewGauge("IdleTimeout") // nanosecond

	poolStateTick := time.NewTicker(5 * time.Second)
	logs.Trace("poolStateTick Start")
	for {
		select {
		case <-poolStateTick.C:
			if pool.IsClosed() {
				logs.Trace("poolStateTick stop")
				poolStateTick.Stop()
				return
			}
			capacity, available, maxCap, waitCount, waitTime, idleTimeout := pool.Stats()
			poolCapacity.Update(capacity)
			poolAvailable.Update(available)
			poolMaxCap.Update(maxCap)
			poolWaitCount.Update(waitCount)
			poolWaitTime.Update(int64(waitTime))
			poolIdleTimeout.Update(int64(idleTimeout))
		}
	}
}
