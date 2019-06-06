package storehelper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"

	"github.com/cenk/backoff"
	"github.com/rubyist/circuitbreaker"
)

func NewRedisPool(server, password string, dbSeleccted, cap int) redispool.IPool {
	cfg := redispool.Config{
		HostAddr:       server,
		Capacity:       cap,
		MaxCapacity:    10000,
		IdleTimeout:    time.Minute,
		DBSelected:     dbSeleccted,
		DBPwd:          password,
		ConnectTimeout: 200 * time.Millisecond,
		ReadTimeout:    0,
		WriteTimeout:   0,
	}

	return redispool.NewRedisPool(cfg, "redisstorage", true)
}

func PingRedisConn(c redis.Conn) error {
	_, err := c.Do("PING")
	return err

}

func ReBuildRedisConn(pool redispool.IPool, retry_max_count int, retry_time_ms int64) (redispool.RedisPoolConn, bool) {
	return ReBuildRedisConn_backoff(pool, retry_time_ms)
}

func ReBuildRedisConn_my(pool redispool.Pool, retry_max_count int, retry_time_ms int64) (redispool.RedisPoolConn, bool) {
	has_retry_time := 0
	next_wait_time := retry_time_ms

	// 重试
	for ; has_retry_time < retry_max_count; has_retry_time++ {
		logs.Trace("ReConnect %d %d ", has_retry_time, next_wait_time)
		new_conn := pool.Get()
		if new_conn.Err() == nil {
			return new_conn, true
		} else {
			new_conn.Close()
			time.Sleep(time.Duration(next_wait_time) * time.Millisecond)
			next_wait_time = next_wait_time * 2
		}
	}

	return redispool.NilRPConn, false
}

func ReBuildRedisConn_circuitbreaker(pool *redis.Pool, retry_max_count int, retry_time_ms int64) (res_conn *redis.Conn, ok bool) {
	cb := circuit.NewThresholdBreaker(2)

	for {
		if !cb.Ready() {
			continue
		}

		logs.Trace("ReConnect")
		new_conn := pool.Get()

		if new_conn.Err() == nil {
			res_conn = &new_conn
			ok = true
			logs.Trace("ReConnect Success")
			cb.Success()
			break
		} else {
			new_conn.Close()
			logs.Trace("ReConnect Fail")
			cb.Fail()
			continue
		}
	}

	logs.Error("ReConnect Error")

	res_conn = nil
	ok = false
	return
}

func ReBuildRedisConn_backoff(pool redispool.IPool, retry_time_ms int64) (res_conn redispool.RedisPoolConn, ok bool) {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Duration(retry_time_ms) * time.Millisecond
	b.MaxElapsedTime = 1 * time.Minute
	// 以下取默认值
	// DefaultMaxInterval         = 60 * time.Second
	// DefaultMaxElapsedTime      = 15 * time.Minute

	ticker := backoff.NewTicker(b)
	defer ticker.Stop()

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for _ = range ticker.C {
		logs.Warn("ReConnect")
		new_conn := pool.Get()
		if !new_conn.IsNil() && new_conn.Err() == nil {
			res_conn = new_conn
			ok = true
			logs.Warn("ReConnect Success")
			break

		} else {
			logs.Error("ReConnect Fail")
			new_conn.Close()
		}
	}

	if err != nil {
		// Operation has failed.
		res_conn = redispool.NilRPConn
		ok = false
		return
	}

	logs.Error("ReConnect Error")

	return
}

func RedisHM2Json(result interface{}, err error) ([]byte, error) {
	values, err := redis.StringMap(result, err)

	if err != nil {
		return []byte{}, err
	}

	chunk, err := json.Marshal(values)
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("marshal obj failed, obj:%v, err:%v", values, err))
	}

	return chunk, nil
}

func Json2RedisHM(data []byte) (map[string]string, error) {
	value_map := map[string]string{}
	err := json.Unmarshal(data, &value_map)

	if err != nil {
		return nil, err
	}

	return value_map, nil
}

type StoreRedis struct {
	pool        redispool.IPool
	server      string
	password    string
	dbSeleccted int
}

func NewStoreRedis(server, password string, dbSeleccted int) *StoreRedis {
	n := &StoreRedis{
		server:      server,
		password:    password,
		dbSeleccted: dbSeleccted,
	}
	return n
}

func (s *StoreRedis) GetConn() redispool.RedisPoolConn {
	return s.pool.Get()
}

func (s *StoreRedis) Clone() (IStore, error) {
	nn := *s
	n := &nn
	n.Open()
	return n, nil
}

func (s *StoreRedis) Open() error {
	//TODO YZH Redis pool size
	s.pool = NewRedisPool(s.server, s.password, s.dbSeleccted, 1000)
	return nil
}

func (s *StoreRedis) Close() error {
	s.pool.Close()
	return nil
}

func (s *StoreRedis) Put(key string, val []byte, rh ReadHandler) error {
	conn := s.GetConn()
	defer conn.Close()
	_, err := conn.Do("SET", key, val)
	return err
}
func (s *StoreRedis) Get(key string) ([]byte, error) {
	conn := s.GetConn()
	defer conn.Close()
	return redis.Bytes(conn.Do("GET", key))
}
func (s *StoreRedis) Del(key string) error {
	conn := s.GetConn()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (s *StoreRedis) StoreKey(key string) string {
	now_time := time.Now()
	return now_time.Format("2006-1-2") + "|" + key
}

func (s *StoreRedis) RedisKey(key_in_store string) (string, bool) {
	// Redis存储路径
	d := strings.Split(key_in_store, "|")
	if len(d) < 2 {
		logs.Error("key_in_store err : %s in |", key_in_store)
		return "", false
	}
	return d[len(d)-1], true
}
