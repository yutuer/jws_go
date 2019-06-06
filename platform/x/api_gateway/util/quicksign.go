package util

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func Decode(src, key string) (string, error) {
	if src == "" {
		return src, nil
	}
	p, _ := regexp.Compile("\\d+")
	ss := p.FindAllString(src, -1)

	ii := make([]int8, 0, len(ss))
	for _, s := range ss {
		i, err := strconv.Atoi(s)
		if err != nil {
			return "", err
		}
		ii = append(ii, int8(i))
	}
	if len(ii) <= 0 {
		return src, nil
	}
	data := make([]byte, len(ii))
	bkey := []byte(key)

	for i := 0; i < len(data); i++ {
		data[i] = byte(ii[i] - int8(0xff&bkey[i%len(bkey)]))
	}
	return string(data), nil
}

func GenSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	kvs := make([]string, 0, len(params)+1)
	for _, k := range keys {
		kvs = append(kvs, k+"="+params[k])
	}
	str := strings.Join(kvs, "&")

	bb := md5.Sum([]byte(str))
	// [1,13]
	bbres := exchgBit(bb, 1, 13)
	// [5,17]
	bbres = exchgBit(bbres, 5, 17)
	// [7,23]
	bbres = exchgBit(bbres, 7, 23)

	//	logs.Warn("gensign str %s %x %x", str, bb, bbres)
	return fmt.Sprintf("%x", bbres)
}

func exchgBit(bb [16]byte, pos1, pos2 int) [16]byte {
	b1 := pos1 / 2
	b2 := pos2 / 2
	b1mask := mask(pos1)
	b2mask := mask(pos2)
	temp1 := bb[b1] & b1mask
	temp2 := bb[b2] & b2mask
	bb[b1] = bb[b1]&^b1mask | byte(temp2)
	bb[b2] = bb[b2]&^b2mask | byte(temp1)
	return bb
}

func mask(pos int) byte {
	if pos%2 == 0 {
		return 0xf0
	}
	return 0x0f
}
