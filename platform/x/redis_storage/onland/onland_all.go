package onland

import (
	"errors"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//
// Redis遍历类
//

type RedisKeysHander func(keys []string) error

type redisScanner struct {
	conn     redis.Conn
	last_idx string

	hander RedisKeysHander
}

func NewScanner(conn redis.Conn, hander RedisKeysHander) *redisScanner {
	return &redisScanner{
		conn:     conn,
		last_idx: "",
		hander:   hander,
	}
}

func (r *redisScanner) Start() error {
	r.last_idx = "0"
	return r.Next()
}

func (r *redisScanner) Next() error {
	res, err := redis.Values(r.conn.Do("SCAN", r.last_idx, "count", 100))
	if err != nil {
		return err
	}

	if len(res) < 2 {
		return errors.New("Scan Res Len less 2")
	}

	ins, ok := res[0].([]byte)
	if !ok {
		return errors.New("res[0] No []byte")
	}

	//logs.Trace("res[0] : %v", string(ins))

	keys, ok := res[1].([]interface{})
	if !ok {
		return errors.New("res[1] No []interface{}")
	}

	r.last_idx = string(ins)

	key_strs := make([]string, 0, len(keys))

	for _, key := range keys {
		k, ok := key.([]byte)
		if !ok {
			logs.Error("key Data %v Error", key)
			continue
		}
		//logs.Trace("res Key %d : %v", i, string(k))
		key_strs = append(key_strs, string(k))
	}

	return r.hander(key_strs)
}

func (r *redisScanner) IsScanOver() bool {
	return r.last_idx == "0"
}

//
// 全量落地逻辑
//

func (o *OnlandStores) OnlandAll() (string, error) {
	conn := o.redis_source_pool.Get()
	defer conn.Close()

	if conn.IsNil() {
		return fmt.Sprintf("OnlandAll DB is nil"),
			fmt.Errorf("OnlandAll DB is nil")
	}
	err, all, num := o.dumpAllKeys(conn.RawConn())

	// 通知失败不重联
	if err == nil {
		return fmt.Sprintf("Sync All: keys %d, Dump %d.", all, num), nil
	} else {
		return fmt.Sprintf("Sync Err: By %s.", err.Error()), err
	}
}

func (o *OnlandStores) dumpAllKeys(conn redis.Conn) (error, int, int) {
	all := 0
	num := 0

	hander := func(keys []string) error {
		//logs.Trace("keys hander")
		for i, key := range keys {
			all++
			if num%1000 == 0 {
				logs.Trace("key has dump %d num %d", i, num)
			}
			if o.isNeedOnLand(key) {
				num++
				o.NewKeyDumpJob(key)
			}

		}
		return nil
	}

	scanner := NewScanner(conn, hander)
	if err := scanner.Start(); err != nil {
		logs.Error("scanner start err %v", err)
	}
	for !scanner.IsScanOver() {
		err := scanner.Next()
		if err != nil {
			logs.Error("scanner next err %v", err)
			return err, 0, 0
		}
	}

	return nil, all, num

}
