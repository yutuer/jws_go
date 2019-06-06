package config

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type SSDBOnlandConfig struct {
	Address   string `mapstructure:"Address"`
	Format    string `mapstructure:"Onland_Format,omitempty"`
	FormatSep string `mapstructure:"Onland_Format_Separator,omitempty"`
}

func (c *SSDBOnlandConfig) HasConfigured() bool {
	if c.Address == "" ||
		len(c.Address) == 0 {
		return false
	}
	return true
}

func (c *SSDBOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	ok, ip, port := mkip(c.Address)
	if ok {
		ssdb_store := storehelper.NewStoreHISSDB(
			ip,
			port,
			c.Format,
			c.FormatSep)

		if ssdb_store != nil {
			addFn(ssdb_store)
		}
	}
	logs.Info("SSDBCfg Config loaded %v.", c)
}

func init() {
	Register("SSDB", &SSDBOnlandConfig{})
}
