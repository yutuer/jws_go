package csdb

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

const defaultRedisPoolSize = 50

var pools = map[uint32]redispool.IPool{}

//GetDBConn ..
func GetDBConn(group uint32) redispool.RedisPoolConn {
	if p, exist := pools[group]; exist {
		return p.GetDBConn()
	}
	return redispool.NilRPConn
}

//RedisConfig ..
type RedisConfig struct {
	Server string `json:"AddrPort"`
	Auth   string `json:"Auth"`
	DB     int    `json:"DB"`
}

//SetupRedis ..
func SetupRedis(mCfg map[uint32]*RedisConfig, devMode bool) {
	for group, cfg := range mCfg {
		if nil == pools[group] {
			poolName := fmt.Sprintf("gamex.redis.%s.%d", "crossservice", group)
			pool := redispool.NewSimpleRedisPool(poolName, cfg.Server, cfg.DB,
				cfg.Auth, devMode, defaultRedisPoolSize, true)
			pools[group] = pool
		}
	}
}
