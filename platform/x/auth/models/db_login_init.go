package models

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const ac_number_base = "ac_number_base"
const ac_atoi = "ac_atoi"
const ac_iota = "ac_iota"

// default use dynamoDB
func GetNumberIDForPlayer(acID string) (int, error) {
	conn := loginRedisPool.Get()
	defer conn.Close()
	if conn.IsNil() {
		return 0, fmt.Errorf("nil redis connection")
	}

	id, err := redis.Int(conn.Do("HGET", ac_atoi, acID))
	if err != nil && err != redis.ErrNil {
		return 0, err
	}
	if id == 0 {
		// register redis
		id, err = redis.Int(conn.Do("INCR", ac_number_base))
		if err != nil {
			return 0, err
		}
		cb := redis.NewCmdBuffer()
		err = cb.Send("HSET", ac_atoi, acID, id)
		if err != nil {
			return 0, err
		}
		err = cb.Send("HSET", ac_iota, id, acID)
		if err != nil {
			return 0, err
		}
		_, err := conn.DoCmdBuffer(cb, true)
		if err != nil {
			return 0, err
		}
	}
	logs.Debug("NumberID is %d", id)
	return id, nil
}

func InitNumberIDInRedis() error {
	conn := loginRedisPool.Get()
	if conn.IsNil() {
		return fmt.Errorf("nil redis connection")
	}
	exist, err := redis.Int(conn.Do("EXISTS", ac_number_base))
	if err != nil {
		return err
	}
	if exist == 1 {
		return nil
	}
	// register redis
	_, err = conn.Do("SET", ac_number_base, 0)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
