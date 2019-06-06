package main

import (
	"os"

	"flag"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/bi_tools/hourly_statis/imp"
)

const (
	logConf  = "conf/log.xml"
	confFile = "conf.toml"
	gidInfo  = "gidinfo.toml"
)

func main() {
	param_time := flag.String("t", "", "2006-01-02 15")
	flag.Parse()

	imp.TimeParam = *param_time
	logs.Debug("param t %s", imp.TimeParam)

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

	if err := imp.Cfg.Init(); err != nil {
		logs.Critical("Config Init Error %s", err.Error())
		logs.Close()
		os.Exit(1)
	}
	logs.Info("Config loaded %v", imp.Cfg)

	var gidConfig struct {
		GidConfig game.GidConfig
	}
	cfgGid := config.NewConfigToml(gidInfo, &gidConfig)
	game.GidCfg = gidConfig.GidConfig
	if cfgGid == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	game.GidCfg.Init()
	logs.Info("Config gidInfo %v", game.Gid2Channel)

	imp.StartImp()

}
