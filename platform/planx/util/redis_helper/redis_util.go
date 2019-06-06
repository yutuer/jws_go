package redis_helper

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type redisKeysHander func(keys []string, values []string) error

type redisScanner struct {
	conn     redis.Conn
	last_idx string
	key      string

	hander redisKeysHander
}

func newScanner(conn redis.Conn, key string, hander redisKeysHander) *redisScanner {
	return &redisScanner{
		conn:     conn,
		last_idx: "",
		hander:   hander,
		key:      key,
	}
}

func (r *redisScanner) Start() {
	r.last_idx = "0"
}

func (r *redisScanner) Next() error {
	res, err := redis.Values(r.conn.Do("HSCAN", r.key, r.last_idx, "count", 100))
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
	values_strs := make([]string, 0, len(keys))
	if len(keys)%2 != 0 {
		return errors.New("keys Len not 2")
	}

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		k, ok := key.([]byte)
		if !ok {
			logs.Error("key Data %v Error", key)
			continue
		}
		//logs.Trace("res Key %d : %v", i, string(k))
		if i%2 == 0 {
			key_strs = append(key_strs, string(k))
		} else {
			values_strs = append(values_strs, string(k))
		}
	}

	return r.hander(key_strs, values_strs)
}

func (r *redisScanner) IsScanOver() bool {
	return r.last_idx == "0"
}

func HScan(conn redis.Conn, key string, f redisKeysHander) error {
	s := newScanner(conn, key, f)
	s.Start()
	for {
		err := s.Next()
		if err != nil {
			return err
		}
		if s.IsScanOver() {
			return nil
		}
	}

	return nil
}
