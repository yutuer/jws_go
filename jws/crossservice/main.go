package main

import (
	"os"
	"runtime"

	_ "net/http/pprof"

	_ "vcs.taiyouxi.net/jws/crossservice/cmds/allinone"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/crossservice/cmds"
)

func main() {
	defer logs.Close()

	app := cli.NewApp()
	app.Version = version.GetVersion()
	app.Name = "crossservice"
	app.Author = "QiaoZhu"
	app.Email = "qiaozhu@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	logs.Info("GOMAXPROCS is %d", runtime.GOMAXPROCS(0))
	logs.Info("Version is %s", app.Version)

	app.Run(os.Args)
}
