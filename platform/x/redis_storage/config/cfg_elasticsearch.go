package config

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type ElasticOnlandConfig struct {
	ElasticUrl string `mapstructure:"Elastic_Url"`
}

func (c *ElasticOnlandConfig) HasConfigured() bool {
	if c.ElasticUrl == "" ||
		len(c.ElasticUrl) == 0 {
		return false
	}
	return true
}

func init() {
	Register("ElasticS", &ElasticOnlandConfig{})
}

func (c *ElasticOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	addFn(storehelper.NewElasticScanRedis(c.ElasticUrl,
		fmt.Sprintf("%d", CommonCfg.ShardId[0])))
}
