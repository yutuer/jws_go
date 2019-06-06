package verupdateurl

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/x/auth/cmds"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"

	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/auth/routers"
)

func init() {
	logs.Trace("verupdateurl cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "verupdateurl",
		Usage:  "verupdateurl",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "app.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

var CommonCfg authConfig.CommonConfig

func Start(c *cli.Context) {
	cfgName := c.String("config")
	var common_cfg struct{ CommonConfig authConfig.CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)

	CommonCfg = common_cfg.CommonConfig
	authConfig.Cfg = common_cfg.CommonConfig

	var limit_cfg struct{ LimitConfig limit.LimitConfig }
	cfgLimit := config.NewConfigToml(cfgName, &limit_cfg)

	iplimitconfig.LoadIPRangeConfig(cfgName, func(ilcfg []iplimitconfig.IPRangeConfig) {
		limit.SetIPLimitCfg(ilcfg)
	})
	limit.SetLimitCfg(limit_cfg.LimitConfig)

	if cfgApp == nil || cfgLimit == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	var waitGroup util.WaitGroupWrapper
	limit.Init()

	if CommonCfg.IsRunModeProdAndTest() {
		gin.SetMode(gin.ReleaseMode)
	}

	logs.Debug("Start Auth Server %v", CommonCfg)

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	waitGroup.Wrap(func() {
		metrics.Start("metrics.toml")
	})

	cmds.InitSentry(CommonCfg.SentryDSN)

	r, exitfun := cmds.MakeGinEngine()
	defer exitfun()
	routers.RegVerUpdateUrl(r)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	if CommonCfg.EnableHttpTLS != "" {
		if err := r.RunTLS(CommonCfg.HttpsPort,
			CommonCfg.HttpCertFile,
			CommonCfg.HttpKeyFile); err != nil {
			logs.Close()
			os.Exit(1)
		}
	} else {
		if err := r.Run(CommonCfg.Httpport); err != nil { // listen and serve on 0.0.0.0:8081
			logs.Close()
			os.Exit(1)
		}
	}

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()
}
