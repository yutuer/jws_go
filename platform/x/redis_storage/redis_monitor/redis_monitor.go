package redis_monitor

import (
	"sync"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/x/redis_storage/onland"
	util2 "vcs.taiyouxi.net/platform/x/redis_storage/util"
)

const (
	MONITOR_STATE_Monitoring = iota
	MONITOR_STATE_Offline
)

type RedisPubMonitor struct {
	psc              *redis.PubSubConn
	onlander         *onland.OnlandStores
	wg               sync.WaitGroup
	db_num           int
	need_store_heads map[string]bool

	pool  redispool.IPool
	state int
}

func (r *RedisPubMonitor) Init(
	pool redispool.IPool,
	onlander *onland.OnlandStores,
	db_num int,
	need_store_heads map[string]bool) {

	r.pool = pool
	r.db_num = db_num
	r.need_store_heads = need_store_heads

	conn := pool.Get()
	// if conn.IsNil() {
	//     logs.Critical("RedisPubMonitor can't build connection with redis!")
	//     panic("RedisPubMonitor can't build connection with redis!")
	// }
	err := r.startPsc(conn.RawConn())
	if err != nil {
		logs.Error("startPsc Err %s", err.Error())
		r.state = MONITOR_STATE_Offline
	} else {
		r.state = MONITOR_STATE_Monitoring
	}

	r.onlander = onlander

	r.wg.Add(1)
}

func (r *RedisPubMonitor) Start() {
	defer r.wg.Done()
	for {
		// 先处理一下是不是连接成功
		if r.state != MONITOR_STATE_Monitoring {
			err := r.retry()
			if err != nil {
				logs.Error("Connect Redis Err, will retry later %s",
					err.Error())
				continue
			}
			r.state = MONITOR_STATE_Monitoring
		}

		switch n := r.psc.Receive().(type) {
		case redis.Message:
			logs.Trace("Message: %s %s", n.Channel, n.Data)
			r.Onland(string(n.Data))

		case redis.PMessage:
			logs.Trace("PMessage: %s %s", n.Channel, n.Data)
			r.Onland(string(n.Data))

		case redis.Subscription:
			logs.Trace("Subscription: %s %s %d", n.Kind, n.Channel, n.Count)
			if n.Count == 0 {
				return
			}
		case error:
			logs.Error("error: %v", n)
			r.state = MONITOR_STATE_Offline
			continue
		}
	}
}

func (r *RedisPubMonitor) Stop() {
	logs.Info("RedisPubMonitor stop")
	r.stopPsc()

	r.wg.Wait()
	r.psc.Close()

	r.onlander.CloseJobQueue()
}

func (r *RedisPubMonitor) Onland(key string) {
	if !r.isNeedOnLand(key) {
		return
	}

	r.wg.Add(1)
	defer r.wg.Done()

	r.onlander.NewKeyDumpJob(key)
}

func (r *RedisPubMonitor) isNeedOnLand(key string) bool {
	return util2.IsNeedOnLand(r.need_store_heads, key)
}
