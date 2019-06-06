package cmds

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/ginhelper"
	"vcs.taiyouxi.net/platform/x/api_gateway/config"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func MakeGinEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	ginhelper.InitSentryTags(
		config.PayCfg.SentryDSN,
		map[string]string{"service": "paygin"})

	logs.InitSentryTags(
		config.PayCfg.SentryDSN,
		map[string]string{"service": "paylog"})

	engine.Use(ginhelper.SentryGinLog())

	return engine
}
