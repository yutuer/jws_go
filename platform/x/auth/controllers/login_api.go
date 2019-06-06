package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/planx/servers/gate"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/models"
)

// APIController is in charge of internal communication with gate servers.
type APIController struct {
	PlanXController
}

// NotifyUserLogin is called by gate/game server
// to notify login server that user logged in
// Gate发现玩家成功登录后调用参数 logintoken, userid
// @router /notifylogin [post]
func (api *APIController) NotifyUserLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			LoginToken string `form:"logintoken"`
			AccountId  string `form:"accountid"`
			Rpc        string `form:"rpcaddrport"`
		}{}
		err := c.BindWith(&data, binding.Form)

		if err != nil {
			logs.Error("NotifyUserLogin Err %s", err.Error())
			c.JSON(500, err.Error())
			return
		}

		logs.Debug("NotifyUserLogin get user: %s, %s, %s",
			data.AccountId, data.LoginToken, data.Rpc)
		_account, err := db.ParseAccount(data.AccountId)
		if err != nil {
			c.String(500, "err : %s", err.Error())
			return
		}
		models.UpdateLoginStatus(_account, data.Rpc, data.LoginToken, models.LSC_LOGIN)

		c.JSON(200, "ok")

	}
}

// Auth服务器通知Login踢人 ?id={uid}&gid={gid}&banTime={banTime}&reason{reason}
// http://127.0.0.1:8081/login/v1/api/kick [get]
func (api *APIController) NotifyKickUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rst struct {
			Result string `json:"result"`
		}

		id := c.DefaultQuery("id", "")
		gid, err := api.GetInt(c, "gid")
		if id == "" || err != nil {
			logs.Error("NotifyKickUser Err by ID nil")
			rst.Result = "NOID"
			c.JSON(500, rst)
			return
		}
		banTime, _ := api.GetInt64(c, "banTime")
		reason := api.GetString(c, "reason")
		avoidMultipleLogin(uint(gid), db.UserIDFromStringOrNil(id), reason, banTime)

		logs.Debug("NotifyKickUser user: %s %d %s", id, banTime, reason)
		rst.Result = "ok"
		c.JSON(200, rst)

	}
}

// Auth服务器通知Login通知禁言信息给客户端 ?id={uid}&gid={gid}&gagt={gag_time}
// http://127.0.0.1:8081/login/v1/api/kick [get]
func (api *APIController) NotifyGagUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rst struct {
			Result string `json:"result"`
		}

		id := c.DefaultQuery("id", "")
		gid, err := api.GetInt(c, "gid")
		gag, err_gag := api.GetInt64(c, "gagt")
		if id == "" || err != nil || err_gag != nil {
			logs.Error("NotifyKickUser Err by ID nil")
			rst.Result = "NOID"
			c.JSON(500, rst)
			return
		}

		info := gate.InfoNotifyToClient{
			GagTime: gag,
		}

		notifyInfo(uint(gid), db.UserIDFromStringOrNil(id), info)

		logs.Debug("NotifyGagUser user: %s", id)
		rst.Result = "ok"
		c.JSON(200, rst)

	}
}

// NotifyUserLogout is called by gate/game server
// to notify login server that user logged out
// Gate发现玩家掉线后调用 或者 玩家主动注销
// @router /notifylogout [post]
func (api *APIController) NotifyUserLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			LoginToken string `form:"logintoken"`
			AccountId  string `form:"accountid"`
			Rpc        string `form:"rpcaddrport"`
		}{}
		err := c.BindWith(&data, binding.Form)

		if err != nil {
			logs.Error("NotifyUserLogin Err %s", err.Error())
			c.JSON(500, err.Error())
			return
		}

		logs.Debug("NotifyUserLogout get user: %s, %s, %s",
			data.AccountId, data.LoginToken, data.Rpc)
		_account, err := db.ParseAccount(data.AccountId)
		if err != nil {
			c.String(500, "err : %s", err.Error())
			return
		}

		models.UpdateLoginStatus(_account, data.Rpc, data.LoginToken, models.LSC_LOGIN)

		c.JSON(200, "ok")
	}
}

