package onland

import (
	"testing"
	"time"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

var (
	pool *redis.Pool
)

func getDBConn() redis.Conn {
	return pool.Get()
}

func newRedisPool(server, password string, dbSeleccted int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			c.Do("SELECT", dbSeleccted)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func SetupRedis(redisServer, redisPassword string, dbSeleccted int) {
	pool = newRedisPool(redisServer, redisPassword, dbSeleccted)
}

func TestScanRedis(t *testing.T) {
	SetupRedis(":6379", "", 0)
	db := getDBConn()

	keys_all := make([]string, 0, 100)
	num := 0

	hander := func(keys []string) error {
		t.Logf("keys hander")
		for i, key := range keys {
			num++
			keys_all = append(keys_all, key)
			t.Logf("key %d %d %s", i, num, key)
		}
		return nil
	}

	scanner := NewScanner(db, hander)
	scanner.Start()
	for !scanner.IsScanOver() {
		err := scanner.Next()
		if err != nil {
			t.Errorf("scanner next err %v", err)
		}
	}

	t.Logf("all keys %d - %v", num, keys_all)
}
