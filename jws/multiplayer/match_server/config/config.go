package config

type CommonConfig struct {
	Runmode      string   `toml:"run_mode"`
	Url          string   `toml:"url"`
	EtcdEndpoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`

	MatchTimeoutSecond int      `toml:"matchTimeoutSecond"`
	TickMax            int      `toml:"tickMax"`
	NewEnterMatchLv    uint32   `toml:"newEnterMatchLv"`
	MatchTicks         []int    `toml:"matchTicks"`
	MatchLvs           []uint32 `toml:"matchLvs"`
	GID                string   `toml:"gid"`

	WaitRobotTimeMax int `toml:"waitRobotTimeMax"`
	WaitRobotTimeMin int `toml:"waitRobotTimeMin"`
}

var (
	Cfg CommonConfig
)
