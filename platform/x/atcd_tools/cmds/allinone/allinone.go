package allinone

import (
	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/atcd_tools/atc_mng"
	"vcs.taiyouxi.net/platform/x/atcd_tools/cmds"
)

func init() {
	logs.Trace("allinone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "开启所有功能",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "config.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

type CommonConfig struct {
	Runmode    string `toml:"runmode"`
	AtcUrlHead string `toml:"AtcUrlHead"`
	DBAddress  string `toml:"DBAddress"`
	Url        string `toml:"Url"`
}

var CommonCfg CommonConfig

func Start(c *cli.Context) {
	cfgName := c.String("config")
	var common_cfg struct{ CommonCfg CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)
	CommonCfg = common_cfg.CommonCfg
	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	if CommonCfg.Runmode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	logs.Info("Start Tool %v", CommonCfg)

	r := gin.Default()

	atc_mng.RouteProfileGet(r, CommonCfg.AtcUrlHead)
	atc_mng.RouteGetClientAll(r, CommonCfg.DBAddress)
	atc_mng.RouteProxy(r, CommonCfg.AtcUrlHead)

	r.Static("/webui", "./public")

	r.Run(CommonCfg.Url) // listen and serve on 0.0.0.0:8080
}
