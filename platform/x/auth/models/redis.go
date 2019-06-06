package models

import (
	"strconv"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func newRedisPool(pooName, server, password string, dbselectCfg string) redispool.IPool {
	db_id, err := strconv.Atoi(dbselectCfg)
	if err != nil {
		logs.Warn("newRedisPool can't read ", dbselectCfg, ". default is 0")
		db_id = 0
	}
	logs.Debug("newRedisPool read %s it is %d", dbselectCfg, db_id)
	return redispool.NewSimpleRedisPool(pooName, server, db_id, password, false,
		redispool.Default_RedisPool_Capacity, true)
}
