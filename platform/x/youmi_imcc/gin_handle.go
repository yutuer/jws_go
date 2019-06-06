package main

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

func acceptHandle(c *gin.Context) {
	netMsg := base.PlayerMsgInfoFromNet{}
	if err := c.BindJSON(&netMsg); err != nil {
		logs.Error("bind json err by %v", err)
	} else {
		if netMsg.MsgType == base.TextMsgTyp {
			base.MsgChan <- netMsg
		}
	}
	c.String(200, "success!")
}
