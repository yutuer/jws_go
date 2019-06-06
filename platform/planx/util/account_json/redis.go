package accountJson

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func NewRedisPool(server, password string, dbSeleccted, cap int) redispool.IPool {
	cfg := redispool.Config{
		HostAddr:    server,
		Capacity:    cap,
		MaxCapacity: 1000,
		IdleTimeout: time.Minute,
		DBSelected:  dbSeleccted,
		DBPwd:       password,

		ConnectTimeout: 200 * time.Millisecond,
		ReadTimeout:    0,
		WriteTimeout:   0,
	}

	return redispool.NewRedisPool(cfg, "redisstorage", true)
}

func MkJson(result interface{}, err error) ([]byte, error) {
	values, err := redis.StringMap(result, err)

	if err != nil {
		return []byte{}, err
	}

	chunk, err := json.Marshal(values)
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("marshal obj failed, obj:%v, err:%v", values, err))
	}

	return chunk, nil
}

func FromJson(conn redis.Conn, key string, data []byte) error {
	value_map := map[string]string{}
	err := json.Unmarshal(data, &value_map)

	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(value_map)*2+1)
	args = append(args, key)
	for k, v := range value_map {
		args = append(args, k)
		args = append(args, v)
	}

	_, err = conn.Do("HMSET", args...)
	return err
}
