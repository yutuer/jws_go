package accountJson

import (
	"errors"

	"github.com/bitly/go-simplejson"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func MkTrueJsonFromRedis(data map[string]string, err error) (*simplejson.Json, error) {
	if err != nil {
		return nil, err
	}
	json := simplejson.New()
	for k, subStr := range data {
		subJson, err := simplejson.NewJson([]byte(subStr))
		if err != nil {
			json.Set(k, subStr)
		} else {
			json.Set(k, subJson.Interface())
		}
	}

	json, err = MkPrueJsonByJson(json)

	return json, nil
}

func FromTrueJsonByJson(conn redis.Conn, key string, json *simplejson.Json) error {
	json, err := FromPureJsonToOldByJson(json)
	if err != nil {
		return err
	}

	jsonMap, err := json.Map()
	strMap := make(map[string]string, len(jsonMap))
	for key, _ := range jsonMap {
		subjson := json.Get(key)
		if subjson != nil {

			b, err := subjson.Encode()
			if err != nil {
				return err
			}
			str := string(b)
			// 会都一层""
			if str[0] == '"' && str[len(str)-1] == '"' {
				str = str[1:]
				str = str[:len(str)-1]
			}
			strMap[key] = str
		}
	}

	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(strMap)*2+1)
	args = append(args, key)
	for k, v := range strMap {
		args = append(args, k)
		args = append(args, v)
	}
	if len(args) <= 1 { // 空的
		return nil
	}

	_, err = conn.Do("HMSET", args...)
	return err
}

const toJsonKeysKey = "__toJsonKeysByMkPrueJsonTool__"

func MkPrueJsonByJson(json *simplejson.Json) (*simplejson.Json, error) {
	js, toJsonKeys, err := mkPrueJsonFromJson("/", json)
	if err != nil {
		return nil, err
	}

	js.Set(toJsonKeysKey, toJsonKeys)

	return js, nil
}

func MkPrueJson(str string) (*simplejson.Json, error) {
	json, err := simplejson.NewJson([]byte(str))
	if err != nil {
		return nil, err
	}
	return MkPrueJsonByJson(json)
}

func mkPrueJsonFromJson(address string, js *simplejson.Json) (*simplejson.Json, []string, error) {
	json := js

	resJson := simplejson.New()
	jsonMap, err := json.Map()

	toJsonKeys := make([]string, 0, 8)

	if err != nil {
		return json, toJsonKeys, nil
	}

	for key, _ := range jsonMap {
		subs := json.Get(key)
		fullKey := address + "/" + key
		//logs.Trace("sub %s %v", fullKey, subs)

		j, isJson := tryJsonStr(fullKey, subs)
		if isJson {
			subs = j
			//logs.Warn("IsJson %s %v", fullKey, j)
			toJsonKeys = append(toJsonKeys, fullKey)
		}

		subjs, subToJsonKeys, err := mkPrueJsonFromJson(fullKey, subs)
		if err != nil {
			return nil, toJsonKeys, err
		} else {
			resJson.Set(key, subjs.Interface())
			for _, toJsonkey := range subToJsonKeys {
				toJsonKeys = append(toJsonKeys, toJsonkey)
			}
		}
	}

	return resJson, toJsonKeys, nil
}

func tryJsonStr(key string, str *simplejson.Json) (*simplejson.Json, bool) {
	s, err := str.String()
	if err != nil {
		//logs.Error("err isJsonStr %s %s", key, err.Error())
		return str, false
	}
	logs.Trace("jsonstr %s %s", key, s)
	subjson, err := simplejson.NewJson([]byte(s))
	if err == nil {
		return subjson, true
	} else {
		return str, false
	}
}

func toJsonStr(key string, subjson *simplejson.Json) (string, error) {
	logs.Trace("tojson %s %v", key, subjson)
	res, err := subjson.Encode()
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func FromPureJsonToOldByJson(json *simplejson.Json) (*simplejson.Json, error) {
	toJsonKeys := make(map[string]bool, 8)
	// 获取被转成字符串的key的列表
	tjsonjs := json.Get(toJsonKeysKey)
	if tjsonjs == nil {
		return nil, errors.New("No toJsonKeysKey")
	}
	tjsonarray, err := tjsonjs.StringArray()
	if err != nil {
		return nil, errors.New("No toJsonKeysKey by" + err.Error())
	}

	for _, s := range tjsonarray {
		toJsonKeys[s] = true
	}

	// 开始递归还原json
	js, err := fromPureJsonToOldSub("/", toJsonKeys, json)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func FromPureJsonToOld(str string) (*simplejson.Json, error) {
	json, err := simplejson.NewJson([]byte(str))

	if err != nil {
		return nil, err
	}

	return FromPureJsonToOldByJson(json)
}

func fromPureJsonToOldSub(address string, toJsonKeys map[string]bool, js *simplejson.Json) (*simplejson.Json, error) {
	json := js

	resJson := simplejson.New()
	jsonMap, err := json.Map()

	if err != nil {
		return json, nil
	}

	// 注意和提纯过程是反的先处理好子节点, 再看看是不是要把子节点转成字符串
	for key, _ := range jsonMap {

		// 这个不需要还原
		if key == toJsonKeysKey {
			continue
		}
		subs := json.Get(key)
		fullKey := address + "/" + key
		logs.Trace("sub %s %v", fullKey, subs)

		subjs, err := fromPureJsonToOldSub(fullKey, toJsonKeys, subs)
		if err != nil {
			return nil, errors.New("Err by" + fullKey + " " + err.Error())
		}

		_, isHasToJson := toJsonKeys[fullKey]
		if isHasToJson {
			str, err := toJsonStr(fullKey, subjs)
			if err != nil {
				return nil, errors.New("Err by toJsonStr " + fullKey + " " + err.Error())
			}
			resJson.Set(key, str)
		} else {
			resJson.Set(key, subjs)
		}
	}

	return resJson, nil
}
