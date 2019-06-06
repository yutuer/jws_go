package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/jws/gamex/modules/herogacharace"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
)

/*
	传入大区id和玩家id，查找玩家在redis的链接
*/
func GetAuthRedisConn(gid, sid uint) (redis.Conn, error) {
	// 203大区使用和200相同的auth
	if gid == 203 {
		gid = 200
	}
	//获取redis_url
	redis_url, err := etcd.Get(fmt.Sprintf("%s/%d/auth/login_redis_addr",
		config.CommonConfig.EtcdRoot,
		gid))
	if err != nil {
		return nil, err
	}
	//获取redis_db
	_redis_db, err := etcd.Get(fmt.Sprintf("%s/%d/auth/login_redis_db",
		config.CommonConfig.EtcdRoot,
		gid))
	if err != nil {
		return nil, err
	}
	redis_db, err := strconv.Atoi(_redis_db)
	if err != nil {
		return nil, err
	}

	//获取redis_auth
	redis_auth, err := etcd.Get(fmt.Sprintf("%s/%d/auth/login_redis_db_pwd",
		config.CommonConfig.EtcdRoot,
		gid))
	if err != nil {
		logs.Error("nil auth db password for")
	}

	len_str := len(redis_auth)
	logs.Debug("redis_auth src: %s, %d", redis_auth, len_str)
	// etcd 转义,方便配置,抛弃etcd中redis_auth所配置的空值的引号
	if len_str >= 2 && strings.HasPrefix(redis_auth, `"`) && strings.HasSuffix(redis_auth, `"`) {
		redis_auth = redis_auth[1 : len(redis_auth)-1]
	}
	logs.Debug("redis_auth dst: %s", redis_auth)
	return getRedisConn(redis_url, redis_db, redis_auth)
}

/*
	传入大区id和玩家id，查找玩家在redis的链接
*/
func GetGamexProfileRedisConn(gid, sid uint) (redis.Conn, error) {
	//获取redis_url
	redisUrl, err := etcd.Get(fmt.Sprintf("%s/%d/%d/gm/redis",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		return nil, err
	}
	//获取redis_db
	_redis_db, err := etcd.Get(fmt.Sprintf("%s/%d/%d/gm/redis_db",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		return nil, err
	}
	redisDB, err := strconv.Atoi(_redis_db)
	if err != nil {
		return nil, err
	}
	//获取redisAuth
	redisAuth, err := etcd.Get(fmt.Sprintf("%s/%d/defaults/redis_auth",
		config.CommonConfig.EtcdRoot,
		gid))
	if err != nil {
		logs.Warn("loadProfile etcd.Get(redis_auth) err %v", err)
		redisAuth = ""
	}
	len_str := len(redisAuth)
	// etcd 转义,方便配置,抛弃etcd中redis_auth所配置的空值的引号
	if len_str >= 2 && strings.HasPrefix(redisAuth, `"`) && strings.HasSuffix(redisAuth, `"`) {
		redisAuth = redisAuth[1 : len(redisAuth)-1]
	}
	logs.Debug("redisUrl %s,redisDB %d,redisAuth: %v", redisUrl, redisDB, redisAuth)
	return getRedisConn(redisUrl, redisDB, redisAuth)
}

func getRedisConn(redisUrl string, redisDB int, redisAuth string) (redis.Conn, error) {
	var redisConn redis.Conn
	err := backoff.Retry(func() error {
		conn, err := redis.Dial("tcp", redisUrl,
			redis.DialConnectTimeout(5*time.Second),
			redis.DialReadTimeout(5*time.Second),
			redis.DialWriteTimeout(5*time.Second),
			redis.DialPassword(redisAuth),
			redis.DialDatabase(redisDB),
		)
		if conn != nil {
			redisConn = conn
		}
		return err
	}, herogacharace.New2SecBackOff())

	if err != nil {
		return nil, err
	}
	if redisConn == nil {
		return nil, fmt.Errorf("redis conn nil for param: %v, %v, %v", redisUrl, redisAuth, redisDB)
	}
	return redisConn, nil
}

func GetGamexRankReidsConn(gid, sid uint) (redis.Conn, error) {
	//获取redis_url
	redisUrl, err := etcd.Get(fmt.Sprintf("%s/%d/%d/rank_redis",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		return nil, err
	}
	//获取redis_db
	_redis_db, err := etcd.Get(fmt.Sprintf("%s/%d/%d/rank_redis_db",
		config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		return nil, err
	}
	redisDB, err := strconv.Atoi(_redis_db)
	if err != nil {
		return nil, err
	}
	//获取redisAuth
	redisAuth, err := etcd.Get(fmt.Sprintf("%s/%d/defaults/rank_redis_auth",
		config.CommonConfig.EtcdRoot,
		gid))
	if err != nil {
		logs.Warn("loadProfile etcd.Get(redis_auth) err %v", err)
		redisAuth = ""
	}
	len_str := len(redisAuth)
	// etcd 转义,方便配置,抛弃etcd中redis_auth所配置的空值的引号
	if len_str >= 2 && strings.HasPrefix(redisAuth, `"`) && strings.HasSuffix(redisAuth, `"`) {
		redisAuth = redisAuth[1 : len(redisAuth)-1]
	}
	logs.Debug("redisAuth:%v, url: %v, db: %v", redisAuth, redisUrl, redisDB)
	return getRedisConn(redisUrl, redisDB, redisAuth)
}
