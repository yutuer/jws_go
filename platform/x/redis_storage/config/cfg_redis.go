package config

import "vcs.taiyouxi.net/platform/planx/util/storehelper"

type RedisOnlandConfig struct {
	Server      string `mapstructure:"Server"`
	Password    string `mapstructure:Password,omitempty`
	DbSeleccted int    `mapstructure:DbSeleccted`
}

func (c *RedisOnlandConfig) HasConfigured() bool {
	if c.Server == "" ||
		len(c.Server) == 0 {
		return false
	}
	return true
}

func init() {
	Register("Redis", &RedisOnlandConfig{})
}

func (c *RedisOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	addFn(storehelper.NewStoreRedis(c.Server, c.Password, c.DbSeleccted))
}
