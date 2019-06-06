package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"runtime"

	_ "net/http/pprof"

	_ "vcs.taiyouxi.net/jws/gamex/cmds/allinone"
	_ "vcs.taiyouxi.net/jws/gamex/cmds/gamemode"
	_ "vcs.taiyouxi.net/jws/gamex/cmds/gatemode"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"

	"vcs.taiyouxi.net/jws/gamex/cmds"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

const running = "RUNNING"

func main() {
	defer logs.Close()

	f, err := os.OpenFile(running, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		f.Close()
	}

	app := cli.NewApp()
	app.Version = version.GetVersion()
	app.Name = "gamex"
	app.Usage = fmt.Sprintf("Ticore game company game server. gate ver(%s), game ver(%s)\n %s \n data ver %s \n",
		gate.GetVersion(), game.GetVersion(), version.GetVersion(), gamedata.GetProtoDataVer())
	app.Author = "YinZeHong"
	app.Email = "yinzehong@taiyouxi.cn"

	cmds.InitCommands(&app.Commands)

	config.NewConfig("log.xml", true, func(lcfgname string, cmd config.LoadCmd) {
		logs.LoadLogConfig(lcfgname)
	})
	logs.Info("GOMAXPROCS is %d", runtime.GOMAXPROCS(0))
	logs.Info("Version is %s", app.Version)
	//runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	app.Run(os.Args)

	os.Remove(running)
}
