package hmt_gift

import (
	"crypto/md5"
	"encoding/base64"
	"strconv"
	"strings"

	"encoding/hex"
	"fmt"
	"net/url"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	legal_gid = []string{"0", "1", "200", "201", "202", "203"}
	swap      = [][]int{{1, 33}, {10, 42}, {18, 50}, {19, 51}}
)

func ParseServerID(numberServerID int) (string, bool) {
	str := strconv.Itoa(numberServerID)
	for _, prefix := range legal_gid {
		if strings.HasPrefix(str, prefix) {
			sid := strings.TrimPrefix(str, prefix)
			serverID := prefix + ":" + sid
			logs.Debug("parse server id: %s", serverID)
			return serverID, true
		}
	}
	return "", false
}

func DecodeData(srcData []byte) ([]byte, error) {
	if len(srcData) > 52 {
		for _, i := range swap {
			srcData[i[0]], srcData[i[1]] = srcData[i[1]], srcData[i[0]]
		}
	}
	dst := make([]byte, 1000)
	n, err := base64.StdEncoding.Decode(dst, srcData)
	if err != nil {
		return nil, err

	}
	dst = dst[0:n]
	return dst, nil
}

func CalSign(data string) (string, error) {
	n, err := url.QueryUnescape(data)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	dataKey := n + "a5ced306175ff1deaff676da872c05c5"

	logs.Debug("DateKey Date: %s",dataKey)

	_md5 := md5.New()
	_md5.Write([]byte(dataKey))
	ret := _md5.Sum(nil)

	return hex.EncodeToString(ret), nil
}
