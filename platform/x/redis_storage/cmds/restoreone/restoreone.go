package restoreone

import (
	"strings"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/platform/x/redis_storage/cmds"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds/helper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {
	logs.Trace("restoreone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "restoreone",
		Usage:  "回档单个用户",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "onland.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "user, u",
				Value: "",
				Usage: "user key in s3 or dynamodb: {user1},{user2},...,{userN}",
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

	users := c.String("user")
	logs.Info("restore %s", users)

	u := strings.Split(users, ",")
	if len(u) > 0 {
		for _, user := range u {
			restorer.RestoreOne(user, nil)
		}
	}

	restorer.Stop()
	logs.Close()

}
