package new_serv

import (
	"strconv"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	all_shard = "all"
)

type NewServConfig struct {
	Shards       []string `toml:"shards"`
	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`
}

func (c *NewServConfig) Check() bool {
	if len(c.Shards) <= 0 {
		logs.Error("cfg shard id is empty")
		return false
	}
	for i, sid := range c.Shards {
		if i == 0 && strings.ToLower(sid) == all_shard {
			return true
		}
		if _, err := strconv.Atoi(sid); err != nil {
			logs.Error("cfg shard id not number %s", c.Shards)
			return false
		}
	}
	return true
}

var Cfg NewServConfig
