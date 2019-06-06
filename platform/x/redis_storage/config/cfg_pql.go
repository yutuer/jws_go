package config

import "vcs.taiyouxi.net/platform/planx/util/storehelper"

type PostgerSQLOnlandConfig struct {
	Host string `mapstructure:"Host"`
	Port int    `mapstructure:"Port"`
	DB   string `mapstructure:"DB"`
	User string `mapstructure:"User"`
	Pass string `mapstructure:"Pass"`
}

func (c *PostgerSQLOnlandConfig) HasConfigured() bool {
	if c.Host == "" ||
		len(c.Host) == 0 {
		return false
	}
	return true
}

func init() {
	Register("PostgreSQL", &PostgerSQLOnlandConfig{})
}

func (c *PostgerSQLOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	addFn(storehelper.NewStorePostgreSQL(
		c.Host,
		c.Port,
		c.DB,
		c.User,
		c.Pass,
	))
}
