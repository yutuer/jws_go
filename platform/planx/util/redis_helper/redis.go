package redis_helper

import (
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/dns_rand"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

type RedisPoolCfg struct {
	RedisServer  string
	DbSelected   int
	DbPwd        string
	DevMode      bool
	DnsValid     bool
	NewRedisPool bool
}

func SetupRedis(cfg RedisPoolCfg) redispool.IPool {
	return SetupRedisByCap(cfg, 10)
}

func SetupRedisForSimple(cfg RedisPoolCfg) redispool.IPool {
	return SetupRedisByCap(cfg, 5)
}

func SetupRedisByCap(cfg RedisPoolCfg, cap int) redispool.IPool {
	redisServer := cfg.RedisServer
	if game.Cfg.RedisDNSValid {
		redisServer = dns_rand.GetAddrByDNS(redisServer)
	}
	return redispool.NewSimpleRedisPool("gamex.redis.save", redisServer,
		cfg.DbSelected, cfg.DbPwd, cfg.DevMode, cap, cfg.NewRedisPool)
}
