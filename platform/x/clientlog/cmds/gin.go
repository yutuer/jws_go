package cmds

import (
	"github.com/gin-gonic/gin"
	ucfg "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/ginhelper"
)

func MakeGinEngine() (engine *gin.Engine, exitfun func()) {
	engine = gin.New()

	engine.Use(gin.Recovery())
	engine.Use(SentryGinLog())
	//engine.Use(ginhelper.LoggerWithWriter())
	//engine.Use(ginhelper.NginxLoggerWithWriter())
	var gh gin.HandlerFunc
	var exitfunlog func()
	ucfg.NewConfig("accesslog.xml", false, func(lcfgname string, cmd ucfg.LoadCmd) {
		gh, exitfunlog = ginhelper.NgixLoggerToFile(lcfgname)
	})
	exitfun = func() {
		if exitfunlog != nil {
			exitfunlog()
		}
	}
	engine.Use(gh)
	//engine.Use(limit.CheckIdentity())
	//engine.Use(limit.RateLimit())
	//engine.Use(statsCCU())
	return engine, exitfun
}
