package routers

import (
	"net/http"
	"strings"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/clientlog/config"
)

type PRLog struct {
	AccountID   string `json:"a" binding:"required"`
	RequestURI  string `json:"r" binding:"required"`
	RequestTime string `json:"t" binding:"required"`
	Problem     string `json:"p" binding:"required"`
}

func RegPR(g *gin.Engine) {
	// 客户端版本强制更新用
	pr := g.Group("/pr/v1/")
	pr.Use(
		limit.CheckIdentity(config.Spec_Header, config.Spec_Header_Content),
		limit.RateLimit())

	//httpie:
	//http POST http://127.0.0.1:8081/pr/v1/log TYX-Request-Id:0a2e427a115aa9a7130d00cb08316455 a="0:10:fdsfds" r="attr/abc" t=1477816181 p="fdfd fdsfs"
	pr.POST("log", func(c *gin.Context) {
		var l PRLog
		if err := c.BindJSON(&l); err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		logs.Debug("log: %v", l)
		c.String(http.StatusOK, "ok")
	})

	//client event
	//httpie test: http POST http://127.0.0.1:8081/pr/v1/ce name=John email=john@example.org TYX-Request-Id:0a2e427a115aa9a7130d00cb08316455
	pr.POST("ce", func(c *gin.Context) {
		decoder := json.NewDecoder(c.Request.Body)
		var abc map[string]interface{}
		if err := decoder.Decode(&abc); err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		logs.Info("ce:%v", abc)
		c.String(http.StatusOK, "ok")
	})

	tools := g.Group("/tools/")
	tools.Use(limit.RateLimit())
	tools.GET("echoip", func(c *gin.Context) {
		clientIP := c.ClientIP()
		clientIPonly := strings.Split(clientIP, ":")[0]
		c.JSON(http.StatusOK, clientIPonly)
	})
}
