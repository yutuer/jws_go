package profile_tool

import (
	"errors"
	"fmt"

	"sort"

	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/redis_storage/restore"
)

func InitDB() {
	serverNameMerge = make(map[string]string, 10)
	for ser, same := range redisSameSever {
		ns := make([]string, 0, 2)
		ns = append(ns, ser)
		if len(same) > 0 {
			ns = append(ns, same...)
		}
		cfg := gmConfig.Cfg.GetServerCfgFromName(ser)
		regRedisConn(ns, cfg.RedisAddress, cfg.RedisAuth, cfg.RedisDB)
		regRedisRankConn(ns, cfg.RedisRankAddress, cfg.RedisAuth, cfg.RedisRankDB)
	}
	logs.Warn("serverNameMerge %v", serverNameMerge)
}

func GetAllServerCfg() []gmConfig.ServerCfg {
	res := make(gmConfig.ServerCfgSlice, 0, len(gmConfig.Cfg.RedisName))
	for _, name := range gmConfig.Cfg.ServerName {
		res = append(res, gmConfig.Cfg.GetServerCfgFromName(name))
	}
	sort.Sort(res)
	return res
}

func GetAllRedis() []string {
	res := make([]string, 0, len(redisConns))
	for name, _ := range redisConns {
		res = append(res, name)
	}
	return res
}

func GetProfileFromRedis(server_name, account string) ([]byte, error) {
	conn := getRedisConn(server_name)
	defer conn.Close()
	if conn.IsNil() {
		return []byte{}, errors.New("No Server Exit")
	}
	return storehelper.RedisHM2Json(conn.Do("HGETALL", account))
}

func GetNickNameFromRedisByACID(redis_name, server_name, account string) (string, error) {
	conn := getRedisConn(server_name)
	defer conn.Close()
	if conn.IsNil() {
		return "", errors.New("No Server Exit")
	}
	// 兼容新老服务器
	server_name = GetServerName(server_name)
	logs.Debug("GetAcountIDFromRedis, %v, %v, %v", redis_name, server_name, account)
	res, err := redis.String(conn.Do("HGET", account, "name"))
	logs.Debug("redis.String, %v, %v", res, err)
	if err != nil || res == "" {
		logs.Debug("redis.String, %v, %v", res, err)
		return res, err
	} else {
		return res, err
	}
}
func GetAcountIDFromRedis(redis_name, server_name, nick string) (string, error) {
	conn := getRedisConn(redis_name)
	defer conn.Close()
	if conn.IsNil() {
		return "", errors.New("No Server Exit")
	}

	// 兼容新老服务器
	server_name = GetServerName(server_name)
	logs.Debug("GetAcountIDFromRedis, %v, %v, %v", redis_name, server_name, nick)
	res, err := redis.String(conn.Do("HGET", fmt.Sprintf("%s:names", server_name), nick))
	logs.Debug("redis.String, %v, %v", res, err)
	if err != nil || res == "" {
		res, err := redis.String(conn.Do("HGET", fmt.Sprintf("names:%s", server_name), nick))
		logs.Debug("redis.String, %v, %v", res, err)
		return res, err
	} else {
		return res, err
	}
}

func ModPrfileFromRedis(server_name, account string, data string) error {
	conn := getRedisConn(server_name)
	defer conn.Close()
	if conn.IsNil() {
		return errors.New("No Server Exit")
	}
	_, err := conn.Do("DEL", account)
	if err != nil {
		logs.Error("DEL ERR By %s", err.Error())
		logs.Error("%s", account)
		logs.Error("%v", data)
		return err
	}
	return restore.FromJson(conn.RawConn(), account, []byte(data))
	return nil
}

func DelRankTrial(server_name, account string) error {
	conn := getRedisRankConn(server_name)
	defer conn.Close()
	if conn.IsNil() {
		return errors.New("No Server Exit")
	}

	i, err := redis.Int(conn.Do("ZREM", fmt.Sprintf("%s:RankCorpTrial", server_name), account))
	if err != nil {
		return err
	}

	if i <= 0 {
		return errors.New("account not found")
	}

	return nil
}

var redisSameSever map[string][]string
var redisConns map[string]redispool.IPool
var serverNameMerge map[string]string

func GetServerName(server_name string) string {
	ser, ok := serverNameMerge[server_name]
	if ok {
		return ser
	}
	return server_name
}

func SameServer(RedisName, MergedServer []string) {
	if len(RedisName) != len(MergedServer) {
		panic(fmt.Errorf("len(RedisName) != len(MergedServer) %v %v",
			RedisName, MergedServer))
	}
	logs.Warn("RedisName MergedServer %v %v ", RedisName, MergedServer)
	if redisSameSever == nil {
		redisSameSever = make(map[string][]string, 16)
	}
	for i, ms := range MergedServer {
		s := RedisName[i]
		ss, ok := redisSameSever[ms]
		if !ok {
			ss = make([]string, 0, 2)
		}
		if ms != "" && s != ms {
			ss = append(ss, s)
		}
		redisSameSever[ms] = ss
	}
	logs.Warn("redisSameSever %v", redisSameSever)
}

func getRedisConn(server_name string) redispool.RedisPoolConn {
	pool, ok := redisConns[server_name]
	if !ok {
		logs.Error("server_name no find by %s", server_name)
		return redispool.NilRPConn
	}

	return pool.Get()
}

func regRedisConn(server_name []string, server, password string, dbSeleccted int) error {
	if redisConns == nil {
		redisConns = make(map[string]redispool.IPool)
	}

	cfg := redispool.Config{
		HostAddr:       server,
		Capacity:       2,
		MaxCapacity:    4,
		IdleTimeout:    time.Minute,
		DBSelected:     dbSeleccted,
		DBPwd:          password,
		ConnectTimeout: 200 * time.Millisecond,
		ReadTimeout:    0,
		WriteTimeout:   0,
	}

	pool := redispool.NewRedisPool(cfg, "gmtools", true)
	if pool.IsNil() {
		return errors.New("RedisPool Create Error")
	}

	for _, s := range server_name {
		redisConns[s] = pool
		serverNameMerge[s] = server_name[0]
	}
	if len(server_name) > 1 {
		logs.Warn("server use same redis pool %v", server_name)
	}
	return nil
}

var redisRankConns map[string]redispool.IPool

func getRedisRankConn(server_name string) redispool.RedisPoolConn {
	pool, ok := redisRankConns[server_name]
	if !ok {
		logs.Error("server_name rank no find by %s", server_name)
		return redispool.NilRPConn
	}

	return pool.Get()
}

func regRedisRankConn(server_name []string, server, password string, dbSeleccted int) error {
	if redisRankConns == nil {
		redisRankConns = make(map[string]redispool.IPool)
	}

	cfg := redispool.Config{
		HostAddr:       server,
		Capacity:       2,
		MaxCapacity:    4,
		IdleTimeout:    time.Minute,
		DBSelected:     dbSeleccted,
		DBPwd:          password,
		ConnectTimeout: 200 * time.Millisecond,
		ReadTimeout:    0,
		WriteTimeout:   0,
	}

	pool := redispool.NewRedisPool(cfg, "gmtools", true)
	if pool.IsNil() {
		return errors.New("RedisPool Create Error")
	}

	for _, s := range server_name {
		redisRankConns[s] = pool
	}
	if len(server_name) > 1 {
		logs.Warn("server use same redis rank pool %v", server_name)
	}
	return nil
}
