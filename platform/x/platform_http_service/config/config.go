package config

import (
	"os"
	"path/filepath"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Config struct {
	DynamoConfig        DynamoDBConfig   `toml:"DynamoConfig"`
	PlatformConfigArray []PlatformConfig `toml:"Platform"`
	EtcdCfg             EtcdConfig       `toml:"EtcdConfig"`
	Gid                 string           `toml:"gid"`
	MarketVersion       string           `toml:"market_version"`
}

type DynamoDBConfig struct {
	AWS_Region       string `toml:"dynamo_region"`
	AWS_KeyId        string `toml:"dynamo_accessKeyID"`
	AWS_AccessKey    string `toml:"dynamo_secretAccessKey"`
	AWS_SessionToken string `toml:"dynamo_sessionToken"`
	AWS_DeviceName   string `toml:"dynamo_deviceName"`
}

type PlatformConfig struct {
	Name string `toml:"name"`
	Host string `toml:"host"`
	Key  string `toml:"key"`
}

type EtcdConfig struct {
	EtcdRoot     string   `toml:"etcd_root"`
	EtcdEndPoint []string `toml:"etcd_endpoint"`
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
	logs.Debug("load config ok %v", CommonConfig)
}

func InitEtcd(endPoints []string) {
	err := etcd.InitClient(endPoints)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
}

func LoadGameData(rootPath string) {
	dataRelPath := "data"
	dataAbsPath := filepath.Join(gamedata.GetDataPath(), dataRelPath)

	load := func(dfilepath string, loadfunc func(string)) {
		loadfunc(filepath.Join(rootPath, dataAbsPath, dfilepath))
		logs.Trace("LoadGameData %s success", dfilepath)
	}

	load("iapmain.data", gamedata.LoadIapMainConfig)
	logs.Debug("load pay data, %v", gamedata.GetIAPIdxMap())
}
