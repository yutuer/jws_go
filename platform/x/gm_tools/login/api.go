package login

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	//∂"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
	"encoding/json"

	"vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type pass struct {
	Passwd string `json:"n"`
}

func RouteLogin(r *gin.Engine) {
	r.POST("/api/v1/login/*gm_name", func(c *gin.Context) {
		gm_name := c.Param("gm_name")
		gm_name = util.DeleteBackslash(gm_name)
		logs.Info("mail get %s", gm_name)

		s := pass{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, err.Error())
			return
		}

		logs.Info("login %s, %v", gm_name, s)

		acc := AccountManager.Get(gm_name)
		if acc == nil {
			c.String(401, "NoAccount")
			return
		}

		passok, cookie := AccountManager.IsPass(gm_name, s.Passwd)

		if passok {
			c.String(200, cookie)
		} else {
			c.String(401, "PassError")
		}
	})

	r.GET("/api/v1/init", func(c *gin.Context) {
		info := struct {
			TimeZone string
		}{
			TimeZone: config.Cfg.TimeLocal,
		}
		str, err := json.Marshal(&info)
		if err != nil {
			logs.Error("/api/v1/init json.Marshal err %s", err.Error())
			c.String(401, err.Error())
			return
		}
		c.String(200, string(str))
	})
}
