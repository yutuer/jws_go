package main

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"

	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/db"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/gin_core"
)

func main() {
	defer logs.Close()
	// need BIlog?
	//logs.LoadLogConfig(logConf)

	config.LoadConfig("config.toml")

	config.LoadGiftData(config.GetDataPath() + "/yybgift.data")
	logs.Debug("Config: %v", config.CommonConfig)

	config.InitErrCode()

	db.Init()

	gin_core.Reg()

	gin_core.Run()
}
