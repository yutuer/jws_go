package cmd

import (
	"os"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/gen_shard_name/new_serv"
)

func init() {
	logs.Trace("newserver cmd loaded")
	register(&cli.Command{
		Name:   "newserv",
		Usage:  "启动游戏newserver模式",
		Action: start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "new_serv.toml",
				Usage: "Gate Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func start(c *cli.Context) {
	confPath := c.String("config")
	var common_cfg struct{ NewServCfg new_serv.NewServConfig }
	cfgApp := config.NewConfigToml(confPath, &common_cfg)
	new_serv.Cfg = common_cfg.NewServCfg
	if cfgApp == nil {
		logs.Critical("NewServCfg Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("NewServ Config loaded %v", new_serv.Cfg)

	if !new_serv.Cfg.Check() {
		logs.Critical("NewServCfg Cfg.Check\n")
		logs.Close()
		os.Exit(1)
	}

	err := etcd.InitClient(new_serv.Cfg.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	new_serv.Imp()
}
