package controllers

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/x/auth/config"

	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/models"

	"strings"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/x/auth/logiclog"
)

// UserController is in charge of registeration.
type UserController struct {
	PlanXController
}

var validID = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{5}`)

func isUserNameFormatOk(userName string) bool {
	if nl := len(userName); nl < 6 || nl > 128 ||
		!validID.MatchString(userName) {
		return false
	}
	return true
}

// LoginAsUser is used for login existing users
func (uc *UserController) LoginAsUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := uc.GetString(c, "name")     //客户端传过来的是经过base64(name)的
		passwd := uc.GetString(c, "passwd") //客户端应该传过来的是pwdbase64(md5(passwd))

		namedb, errDec1 := base64.URLEncoding.DecodeString(name)
		md5pwd, errDec2 := secure.DefaultEncode.Decode64FromNet(passwd)

		if !isUserNameFormatOk(string(namedb)) {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("err: username(%s) is illegal", namedb),
				errorctl.ClientErrorUsernameRegFormatIllegal)
			return
		}

		if errDec1 != nil || errDec2 != nil {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("err1: %v, err2 %v", errDec1, errDec2),
				errorctl.ClientErrorFormatVerifyFailed)
			return
		}

		authToken, uid, info_to_client, ban_time, ban_reason, err := models.AuthUserPass(strings.ToLower(string(namedb)), string(md5pwd))
		if err != nil {
			switch err {
			case models.XErrAuthUsernameNotFound:
				errorctl.CtrlErrorReturn(c, "[Auth.Device]", err, errorctl.ClientErrorUsernameNotFound)
			case models.XErrAuthUserPasswordInCorrect:
				errorctl.CtrlErrorReturn(c, "[Auth.Device]", err, errorctl.ClientErrorUserPasswordIncorrect)
			case models.XErrByBan:
				errorctl.CtrlBanReturn(c, "[Auth.Device]", err, errorctl.ClientErrorBanByGM, ban_time, ban_reason)
			}
			return
		}

		shardHasRoleStr := []string{}
		shardHasRole, err := models.GetUserShardHasRole(uid.String())
		if err != nil {
			logs.Error("[LoginAsUser] login GetUserShard err: %v", err)
			shardHasRole = []models.ShardHasRole{}
		} else {
			for _, st := range shardHasRole {
				shardHasRoleStr = append(shardHasRoleStr, st.Shard)
			}
		}

		lastShard, err := models.GetUserLastShard(uid.String())
		if err != nil {
			logs.Error("[quick] login GetUserLastShard err: %v", err)
		}

		logiclog.LogLogin(uid.String(), strings.ToLower(string(namedb)), "", "", false)

		c.JSON(200, map[string]interface{}{
			"result": "ok",
			//"expire":    60 * 5,    //5 min
			"authtoken":     authToken, //Auth token 返回给客户端
			"info":          *info_to_client,
			"shardrole":     shardHasRoleStr,
			"shardrolev250": shardHasRole,
			"lastshard":     lastShard,
		})

		//Auth token 发送到登录校验服务器，并设置有效时间
		notifyLoginServer(authToken, uid)
	}
}

// http://127.0.0.1:8789/auth/v1/user/ban/{userid}?time={timetoban}&gid={gid}&reason={reason}
func (uc *UserController) BanUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("id")
		ban_time, err := uc.GetInt64(c, "time")
		gid := uc.GetString(c, "gid")
		reason := uc.GetString(c, "reason")
		if err != nil || gid == "" {
			c.JSON(401, map[string]interface{}{
				"ok":     "false",
				"reason": err.Error(),
			})
			return
		}

		logs.Debug("BanUser %s %s %d %s", uid, gid, ban_time, reason)
		err = BanUser(uid, gid, ban_time, reason)
		if err != nil {
			c.JSON(401, map[string]interface{}{
				"ok":     "false",
				"reason": err.Error(),
			})
			return
		}

		c.JSON(200, map[string]interface{}{
			"result": "ok",
		})
		return
	}
}

// http://127.0.0.1:8789/auth/v1/user/gag/{userid}?time={timetoban}&gid={gid}
func (uc *UserController) GagUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("id")
		gag_time, err := uc.GetInt64(c, "time")
		gid := uc.GetString(c, "gid")

		if err != nil || gid == "" {
			c.JSON(401, map[string]interface{}{
				"ok":     "false",
				"reason": err.Error(),
			})
			return
		}

		logs.Debug("Gag %s %s %d", uid, gid, gag_time)
		err = GagUser(uid, gid, gag_time)
		if err != nil {
			c.JSON(401, map[string]interface{}{
				"ok":     "false",
				"reason": err.Error(),
			})
			return
		}

		c.JSON(200, map[string]interface{}{
			"result": "ok",
		})
		return
	}
}

// RegisterAndLogin : register a user with email and password,
// and login at the same time
func (uc *UserController) RegisterAndLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("id")
		name := uc.GetString(c, "name")     //客户端传过来的是经过base64(name)的
		passwd := uc.GetString(c, "passwd") //客户端应该传过来的是pwdbase64(md5(passwd))过的
		email := uc.GetString(c, "email")

		//传输解密过程
		byte_name, errDec1 := base64.URLEncoding.DecodeString(name)
		byte_md5pwd, errDec2 := secure.DefaultEncode.Decode64FromNet(passwd)

		if errDec1 != nil || errDec2 != nil {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("err1: %v, err2 %v", errDec1, errDec2),
				errorctl.ClientErrorFormatVerifyFailed)
			return
		}

		namedb := strings.ToLower(string(byte_name))
		passwddb := string(byte_md5pwd)

		//用户名密码必须符合规范
		if !isUserNameFormatOk(namedb) {
			errorctl.CtrlErrorReturn(c, "[Auth.Device]",
				fmt.Errorf("err: username(%s) is illegal", namedb),
				errorctl.ClientErrorUsernameRegFormatIllegal)
			return
		}

		//注册用户名密码
		uid, err := models.RegisterWithPassword(namedb, passwddb, email, deviceID)
		if err != nil {
			switch err {
			case models.XErrRegPwdUserNameUsed:
				errorctl.CtrlErrorReturn(c, "[Auth.Device]", err, errorctl.ClientErrorUsernameHasBeenUsed)
			case models.XErrRegPwdUserHasName:
				errorctl.CtrlErrorReturn(c, "[Auth.Device]", err, errorctl.ClientErrorDeviceAlreadyBinded)
			}
			return
		}

		//使用uuid生成auth token
		authToken, err := models.AuthToken(uid, "")
		if err != nil {
			//AuthToken只返回DB级Error
			errorctl.CtrlErrorReturn(c, "[Auth.Device]", err, errorctl.ClientErrorMaybeDBProblem)
			return
		}

		c.JSON(200, map[string]interface{}{
			"result": "ok",
			//"expire":    60 * 5,    //5 min
			"authtoken": authToken, //Auth token 返回给客户端
		})

		//Auth token 发送到登录校验服务器，并设置有效时间
		notifyLoginServer(authToken, uid)
	}
}

// http://127.0.0.1:8789/auth/v1/user/isHasReg
func (uc *UserController) IsHasReg() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := struct {
			Uid string `json:"uid"`
		}{
			"",
		}

		err := c.Bind(&r)

		if err != nil {
			c.String(401, err.Error())
			logs.Trace("IsHasReg Err By %s", err.Error())
			return
		}

		acc, err := db.ParseAccount(r.Uid)
		if err != nil {
			c.String(402, err.Error())
			return
		}

		userShard := []string{}
		userShardSt, err := models.GetUserShardHasRole(acc.UserId.String())
		if err != nil {
			logs.Error("[quick] login GetUserShard err: %v", err)
			c.String(403, err.Error())
			return
		} else {
			for _, st := range userShardSt {
				userShard = append(userShard, st.Shard)
			}
		}

		if userShard == nil || len(userShard) == 0 {
			c.String(200, string("no"))
		} else {
			shardName, err := models.GetShardName(acc.ShardId, acc.GameId)
			logs.Trace("GetUserShardHasRole %s %v", shardName, userShard)
			if err != nil {
				c.String(404, err.Error())
				return
			}
			for _, s := range userShard {
				if s == shardName {
					c.String(200, string("yes"))
					return
				}
			}
			c.String(200, string("no"))
		}

	}
}

func (uc *UserController) TestSentry() gin.HandlerFunc {
	return func(c *gin.Context) {
		panic("########### test sentry panic ... ")
	}
}

func (uc *UserController) GetJson() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"result": "ok",
		})
	}
}
