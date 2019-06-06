package cmds

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/ginhelper"
)

func InitSentry(DSN string) error {
	return ginhelper.InitSentry(DSN)
}

func SentryGinLog() gin.HandlerFunc {
	return ginhelper.SentryGinLog()
}
