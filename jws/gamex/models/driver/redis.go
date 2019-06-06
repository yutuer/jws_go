package driver

import (
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/dns_rand"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

var (
	pool redispool.IPool
)

func GetDBConn() redispool.RedisPoolConn {
	return pool.GetDBConn()
}

func SetupRedis(redisServer string, dbSeleccted int, dbPwd string, devmode bool) {
	if game.Cfg.RedisDNSValid {
		redisServer = dns_rand.GetAddrByDNS(redisServer)
	}
	pool = redispool.NewSimpleRedisPool("gamex.redis.save", redisServer,
		dbSeleccted, dbPwd, devmode, 10, game.Cfg.NewRedisPool)
}

func SetupRedisForSimple(redisServer string, dbSeleccted int, dbPwd string, devmode bool) {
	if game.Cfg.RedisDNSValid {
		redisServer = dns_rand.GetAddrByDNS(redisServer)
	}
	pool = redispool.NewSimpleRedisPool("gamex.redis.save", redisServer,
		dbSeleccted, dbPwd, devmode, 5, game.Cfg.NewRedisPool)
}

//TODO SetupRedis pool close: pool.Close()
func ShutdownRedis() {
	if pool != nil {
		logs.Debug("pool quit chan")
		pool.Close()
	}
}

func SetupRedisByCap(redisServer string, dbSeleccted int, dbPwd string, cap int) {
	if game.Cfg.RedisDNSValid {
		redisServer = dns_rand.GetAddrByDNS(redisServer)
	}
	pool = redispool.NewSimpleRedisPool("gamex.redis.save", redisServer,
		dbSeleccted, dbPwd, false, cap, game.Cfg.NewRedisPool)
}
