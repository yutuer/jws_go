package restoreall

import (
	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/platform/x/redis_storage/cmds"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds/helper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {
	logs.Trace("restoreall cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "restoreall",
		Usage:  "回档对应数据源中的所有用户存档",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "onland.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func Start(c *cli.Context) {
	config.ReadConfig(c)

	restorer := helper.InitRestore()
	if restorer == nil {
		return
	}
	// 启动
	go restorer.Start()

	logs.Info("restore all")

	err := restorer.RestoreAll()
	if err != nil {
		logs.Error("restore all err %s", err.Error())
	}

	restorer.Stop()
	logs.Close()

}
