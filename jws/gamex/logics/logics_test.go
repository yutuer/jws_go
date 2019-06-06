package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
)

// logics目录下所有用例共用这个init
func init() {
	account.InitDebuger()
	gamedata.DebugLoadLocalGamedata()
	driver.SetupRedis(":6379", 15, "", true)
	etcd.InitClient([]string{"http://127.0.0.1:2379/"})
}
