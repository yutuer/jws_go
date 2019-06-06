package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/x/hero_h5/serverinfo"
	//"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	h5Config "vcs.taiyouxi.net/platform/x/hero_h5/config"
	"vcs.taiyouxi.net/platform/x/hero_h5/roleinfo"
)

func main() {

	h5Config.LoadConfig("conf/config.toml")
	err := etcd.InitClient(h5Config.CommonConfig.Etcd_Endpoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("etcd Init client done.")
	authConfig.Cfg.Runmode = "test"
	authConfig.Cfg.EtcdRoot = h5Config.CommonConfig.Etcd_Root
	gin := gin.Default()
	serverinfo.RegisterAllServerInfo(gin)
	roleinfo.Init(gin)
	gin.Run(h5Config.CommonConfig.Port)
}
