package allinone

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gamex_merge/cmds"
	"vcs.taiyouxi.net/platform/x/gamex_merge/merge"
)

func init() {
	logs.Trace("allinone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "启动游戏AllInOne模式",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "merge.toml",
				Usage: "merge Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func Start(c *cli.Context) {
	cfgName := c.String("config")

	var common_cfg struct{ MergeCfg merge.MergeConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)
	merge.Cfg = common_cfg.MergeCfg
	if cfgApp == nil {
		logs.Critical("MergeConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Merge Config loaded %v", merge.Cfg)

	err := etcd.InitClient(merge.Cfg.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	if !merge.Cfg.Init() {
		logs.Error("merge.Cfg.Init fail")
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	merge.RedisInit()

	// 开始合服
	st := time.Now()
	merge.Prepare()
	logs.Info("Prepare %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.MergeRoleName(); err != nil {
		logs.Error("MergeRoleName err %s", err.Error())
		return
	}
	logs.Info("MergeRoleName %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.GuildSimpleMerge(); err != nil {
		logs.Error("GuildSimpleMerge err %s", err.Error())
		return
	}
	logs.Info("GuildSimpleMerge %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.MergeGuildName(); err != nil {
		logs.Error("MergeGuildName err %s", err.Error())
		return
	}
	logs.Info("MergeGuildName %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.DestGenFirstMerge(); err != nil {
		logs.Error("DestGenFirstMerge err %s", err.Error())
		return
	}
	logs.Info("DestGenFirstMerge %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.MergeTeamPvp(); err != nil {
		logs.Error("MergeTeamPvp err %s", err.Error())
		return
	}
	logs.Info("MergeTeamPvp %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.MergeRank(); err != nil {
		logs.Error("MergeRank err %s", err.Error())
		return
	}
	logs.Info("MergeRank %v", time.Now().Sub(st).Seconds())

	st = time.Now()
	if err := merge.MergeGVG(); err != nil {
		logs.Error("MergeGVG err %s", err.Error())
		return
	}
	logs.Info("MergeGVG %v", time.Now().Sub(st).Seconds())

	merge.OutPutRes()

	logs.Info("Merge Finish !!!")
}
