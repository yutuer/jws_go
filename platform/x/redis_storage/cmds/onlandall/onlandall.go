package onlandall

import (
	"github.com/codegangsta/cli"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds/helper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"
)

const ()

func init() {
	logs.Trace("onlandall cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "onlandall",
		Usage:  "将库中所有账号落地",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "onland.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "nowtime, t",
				Value: "",
				Usage: "bilog today time",
			},
		},
	})
}

func Start(c *cli.Context) {
	config.ReadConfig(c)

	onlandStores := helper.InitOnLandStores()
	helper.InitBackends(func(s storehelper.IStore) {
		onlandStores.Add(s)
	})

	// 启动
	go onlandStores.Start()

	if _, err := onlandStores.OnlandAll(); err != nil {
		logs.Error("onlandStores.OnlandAll err: %v", err)
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	onlandStores.CloseJobQueue()
	onlandStores.Stop()
	logs.Close()
}
