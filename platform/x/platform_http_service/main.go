package main

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util"
	pConfig "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/platform_http_service/config"
	"vcs.taiyouxi.net/platform/x/platform_http_service/db"
	gin2 "vcs.taiyouxi.net/platform/x/platform_http_service/gin"
)

func main() {
	defer logs.Close()

	logxml := pConfig.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	config.LoadConfig("config.toml")
	config.InitEtcd(config.CommonConfig.EtcdCfg.EtcdEndPoint)

	config.LoadGameData("")

	db.InitDynamoDB()

	waiter := util.WaitGroupWrapper{}
	for _, cfg := range config.CommonConfig.PlatformConfigArray {
		waiter.Wrap(func() {
			gin := gin.Default()
			plat := gin2.NewPlatform(cfg)
			plat.RegHandler(gin)
			gin.Run(cfg.Host)
		})
	}
	waiter.Wait()
}
