package controllers

import (
	"errors"
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/models"

	"github.com/astaxie/beego/httplib"
)

func GagUser(uid, gid string, time_to_ban int64) error {
	// 两方面 第一要设置禁言的标签 第二要通知在线的账号禁言
	err := models.SetGagUnInfo(uid, time_to_ban)
	if err != nil {
		return err
	}

	// 再通知login把在线的
	return notifyLoginServerSendGagInfo(uid, gid, time_to_ban)
}

func notifyLoginServerSendGagInfo(uid, gid string, time_to_ban int64) error {
	URL := fmt.Sprintf("%s?id=%s&gid=%s&gagt=%d", config.Cfg.LoginGagApiAddr, uid, gid, time_to_ban)
	var rst struct {
		Result string `json:"result"`
	}
	req := httplib.Get(URL)
	err := req.ToJson(&rst)
	if err != nil {
		logs.Error("notifyLoginServerSendGagInfo failed %s", err.Error())
		return err
	}

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if rst.Result != "ok" {
		return errors.New(rst.Result)
	}

	return nil
}
