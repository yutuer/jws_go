package onlandone

import (
	"strings"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/platform/x/redis_storage/cmds"
	"vcs.taiyouxi.net/platform/x/redis_storage/cmds/helper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

const ()

func init() {
	logs.Trace("onlandone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "onlandone",
		Usage:  "落地库中指定账号",
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
				Usage: "user key in redis to onland: {user1},{user2},...,{userN}",
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

	users := c.String("user")
	logs.Info("onland %s", users)

	u := strings.Split(users, ",")
	if len(u) > 0 {
		for _, user := range u {
			onlandStores.NewKeyDumpJob(user)
		}
	}

	//time.Sleep(time.Second * 5)
	onlandStores.CloseJobQueue()
	onlandStores.Stop()

	logs.Close()

}
