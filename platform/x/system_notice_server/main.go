package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/system_notice_server/cmds"
	_ "vcs.taiyouxi.net/platform/x/system_notice_server/cmds/allinone"
)

func main() {
	defer logs.Close()
	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "SystemNoticeServer"
	app.Usage = fmt.Sprintf("Notice Http Api Server. version: %s", version.GetVersion())
	app.Author = "LBB"
	app.Email = "libingbing@taiyouxi.net"

	cmds.InitCommands(&app.Commands)

	app.Run(os.Args)
}
