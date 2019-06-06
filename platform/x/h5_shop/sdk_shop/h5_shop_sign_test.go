package sdk_shop

import (
	"testing"
	"fmt"
	"strings"
	"crypto/md5"
	"encoding/hex"
	"net/url"
)

func TryUnmarshalzzz(params map[string]string, key string) (string ,error) {
	kvs := make([]string, 0, len(params)+1)

	kvs = append(kvs, "data"+"="+params["data"])
	kvs = append(kvs, "timestamp"+"="+params["timestamp"])
	//kvs = append(kvs, key)
	str := strings.Join(kvs, "&")
	fmt.Printf("srcStr is: %s\n", str)
	dest, err := url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	fmt.Printf("decode url is: %s\n", dest)
	_md5 := md5.New()
	_md5.Write([]byte(dest))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	fmt.Printf("final str: %s\n", string(hexText))
	return string(hexText), nil
}


func TestRandaaa(t *testing.T) {
	key := "b0802611d6a412603f7dfc3bf99494be"
	param := make(map[string]string , 0)
	param["data"] = "eTJnYW1laWViOjgsInQxa2VuIjoiNTc5OydmOTI1YTQkZDUyZWRvNDYwYzg3YzUzYjBmOTAifQ=="
	param["timestamp"] = "1482460056"
	signOrg := "de5291ba952058b0589b56479cc450ed"
	fmt.Printf("%s\n",signOrg)
	dst,_ := TryUnmarshalzzz(param , key)

	fmt.Printf("%s\n",dst)
}
