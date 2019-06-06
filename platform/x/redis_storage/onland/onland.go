package onland

import (
	// "encoding/base64"
	"fmt"
	"sync"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/redis_storage/util"
)

type RedisDumpHandler func(con redispool.RedisPoolConn, key string) ([]byte, error)

type OnlandStores struct {
	land              []storehelper.IStore
	key_queue         chan string
	redis_source_pool redispool.IPool
	source            storehelper.IStore
	wg                sync.WaitGroup
	wg_key_queue      sync.WaitGroup

	redisDumpFun RedisDumpHandler

	need_store_heads map[string]bool

	workers_num int

	need_log_size bool

	format string
	seq    string
}

func (o *OnlandStores) Init(
	pool redispool.IPool,
	need_store_heads map[string]bool,
	store_mode string,
	workers_num int,
	need_log_size bool) {

	o.redis_source_pool = pool
	o.land = []storehelper.IStore{}
	o.wg.Add(1)
	o.key_queue = make(chan string, workers_num*16)
	o.need_store_heads = need_store_heads
	o.workers_num = workers_num
	o.need_log_size = need_log_size

	switch store_mode {
	case "HGETALL":
		o.redisDumpFun = readDataToOnLand
	case "DUMP":
		o.redisDumpFun = dumpRedis
	}

}

func (o *OnlandStores) Add(s storehelper.IStore) {
	logs.Info("new store %v", s)

	//		ns, err := s.Clone()
	//		if err != nil {
	//		logs.Error("open store err %v", err)
	//	} else {
	o.land = append(o.land, s)
	//	}
}

func dump(key string, val []byte, stores []storehelper.IStore, rh storehelper.ReadHandler) {

	for i := 0; i < len(stores); i++ {
		if stores[i] == nil {
			logs.Error("land nil %d", i)
			continue
		}

		key_store := stores[i].StoreKey(key)

		//logs.Info("Dump %v To %s", key, key_store)
		//TODO 这个Del是必须的吗？ by YZH
		stores[i].Del(key_store)
		err := stores[i].Put(key_store, val, rh)
		if err != nil {
			logs.Error("Onland Dump Err: %v", err.Error())
		}
	}
}

func (o *OnlandStores) close(stores []storehelper.IStore) {
	for i := 0; i < len(stores); i++ {
		if stores[i] == nil {
			logs.Error("land nil %d", i)
			continue
		}
		stores[i].Close()
	}
}

func (o *OnlandStores) NewKeyDumpJob(key string) {
	o.wg_key_queue.Add(1)
	o.key_queue <- key
}

func (o *OnlandStores) CloseJobQueue() {
	// 这回引起OnlandStores的退出
	close(o.key_queue)
}

func (o *OnlandStores) Start() {
	defer o.wg.Done()

	for i := 0; i < o.workers_num; i++ {
		go func() {
			o.wg.Add(1)
			defer o.wg.Done()

			//每个worker都拿着独立的db hanlder
			//go routine safe + performance
			stores := []storehelper.IStore{}
			for _, s := range o.land {
				b, berr := s.Clone()
				// berr := b.Open()
				if berr != nil {
					logs.Error("Open store err: %v", berr.Error())
					return
				}
				stores = append(stores, b)
			}

			for key := range o.key_queue {
				o.wg_key_queue.Done()
				o.restore(key, stores)
			}
			logs.Info("key queue close")
			o.close(stores)
		}()
	}
}

func (o *OnlandStores) Stop() {
	logs.Info("OnlandStores stop")
	o.wg_key_queue.Wait()
	o.wg.Wait()
}

//TODO YZH OnlandStores restore 可以通过pipeline增加读性能, 但是也要小心存档过大导致传输问题
func (o *OnlandStores) restore(key string, stores []storehelper.IStore) {
	res, ok := o.readData(key)
	if !ok {
		return
	}

	// 包个访问redis的方法进去，给需要二次访问redis的IStore用，不用的IStore可以无视
	dump(key, res, stores,
		func(rkey string) ([]byte, bool) {
			return o.readData(rkey)
		})
	// 记录存档大小日志
	if o.need_log_size {
		logSize(key, res)
	}
}

func (o *OnlandStores) readData(key string) ([]byte, bool) {
	conn := o.redis_source_pool.Get()
	defer conn.Close()

	res, err := o.redisDumpFun(conn, key)
	if err != nil {
		conn_re, ok := storehelper.ReBuildRedisConn(
			o.redis_source_pool, 6, 1000)
		if !ok {
			logs.Error("ReBuildRedisConn Err Work %s Cannot Done", key)
			return nil, false
		}
		conn = conn_re
		res, err = o.redisDumpFun(conn, key)
		if err != nil {
			logs.Error("ReBuildRedisConn and o.redisDumpFun Err Work %s Cannot Done, err: %s",
				key, err.Error())
			return nil, false
		}
	}
	return res, true
}

func (o *OnlandStores) isNeedOnLand(key string) bool {
	return util.IsNeedOnLand(o.need_store_heads, key)
}

func logSize(key string, val []byte) {
	// TO YinZeHong 存档大小日志
	// 注意这个函数会在多个协程中调用，如果需要线程安全，可以考虑加锁
	// 这个函数只会在监控模式下调用，这时压力不会很大，加锁不会影响性能
	logs.Info("Test Size Log %s : Size %d", key, len(val))
}

func readDataToOnLand(conn redispool.RedisPoolConn, key_in_redis string) ([]byte, error) {
	if conn.IsNil() {
		return nil, fmt.Errorf("conn is nil now.")
	}
	return storehelper.RedisHM2Json(conn.Do("HGETALL", key_in_redis))
}

func dumpRedis(conn redispool.RedisPoolConn, key_in_redis string) ([]byte, error) {
	if conn.IsNil() {
		return nil, fmt.Errorf("conn is nil now.")
	}
	// vv, er := redis.Bytes(conn.Do("DUMP", key_in_redis))
	// v := base64.StdEncoding.EncodeToString(vv)
	// return []byte(v), er

	return redis.Bytes(conn.Do("DUMP", key_in_redis))

}
