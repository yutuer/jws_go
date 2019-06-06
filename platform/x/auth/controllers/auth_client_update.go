package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
)

func (dic *DeviceIDController) UpdateVer() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Query("channel")
		sub_channel := c.Query("subchannel")
		if channel == "" || sub_channel == "" {
			errorctl.CtrlErrorReturn(c, "[Auth.UpdateVer]",
				fmt.Errorf(":-)"), errorctl.ClientErrorUpdateVerParamErr,
			)
			return
		}
		url := config.GetChannelUrl(channel, sub_channel)
		if url == "" {
			errorctl.CtrlErrorReturn(c, "[Auth.UpdateVer]",
				fmt.Errorf(":-)"), errorctl.ClientErrorUpdateVerUrlNotFound,
			)
			return
		}

		c.Redirect(http.StatusMovedPermanently, url)
	}
}
