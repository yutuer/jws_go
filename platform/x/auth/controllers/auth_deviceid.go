package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/x/auth/config"

	"strings"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/logiclog"
	"vcs.taiyouxi.net/platform/x/auth/models"
)

// DeviceIDController about object
type DeviceIDController struct {
	PlanXController
}

// Test ... just make sure test framework work
// /auth/v1/device/test
func (dic *DeviceIDController) Test() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, "ok")
	}
}

// RegDeviceAndLogin 注册设备ID并登录，如果已经存在则直接登录
// /auth/v1/device/:id
//目前Unity客户端的实现是SystemInfo.deviceUniqueIdentifier@timestamp@channeid
//C# SystemInfo.deviceUniqueIdentifier + "@" + Utils.GetTimeStamp() + "@apple.com";
//例如:eb6b9e72e0a54dc1b9fe9f4f5d5992d102337525@07301558224760@apple.com
// TODO 如何防止撞库？安全性极低，没有合理的方案能够确保ID冲突是什么引起的
// TODO deviceID应该体现game id, did:{gid}:{deviceID}
func (dic *DeviceIDController) RegDeviceAndLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := struct {
			Result      string `json:"result"`
			Authtoken   string `json:"authtoken,omitempty"`
			DisplayName string `json:"display,omitempty"`
			Err         string `json:"error,omitempty"`
		}{
			"no",
			"", "", "",
		}

		deviceID := c.Param("id")
		crc := dic.GetString(c, "crc")
		if crc == "" || deviceID == "" {
			//无法校验的请求一概放弃
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf(":-)"), errorctl.ClientErrorFormatVerifyFailed,
			)
			return
		}

		//校验crc加密一致性
		bSafeID, _ := secure.DefaultEncode.Decode64FromNet(crc)
		safeID := string(bSafeID)
		if safeID != deviceID {
			//校验失败
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf(":-)"), errorctl.ClientErrorFormatVerifyFailed,
			)
			return
		}

		isReg := false
		//校验成功
		// 如果设备已有用户名密码，返回错误，要求用用户名密码
		di, err, _ := models.TryToCheckDeviceInfo(deviceID)
		switch err {
		case models.XErrUserNotExist:
			//可以创建新帐号
			di, err = models.AddDevice(deviceID, "", "", "")
			// if err == nil && di == nil {
			// 	err = fmt.Errorf("AddDevice got unknown result")
			// }
			uid := db.InvalidUserID
			if di != nil {
				uid = di.UserId
			}
			if err != nil {
				errorctl.CtrlErrorReturn(c, "[Auth.Device]",
					fmt.Errorf("AddDevice error: uid(%d), error(%s)", uid, err),
					errorctl.ClientErrorMaybeDBProblem,
				)
				return
			}
			isReg = true
		case models.XErrUserExistButNeedPassword:
			//要求强制用用户名密码登录
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("You need use username&password to login"),
				errorctl.ClientErrorHaveToUsePassword,
			)
			return

		default:
			if err != nil {
				errorctl.CtrlErrorReturn(c, "[Auth.Device]",
					err,
					errorctl.ClientErrorMaybeDBProblem,
				)
				return
			}
		}

		//下面逻辑err == nil
		uid := di.UserId
		authToken, err := models.AuthToken(uid, deviceID)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				err,
				errorctl.ClientErrorMaybeDBProblem,
			)
			return
		}
		logiclog.LogLogin(uid.String(), deviceID, "", "", isReg)

		r.Result = "ok"
		r.Authtoken = authToken
		r.DisplayName = di.Display
		//Auth token 发送到登录校验服务器，并设置有效时间
		notifyLoginServer(authToken, uid)

		c.JSON(200, r)
	}
}

func (dic *DeviceIDController) EchoIp() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		clientIPonly := strings.Split(clientIP, ":")[0]
		c.JSON(200, clientIPonly)
	}
}

func (dic *DeviceIDController) DebugLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.PostForm("errorCode")
		auth := c.PostForm("auth")
		gamex := c.PostForm("gamex")
		isFirstEnter := c.PostForm("isFirstEnter")
		operation := c.PostForm("operation")
		accountId := c.PostForm("accountId")
		exception := c.PostForm("exception")
		stackTrace := c.PostForm("stackTrace")
		time := c.PostForm("time")
		deviceId := c.PostForm("DeviceId")

		p := struct {
			AccountId    string
			ErrorCode    string
			IP           string
			Auth         string
			Gamex        string
			IsFirstEnter string
			Operation    string
			Exception    string
			StackTrace   string
			Time         string
			DeviceId     string
		}{
			accountId, err, c.ClientIP(), auth,
			gamex, isFirstEnter, operation, exception,
			stackTrace, time, deviceId,
		}
		logiclog.LogDebug(p)

		r := struct {
			Res string
		}{
			"ok",
		}
		c.JSON(200, r)
	}
}
