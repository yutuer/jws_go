package config

import (
	"os"

	"path/filepath"

	"fmt"

	"github.com/BurntSushi/toml"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Config struct {
	Host          map[string]HostInfo   `toml:"host"`
	EtcdRoot      string                `toml:"etcd_root"`
	EtcdEndPoint  []string              `toml:"etcd_endpoint"`
	Key           string                `toml:"key"`
	AWS_Region    string                `toml:"aws_region"`
	MailDB        map[string]MailDBName `toml:"mail_db"`
	AWS_AccessKey string                `toml:"aws_access_key"`
	AWS_SecretKey string                `toml:"aws_secret_key"`
	SecretKey     string                `toml:"secret_key"`
	UCConfig      UCConfig              `toml:"uc"`
}

type MailDBName struct {
	Name string
}

type HostInfo struct {
	Host string
}

type UCConfig struct {
	Dynamo_db_Device        string `toml:"dynamo_db_Device"`
	Dynamo_db_Name          string `toml:"dynamo_db_Name"`
	Dynamo_db_UserInfo      string `toml:"dynamo_db_UserInfo"`
	Dynamo_db_GM            string `toml:"dynamo_db_GM"`
	Dynamo_db_UserShardInfo string `toml:"dynamo_db_UserShardInfo"`
	SecretKey               string `toml:"secretKey"`
	APIKey                  string `toml:"apiKey"`
	Caller                  string `toml:"caller"`
	GID                     uint   `toml:"gid"`
}

var ucGiftData map[string]ProtobufGen.UCGIFT

var CommonConfig Config

func LoadConfig(confStr string) {
	if _, err := toml.DecodeFile(confStr, &CommonConfig); err != nil {
		fmt.Println(err)
		logs.Critical("Config Read Error\n")
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

	load("item.data", gamedata.LoadItemData)
	load("ucgift.data", gamedata.LoadUCGiftData)
	load("hmtitemlist.data", gamedata.LoadHMTGiftData)
}

func GetUCGiftData(id string) *ProtobufGen.UCGIFT {
	return gamedata.GetUCGiftData(id)
}

func IsHmtGiftItem(ids []string) bool {
	for _, id := range ids {
		if !gamedata.IsHMTGiftItem(id) {
			return false
		}
	}
	return true
}
