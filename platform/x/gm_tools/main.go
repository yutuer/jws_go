package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	_ "vcs.taiyouxi.net/platform/x/gm_tools/cmds/allinone"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/gm_tools/cmds"
)

func main() {
	defer logs.Close()

	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "GM_Tools"
	app.Usage = fmt.Sprintf("GM Tool web server: version:%s", version.GetVersion())
	app.Author = "FanYang"
	app.Email = "fanyang@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	app.Run(os.Args)
}
