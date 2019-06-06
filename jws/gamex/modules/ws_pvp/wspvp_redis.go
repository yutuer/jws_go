package ws_pvp

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

var pool redispool.IPool
var redisCfg RedisDBSetting

func getDBConn() redispool.RedisPoolConn {
	return pool.GetDBConn()
}

type RedisDBSetting struct {
	AddrPort string
	Auth     string
	DB       int
}

func InitRedis(devMode bool) {
	if devMode {
		redisCfg = RedisDBSetting{
			AddrPort: game.Cfg.Redis,
			Auth:     game.Cfg.RedisAuth,
			DB:       2,
		}
		logs.Debug("init wspvp redis config by default, %v", redisCfg)
	} else {
		err := InitRedisCfg()
		if err != nil {
			panic(err)
		}
	}
	SetupRedis(redisCfg.AddrPort, redisCfg.DB, redisCfg.Auth, devMode)
}

func SetupRedis(redisServer string, dbSeleccted int, dbPwd string, devmode bool) {
	if nil == pool {
		poolName := fmt.Sprintf("gamex.redis.%s", "wspvp")
		pool = redispool.NewSimpleRedisPool(poolName, redisServer, dbSeleccted,
			dbPwd, devmode, Wspvp_Redis_Pool_Size, game.Cfg.NewRedisPool)
	}
}

func InitRedisCfg() error {
	etcdMap, err := getEtcdCfg()
	if err != nil {
		return err
	}
	groupId := gamedata.GetWSPVPGroupId(uint32(game.Cfg.ShardId[0]))
	groupString := fmt.Sprintf("%d", groupId)
	rcfg, ok := etcdMap[groupString]
	if !ok {
		rcfg, ok = etcdMap["default"]
		if !ok {
			return fmt.Errorf("wspvpetcdmap key not found. %s", groupString)
		} else {
			redisCfg = rcfg
		}
	} else {
		redisCfg = rcfg
	}
	logs.Debug("init redis cfg from etcd, %v", redisCfg)
	return nil
}

func getEtcdCfg() (map[string]RedisDBSetting, error) {
	var etcdMap map[string]RedisDBSetting
	if jsonValue, err := etcd.Get(getEtcdKey()); err != nil {
		return nil, fmt.Errorf("wspvp etcd get key failed. %s", getEtcdKey())
	} else {

		if err := json.Unmarshal([]byte(jsonValue), &etcdMap); err != nil {
			return nil, fmt.Errorf("wspvp json.Unmarshal key failed. %s", getEtcdKey())
		}
	}
	return etcdMap, nil
}

func getEtcdKey() string {
	return fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, game.Cfg.Gid, "WSPVP/dbs")
}
