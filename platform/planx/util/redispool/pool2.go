package redispool

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

/*
	使用redigo/redis，多线程安全，redis连接池，可增可减
*/

type Pool2 struct {
	Name    string
	resPool *redis.Pool
	conf    Config
}

func (p *Pool2) IsNil() bool {
	return p.resPool == nil
}

func (p *Pool2) IsClosed() bool {
	return p.resPool == nil
}

func (p *Pool2) Close() {
	if p.IsNil() {
		return
	}
	p.resPool.Close()
	p.resPool = nil
}

func (p *Pool2) Get() RedisPoolConn {
	return p.GetDBConn()
}

func (p *Pool2) GetDBConn() RedisPoolConn {
	return RedisPoolConn{VrConn{p.resPool.Get()}, p}
}

func (p *Pool2) CloseConn(c RedisPoolConn) {
	c.VrConn.Close()
}

func (p *Pool2) Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration) {
	return int64(p.resPool.ActiveCount()), 0, 0, 0, 0, 0
}
