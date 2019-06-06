package main

import (
	"encoding/base64"
	"encoding/json"
	//"fmt"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type receiptData struct {
	Data []byte `json:"data"`
}

type receiptDataToApple struct {
	Data string `json:"receipt-data"`
}

func VerifyData(data []byte) error {
	res := base64.StdEncoding.EncodeToString(data)
	logs.Info("verify base64 %v", res)

	b, err := json.Marshal(receiptDataToApple{res})
	if err != nil {
		return err
	}

	re, err := util.HttpPost("https://sandbox.itunes.apple.com/verifyReceipt",
		"application/json; charset=utf-8", b)

	if err != nil {
		return err
	}

	logs.Warn("post res %v", re)
	return nil
}

func main() {
	r := gin.Default()
	r.POST("/verify", func(c *gin.Context) {
		s := receiptData{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, (err.Error()))
			return
		}

		rdata := s.Data

		logs.Info("verify %v", rdata)

		err = VerifyData(rdata)

		if err == nil {
			c.String(200, "{}")
		} else {
			c.String(401, (err.Error()))
		}
	})
	r.Run(":7788")
}
