package redispool

import (
	"time"

	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
	"vcs.taiyouxi.net/platform/planx/youtube/vitess/pools"
)

// 这个Pool是老版本
type Pool struct {
	Name     string
	resPool  *pools.ResourcePool
	conf     Config
	QuitChan chan bool
}

func (p *Pool) IsNil() bool {
	return p.resPool == nil
}

func (p *Pool) IsClosed() bool {
	if p.resPool == nil {
		return true
	}
	return p.resPool.IsClosed()
}

func (p *Pool) Close() {
	if p.IsNil() {
		return
	}
	p.QuitChan <- true
	p.resPool.Close()
	p.resPool = nil
}

func (p *Pool) Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration) {
	return p.resPool.Stats()
}

//Get 目前为了兼容Auth, storehelper, gmtools, redis_storage
func (p *Pool) Get() RedisPoolConn {
	return p.GetDBConn()
}

func (p *Pool) GetDBConn() RedisPoolConn {
	return p.getDBConn(true, false)
}

func (p *Pool) CloseConn(c RedisPoolConn) {
	p.resPool.Put(c.VrConn)
}

func (p *Pool) getDBConn(retry, ping bool) RedisPoolConn {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	r, err := p.resPool.Get(ctx)
	if err == pools.ErrTimeout {
		hasValid := false
		mutex.Lock()
		if p.resPool.Available() <= 0 {
			hasValid = p.dynamicAddCapacity()
			logs.Warn("Pool ErrTimeout success dynamicAddCapacity")
		} else {
			hasValid = true
			logs.Warn("Pool ErrTimeout but no need dynamicAddCapacity")
		}
		mutex.Unlock()

		if hasValid {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 600*time.Millisecond)
			defer cancel2()
			r, err = p.resPool.Get(ctx2)
		}
	}
	switch err {
	case nil:
		//normal
	case pools.ErrTimeout:
		logs.Error("GetDBConn get a timeout, err %s", ctx.Err())
		return NilRPConn
	case pools.ErrClosed:
		logs.Error("GetDBConn RedisPoolConn is closed by mistake? err:%s", err)
		return NilRPConn
	default:
		//通常这里是数据库开启链接出现的错误
		logs.Error("GetDBConn error will retry: %s", err.Error())
	}

	if err != nil {
		if retry {
			nc, err := createRedisConn(p.conf)
			if err != nil {
				logs.Error("GetDBConn error and retry still err %s", err.Error())
				return NilRPConn
			}
			return RedisPoolConn{
				VrConn: VrConn{
					Conn: nc,
				},
				p: p,
			}
		} else {
			return NilRPConn
		}
	} else {
		c, ok := r.(VrConn)
		if ok {
			if ping {
				if _, err := c.Do("PING"); err != nil {
					logs.Error("redispool get conn but ping err, replace new one") // TBD 加个log看是否发生，之后删除
					nc, err := createRedisConn(p.conf)
					if err != nil {
						logs.Error("redispool get conn but ping err, try create new redis conn but err %s", err.Error())
						return NilRPConn
					}
					c = VrConn{
						Conn: nc,
					}
				}
			}
			return RedisPoolConn{
				VrConn: c,
				p:      p,
			}
		} else {
			logs.Trace("GetDBConn should not happen.%t, %v", r, r)
			return NilRPConn
		}
	}
}

// 动态增加容量
// TODO: by YZH 如何实现redispool动态减容量？
// TODO: by YZH 根据github.com/youtube/vitness/go/vt/dbconnpool Capacity设置需要和Close等操作之间加锁
func (p *Pool) dynamicAddCapacity() bool {
	cfgCap := p.conf.Capacity
	cfgMaxCap := p.conf.MaxCapacity
	oldPoolCap := int(p.resPool.Capacity())
	if cfgMaxCap <= cfgCap || oldPoolCap >= cfgMaxCap {
		return false
	}
	newCap := oldPoolCap * 2
	if newCap >= cfgMaxCap {
		newCap = cfgMaxCap
	}
	if err := p.resPool.SetCapacity(newCap); err != nil {
		return false
	}
	logs.Warn("[redispool] dynamicAddCapacity for %s  %d -> %d", p.Name, oldPoolCap, newCap)
	return true
}

func (p *Pool) poolCheckConn() {
	p.QuitChan = make(chan bool, 1)
	go func() {
		for {
			select {
			case <-timing.After(time.Second * 3):
				go func() {
					r := p.getDBConn(true, true)
					defer r.Close()
				}()
			case <-p.QuitChan:
				logs.Debug("check pool conn quit")
				return
			}
		}
	}()
}

var (
	timing = timingwheel.NewTimingWheel(time.Second, 10)
	mutex  sync.Mutex
)
