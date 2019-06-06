package storehelper

import (
	"testing"
)

var (
	OSS_Bucket    = "jws-teamcity"
	OSS_EndPoint  = "oss-ap-southeast-1.aliyuncs.com"
	OSS_AccessKey = "LTAIMYKZiHZ2wrGq"
	OSS_SecretKey = "o1UCzrPk2alO5cO1NPLE1z7P7FnqGn"
)

func TestStoreOSS(t *testing.T) {
	ossStore := NewStore(OSS, OSSInitCfg{
		EndPoint:  OSS_EndPoint,
		AccessId:  OSS_AccessKey,
		AccessKey: OSS_SecretKey,
		Bucket:    OSS_Bucket,
	})
	ossStore.Open()
	key := "test:0:10"
	err := ossStore.Put(key, []byte("北国风光，千里冰封，万里雪飘"), nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	getRes, err := ossStore.Get(key)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(getRes) != "北国风光，千里冰封，万里雪飘" {
		t.Error(string(getRes))
		t.FailNow()
	}
	err = ossStore.Del(key)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
