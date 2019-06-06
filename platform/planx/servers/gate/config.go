package gate

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/gate/ccumetrics"
	"vcs.taiyouxi.net/platform/planx/servers/gate/rpc"
)

// Service ...
//type Service struct {
//Name string `toml:"name"`
//IP   string `toml:"ip"`
//}

// Config ...
type BasicCfg struct {
	//请指定内网公网IP，必须完整ip:port
	RunMode      string             `toml:"run_mode"`
	PublicIP     string             `toml:"publicip"`
	Listen       string             `toml:"listen"`
	MaxConn      uint               `toml:"maxconn"`
	GameServers  []string           `toml:"gameservers"`
	ShardID      []uint             `toml:"shard_id"`
	MergeRel     []uint             `toml:"merge_rel"`
	ConnServer   string             `toml:"connetion_server"`
	NAcceptor    uint               `toml:"acceptors"`
	NWaitingConn uint               `toml:"waiting_queue"`
	SslCfg       servers.SSLCertCfg `toml:"SslCfg"`
	SslCaCfg     servers.SSLCertCfg `toml:"SslCaCfg"`
}

type Config struct {
	GateConfig       BasicCfg
	RPCConfig        rpc.Config
	CCUMetricsConfig ccumetrics.Config
}

const (
	RunModeLocal = "local"
	//RunModeDev 开发模式
	RunModeDev = "dev"
	// RunModeTest 测试模式
	RunModeTest = "test"
	//RunModeProd 生产模式
	RunModeProd = "prod"
)

//IsRunModeProd 是否当前处于生产模式
func (c *Config) IsRunModeProd() bool {
	return c.GateConfig.RunMode == RunModeProd
}

func (c *Config) IsRunModeLocal() bool {
	return c.GateConfig.RunMode == RunModeLocal
}
