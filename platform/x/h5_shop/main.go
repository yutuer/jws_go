package main

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	h5Config "vcs.taiyouxi.net/platform/x/h5_shop/config"
	"vcs.taiyouxi.net/platform/x/h5_shop/sdk_shop"

	"vcs.taiyouxi.net/platform/x/gift_sender/config"
)

func main() {
	h5Config.LoadConfig("conf/config.toml")
	logs.Debug("%v", h5Config.CommonConfig)
	err := etcd.InitClient(h5Config.CommonConfig.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		return
	}
	authConfig.Cfg.Runmode = "test"
	authConfig.Cfg.EtcdRoot = h5Config.CommonConfig.EtcdRoot
	config.CommonConfig.EtcdRoot = h5Config.CommonConfig.EtcdRoot
	logs.Debug("load from h5shop config %v", h5Config.CommonConfig)
	gin := gin.Default()
	gin.POST("user/costdiamond.html", sdk_shop.CostDiamond)
	gin.Run(h5Config.CommonConfig.Port)
}
