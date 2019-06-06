package main

import (
	"os"

	//A visual interface to work with runtime profiling data from Go programs.
	//instead of pprof/http
	"fmt"

	_ "github.com/rakyll/gom/http"
	_ "vcs.taiyouxi.net/platform/x/chat/cmds/allinone"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/chat/cmds"
)

const (
	logConf = "conf/log.xml"
)

/*
	chattown配套机器人工具请使用 titools/chatroomrobot/client
*/
func main() {
	defer logs.Close()
	logs.LoadLogConfig(logConf)

	util.PProfStart()

	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "Chat"
	app.Usage = fmt.Sprintf("Chat Http Api Server. version: %s", version.GetVersion())
	app.Author = "YZH"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	app.Run(os.Args)
}
