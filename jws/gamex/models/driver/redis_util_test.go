package driver

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func TestMain(m *testing.M) {
	SetupRedis(":6379", 4, "", true)
	mkTestSet("keyfortesthscan", 2000)

	exitc := m.Run()
	// TearDown
	mvTestSet("keyfortesthscan")
	os.Exit(exitc)
}

func mkTestSet(key string, c int) {
	conn := GetDBConn()
	for i := 0; i < c; i++ {
		conn.Do("hset", key, fmt.Sprintf("%d:%s", i, uuid.NewV4().String()), i)
	}
}

func mvTestSet(key string) {
	conn := GetDBConn()
	_, err := conn.Do("del", key)
	if err != nil {
		logs.Error("Remove kefortesthscan err: %s", err.Error())
	}
}

func TestHScan(t *testing.T) {
	conn := GetDBConn()
	all := 0
	err := HScan(conn.Conn, "keyfortesthscan", func(keys, values []string) error {
		all += len(keys)
		for j := 0; j < len(keys); j++ {
			//logs.Trace("keyfortesthscan %s -> %s", keys[j], values[j])
			// 为了将来集成测试，求不刷屏
			assert.NotEmpty(t, keys[j])
			assert.NotEmpty(t, values[j])
		}
		return nil
	})
	if err != nil {
		t.Errorf("TestHScan Err")
	}
	logs.Warn("keyfortesthscan %d", all)
	return
}

func debugAddName(name string, acid string, sid uint) {
	conn := GetDBConn()
	_, err := conn.Do("HSETNX", TableChangeName(sid), name, acid)
	if err != nil {
		logs.Error("debugAddName err %s", err.Error())
	}
}

func debugDelName(name string, acid string, sid uint) {
	conn := GetDBConn()
	_, err := conn.Do("HDEL", TableChangeName(sid), name, acid)
	if err != nil {
		logs.Error("debugDelName err %s", err.Error())
	}
}

func TestRenameToRedis(t *testing.T) {
	res, err := RenameToRedis("lbb001", "lbb002", "111", 1)
	assert.Nil(t, err)
	assert.NotEmpty(t, res)

	/*
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(res)
		}
	*/

	res, err = RenameToRedis("lbb002", "lbb003", "111", 1)
	assert.Nil(t, err)
	assert.NotEmpty(t, res)

	/*
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(res)
		}*/

	debugDelName("lbb003", "111", 1)
}
