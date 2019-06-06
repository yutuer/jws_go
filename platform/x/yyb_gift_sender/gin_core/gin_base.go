package gin_core

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/db"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/encode"
)

var engine *gin.Engine

const (
	delay_time = 300
)

func Reg() {
	engine = gin.Default()
	engine.SetHTMLTemplate(html)
	engine.GET(config.CommonConfig.GiftUrl, giftHandle)
	engine.GET(config.CommonConfig.ShardUrl, shardHandle)
}

func Run() {
	engine.Run(config.CommonConfig.Host)
}

type BaseResp struct {
	RetCode int    `json:"ret"`
	Msg     string `json:"msg"`
}

func giftHandle(c *gin.Context) {
	logs.Debug("Success %v", c.Request.URL.Query())
	timestamp, err := strconv.ParseInt(c.Query("timestamp"), 10, 0)
	if err != nil {
		logs.Error("Invalid Args(timestamp)")
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     config.RetCode_ErrorArg,
			"message": config.ErrMap[config.RetCode_ErrorArg],
		})
		return
	}
	if time.Now().Unix() > int64(timestamp+delay_time) {
		logs.Error("timeout")
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     config.RetCode_InvalidRequest,
			"message": config.ErrMap[config.RetCode_InvalidRequest],
		})
		return
	}

	if !checkArgs(c) {
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     config.RetCode_ErrorArg,
			"message": config.ErrMap[config.RetCode_ErrorArg],
		})
		return
	}

	// 验证sig
	args := make(map[string]string, 10)
	var sig string
	for k, v := range c.Request.URL.Query() {
		if k == "sig" {
			sig = c.Query(k)
			continue
		}
		args[k] = v[0]
	}
	my_sig, err := encode.CalSig(args, config.CommonConfig.GiftUrl, config.CommonConfig.Appkey)
	if err != nil {
		panic(err)
	}
	if sig == my_sig {
		logs.Info("Right Sig is %s", my_sig)
	} else {
		logs.Error("Invalid Sig")
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     config.RetCode_ErrorSig,
			"message": config.ErrMap[config.RetCode_ErrorSig],
		})
		return
	}

	openid := c.Query("openid")
	taskid, _ := strconv.Atoi(c.Query("taskid"))

	// 查询用户领奖状态并发奖
	ret := db.CheckSend(openid+config.CommonConfig.DeviceSuffix, uint32(taskid))
	if ret == config.RetCode_FinishAndSuccess {
		logs.Info("Send Success(task: %d)", taskid)
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     ret,
			"message": config.ErrMap[ret],
		})
	} else {
		c.HTML(200, "yyb_gift", gin.H{
			"ret":     ret,
			"message": config.ErrMap[ret],
		})
	}
}

func checkArgs(c *gin.Context) bool {
	if c.Query("openid") == "" || c.Query("taskid") == "" {
		logs.Error("openid or taskid error")
		return false
	}

	if c.Query("appid") != config.CommonConfig.AppID {
		logs.Error("appid error")
		return false
	}
	if c.Query("action") != "check_send" {
		logs.Error("action error")
		return false
	}
	if c.Query("area") != "wx" && c.Query("area") != "qq" {
		logs.Error("area error")
		return false
	}
	if c.Query("pkey") != encode.GetMD5Key(c.Query("openid")+config.CommonConfig.Appkey+c.Query("timestamp")) {
		logs.Error("pkey error")
		return false
	}
	return true
}
