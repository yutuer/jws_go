package auth

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/auth/cmds"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/auth/routers"
)

func init() {
	logs.Trace("auth cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "auth",
		Usage:  "Auth",
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

	var sdk_hero_cfg struct{ SdkHeroConfig authConfig.SdkHeroConfig }
	cfgSdkHero := config.NewConfigToml(cfgName, &sdk_hero_cfg)

	authConfig.HeroSdkCfg = sdk_hero_cfg.SdkHeroConfig

	var sdk_quick_cfg struct{ SdkQuickConfig authConfig.SdkQuickConfig }
	cfgSdkQuick := config.NewConfigToml(cfgName, &sdk_quick_cfg)
	authConfig.QuickSdkCfg = sdk_quick_cfg.SdkQuickConfig

	var sdk_vivo_cfg struct{ SdkVivoConfig authConfig.SdkVivoConfig }
	cfgSdkVivo := config.NewConfigToml(cfgName, &sdk_vivo_cfg)
	authConfig.VivoSdkCfg = sdk_vivo_cfg.SdkVivoConfig

	var sdk_6waves_cfg struct{ Sdk6wavesConfig authConfig.Sdk6wavesConfig }
	cfgSdk6waves := config.NewConfigToml(cfgName, &sdk_6waves_cfg)
	authConfig.WavesSdkCfg = sdk_6waves_cfg.Sdk6wavesConfig

	iplimitconfig.LoadIPRangeConfig(cfgName, func(ilcfg []iplimitconfig.IPRangeConfig) {
		limit.SetIPLimitCfg(ilcfg)
	})
	var limit_cfg struct{ LimitConfig limit.LimitConfig }
	cfgLimit := config.NewConfigToml(cfgName, &limit_cfg)

	limit.SetLimitCfg(limit_cfg.LimitConfig)

	if cfgApp == nil || cfgSdkHero == nil || cfgSdkQuick == nil ||
		cfgSdkVivo == nil || cfgLimit == nil || cfgSdk6waves == nil {
		logs.Critical("Config Read Error\n")
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
	authConfig.InitAuthMetrics()

	if err := models.InitDynamo(&CommonCfg); err != nil {
		logs.Critical("InitDynamo Error\n")
		logs.Close()
		os.Exit(1)
	}

	cmds.InitSentry(CommonCfg.SentryDSN)

	// client ver update conf
	signalhandler.SignalKillFunc(func() { authConfig.Stop() })
	waitGroup.Wrap(func() {
		authConfig.Start()
	})

	r, exitfun := cmds.MakeGinEngine()
	defer exitfun()
	routers.RegAuth(r)
	routers.RegVerUpdateUrl(r)

	gm := cmds.MakeGinGMEngine()
	routers.RegAuthGMCommand(gm)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	go func() {
		if CommonCfg.EnableHttpTLS != "" {
			if err := r.RunTLS(CommonCfg.HttpsPort,
				CommonCfg.HttpCertFile,
				CommonCfg.HttpKeyFile); err != nil {
				logs.Close()
				os.Exit(1)
			}
		} else {
			if err := r.Run(CommonCfg.Httpport); err != nil { // listen and serve on 0.0.0.0:8080
				logs.Close()
				os.Exit(1)
			}
		}
	}()

	go func() {
		if err := gm.Run(CommonCfg.GMHttpport); err != nil { // listen and serve on 0.0.0.0:8789
			logs.Close()
			os.Exit(1)
		}
	}()

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()
	logs.Close()
}
