package qiniu

import . "github.com/qiniu/api.v6/conf"

func SetCfg(acckey, seckey string) {
	ACCESS_KEY = acckey
	SECRET_KEY = seckey
	UP_HOST = "http://up-z2.qiniu.com"
}

func IsActivity() bool {
	return ACCESS_KEY != "" && SECRET_KEY != ""
}
