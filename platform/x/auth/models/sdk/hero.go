package sdk

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"

	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
)

//const (
//	connectTimeout   = 5 * time.Second
//	readWriteTimeout = 5 * time.Second
//)

func GetToken(authcode string) (string, error) {
	params := map[string]string{}
	params["productId"] = authConfig.HeroSdkCfg.ProductId
	params["projectId"] = authConfig.HeroSdkCfg.ProjectId
	params["__e__"] = "1"
	params["client_id"] = authConfig.HeroSdkCfg.AppKey
	params["client_secret"] = authConfig.HeroSdkCfg.ClientSecret
	params["grant_type"] = "authorization_code"
	params["redirect_uri"] = "1"
	params["serverId"] = authConfig.HeroSdkCfg.ServerId
	params["code"] = authcode
	params["sign"] = genSign(params, authConfig.HeroSdkCfg.AppKey)

	// .SetTimeout(connectTimeout, readWriteTimeout)
	req := httplib.Post(authConfig.HeroSdkCfg.Url + authConfig.HeroSdkCfg.UrlToken)
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
		return "", fmt.Errorf("hero gettoken response errcode %d", r.Codes)
	}
	return r.AccessToken, nil
}

func GetUserInfo(token string) (int64, string, error) {
	params := map[string]string{}
	params["__e__"] = "1"
	params["access_token"] = token
	params["productId"] = authConfig.HeroSdkCfg.ProductId
	params["sign"] = genSign(params, authConfig.HeroSdkCfg.AppKey)

	// .SetTimeout(connectTimeout, readWriteTimeout)
	req := httplib.Post(authConfig.HeroSdkCfg.Url + authConfig.HeroSdkCfg.UrlUserinfo)
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

	var r retUser
	err := req.ToJson(&r)
	if err != nil {
		return -1, "", err
	}
	if r.Codes != 0 {
		return -1, "", fmt.Errorf("hero getuser response errcode %d", r.Codes)
	}
	logs.Debug("hero sdk userinfo %v", r)
	return r.Id, r.Sdkuserid, nil
}

type retToken struct {
	Codes       int    `json:"codes"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type retUser struct {
	Codes     int    `json:"codes"`
	Username  string `json:"username"`
	Id        int64  `json:"id"`
	Sdkuserid string `json:"sdkuserid"`
} // {"id":28427284,"username":"U58369566A","codes":"0","cmStatus":null,"sdkuserid":"U58369566A"}

func genSign(params map[string]string, appKey string) string {
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	kvs := make([]string, 0, len(params)+1)
	for _, k := range keys {
		kvs = append(kvs, k+"="+params[k])
	}
	kvs = append(kvs, appKey)
	str := strings.Join(kvs, "&")

	bb := md5.Sum([]byte(str))
	// [1,13]
	bbres := exchgBit(bb, 1, 13)
	// [5,17]
	bbres = exchgBit(bbres, 5, 17)
	// [7,23]
	bbres = exchgBit(bbres, 7, 23)

	//	logs.Warn("gensign str %s %x %x", str, bb, bbres)
	return fmt.Sprintf("%x", bbres)
}

func exchgBit(bb [16]byte, pos1, pos2 int) [16]byte {
	b1 := pos1 / 2
	b2 := pos2 / 2
	b1mask := mask(pos1)
	b2mask := mask(pos2)
	temp1 := bb[b1] & b1mask
	temp2 := bb[b2] & b2mask
	bb[b1] = bb[b1]&^b1mask | byte(temp2)
	bb[b2] = bb[b2]&^b2mask | byte(temp1)
	return bb
}

func mask(pos int) byte {
	if pos%2 == 0 {
		return 0xf0
	}
	return 0x0f
}
