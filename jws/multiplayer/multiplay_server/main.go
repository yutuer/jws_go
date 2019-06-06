package main

import (
	"os"
	"runtime"

	_ "net/http/pprof"

	_ "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/cmds/allinone"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/cmds"
)

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
