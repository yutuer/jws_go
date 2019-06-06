package imp

import "time"

const gameId = "134"

type CommonConfig struct {
	ElasticUrl string
	Index      string
	TimeLocal  string `toml:"time_local"`
	LogPrefix  string `toml:"log_prefix"`
	LogOutput  string `toml:"log_output"`
	Graphite   string
	Gid2Shard  map[string]ShardRange `toml:"Gid2Shard"`
}

type ShardRange struct {
	From int `toml:"From"`
	To   int `toml:"To"`
}

var (
	Cfg       CommonConfig
	TimeLocal *time.Location
	TimeParam string
)

func (c *CommonConfig) Init() error {
	l, err := time.LoadLocation(c.TimeLocal)
	if err != nil {
		return err
	}
	TimeLocal = l
	return nil
}
