package account2json

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"os"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
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

func Imp(conn redispool.RedisPoolConn, acID string) string {
	res := simplejson.New()

	redisKeys := []string{"profile", "bag", "store", "pguild", "tmp", "simpleinfo", "general", "anticheat", "friend"}

	for _, key := range redisKeys {
		profile_res, err := accountJson.MkTrueJsonFromRedis(redis.StringMap(conn.Do("HGETALL", key+":"+acID)))
		if err != nil {
			return PrintErrInfo("Redis redisKeys %s Err Nil %s", key, err.Error())
		}

		res.Set(key, profile_res)
	}

	str, err := res.Encode()

	if err != nil {
		return PrintErrInfo("Redis to string %v By %s", os.Args, err.Error())
	}

	return string(str)
}
