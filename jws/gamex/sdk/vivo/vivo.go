package vivo

import (
	"time"

	"fmt"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	payutil "vcs.taiyouxi.net/platform/x/api_gateway/util"
)

func TryPay(cpOrderNumber string, money uint32, ExtrInfo string, goodTitle, goodDesc string) map[string]string {
	req := httplib.Post(Cfg.Url).SetTimeout(8*time.Second, 8*time.Second)

	param := make(map[string]string, 16)
	t := time.Now().In(util.ServerTimeLocal)
	ts := fmt.Sprintf("%04d%02d%02d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())

	param["version"] = "1.0.0"
	param["signMethod"] = "MD5"
	param["cpId"] = Cfg.CPId
	param["appId"] = Cfg.AppID
	param["cpOrderNumber"] = cpOrderNumber
	param["notifyUrl"] = Cfg.SdkNotifyUrl
	param["orderTime"] = ts
	param["orderAmount"] = fmt.Sprintf("%d", money*100)
	param["orderTitle"] = goodTitle
	param["orderDesc"] = goodDesc
	param["extInfo"] = ExtrInfo
	param["signature"] = payutil.GetVivoSign(payutil.Para(param), Cfg.CPKey)

	for k, v := range param {
		req.Param(k, v)
	}

	logs.Debug("vivo TryPay param %v", param)
	var r ret
	err := req.ToJSON(&r)
	if err != nil {
		logs.Error("vivo TryPay req.ToJson failed err with %v", err)
		return nil
	}
	logs.Debug("vivo TryPay resp %v", r)

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 || r.RespCode != "200" {
		logs.Error("vivo TryPay res failed with %v", r)
		return nil
	}

	res := make(map[string]string, 8)
	res["accessKey"] = r.AccessKey
	res["orderNumber"] = r.OrderNumber
	res["orderAmount"] = r.OrderAmount
	return res
}

type ret struct {
	RespCode    string `json:"respCode"`
	RespMsg     string `json:"respMsg"`
	SignMethod  string `json:"signMethod"`
	Signature   string `json:"signature"`
	AccessKey   string `json:"accessKey"`
	OrderNumber string `json:"orderNumber"`
	OrderAmount string `json:"orderAmount"`
}
