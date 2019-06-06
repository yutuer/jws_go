package frometcd

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
)

//InitRedis ..
func InitRedis(devMode bool) {
	cfg := map[uint32]*csdb.RedisConfig{}
	if devMode {
		for _, group := range config.Cfg.GroupIDs {
			cfg[group] = &csdb.RedisConfig{
				Server: config.Cfg.DBRedis,
				Auth:   config.Cfg.DBRedisAuth,
				DB:     config.Cfg.DBRedisDB,
			}
		}
	} else {
		loaded, err := loadRedisCfg()
		if nil != err {
			panic(err)
		}
		cfg = loaded
		logs.Info("Load DB Config: %+v", cfg)
	}

	csdb.SetupRedis(cfg, devMode)
}

func loadRedisCfg() (map[uint32]*csdb.RedisConfig, error) {
	etcdMap, err := etcdCfg()
	if err != nil {
		return nil, fmt.Errorf("get etcd config failed, %v", err)
	}

	cfg := map[uint32]*csdb.RedisConfig{}
	for _, group := range config.Cfg.GroupIDs {
		if sub, ok := etcdMap[fmt.Sprint(group)]; ok {
			cfg[group] = sub
		} else {
			if def, ok := etcdMap["default"]; ok {
				cfg[group] = def
			} else {
				return nil, fmt.Errorf("crossservice etcdmap is empty. group[%d]", group)
			}
		}
	}

	return cfg, nil
}

//TODO etcd的内容待确认
func etcdCfg() (map[string]*csdb.RedisConfig, error) {
	var etcdMap map[string]*csdb.RedisConfig
	key := etcdKey()
	logs.Info("Load DB Config From Etcd key: %+v", key)
	jsonValue, err := etcd.Get(key)
	if err != nil {
		return nil, fmt.Errorf("csrob etcd get key failed. %s", key)
	}
	logs.Info("Load DB Config From Etcd value: %+v", jsonValue)

	if err := json.Unmarshal([]byte(jsonValue), &etcdMap); err != nil {
		return nil, fmt.Errorf("csrob json.Unmarshal key failed. %s", key)
	}

	return etcdMap, nil
}

func etcdKey() string {
	return fmt.Sprintf("%s/%d/%s", config.Cfg.EtcdRoot, config.Cfg.Gid, config.Cfg.DBKeyEtcdPath)
}
