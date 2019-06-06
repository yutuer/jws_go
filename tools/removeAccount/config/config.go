package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RmoveInfoConfig struct {
	RmoveInfoConfig Config
}

type Config struct {
	PerInfo []PerInfo
}

type PerInfo struct {
	Redis        string             `toml:"redis"`
	RedisDb      int                `toml:"redis_db"`
	Acid         string             `toml:"acid"`
	Password     string             `toml:"password"`
	Mat          []string           `toml:"mat"`
	Xp           []string           `toml:"xp"`
	Jade         map[string]JadeNum `toml:"jade"`
	HeroPiece    map[string]JadeNum    `toml:"heropiece"`
	VirtualMoney []string           `toml:"virtual_money"`
}

var RemoveConfig RmoveInfoConfig

type JadeNum struct {
	Number int64 `toml:"number"`
}

func LoadConfig(configpath string) {
	if _, err := toml.DecodeFile(configpath, &RemoveConfig); err != nil {
		fmt.Println(err)
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
}
