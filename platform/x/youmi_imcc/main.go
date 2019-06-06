package main

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/rules"
)

/*
	it will discard the im data after restart server
*/

func main() {

	logs.LoadLogConfig(base.LogConfPath)

	_, err := toml.DecodeFile(base.ConfPath, &base.Cfg)
	if err != nil {
		logs.Error("load config error by %v", err)
		return
	}
	logs.Debug("load config: %v", base.Cfg)
	rules.Reg()

	time.LoadLocation("Asia/Shanghai")

	engine := gin.Default()
	engine.POST(base.Cfg.Url, acceptHandle)
	go engine.RunTLS(base.Cfg.Host, base.ServerCRT, base.ServerKEY)

	startService()
}
