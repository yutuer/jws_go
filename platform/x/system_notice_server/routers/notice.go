package routers

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/system_notice_server/config"
	"vcs.taiyouxi.net/platform/x/system_notice_server/logic"
)

func RegNotice(g *gin.Engine) {
	auth_public := g.Group("/notice/v1/")
	//auth_public.Use(limit.RateLimit())

	//
	auth_public.GET("getnotice", GetNoticeHandler)

}

func GetNoticeHandler(c *gin.Context) {
	config.AddNoticeCCUCount()
	gid := c.Query("gid")
	version := c.Query("version")
	retNotice := logic.GetNotice(gid, version)
	c.String(200, retNotice)
}
