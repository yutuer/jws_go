package main

import (
	"os"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/gen_shard_name/cmd"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "tools"
	app.Usage = "Ticore game company game server maintenance tool."
	app.Author = "YinZeHong"
	app.Email = "yinzehong@taiyouxi.cn"

	cmd.InitCommands(&app.Commands)
	defer logs.Close()
	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	app.Run(os.Args)
}
