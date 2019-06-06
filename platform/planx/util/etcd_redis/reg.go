//在etcd中注册游戏分服当前正在使用的redis数据库的地址端口等信息。
package etcd_redis

import (
	"fmt"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	RedisInfoEtcdKeyPathRoot = "%s/redis/"
	RedisInfoEtcdKeyPath     = "%s/redis/%d/"
	RedisEtcdKeyAddr         = "address"
	RedisEtcdKeyAuth         = "auth"
	RedisEtcdKeyDBNum        = "db"
)

type redisInfo struct {
	ShardID      uint
	ProfileName  string
	RedisAddress string
	RedisDB      int
	RedisAuth    string
}

func RegRedisInfo(
	etcdRoot string, shardID uint, redisAddress string, redisDB int, redisAuth string) error {

	// path /a4k/redis/{sid}/{profile}/redisAddr -> :6379
	logs.Trace("RegRedisInfo %d - %s %d %s",
		shardID,
		redisAddress, redisDB, redisAuth)
	path := fmt.Sprintf(RedisInfoEtcdKeyPath,
		etcdRoot, shardID)
	err := etcd.Set(path+RedisEtcdKeyAddr, redisAddress, 0)
	if err != nil {
		return err
	}
	err = etcd.Set(path+RedisEtcdKeyAuth, redisAuth, 0)
	if err != nil {
		return err
	}
	err = etcd.Set(path+RedisEtcdKeyDBNum, strconv.Itoa(redisDB), 0)
	if err != nil {
		return err
	}
	return nil
}

// GetRedisInfoAll 目前只在RedisStorage中使用
func GetRedisInfoAll(etcdRoot string) (map[uint]redisInfo, error) {
	res := make(map[uint]redisInfo, 32)
	path := fmt.Sprintf(RedisInfoEtcdKeyPathRoot,
		etcdRoot)
	allShards, err := etcd.GetAllSubKeys(path)
	if err != nil {
		return res, err
	}
	logs.Trace("allProfiles %v", allShards)
	for _, shardIDPath := range allShards {
		info, err := etcd.GetAllSubKeyValue(shardIDPath)
		if err != nil {
			continue
		}
		logs.Trace("info %s ->  %v", shardIDPath, info)
		shardIDPaths := strings.Split(shardIDPath, "/")
		if len(shardIDPaths) <= 0 {
			continue
		}
		shardIDStr := shardIDPaths[len(shardIDPaths)-1]
		address, addOk := info[RedisEtcdKeyAddr]
		num, dbOk := info[RedisEtcdKeyDBNum]
		auth, authOk := info[RedisEtcdKeyAuth]
		if !(addOk && dbOk && authOk) {
			continue
		}
		db, err := strconv.Atoi(num)
		if err != nil {
			continue
		}
		sid, err := strconv.Atoi(shardIDStr)
		if err != nil {
			continue
		}
		res[uint(sid)] = redisInfo{
			RedisAddress: address,
			RedisAuth:    auth,
			RedisDB:      db,
			ShardID:      uint(sid),
		}
	}
	return res, nil
}
