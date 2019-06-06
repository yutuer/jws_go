package config

//CommonConfig ..
type CommonConfig struct {
	RunMode          string `toml:"run_mode"`
	PublicIP         string `toml:"publicip"`
	PublicPort       int    `toml:"publicport"`
	InternalHTTPPort int    `toml:"internal_http_port"`

	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`

	DBRedis       string `toml:"db_redis"`
	DBRedisDB     int    `toml:"db_redis_db"`
	DBRedisAuth   string `toml:"db_redis_Auth"`
	DBKeyEtcdPath string `toml:"db_key_etcd_path"`

	DiscoverRedis       string `toml:"discover_redis"`
	DiscoverRedisDB     uint32 `toml:"discover_redis_db"`
	DiscoverRedisAuth   string `toml:"discover_redis_Auth"`
	DiscoverKeyEtcdPath string `toml:"discover_key_etcd_path"`

	Gid      uint32   `toml:"gid"`
	GroupIDs []uint32 `toml:"group_ids"`

	ShardRange   [][]uint32 `toml:"shard_range"`
	ExclusionNum uint32     `toml:"exclusion_num"`

	DSN string `toml:"DSN"`

	IPFilter []string `toml:"ip_filter"`
}

//..
var (
	Cfg   CommonConfig
	Index string
)

//IsDevMode ..
func IsDevMode() bool {
	return Cfg.RunMode == "dev"
}

//IsDevProd ..
func IsDevProd() bool {
	return Cfg.RunMode == "prod"
}

//SetIndex ..
func SetIndex(i string) {
	Index = i
}

//GetIndex ..
func GetIndex() string {
	return Index
}
