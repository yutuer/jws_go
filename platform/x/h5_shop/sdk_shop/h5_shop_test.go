package sdk_shop

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func TryUnmarshalqqq(bytes []byte) []byte {
	logs.Debug("%s\n",bytes)
	bytLen := len(bytes)
	if bytLen > 50 {
		bytes[1], bytes[33] = bytes[33], bytes[1]
		bytes[10], bytes[42] = bytes[42], bytes[10]
		bytes[18], bytes[50] = bytes[50], bytes[18]
		bytes[19], bytes[51] = bytes[51], bytes[19]
	}
	bsOut, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		logs.Error("json: error decoding base64 binary '%s': %v", bsOut, err)
		return nil
	}
	logs.Debug("TryUnmarshal %s\n", bsOut)
	return bsOut
}

type QAQ struct {
	Gameid   int `json:"gameid"`
	Roleid   int `json:"roleid"`
	Serverid int `json:"serverid"`
}

func TestRandSet(t *testing.T) {
	bytes := []byte(string("eTJnYW1laWViOjgsInQxa2VuIjoiNTc5OydmOTI1YTQkZDUyZWRvNDYwYzg3YzUzYjBmOTAifQ=="))
	fmt.Printf("%s", bytes)
	fmt.Println()
	fmt.Println(len(bytes))
	dst := TryUnmarshalqqq(bytes)

	fmt.Printf("%s", dst)
	fmt.Println()
	fmt.Println(len(dst))
	tq := QAQ{}
	json.Unmarshal(dst, &tq)
	fmt.Println(tq)
}
