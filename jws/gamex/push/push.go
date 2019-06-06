package push

import (
	"encoding/json"

	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	singleDev_Push_Url_ForSign = "openapi.xg.qq.com/v2/push/single_device"
	singleDev_Push_Url         = "http://" + singleDev_Push_Url_ForSign
	android_access_id          = "2100163221"
	android_secret_key         = "a82deafbe3ddbc3622ae109d7629d86a"
	ios_access_id              = "2200152430"
	ios_secret_key             = "13864a53e9b9a04b768bc8c6570da0df"
	Platform_ios               = "0"
	Platform_android           = "1"
)

type result struct {
	RetCode int    `json:"ret_code"`
	ErrMsg  string `json:"err_msg,omitempty"`
	Result  string `json:"result,omitempty"`
}

type android_msg struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	BuilderId int    `json:"builder_id"`
	Ring      int    `json:"ring"`
}
type ios_msg struct {
	Aps ios_content `json:"aps"`
}
type ios_content struct {
	Alert string `json:"alert"`
}

func Push2Device(platform, deviceToken, title, content string) (retry bool) {
	if deviceToken == "0" || deviceToken == "" || deviceToken == "null" {
		logs.Warn("Push2Device deviceToken=%s %v %v %v", deviceToken, platform, title, content)
		return false
	}

	access_id := ""
	secret_key := ""
	msg_str := ""
	message_type := "0"
	environment := "0"
	switch platform {
	case Platform_ios:
		access_id = ios_access_id
		secret_key = ios_secret_key
		ioscontent := ios_content{content}
		msg := ios_msg{ioscontent}
		msg_b, _ := json.Marshal(&msg)
		msg_str = string(msg_b)
		environment = "2"
	case Platform_android:
		access_id = android_access_id
		secret_key = android_secret_key
		msg := android_msg{}
		msg.Title = title
		msg.Content = content
		msg.Ring = 1
		if l, err := time.LoadLocation(game.Cfg.TimeLocal); err != nil {
			t := time.Now()
			t = t.In(l)
			if t.Hour() >= 9 && t.Hour() <= 21 {
				msg.Ring = 1
			} else {
				msg.Ring = 0
			}
		}
		msg_b, _ := json.Marshal(&msg)
		msg_str = string(msg_b)
		message_type = "1"
	default:
		logs.Error("[Push] platform err: %s", platform)
		return false
	}

	req := httplib.Post(singleDev_Push_Url)
	req.Header("Content-type", "application/x-www-form-urlencoded")
	// params
	params := make(map[string]string, 7)
	params["access_id"] = access_id
	params["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	params["device_token"] = deviceToken
	params["message_type"] = message_type
	params["multi_pkg"] = "1"
	params["message"] = string(msg_str)
	params["environment"] = environment
	sign := genSign(singleDev_Push_Url_ForSign, "POST", params, secret_key)
	params["sign"] = sign
	for k, v := range params {
		req.Param(k, v)
	}

	res := result{}

	data, err := req.Bytes()
	if err != nil {
		logs.Error("[Push] req err %s", err.Error())
		return false
	}

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	err = json.Unmarshal(data, &res)
	if err != nil {
		logs.Error("[Push] req json.Unmarshal err %s %s %v", err.Error(), string(data), params)
		return false
	}

	if res.RetCode != 0 {
		logs.Error("[Push] res fail: %v %s %s %v", res, platform, deviceToken, params)
		if res.RetCode == 71 { // 71	APNS服务器繁忙 retry
			logs.Warn("[Push] will retry %v %v", res, params)
			return true
		}
	}

	logs.Trace("[Push] send success %v", params)
	return false
}

func Push2Account(acid, content string) bool {
	//TODO by YZH: 从数据库中提取,或者直接再这里传入玩家Token（通过gmtools查找）。
	//res := account_info.AccountInfoDynamo.GetAccountDeviceInfo(acid)
	//if res != nil {
	//	logs.Trace("Push2Account acid %s res %v", acid, res)
	//	Push2Device(res.PlatformType, res.DeviceInfo, "", content)
	//	return true
	//}
	//logs.Error("Push2Account acid %s not found device token", acid)
	return false
}

func genSign(url, method string, kv map[string]string, secret string) string {
	keys := make([]string, 0, len(kv))
	for k, _ := range kv {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	kvs := make([]string, 0, len(kv)+3)
	kvs = append(kvs, method)
	kvs = append(kvs, url)
	for _, k := range keys {
		kvs = append(kvs, k+"="+kv[k])
	}
	kvs = append(kvs, secret)
	str := strings.Join(kvs, "")
	bb := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", bb)
}
