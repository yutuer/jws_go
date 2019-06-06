package main

import (
	"fmt"
	"os"

	_ "vcs.taiyouxi.net/platform/x/auth/cmds/allinone"
	_ "vcs.taiyouxi.net/platform/x/auth/cmds/auth"
	_ "vcs.taiyouxi.net/platform/x/auth/cmds/login"
	_ "vcs.taiyouxi.net/platform/x/auth/cmds/verupdateurl"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/auth/cmds"
)

/*
Auth系统使用Gin作为API访问框架
*/
func main() {
	defer logs.Close()
	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "AuthApix"
	app.Usage = fmt.Sprintf("Auth Http Api Server. version: %s", version.GetVersion())
	app.Author = "YZH"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	app.Run(os.Args)
}
