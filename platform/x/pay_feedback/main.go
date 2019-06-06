package main

import (
	"os"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/pay_feedback/imp"
)

const (
	logConf  = "conf/log.xml"
	confFile = "conf.toml"
)

func main() {
	defer logs.Close()
	logs.LoadLogConfig(logConf)

	var common_cfg struct{ CommonCfg imp.CommonConfig }
	cfgApp := config.NewConfigToml(confFile, &common_cfg)
	imp.Cfg = common_cfg.CommonCfg
	if cfgApp == nil {
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("load config %v", imp.Cfg)

	if imp.Cfg.CorrectFromRedis {
		driver.SetupRedis(
			imp.Cfg.Redis,
			imp.Cfg.RedisDB,
			imp.Cfg.RedisAuth,
			false,
		)
	}
	if err := imp.Run(); err != nil {
		logs.Critical("Run Error %v", err)
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Success !")
}
