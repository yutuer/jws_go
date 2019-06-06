package routers

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/controllers"
	"vcs.taiyouxi.net/platform/x/auth/limit"
)

func RegVerUpdateUrl(g *gin.Engine) {
	device_id_controller := controllers.DeviceIDController{}
	// 客户端版本强制更新用
	ver_update := g.Group("/verupdate/v1/")
	ver_update.Use(limit.RateLimit())
	ver_update.GET("getUrl", device_id_controller.UpdateVer())

	tools := g.Group("/tools/")
	tools.Use(limit.RateLimit())
	tools.GET("echoip", device_id_controller.EchoIp())

	debug := g.Group("/debug/")
	debug.Use(limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content), limit.RateLimit())
	debug.POST("log", device_id_controller.DebugLog())
}
