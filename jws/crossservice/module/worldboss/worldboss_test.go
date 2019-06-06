package worldboss

import (
	"fmt"
	"os"
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

var (
	testGroupId        uint32 = 10001
	unavailableGroupId uint32 = 404404
	notExistGroupId    uint32 = 503503

	testSid uint32 = 255
)

func TestMain(m *testing.M) {
	gamedata.DebugLoadLocalGamedata()
	debugSetCSConfig()

	retCode := m.Run()

	os.Exit(retCode)
}

// debugSetConfig 设置CrossServer测试参数
func debugSetCSConfig() {
	rc := &csdb.RedisConfig{Server: "127.0.0.1:6379", Auth: "", DB: 8}
	// 修改Server用于测试redis无法连接
	rc_redis_down := &csdb.RedisConfig{Server: "127.0.0.1:26379", Auth: "", DB: 8}

	mCfg := make(map[uint32]*csdb.RedisConfig)
	mCfg[testGroupId] = rc
	mCfg[unavailableGroupId] = rc_redis_down

	csdb.SetupRedis(mCfg, true)
}

// debugSetResource 每次返回一个新的带固定参数的resouce
func debugGetResource() *resources {
	r := newResources(testGroupId, &WorldBoss{})
	r.ticker = newTickerHolder(r)
	r.ticker.roundStatus = &RoundStatus{BatchTag: fmt.Sprintf("%04d%02d%02d", 2017, 7, 25)}

	return r
}
