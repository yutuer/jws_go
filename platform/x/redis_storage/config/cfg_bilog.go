package config

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/jws/gamex/bdclog"
)

type RedisBiLogConfig struct {
	GidInfo []game.GidInfo `mapstructure:"GidInfo"`
}

func (c *RedisBiLogConfig) HasConfigured() bool {
	return true
}

func init() {
	Register("RedisBiLog", &RedisBiLogConfig{})
}

func (c *RedisBiLogConfig) Setup(addFn func(storehelper.IStore)) {
	ss := make([]string, len(CommonCfg.ShardId))
	for i, sid := range CommonCfg.ShardId {
		ss[i] = fmt.Sprintf("%d", sid)
	}
	biLog_store := bdclog.NewBiScanRedis(CommonCfg.Gid,
		ss, Cfg_Time, c.GidInfo)
	addFn(biLog_store)
}
