package config

type CommonConfig struct {
	Runmode          string   `toml:"run_mode"`
	Listen           string   `toml:"listen"`
	EtcdEndpoint     []string `toml:"etcd_endpoint"`
	EtcdRoot         string   `toml:"etcd_root"`
	ListenNotifyAddr string   `toml:"listen_notify_addr"`
	PublicIP         string   `toml:"publicip"`
	GID              string   `toml:"gid"`
	MatchToken       string   `toml:"match_token"`
}

var (
	Cfg CommonConfig
)
