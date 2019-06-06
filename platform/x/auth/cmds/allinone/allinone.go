package allinone

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	ucfg "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/auth/cmds"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/auth/routers"
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
				Value: "app.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "iplimit, ip",
				Value: "iplimit.toml",
				Usage: "IpLimit Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "superuid, uid",
				Value: "superuid.toml",
				Usage: "SuperUid Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "logiclog, ll",
				Value: "logiclog.xml",
				Usage: "log player logic logs, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

var CommonCfg authConfig.CommonConfig
var IPLimit iplimitconfig.IPLimits
var SuperUidCfg authConfig.SuperUidConfig

func Start(c *cli.Context) {
	var logiclogName string = c.String("logiclog")
	ucfg.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

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

	var sdk_enjoy_cfg struct{ SdkEnjoyConfig authConfig.SdkEnjoyConfig }
	cfgSdkEnjoy := config.NewConfigToml(cfgName, &sdk_enjoy_cfg)
	authConfig.EnjoySdkCfg = sdk_enjoy_cfg.SdkEnjoyConfig

	var sdk_vn_cfg struct{ SdkVNConfig authConfig.SdkVNConfig }
	cfgSdkVN := config.NewConfigToml(cfgName, &sdk_vn_cfg)
	authConfig.VNSdkCfg = sdk_vn_cfg.SdkVNConfig

	var sdk_6waves_cfg struct{ Sdk6WavesConfig authConfig.Sdk6wavesConfig }
	cfgSdk6waves := config.NewConfigToml(cfgName, &sdk_6waves_cfg)
	authConfig.WavesSdkCfg = sdk_6waves_cfg.Sdk6WavesConfig

	logs.Debug("sdk conf %v %v %v %v %v", authConfig.QuickSdkCfg, authConfig.VivoSdkCfg,
		authConfig.EnjoySdkCfg, authConfig.VNSdkCfg, authConfig.WavesSdkCfg)

	// iplimit
	iplimitconfig.LoadIPRangeConfig(cfgName, func(ilcfg []iplimitconfig.IPRangeConfig) {
		limit.SetIPLimitCfg(ilcfg)
	})
	var limit_cfg struct{ LimitConfig limit.LimitConfig }
	cfgLimit := config.NewConfigToml(cfgName, &limit_cfg)

	ipCfgName := c.String("iplimit")
	config.NewConfigToml(ipCfgName, &IPLimit)

	limit.SetIPLimitCfg(IPLimit.IPLimit)
	logs.Debug("IpLimit cfg %v", IPLimit)

	limit.SetLimitCfg(limit_cfg.LimitConfig)

	// super uid
	suidCfgName := c.String("superuid")
	cfgUid := config.NewConfigToml(suidCfgName, &SuperUidCfg)
	logs.Debug("SuperUid %v", SuperUidCfg)

	if cfgApp == nil || cfgSdkHero == nil || cfgSdkQuick == nil ||
		cfgSdkVivo == nil || cfgSdkEnjoy == nil || cfgLimit == nil || cfgUid == nil || cfgSdkVN == nil || cfgSdk6waves == nil {
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	limit.Init()
	authConfig.SuperUidsCfg = SuperUidCfg

	if CommonCfg.IsRunModeProdAndTest() {
		gin.SetMode(gin.ReleaseMode)
	}

	var waitGroup util.WaitGroupWrapper

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
	models.InitLoginRedis(&CommonCfg)
	err := models.InitNumberIDInRedis()
	if err != nil {
		logs.Critical("init numberID in redis err by %v", err.Error())
		logs.Close()
		os.Exit(1)
	}
	err = etcd.InitClient(CommonCfg.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient %s", err.Error())
	}

	err = models.InitShardInfo()
	if err != nil {
		logs.Critical("init shard info err by %v", err.Error())
		logs.Close()
		os.Exit(1)
	}

	cmds.InitSentry(CommonCfg.SentryDSN)

	r, exitfun := cmds.MakeGinEngine()
	defer exitfun()
	routers.RegAuth(r)
	routers.RegLogin(r)
	routers.RegVerUpdateUrl(r)

	gm := cmds.MakeGinGMEngine()
	routers.RegAuthGMCommand(gm)

	// client ver update conf
	signalhandler.SignalKillFunc(func() { authConfig.Stop() })
	waitGroup.Wrap(func() {
		authConfig.Start()
	})

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	util.PProfStart()

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
