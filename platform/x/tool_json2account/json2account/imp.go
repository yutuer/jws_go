package json2account

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"vcs.taiyouxi.net/platform/planx/util/account_json"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func PrintErrInfo(format string, v ...interface{}) (res string) {
	info := fmt.Sprintf(format, v...)
	infoJson := struct {
		ErrInfo string
	}{
		ErrInfo: info,
	}
	infoJsonStr, _ := json.Marshal(infoJson)
	res = string(infoJsonStr)
	fmt.Println(res)
	return res
}

func Imp(conn redispool.RedisPoolConn, acID string, bytes []byte) string {

	acc, err := simplejson.NewJson(bytes)
	if err != nil {
		return PrintErrInfo("simplejson.NewJson Err %v By %s", acID, err.Error())
	}

	accMap, err := acc.Map()
	for key, _ := range accMap {
		profilejson := acc.Get(key)
		full := key + ":" + acID
		if profilejson != nil {
			err = accountJson.FromTrueJsonByJson(conn.RawConn(), full, profilejson)
			if err != nil {
				return PrintErrInfo("FromTrueJson Err %v By %s", full, err.Error())
			}
		}
	}
	return ""
}
