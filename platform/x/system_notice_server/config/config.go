package config

import "vcs.taiyouxi.net/platform/planx/servers"

type CommonConfig struct {
	Runmode string `toml:"runmode"`

	//EnableHttpListen string `toml:"EnableHttpListen"`
	EnableHttpTLS string `toml:"EnableHttpTLS"`
	HttpsPort     string `toml:"HttpsPort"`
	HttpCertFile  string `toml:"HttpCertFile"`
	HttpKeyFile   string `toml:"HttpKeyFile"`

	Dynamo_endpoint        string `toml:"dynamo_endpoint"`
	Dynamo_region          string `toml:"dynamo_region"`
	Dynamo_accessKeyID     string `toml:"dynamo_accessKeyID"`
	Dynamo_secretAccessKey string `toml:"dynamo_secretAccessKey"`
	Dynamo_sessionToken    string `toml:"dynamo_sessionToken"`

	Httpport string `toml:"httpport"`

	SentryDSN string `toml:"SentryDSN"`

	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`

	StoreDriver  string `toml:"store_driver"`
	OSSEndPoint  string `toml:"oss_endpoint"`
	OSSAccessId  string `toml:"oss_access_id"`
	OSSAccessKey string `toml:"oss_access_key"`
	OSSBucket    string `toml:"oss_bucket"`

	WatchKey          string `toml:"watch_key"`
	S3_Buckets_Notice string `toml:"s3_bucket_notice"`

	NAcceptor    uint               `toml:"acceptors"`
	NWaitingConn uint               `toml:"waiting_queue"`
	MaxConn      uint               `toml:"maxconn"`
	ListenTrySsl string             `toml:"listen_try_ssl"`
	SslCfg       servers.SSLCertCfg `toml:"SslCfg"`
	SslCaCfg     servers.SSLCertCfg `toml:"SslCaCfg"`
}

type SuperUidConfig struct {
	Uid []string
}

type GonggaoInfo struct {
	WhitelistPwd      string
	MaintainStartTime int64
	MaintainEndTime   int64
}

func (c *CommonConfig) IsRunModeProdAndTest() bool {
	return c.Runmode == "prod" || c.Runmode == "test"
}

func (c *CommonConfig) IsRunModeLocal() bool {
	return c.Runmode == "local"
}

func (c *SuperUidConfig) IsSuperUid(uid string) bool {
	for _, _uid := range c.Uid {
		if uid == _uid {
			return true
		}
	}
	return false
}
