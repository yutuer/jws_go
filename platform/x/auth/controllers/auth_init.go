package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/db"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/config"

	"github.com/astaxie/beego/httplib"
)

func notifyLoginServer(authToken string, userID db.UserID) bool {
	// TBD 需要考虑重发机制么？
	var authNotify struct {
		AuthToken string    `json:"at"`
		UserID    db.UserID `json:"uid"`
	}

	authNotify.AuthToken = authToken
	authNotify.UserID = userID

	bytejson, err := json.Marshal(authNotify)
	logs.Trace("notifyLoginServer userID:", userID)
	if err != nil {
		logs.Critical("Auth notify Login, it should not happen!", err.Error())
		return false
	}
	data := base64.URLEncoding.EncodeToString(bytejson)
	URL := fmt.Sprintf("%s?info=%s", config.Cfg.LoginApiAddr, data)
	//
	var rst struct {
		Result string `json:"result"`
	}
	req := httplib.Get(URL)
	err = req.ToJson(&rst)
	if err != nil {
		logs.Critical("Auth notify Login server failed, user can't log in", err.Error())
		return false
	}

	//err = json.Unmarshal(result, &rst)
	//if err != nil {
	//logs.Error("Auth notify Login server success, but return value decode failed!")
	//return false
	//}

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if rst.Result != "ok" {
		return false
	}

	return true
}