// GateRegister is called by gate/game server, while it's alive.
// Game server should trigger this API regular in N seconds,
// to update information of game server.
// Otherwise, the gate should be something wrong, removed from list.
// Gate服务调用， 推送当前Gate上的玩家数量，更新自身IP地址
// @router /gateregister [post]
func (api *APIController) GateRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			Gameipaddrport string `form:"gameipaddrport"`
			Rpcipaddrport  string `form:"rpcipaddrport"`
			Ccu            string `form:"ccu"`
			Shardid        string `form:"shardid"`
		}{}
		//b, _ := ioutil.ReadAll(c.Request.Body)
		//logs.Trace("req %v", string(b))

		err := c.BindWith(&data, binding.Form)

		if err != nil {
			logs.Error("GateRegister Err %s", err.Error())
			c.JSON(200, err.Error())
			return
		}

		//logs.Trace("GateRegister res %v", data)

		ccu, err := strconv.ParseInt(data.Ccu, 10, 64)
		if err != nil {
			logs.Error("GateRegister Err1 %s", err.Error())
			c.JSON(200, err.Error())
			return
		}
		shardid, err := strconv.Atoi(data.Shardid)
		if err != nil {
			logs.Error("GateRegister Err2 %s", err.Error())
			c.JSON(200, err.Error())
			return
		}

		gate := models.GateInfo{
			GameIPAddrPort: data.Gameipaddrport,
			RPCIPAddrPort:  data.Rpcipaddrport,
			CCU:            ccu,
			ShardID:        uint(shardid),
		}

		succ, err := models.AddOneGate(gate)

		//logs.Trace("AddOneGate %v", gate)

		res := ""
		if succ {
			res = "ok"
		} else {
			res = err.Error()
		}
		c.JSON(200, res)
	}
}

func (api *APIController) NotifyUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			Gid         string `form:"gid"`
			ShardId     string `form:"shardid"`
			Uid         string `form:"uid"`
			NewRole     string `form:"newrole"`
			PayFeedBack string `form:"PayFeedBack"`
		}{}

		ret := struct {
			Res            string
			HadGotFeedBack bool
		}{}
		err := c.BindWith(&data, binding.Form)
		if err != nil {
			logs.Error("NotifyUserInfo Err %s", err.Error())
			ret.Res = err.Error()
			c.JSON(200, ret)
			return
		}

		gid, err := strconv.Atoi(data.Gid)
		if err != nil {
			logs.Error("NotifyUserInfo Err1 %s", err.Error())
			ret.Res = err.Error()
			c.JSON(200, ret)
			return
		}
		sid, err := strconv.Atoi(data.ShardId)
		if err != nil {
			logs.Error("NotifyUserInfo Err2 %s", err.Error())
			ret.Res = err.Error()
			c.JSON(200, ret)
			return
		}

		logs.Debug("NotifyUserInfo rec %v", data)
		if data.NewRole != "" {
			err := models.SetUserShardHasRole(gid, sid, data.Uid, data.NewRole)
			if err != nil {
				logs.Error("NotifyUserInfo Err3 %s", err.Error())
				ret.Res = err.Error()
				c.JSON(200, ret)
				return
			}
			logs.Debug("NotifyUserInfo %d:%d:%s", gid, sid, data.Uid)
		}

		if err := models.SetUserLastShard(data.Uid, gid, sid); err != nil {
			logs.Error("NotifyUserInfo SetUserLastShard Err4 %s", err.Error())
			ret.Res = err.Error()
			c.JSON(200, ret)
			return
		}

		if data.PayFeedBack == "wantGet" {
			err, canGot := models.SetUserPayFeedBackIfNot(data.Uid)
			if err != nil {
				logs.Error("NotifyUserInfo SetUserPayFeedBackIfNot Err5 %s", err.Error())
				ret.Res = err.Error()
				c.JSON(200, ret)
				return
			}
			ret.HadGotFeedBack = canGot
		}
		ret.Res = "ok"
		c.JSON(200, ret)
	}
}

// OnlineStatusQuery is used for find user login status
// 在线登录状态查询
// @router /onlinestatusquery [get]
func (api *APIController) OnlineStatusQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, "ok")
	}
}

// AuthTokenNotify called by Auth server, with base64 encoded content of data
// data content is base64(json{ at:authToken, uid:user_id })
func (api *APIController) AuthTokenNotify() gin.HandlerFunc {
	return func(c *gin.Context) {
		var authToken struct {
			AuthToken string    `json:"at"`
			UserID    db.UserID `json:"uid"`
		}
		info := api.GetString(c, "info")
		if info == "" {
			errorctl.CtrlErrorReturn(c, "[AuthToken0]", fmt.Errorf("info param is empty"), -1)
			return
		}

		logs.Trace("[AuthToken0] raw info param:", info)
		js, err := base64.URLEncoding.DecodeString(info)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[AuthToken1]", err, -1)
			return
		}
		logs.Trace("[AuthToken1] info param decoded:", string(js))
		err = json.Unmarshal(js, &authToken)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[AuthToken2]", err, -2)
			return
		}
		err = models.LoginRegAuthToken(authToken.AuthToken, authToken.UserID)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[AuthToken3]", err, -3)
			return
		}

		c.JSON(200, struct {
			Result string `json:"result"`
		}{
			Result: "ok",
		})
	}
}
