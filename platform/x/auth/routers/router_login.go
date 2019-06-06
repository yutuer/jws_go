package routers

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/controllers"
	"vcs.taiyouxi.net/platform/x/auth/limit"
)

func RegLogin(g *gin.Engine) {
	loginController := controllers.LoginController{}
	apiController := controllers.APIController{}

	login_public := g.Group("/login/v1/")
	login_public.Use(limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content), limit.RateLimit())

	if config.Cfg.Runmode != "prod" {
		//DebugKick是调试用API，协助客户端测试踢人操作
		login_public.GET("debug/kick", loginController.DebugKick())
	}

	login_public.GET("shards/:gid", loginController.FetchShardsInfo())
	login_public.GET("getgate", loginController.GetGate())

	login_public_v2 := g.Group("/login/v2/")
	login_public_v2.Use(limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content), limit.RateLimit())
	login_public_v2.GET("shards/:gid", loginController.FetchShardsInfoV2())

	login_api := g.Group("/login/v1/api/", limit.InternalOnlyLimit())

	login_api.POST("notifylogin", apiController.NotifyUserLogin())
	login_api.POST("notifylogout", apiController.NotifyUserLogout())
	login_api.POST("notifyuserinfo", apiController.NotifyUserInfo())
	login_api.POST("gateregister", apiController.GateRegister())
	login_api.GET("onlinestatusquery", apiController.OnlineStatusQuery())
	login_api.GET("authtoken", apiController.AuthTokenNotify())
	login_api.GET("kick", apiController.NotifyKickUser())
	login_api.GET("gag", apiController.NotifyGagUser())
}
