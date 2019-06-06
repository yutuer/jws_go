package main

import (
	"fmt"
	"os"

	"runtime"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/api_gateway/cmds"
	_ "vcs.taiyouxi.net/platform/x/api_gateway/cmds/pay"
)

func main() {
	defer logs.Close()

	app := cli.NewApp()
	app.Version = version.GetVersion()
	app.Name = "api_gateway"
	app.Usage = fmt.Sprintf("Ticore game company api gateway server. version:%s", version.GetVersion())
	app.Author = "ZhangZhen"
	app.Email = "zhangzhen@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	logs.Info("GOMAXPROCS is %d", runtime.GOMAXPROCS(0))

	app.Run(os.Args)
}
