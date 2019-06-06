package sdk

import (
	"fmt"

	"time"

	"github.com/astaxie/beego/httplib"
	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
)

const (
	Typ_Android = "0"
	Typ_IOS     = "1"
)

func CheckUser(uid, token, typ string, isMuBao1 bool) error {
	err := backoff.Retry(
		func() error {
			req := httplib.Post(authConfig.QuickSdkCfg.Url).SetTimeout(
				500*time.Millisecond, 500*time.Millisecond)
			req.Param("uid", uid)
			req.Param("token", token)
			if typ == Typ_IOS {
				req.Param("product_code", authConfig.QuickSdkCfg.IOSProductCode)
			} else {
				if isMuBao1 {
					req.Param("product_code", authConfig.QuickSdkCfg.AndroidProductCode_MuBao1)
				} else {
					req.Param("product_code", authConfig.QuickSdkCfg.AndroidProductCode)
				}
			}

			str, err := req.String()
			if err != nil {
				logs.Warn("quick sdk CheckUser req.String() err %s", err.Error())
				return err
			}
			if str != "1" {
				logs.Warn("quick sdk CheckUser fail ret %s", str)
				return fmt.Errorf("quick sdk checkuser err, receive: %s", str)
			}
			return nil
		},
		util.New2SecBackOff(),
	)

	if err != nil {
		return fmt.Errorf("quick sdk CheckUser backoff err %s", err.Error())
	}

	return nil
}
