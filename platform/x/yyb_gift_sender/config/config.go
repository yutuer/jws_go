package config

import (
	"os"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Config struct {
	GiftUrl             string   `toml:"GiftUrl"`
	ShardUrl            string   `toml:"ShardUrl"`
	Host                string   `toml:"Host"`
	Appkey              string   `toml:"Appkey"`
	EtcdRoot            string   `toml:"etcd_root"`
	EtcdEndPoint        []string `toml:"etcd_endpoint"`
	GiftMongoDBUrl      string   `toml:"GiftMongoDBUrl"`
	MailMongoDBUrl      string   `toml:"MailMongoDBUrl"`
	AuthMongoDBUrl      string   `toml:"AuthMongoDBUrl"`
	GiftDBName          string   `toml:"GiftDBName"`
	MailDBName          string   `toml:"MailDBName"`
	AuthDBName          string   `toml:"AuthDBName"`
	DeviceSuffix        string   `toml:"DeviceSuffix"`
	AppID               string   `toml:"AppID"`
	RechargeEtcdRoot    string   `toml:"recharge_etcd_root"`
	DynamoRegion        string   `toml:"dynamo_region"`
	DynamoAccessKey     string   `toml:"dynamo_accessKeyID"`
	DynamoSecretAccess  string   `toml:"dynamo_secretAccessKey"`
	DynamoSessionToken  string   `toml:"dynamo_sessionToken"`
	DB                  string   `toml:"db"`
	Dynamo_NameDevice   string   `toml:"dynamo_db_Device"`
	Dynamo_NameName     string   `toml:"dynamo_db_Name"`
	Dynamo_NameUserInfo string   `toml:"dynamo_db_UserInfo"`
	Dynamo_GM           string   `toml:"dynamo_db_GM"`
}

var CommonConfig Config

func LoadConfig(confStr string) {
	var common_cfg struct{ CommonConfig Config }
	cfgApp := config.NewConfigToml(confStr, &common_cfg)

	CommonConfig = common_cfg.CommonConfig

	if cfgApp == nil {
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
}
