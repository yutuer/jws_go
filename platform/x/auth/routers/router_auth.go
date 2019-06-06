package routers

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/controllers"
	"vcs.taiyouxi.net/platform/x/auth/limit"
)

func statsCCU() gin.HandlerFunc {
	return func(c *gin.Context) {
		config.AddAuthCCUCount()
	}
}
func RegAuth(g *gin.Engine) {
	device_id_controller := controllers.DeviceIDController{}
	user_controller := controllers.UserController{}

	auth_public := g.Group("/auth/v1/")
	auth_public.Use(limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content), limit.RateLimit(), statsCCU())

	// 暂不用
	auth_public.GET("device/:id", device_id_controller.RegDeviceAndLogin())
	auth_public.GET("deviceWithCode/:id", device_id_controller.RegDeviceWithCodeAndLogin())

	// unity3d登录
	auth_public.GET("user/login", user_controller.LoginAsUser())
	// unity3d注册并登录
	auth_public.GET("user/reg/:id", user_controller.RegisterAndLogin())

	// 手机sdk登录
	auth_publicV2 := g.Group("/auth/v2/")
	auth_publicV2.Use(limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content), limit.RateLimit(), statsCCU())
	auth_publicV2.POST("deviceWithQuick", device_id_controller.RegDeviceWithQuickAndLogin_v2())
	auth_publicV2.POST("deviceWithQuickMuBao1", device_id_controller.RegDeviceWithQuickAndLogin_MuBao1())
	auth_publicV2.POST("deviceWithVivo", device_id_controller.RegDeviceWithVivoAndLogin_v2())
	auth_publicV2.POST("deviceWithHero", device_id_controller.RegDeviceWithHeroAndLogin_v2())
	auth_publicV2.POST("deviceWithEnjoy", device_id_controller.RegDeviceWithEnjoyAndLogin_v2())
	auth_publicV2.POST("deviceWithJwsvn", device_id_controller.RegDeviceWithVNAndLogin_v2())
	auth_publicV2.POST("deviceWithEnjoyKo", device_id_controller.RegDeviceWithEnjoyKoAndLogin_v2())
	auth_publicV2.POST("deviceWithJwsja",device_id_controller.RegDeviceWith6wavesAndLogin_v2())
	// 下面是测试接口
	if config.Cfg.Runmode != "prod" {
		g.GET("/auth/testsentry", user_controller.TestSentry())
	}
	g.GET("user/getjson", user_controller.GetJson())
}

func RegAuthGMCommand(g *gin.Engine) {
	user_controller := controllers.UserController{}

	auth_api := g.Group("/auth/v1/api/", limit.InternalOnlyLimit())
	auth_api.GET("user/ban/:id", user_controller.BanUser())
	auth_api.GET("user/gag/:id", user_controller.GagUser())    // 禁言，应该没用了
	auth_api.POST("user/isHasReg", user_controller.IsHasReg()) // 查询acid在此服是否有角色
}
