package etcd_redis

import (
	"testing"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func TestToGet(t *testing.T) {
	etcd.InitClient([]string{"http://127.0.0.1:2379/"})
	RegRedisInfo("/a4k", 0, "10.0.1.69:6379", 1, "")
	RegRedisInfo("/a4k", 1, "10.0.1.69:6379", 4, "")
	RegRedisInfo("/a4k", 2, "10.0.1.69:6379", 6, "")
	RegRedisInfo("/a4k", 3, "10.0.1.69:6379", 8, "")

	res, err := GetRedisInfoAll("/a4k")
	if err != nil {
		t.Errorf("err by %s", err.Error())
	}
	logs.Trace("res %v", res)
	logs.Close()
}
