package redispool

import (
	"vcs.taiyouxi.net/platform/planx/youtube/vitess/pools"
	//"github.com/youtube/vitess/go/pools"

	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type IPool interface {
	IsNil() bool
	IsClosed() bool
	Close()
	Get() RedisPoolConn
	GetDBConn() RedisPoolConn
	CloseConn(RedisPoolConn)
	Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration)
}

type RedisPoolConn struct {
	VrConn
	p IPool
}

var NilRPConn = RedisPoolConn{VrConn: NilConn}

func (rpc RedisPoolConn) RawConn() redis.Conn {
	return rpc.VrConn.RawConn()
}

func (rpc RedisPoolConn) Close() {
	if rpc.IsNil() {
		logs.Warn("RedisPoolConn send Nil conn back to pool[1] %v", rpc)
		return
	}
	if rpc.p == nil {
		logs.Warn("RedisPoolConn send Nil conn back to pool[2] %v", rpc)
		return
	}
	rpc.p.CloseConn(rpc)
}

// ResourceConn adapts a Redigo connection to a Vitess Resource.
type VrConn struct {
	redis.Conn
}

var NilConn = VrConn{Conn: nil}

func (r VrConn) IsNil() bool {
	return r.Conn == nil
}

func (r VrConn) Close() {
	if r.IsNil() {
		return
	}
	r.Conn.Close()
}

func (r VrConn) RawConn() redis.Conn {
	return r.Conn
}

func NewRedisPool(cfg Config, name string, use_new_pool bool) (p IPool) {
	if use_new_pool {
		p = &Pool2{
			Name: name,
			conf: cfg,
			resPool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					factorycfg := cfg
					c, err := createRedisConn(factorycfg)
					if err != nil {
						return nil, err
					}
					return c, err
				},
				MaxIdle:     cfg.Capacity,    //Capacity
				MaxActive:   cfg.MaxCapacity, //MaxCapacity
				IdleTimeout: cfg.IdleTimeout, //idleTimeout
				Wait:        true,
			},
		}
		logs.Warn("New RedisPool Valid")
	} else {
		pool := &Pool{
			Name:     name,
			conf:     cfg,
			QuitChan: make(chan bool, 1),
			resPool: pools.NewResourcePool(
				func() (pools.Resource, error) {
					factorycfg := cfg
					c, err := createRedisConn(factorycfg)
					if err != nil {
						return nil, err
					}
					return VrConn{
						Conn: c,
					}, err
				},
				cfg.Capacity,    //Capacity
				cfg.MaxCapacity, //MaxCapacity
				cfg.IdleTimeout, //idleTimeout
			)}
		pool.poolCheckConn()
		p = pool
		logs.Warn("Old RedisPool Valid")
	}

	return p
}

func createRedisConn(cfg Config) (redis.Conn, error) {
	c, err := redis.Dial("tcp", cfg.HostAddr,
		redis.DialConnectTimeout(cfg.ConnectTimeout),
		redis.DialReadTimeout(cfg.ReadTimeout),
		redis.DialWriteTimeout(cfg.WriteTimeout),
		redis.DialDatabase(cfg.DBSelected),
		redis.DialPassword(cfg.DBPwd),
	)
	return c, err
}
