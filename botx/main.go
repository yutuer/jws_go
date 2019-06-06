package main

import (
	"fmt"
	"os"

	//"net/http"
	//_ "net/http/pprof"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/botx/cmds"
	_ "vcs.taiyouxi.net/botx/cmds/abot"
	_ "vcs.taiyouxi.net/botx/cmds/mbot"

	"github.com/codegangsta/cli"
)

func main() {
	defer logs.Close()

	//go func() {
	//http.ListenAndServe("localhost:6661", nil)
	//}()

	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "botx"
	app.Usage = fmt.Sprintf("Ticore bot for load test. version:%s.", GetVersion())
	app.Author = "YinZeHong"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	app.Run(os.Args)
}
