package redisStorageApi

import (
	"fmt"

	"strconv"
	"sync"
	"time"

	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/etcd_redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/redis_storage/api/interface"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds/helper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"
	"vcs.taiyouxi.net/platform/x/redis_storage/restore"
)

const (
	WarmKeyUrl = "/api/v1/warm"
)

var (
	ErrTimeout = errors.New("WarmKeyUrlErrTimeout")
)

type ApiService struct {
	bindAddr     string
	callUrl      string
	etcdKeyRoot  string
	etcdRoot     string
	shardID      string
	redisAddr    string
	redisNum     int
	serviceKeyID string
	restorer     *restore.Restore
	wg           sync.WaitGroup
	stopChan     chan bool
}

func (a *ApiService) Init(etcdRoot, bindAddr, shardID string) {
	a.bindAddr = bindAddr
	a.callUrl = fmt.Sprintf("http://%s%s", bindAddr, WarmKeyUrl)
	a.etcdKeyRoot = fmt.Sprintf(
		postService.RedisStorageWarmServiceEtcdKey,
		etcdRoot)
	a.etcdRoot = etcdRoot
	a.shardID = shardID

}

func (a *ApiService) initRedisInfo(address, auth string, db int) {
	logs.Warn("Load Redis Info From Etcd %s %s %d", address, auth, db)
	a.redisAddr = address
	a.redisNum = db
	a.serviceKeyID = redisStorageApiInterface.GetServiceKeyID(a.shardID, address, db)
	if a.restorer != nil {
		a.restorer.Stop()
	}

	// 直接改配置
	config.RestoreCfg.Redis_Restore = address
	config.RestoreCfg.RedisAuth_Restore = auth
	config.RestoreCfg.RedisDB_Restore = db

	a.restorer = helper.InitRestore()
	go a.restorer.Start()
	return
}

func (a *ApiService) waitRedisInfoFromEtcdInBegin() {
	timeChan := time.After(1 * time.Millisecond)
	logs.Warn("Load Redis Info From Etcd")
	sid, err := strconv.Atoi(a.shardID)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case <-timeChan:
			timeChan = time.After(1 * time.Second)
			res, err := etcd_redis.GetRedisInfoAll(a.etcdRoot)
			if err != nil {
				logs.Error("From Etcd Err By %s", err.Error())
				continue
			}

			info, ok := res[uint(sid)]
			if !ok {
				logs.Error("From Etcd Err By No SID %d info ", sid)
				continue
			}
			a.initRedisInfo(info.RedisAddress, info.RedisAuth, info.RedisDB)
			return
		}
	}
}

func (a *ApiService) checkRedisInfoFromEtcdInRunning(shardID string) {
	logs.Warn("Waitting Redis Info From Etcd")
	sid, err := strconv.Atoi(a.shardID)
	if err != nil {
		logs.Error("shardID Err By %s %s", shardID, err.Error())
		return
	}
	res, err := etcd_redis.GetRedisInfoAll(a.etcdRoot)
	if err != nil {
		logs.Error("From Etcd Err By %s", err.Error())
		return
	}

	info, ok := res[uint(sid)]
	if !ok {
		logs.Error("From Etcd Err By No SID %d info ", sid)
		return
	}

	if a.redisAddr != info.RedisAddress ||
		a.redisNum != info.RedisDB {
		logs.Warn("Redis Info Change From Etcd To %v", info)
		a.initRedisInfo(info.RedisAddress, info.RedisAuth, info.RedisDB)
	}
	return
}

type postCmd struct {
	key string
	res chan error
}

func (a *ApiService) Start() {
	a.waitRedisInfoFromEtcdInBegin()

	cmdChannel := make(chan postCmd, 4096)

	g := gin.New()
	g.POST(WarmKeyUrl, func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		s := new(redisStorageApiInterface.WarmData)
		err := c.Bind(s)

		if err != nil {
			c.String(400, err.Error())
			return
		}
		resChan := make(chan error, 1)
		logs.Trace("warm  %s", s.Key)
		select {
		case cmdChannel <- postCmd{
			key: s.Key,
			res: resChan,
		}:
		case <-ctx.Done():
			err = ErrTimeout
		}

		if err != nil {
			c.String(402, err.Error())
			return
		}

		select {
		case err = <-resChan:
		case <-ctx.Done():
			err = ErrTimeout
		}

		if err == nil {
			c.String(200, string("ok"))
		} else {
			c.String(401, err.Error())
		}
	})

	a.wg.Add(1)
	tinc := 15 * time.Second
	timeChan := time.After(tinc)
	timeGetInfoChan := time.After(15 * time.Second)
	a.stopChan = make(chan bool, 1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case cmd := <-cmdChannel:
				err := a.restorer.RestoreOneWaitRes(cmd.key, nil)
				if cmd.res != nil {
					cmd.res <- err
				}
			case <-timeChan:
				postService.RegService(a.etcdKeyRoot,
					redisStorageApiInterface.GetServiceKeyID(
						a.shardID,
						a.redisAddr,
						a.redisNum),
					a.callUrl, 0, tinc*10)
				timeChan = time.After(tinc)
			case <-timeGetInfoChan:
				timeGetInfoChan = time.After(15 * time.Second)
				a.checkRedisInfoFromEtcdInRunning(a.shardID)
			case <-a.stopChan:
				return
			}
		}
	}()

	g.Run(a.bindAddr)
}

func (a *ApiService) Stop() {
	a.stopChan <- true
	a.wg.Wait()
	a.restorer.Stop()
}

func NewAPIService() *ApiService {
	return new(ApiService)
}
