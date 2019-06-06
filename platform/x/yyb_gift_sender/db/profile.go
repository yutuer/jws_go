package db

import (
	"fmt"

	"strconv"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
)

func loadProfile(acid string) *account.Account {
	id, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("loadProfile db.ParseAccount %s", acid)
		return nil
	}
	logs.Warn("key: %s", fmt.Sprintf("%s/%d/%d/redis",
		config.CommonConfig.EtcdRoot,
		id.GameId, id.ShardId))
	redis_url, err := etcd.Get(fmt.Sprintf("%s/%d/%d/redis",
		config.CommonConfig.EtcdRoot,
		id.GameId, id.ShardId))
	if err != nil {
		logs.Error("loadProfile etcd.Get(redis) err %v", err)
		return nil
	}

	_redis_db, err := etcd.Get(fmt.Sprintf("%s/%d/%d/redis_db",
		config.CommonConfig.EtcdRoot,
		id.GameId, id.ShardId))
	if err != nil {
		logs.Error("loadProfile etcd.Get(redis_db) err %v", err)
		return nil
	}
	redis_db, err := strconv.Atoi(_redis_db)
	if err != nil {
		logs.Error("loadProfile strconv.Atoi err %v", err)
		return nil
	}

	redis_auth, err := etcd.Get(fmt.Sprintf("%s/%d/defaults/redis_auth",
		config.CommonConfig.EtcdRoot,
		id.GameId))
	if err != nil {
		logs.Warn("loadProfile etcd.Get(redis_auth) err %v", err)
		redis_auth = ""
	}
	len_str := len(redis_auth)
	logs.Debug("redis_auth src: %s, %d", redis_auth, len_str)
	// etcd 转义,方便配置,抛弃etcd中redis_auth所配置的空值的引号
	if len_str >= 2 && strings.HasPrefix(redis_auth, `"`) && strings.HasSuffix(redis_auth, `"`) {
		redis_auth = redis_auth[1 : len(redis_auth)-1]
	}
	logs.Debug("redis_auth dst: %s", redis_auth)
	driver.SetupRedisForSimple(
		redis_url,
		redis_db,
		redis_auth,
		false,
	)

	defer driver.ShutdownRedis()
	logs.Warn("ID: %v", id)
	acc, err := account.LoadFullAccount(id, false)
	if err != nil {
		logs.Error("loadProfile LoadFullAccount err %v", err)
		return nil
	}

	return acc
}
