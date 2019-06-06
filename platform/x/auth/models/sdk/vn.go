package sdk

import (
	"fmt"

	"encoding/json"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
)

//const (
//	connectTimeout   = 5 * time.Second
//	readWriteTimeout = 5 * time.Second
//)

const (
	VNAndroidChannel = "5003"
	VNIOSChannel     = "5004"
)

func GetVNToken(authcode string, channelID string) (string, error) {
	params := map[string]string{}

	params["__e__"] = "1"
	params["grant_type"] = "authorization_code"
	params["redirect_uri"] = "1"
	params["code"] = authcode
	if channelID == VNAndroidChannel {
		params["productId"] = authConfig.VNSdkCfg.ProductId
		params["client_id"] = authConfig.VNSdkCfg.AppKey
		params["client_secret"] = authConfig.VNSdkCfg.ClientSecret
		params["sign"] = genSign(params, authConfig.VNSdkCfg.AppKey)
	} else if channelID == VNIOSChannel {
		params["productId"] = authConfig.VNSdkCfg.IOSProductId
		params["client_id"] = authConfig.VNSdkCfg.IOSAppKey
		params["client_secret"] = authConfig.VNSdkCfg.IOSClientSecret
		params["sign"] = genSign(params, authConfig.VNSdkCfg.IOSAppKey)
	}

	// .SetTimeout(connectTimeout, readWriteTimeout)
	req := httplib.Post(authConfig.VNSdkCfg.Url + authConfig.VNSdkCfg.UrlToken)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range params {
		req.Param(k, v)
	}

	str, err := req.String()
	if err != nil {
		return "", err
	}
	
	var r retToken
	if err := json.Unmarshal([]byte(str), &r); err != nil {
		return "", fmt.Errorf("%s -- recmsg: %s", err.Error(), str)
	}
	if r.Codes != 0 {
		return "", fmt.Errorf("vn gettoken response %v", r)
	}
	return r.AccessToken, nil
}

func GetVNUserInfo(token string, channelID string) (*vnRetUser, error) {
	params := map[string]string{}
	params["access_token"] = token
	if channelID == VNAndroidChannel {
		params["productId"] = authConfig.VNSdkCfg.ProductId
		params["sign"] = genSign(params, authConfig.VNSdkCfg.AppKey)
	} else if channelID == VNIOSChannel {
		params["productId"] = authConfig.VNSdkCfg.IOSProductId
		params["sign"] = genSign(params, authConfig.VNSdkCfg.IOSAppKey)
	}

	// .SetTimeout(connectTimeout, readWriteTimeout)
	req := httplib.Post(authConfig.VNSdkCfg.Url + authConfig.VNSdkCfg.UrlUserinfo)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range params {
		req.Param(k, v)
	}

	//	str, err := req.String()
	//	if err != nil {
	//		return -1, err
	//	}
	//	logs.Error("############# getuserinfo %s", str)
	//	return -1, fmt.Errorf("test")

	var r vnRetUser
	err := req.ToJson(&r)
	if err != nil {
		return nil, err
	}
	if r.Codes != 0 {
		return nil, fmt.Errorf("vn getuser response %v", r)
	}
	logs.Debug("vn sdk userinfo %v", r)
	return &r, nil
}

type vnRetUser struct {
	Codes     int    `json:"codes"`
	Username  string `json:"username"`
	Id        int64  `json:"id"`
	Sdkuserid int64  `json:"sdkuserid"`
} // {"id":28427284,"username":"U58369566A","codes":"0","cmStatus":null,"sdkuserid":"U58369566A"}
