package sdk

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/httplib"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
)

const (
	from_name = "taiyouxi"
)

func CheckVivo(token string) (error, string) {
	req := httplib.Post(authConfig.VivoSdkCfg.Url).SetTimeout(
		500*time.Millisecond, 500*time.Millisecond)
	req.Param("authtoken", token)
	req.Param("from", from_name)

	var r ret
	err := req.ToJson(&r)
	if err != nil {
		return err, ""
	}

	if r.RetCode != 0 {
		return fmt.Errorf("vivo sdk CheckVivo err, receive: %s", r), ""
	}
	return nil, r.Data.OpenId
}

type ret struct {
	RetCode int     `json:"retcode"`
	Data    retdata `json:"data"`
}

type retdata struct {
	Success bool   `json:"success"`
	OpenId  string `json:"openid"`
}
