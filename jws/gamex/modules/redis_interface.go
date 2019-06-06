package modules

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/dns_rand"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

// 主要给排行榜相关的db用

var (
	pool redispool.IPool
	Cfg  Config
)

type Config struct {
	Redis         string `toml:"rank_redis"`
	RedisDNSValid bool   `toml:"redis_dns_valid"`
	RedisAuth     string `toml:"rank_redis_auth"`
	RedisDB       int    `toml:"rank_redis_db"`
}

func (c *Config) Sync2Etcd() bool {
	gameCfg := game.Cfg
	_ss := make([]uint, 0, 4)
	_ss = append(_ss, gameCfg.ShardId...)
	_ss = append(_ss, gameCfg.MergeRel...)
	for _, shardId := range _ss {
		key_parent := fmt.Sprintf("%s/%d/%d/gm/", gameCfg.EtcdRoot, gameCfg.Gid, shardId)
		// rank reload url
		if err := etcd.Set(key_parent+"redis_rank", c.Redis, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"redis_rank", err)
			return false
		}
		if err := etcd.Set(key_parent+"redis_rank_db", fmt.Sprintf("%d", c.RedisDB), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"redis_rank_db", err)
			return false
		}
	}
	return true
}

func GetDBConn() redispool.RedisPoolConn {
	return pool.GetDBConn()
}
func SetupRedis(redisServer string, dbSeleccted int, dbPwd string, devmode bool) {
	if pool == nil {
		poolName := fmt.Sprintf("gamex.redis.%s", "rank")
		if Cfg.RedisDNSValid {
			redisServer = dns_rand.GetAddrByDNS(redisServer)
		}
		pool = redispool.NewSimpleRedisPool(poolName, redisServer,
			dbSeleccted, dbPwd, devmode,
			redispool.Default_RedisPool_Capacity,
			game.Cfg.NewRedisPool)
	}
}
