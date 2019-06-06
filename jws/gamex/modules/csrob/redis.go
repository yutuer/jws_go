package csrob

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

var pool redispool.IPool

func getDBConn() redispool.RedisPoolConn {
	return pool.GetDBConn()
}

type redisConfig struct {
	Server string `json:"AddrPort"`
	Auth   string `json:"Auth"`
	DB     int    `json:"DB"`
}

func InitRedis(devMode bool) {
	var cfg *redisConfig
	if devMode {
		cfg = &redisConfig{
			Server: game.Cfg.Redis,
			Auth:   game.Cfg.RedisAuth,
			DB:     3,
		}
	} else {
		loaded, err := loadRedisCfg()
		if nil != err {
			panic(err)
		}
		cfg = loaded
	}

	setupRedis(cfg, devMode)
}

func setupRedis(cfg *redisConfig, devMode bool) {
	if nil == pool {
		poolName := fmt.Sprintf("gamex.redis.%s", "csrob")
		pool = redispool.NewSimpleRedisPool(poolName, cfg.Server, cfg.DB,
			cfg.Auth, devMode, RedisPoolSize, game.Cfg.NewRedisPool)
	}
}

func loadRedisCfg() (*redisConfig, error) {
	etcdMap, err := etcdCfg()
	if err != nil {
		return nil, fmt.Errorf("get etcd config failed, %v", err)
	}

	//TODO 现在不支持gamex多shard启动
	groupID := gamedata.GetCSRobGroupId(uint32(game.Cfg.ShardId[0]))
	cfg, ok := etcdMap[fmt.Sprint(groupID)]
	if false == ok {
		cfg, ok = etcdMap["default"]
		if false == ok {
			return nil, fmt.Errorf("csrob etcdmap is empty. group[%d]", groupID)
		}
	}

	if nil == cfg {
		return nil, fmt.Errorf("csrob etcdmap is nil. group[%d]", groupID)
	}

	return cfg, nil
}

//TODO etcd的内容待确认
func etcdCfg() (map[string]*redisConfig, error) {
	var etcdMap map[string]*redisConfig
	jsonValue, err := etcd.Get(etcdKey())
	if err != nil {
		return nil, fmt.Errorf("csrob etcd get key failed. %s", etcdKey())
	}

	if err := json.Unmarshal([]byte(jsonValue), &etcdMap); err != nil {
		return nil, fmt.Errorf("csrob json.Unmarshal key failed. %s", etcdKey())
	}

	return etcdMap, nil
}

func etcdKey() string {
	return fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, game.Cfg.Gid, "CSROB/dbs")
}
