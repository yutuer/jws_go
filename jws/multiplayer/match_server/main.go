package main

import (
	"os"
	"runtime"

	_ "net/http/pprof"

	_ "vcs.taiyouxi.net/jws/multiplayer/match_server/cmds/allinone"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/cmds"
)

//DONE: timeout放到1h, 客户端会收到服务器Cancel消息。
//DONE: 匹配登机的登机宽度需要放大,增匹配成功率, 可匹配的人已经使用了排序,所以可以保证就近
//DONE: 时间分片需要放大, 因为每次匹配都是匹配当前时间片中的人

//TODO: Matchplay的等级选择应该用70 percentile
//DONE: 客户端可以Cancel的逻辑需要实现
//DONE: 不同boss的匹配放到独立的goroutine中

//TODO: 根据当前参加Match的人数来选择策略,根据配置的策略启动不同service.
//TODO: 能够使1-3个人都可以开始匹配

func main() {
	defer logs.Close()

	app := cli.NewApp()
	app.Version = version.GetVersion()
	app.Name = "multiplay"
	app.Author = "YinZeHong"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	logs.Info("GOMAXPROCS is %d", runtime.GOMAXPROCS(0))
	logs.Info("Version is %s", app.Version)

	app.Run(os.Args)
}
