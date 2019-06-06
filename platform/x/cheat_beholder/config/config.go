package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Config struct {
	Etcd_Root     string
	Etcd_Endpoint []string
	Port          string
	AndGid        []uint
}

var CommonConfig Config

func LoadConfig(configpath string) {
	if _, err := toml.DecodeFile(configpath, &CommonConfig); err != nil {
		fmt.Println(err)
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
}
