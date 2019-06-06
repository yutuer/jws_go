package encode

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"github.com/qiniu/api.v6/url"
	"sort"
)

const HTTP_Typ = "GET"

func CalSig(args map[string]string, URL string, appKey string) (string, error) {
	srcStr, err := getSrcStr(args, URL, HTTP_Typ)
	if err != nil {
		return "", err
	}
	private_key := appKey + "&"

	hmac_str := hmac_sha1(private_key, srcStr)
	return base64.StdEncoding.EncodeToString(hmac_str), nil
}

func hmac_sha1(key string, srcStr string) []byte {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(srcStr))
	return mac.Sum(nil)
}

func getSrcStr(args map[string]string, URL, typeHttp string) (string, error) {
	encode_url := url.QueryEscape(URL)
	keys := make([]string, 0, len(args))
	for k, _ := range args {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	var kvs string
	for _, k := range keys {
		if kvs == "" {
			kvs = k + "=" + args[k]
		} else {
			kvs = kvs + "&" + k + "=" + args[k]
		}
	}
	url_encode_key := url.QueryEscape(kvs)
	return typeHttp + "&" + encode_url + "&" + url_encode_key, nil
}

func GetMD5Key(srcStr string) string {
	_md5 := md5.New()
	_md5.Write([]byte(srcStr))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	return string(hexText)
}
