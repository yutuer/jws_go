package main

import (
	"log"

	"github.com/BurntSushi/toml"
	"vcs.taiyouxi.net/tools/mirror_account/imp"
)

func main() {
	if _, err := toml.DecodeFile("conf/config.toml", &imp.Cfg); err != nil {
		log.Fatalf("toml err %v", err)
		return
	}
	log.Println("conf ", imp.Cfg)

	err := imp.Init()
	if err != nil {
		log.Fatalf("imp.init err %v", err.Error())
		return
	}

	err = imp.GetRank()
	if err != nil {
		log.Fatalf("imp.GetRank err %v", err.Error())
		return
	}

	err = imp.FindAccount()
	if err != nil {
		log.Fatalf("imp.FindAccount err %v", err.Error())
		return
	}
	err = imp.FindDevice()
	if err != nil {
		log.Fatalf("imp.FindDevice err %v", err.Error())
		return
	}

	imp.ShowRes()

	err = imp.WriteDynamo()
	if err != nil {
		log.Fatalf("imp.WriteDynamo err %v", err.Error())
		return
	}
}
