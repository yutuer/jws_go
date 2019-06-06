/*
重构的ws_pvp测试用例，使用了SubTest
请使用GO1.8以上版本进行测试
并确保本地Redis服务正常
*/
package ws_pvp

import (
	"testing"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
)

func TestMain(m *testing.M) {
	gamedata.DebugLoadLocalGamedata()

	SetupRedis("127.0.0.1:6379", 15, "", true)
	etcd.InitClient([]string{"http://127.0.0.1:2379/"})
	game.Cfg.ShardId = []uint{203}

	m.Run()
}
