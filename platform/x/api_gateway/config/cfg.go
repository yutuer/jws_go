package config

import (
	"os"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

var (
	PayCfg         PayConfig
	PayServicesCfg PayServices
)

type PayConfig struct {
	SdkHttpPort      string `toml:"sdk_http_port"`
	Dynamo_Pay_Index string `toml:"dynamo_db_pay_index"`
	SentryDSN        string `toml:"SentryDSN"`
}

type PaySDK struct {
	SdkUrlRelPath string   `toml:"sdk_url_rel_path"`
	Params        []string `toml:"params"`
}

type PayService struct {
	Name string              `toml:"name"`
	Sdk  map[string]PaySDK   `toml:"SDK"`
	Mail map[string]DBConfig `toml:"Mail"`
	Pay  map[string]DBConfig `toml:"Pay"`
}

type PayServices struct {
	Services []PayService `toml:"PayService"`
}

type DBConfig struct {
	DBName                 string `toml:"db_name"`
	Dynamo_Region          string `toml:"dynamo_region"`
	Dynamo_AccessKeyID     string `toml:"dynamo_accessKeyID"`
	Dynamo_SecretAccessKey string `toml:"dynamo_secretAccessKey"`
	MongoDBUrl             string `toml:"mongo_url"`
	DBDriver               string `toml:"db_driver"`
}

func (dbc DBConfig) GetMailConfig() mailhelper.MailConfig {
	return mailhelper.MailConfig{
		AWSRegion:    dbc.Dynamo_Region,
		DBName:       dbc.DBName,
		AWSAccessKey: dbc.Dynamo_AccessKeyID,
		AWSSecretKey: dbc.Dynamo_SecretAccessKey,
		MongoDBUrl:   dbc.MongoDBUrl,
		DBDriver:     dbc.DBDriver,
	}
}

func (dbc DBConfig) GetPayDBConfig() pay.PayDBConfig {
	return pay.PayDBConfig{
		AWSRegion:    dbc.Dynamo_Region,
		DBName:       dbc.DBName,
		AWSAccessKey: dbc.Dynamo_AccessKeyID,
		AWSSecretKey: dbc.Dynamo_SecretAccessKey,
		MongoDBUrl:   dbc.MongoDBUrl,
		DBDriver:     dbc.DBDriver,
	}
}

type MongoConfig struct {
}

func ReadConfig(c *cli.Context) {
	cfgName := c.String("config")

	var pay_cfg struct{ PayConfig PayConfig }
	cfgApp := config.NewConfigToml(cfgName, &pay_cfg)
	PayCfg = pay_cfg.PayConfig
	if cfgApp == nil {
		logs.Critical("PayConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	cfgApp = config.NewConfigToml(cfgName, &PayServicesCfg)
	if cfgApp == nil {
		logs.Critical("PayServices Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Config Loaded PayServices:%v", PayServicesCfg)
	logs.Info("Config Loaded PayCfg:%v", PayCfg)
}
