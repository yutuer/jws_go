package warm

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"sync"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd_redis"
	apiInterface "vcs.taiyouxi.net/platform/x/redis_storage/api/interface"
)

type module struct {
	api      apiInterface.WarmApi
	wg       sync.WaitGroup
	stopChan chan bool
}

func New(sid uint) *module {
	m := new(module)
	m.api.Init(game.Cfg.EtcdRoot, int(sid), game.Cfg.EtcdEndPoint)
	return m
}

func (r *module) AfterStart(g *gin.Engine) {
}

func (r *module) BeforeStop() {
}

func (r *module) Start() {
	r.api.Start()

	r.wg.Add(1)
	timeChan := time.After(60 * time.Second)
	r.stopChan = make(chan bool, 1)

	etcd_redis.RegRedisInfo(
		game.Cfg.EtcdRoot,
		uint(r.api.SId()),
		game.Cfg.Redis,
		game.Cfg.RedisDB,
		game.Cfg.RedisAuth)
	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-timeChan:
				timeChan = time.After(60 * time.Second)
				etcd_redis.RegRedisInfo(
					game.Cfg.EtcdRoot,
					uint(r.api.SId()),
					game.Cfg.Redis,
					game.Cfg.RedisDB,
					game.Cfg.RedisAuth)
			case <-r.stopChan:
				return
			}
		}
	}()
}

func (r *module) Stop() {
	r.stopChan <- true
	r.api.Stop()
	r.wg.Wait()
}

func (r *module) WarmKey(key string) error {
	warmServiceKey := apiInterface.GetServiceKeyID(
		strconv.Itoa(r.api.SId()),
		game.Cfg.Redis,
		game.Cfg.RedisDB)

	return r.api.WarmKey(warmServiceKey, key)
}
