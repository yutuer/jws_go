package redis_helper

import (
	"fmt"
	"testing"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func mkTestSet(key string, c int) {
	SetupRedis(":6379", 4, "", true)
	conn := GetDBConn()
	for i := 0; i < c; i++ {
		conn.Do("hset", key, fmt.Sprintf("%d:%s", i, uuid.NewV4().String()), i)
	}
}

func TestHScan(t *testing.T) {
	mkTestSet("keyfortesthscan", 2000)
	conn := GetDBConn()
	all := 0
	err := HScan(conn.Conn, "keyfortesthscan", func(keys, values []string) error {
		all += len(keys)
		for j := 0; j < len(keys); j++ {
			logs.Trace("keyfortesthscan %s -> %s", keys[j], values[j])
		}
		return nil
	})
	if err != nil {
		t.Errorf("TestHScan Err")
	}
	logs.Warn("keyfortesthscan %d", all)
	logs.Close()
	return
}

func TestRename(t *testing.T) {
	SetupRedis(":6379", 4, "", true)
	res, err := RenameToRedis("lbb001", "lbb002", "111", 1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}

	res, err = RenameToRedis("lbb002", "lbb003", "111", 1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
}
