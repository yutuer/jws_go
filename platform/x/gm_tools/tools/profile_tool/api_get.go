package profile_tool

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type ProfileValue struct {
	Value string `json:"v"`
}

func RouteProfileGet(r *gin.Engine) {

	r.GET("/api/v1/server_all", func(c *gin.Context) {
		logs.Info("GetAllRedis")

		data, err := util.ToJson(GetAllRedis())
		if err == nil {
			c.String(200, string(data))
		} else {
			c.String(400, err.Error())
		}
	})

	r.GET("/api/v1/server_cfg_all", func(c *gin.Context) {
		logs.Info("GetAllServerCfg")

		data, err := util.ToJson(GetAllServerCfg())
		if err == nil {
			c.String(200, string(data))
		} else {
			c.String(400, err.Error())
		}
	})

	r.GET("/api/v1/profile_get/:server_name/*account", func(c *gin.Context) {
		server_name := c.Param("server_name")
		account := c.Param("account")

		account = util.DeleteBackslash(account)

		logs.Info("GetProfileFromRedis %s, %s", server_name, account)

		data, err := GetProfileFromRedis(server_name, account)
		if err == nil {
			c.String(200, string(data))
		} else {
			c.String(400, err.Error())
		}
	})

	r.POST("/api/v1/profile_get/:server_name/*account", func(c *gin.Context) {
		server_name := c.Param("server_name")
		account := c.Param("account")

		account = util.DeleteBackslash(account)

		s := ProfileValue{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, err.Error())
			return
		}

		logs.Info("ModPrfileFromRedis %s, %s - %v", server_name, account, s)

		err = ModPrfileFromRedis(server_name, account, s.Value)
		if err == nil {
			c.String(200, string("ok"))
		} else {
			c.String(401, err.Error())
		}
	})
}
