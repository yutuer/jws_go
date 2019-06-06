package gin

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/platform_http_service/config"
)

type Platform interface {
	RegHandler(gin *gin.Engine)
	GetConfig() config.PlatformConfig
}

func NewPlatform(pCfg config.PlatformConfig) Platform {
	switch pCfg.Name {
	case "EG":
		return &EGPlatform{cfg: pCfg}
	}
	return nil
}
