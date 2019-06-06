package main

import (
	"os"

	"runtime"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/gamex_merge/cmds"
	_ "vcs.taiyouxi.net/platform/x/gamex_merge/cmds/allinone"
)

func main() {
	defer logs.Close()
	app := cli.NewApp()
	app.Version = version.GetVersion()
	app.Name = "gamex_merge"
	app.Author = "ZhangZhen"
	app.Email = "zhangzhen@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	logs.Info("GOMAXPROCS is %d", runtime.GOMAXPROCS(0))
	logs.Info("Version is %s", app.Version)

	app.Run(os.Args)

}
