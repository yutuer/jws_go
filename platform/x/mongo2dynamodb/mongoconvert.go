package main

import (
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func convertMongoKV(tableName string, item string, json *simplejson.Json) interface{} {
	var key = item
	if _, err := json.Get(item).Map(); err == nil {
		if v, err := json.Get(key).Map(); err == nil {
			for k, nestedV := range v {
				if strings.HasPrefix(k, "$numberLong") {
					i, err := strconv.Atoi(nestedV.(string))
					if err != nil {
						continue
					}
					return i
				}
			}
		} else {
			logs.Error("json get key: %v err by %v", key, err)
		}
	} else {
		if item == "isread" {
			if v, err := json.Get(item).Bool(); err == nil {
				if v == true {
					return "read"
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
		if item == "isget" {
			if v, err := json.Get(item).Bool(); err == nil {
				if v == true {
					return "getted"
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
		if tableName == "UserInfo_G203" {
			if item == "user_id" {
				if v, err := json.Get(item).String(); err == nil {
					return "uid:" + v
				} else {
					return nil
				}
			}
		}
		if v, err := json.Get(item).Int(); err == nil {
			return v
		}
		if v, err := json.Get(item).String(); err == nil {
			if v == "" {
				return nil
			}
			return v
		}
		if v := json.Get(item).Interface(); v != nil {
			return v
		}
	}
	return nil
}
