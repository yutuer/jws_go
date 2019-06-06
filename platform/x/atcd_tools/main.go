package main

import (
	"os"

	_ "vcs.taiyouxi.net/platform/x/atcd_tools/cmds/allinone"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/atcd_tools/cmds"
)

func main() {
	defer logs.Close()
	app := cli.NewApp()

	app.Version = ""
	app.Name = "atcd_tool"
	app.Usage = "Atcd Config Tool web server"
	app.Author = "FanYang"
	app.Email = "fanyang@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	app.Run(os.Args)
}
