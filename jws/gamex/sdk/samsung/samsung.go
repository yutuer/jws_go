package samsung

import (
	"encoding/json"
	"time"

	"net/url"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

const (
	SIGN            = "sign"
	SIGN_TYPE       = "signtype"
	SIGN_TYPE_VALUE = "RSA"
	TRANSDATA       = "transdata"
	GoodSamsungId   = 12
)

func TryPay(acid string, iapIndex uint32, cpOrderNumber string, money uint32,
	ExtrInfo string, goodTitle, goodDesc string) map[string]string {
	req := httplib.Post(Cfg.Url).SetTimeout(8*time.Second, 8*time.Second)

	param := make(map[string]interface{}, 16)
	param["appid"] = Cfg.AppID
	param["waresid"] = GoodSamsungId
	param["waresname"] = goodTitle
	param["cporderid"] = cpOrderNumber
	param["price"] = float32(money)
	param["currency"] = "RMB"
	param["appuserid"] = acid[:32]
	param["cpprivateinfo"] = ExtrInfo
	param["notifyurl"] = Cfg.SdkNotifyUrl

	logs.Debug("samsung TryPay req param %v", param)
	transdataStr, err := json.Marshal(param)
	if err != nil {
		logs.Error("samsung TryPay json.Marshal err %v %v", err, param)
		return nil
	}

	sign, err := util.Sign(transdataStr)
	if err != nil {
		logs.Error("samsung TryPay Sign err %v %v", err, param)
		return nil
	}
	reqParam := make(map[string]string, 3)
	reqParam[TRANSDATA] = string(transdataStr)
	reqParam[SIGN] = sign
	reqParam[SIGN_TYPE] = SIGN_TYPE_VALUE

	for k, v := range reqParam {
		req.Param(k, v)
	}

	logs.Debug("samsung TryPay param %v", reqParam)

	data, err := req.Bytes()
	if err != nil {
		logs.Error("samsung TryPay req.Bytes failed err with %v", err)
		return nil
	}
	logs.Debug("samsung TryPay resp %s", string(data))
	m, err := url.ParseQuery(string(data))
	if err != nil || len(m) <= 0 || len(m["transdata"]) <= 0 {
		logs.Error("samsung TryPay url.ParseQuery failed err with %v", err)
		return nil
	}
	logs.Debug("samsung TryPay resp %v", m)
	var r ret
	err = json.Unmarshal([]byte(m["transdata"][0]), &r)
	if err != nil {
		logs.Error("samsung TryPay json.Unmarshal failed err with %v", err)
		return nil
	}

	logs.Debug("samsung TryPay resp %v", r)

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 || r.Code != "" {
		logs.Error("samsung TryPay res failed with %v", r)
		return nil
	}

	res := make(map[string]string, 8)
	res["orderNumber"] = r.TransId
	return res
}

type ret struct {
	TransId string `json:"transid,omitempty"`
	Code    string `json:"code"`
	ErrMsg  string `json:"errmsg"`
}
