package main

import (
	"fmt"
	"os"

	_ "vcs.taiyouxi.net/platform/x/clientlog/cmds/allinone"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/clientlog/cmds"
)

/*
Auth系统使用Gin作为API访问框架
*/
func main() {
	defer logs.Close()
	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "ClientLog"
	app.Usage = fmt.Sprintf("ClientLog Server. version: %s", version.GetVersion())
	app.Author = "YZH"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	app.Run(os.Args)
}
