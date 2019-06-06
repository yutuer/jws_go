package util

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	EnjoyAndroidChannel = "5001"
	EnjoyIOSChannel     = "5002"
)

func GenEnjoySign(params map[string]string, key string) (string, error) {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	kvs := make([]string, 0, len(params)+1)
	for _, k := range keys {
		kvs = append(kvs, k+"="+params[k])
	}
	str := strings.Join(kvs, "&")
	logs.Debug("srcStr is: %s", str)
	dest, err := url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	logs.Debug("decode url is: %s", dest)
	_md5 := md5.New()
	_md5.Write([]byte(dest + key))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	logs.Debug("final str: %s", string(hexText))
	return string(hexText), nil
}
