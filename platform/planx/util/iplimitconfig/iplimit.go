package iplimitconfig

import (
	"github.com/BurntSushi/toml"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type IPRangeConfig struct {
	From string `toml:"from"`
	To   string `toml:"to"`
}

type IPLimits struct {
	IPLimit []IPRangeConfig
}

func LoadIPRangeConfig(ipcfg string, reloadFunc func([]IPRangeConfig)) {
	var limits IPLimits

	config.NewConfig(ipcfg, true, func(lcfgname string, loadStatus config.LoadCmd) {
		switch loadStatus {
		case config.Load, config.Reload:
			if _, err := toml.DecodeFile(lcfgname, &limits); err != nil {
				logs.Critical("App config load failed. %s, %s\n", lcfgname, err.Error())
			} else {
				logs.Info("IPLimit Config loaded: %s\n", lcfgname)
				if reloadFunc != nil {
					reloadFunc(limits.IPLimit[:])
				}
			}
		case config.Unload:

		}
	})
}
