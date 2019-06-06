package db

import (
	"fmt"
	"testing"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
)

func TestMap2Struct(t *testing.T) {
	db := &DBByDynamoDB{}
	dataMap := make(map[string]interface{})
	dataMap["id"] = "fsdfsdfsd"
	dataMap["gift_id"] = 1
	dataMap["times"] = 1323121
	dataMap["get_time"] = 1493000132
	dataMap["int_param"] = 2

	yybInfo, err := db.map2struct(dataMap)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	fmt.Println(yybInfo)

	dataMap2, err := db.struct2map(yybInfo)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	fmt.Println(dataMap2)
	yybInfo2, err := db.map2struct(dataMap2)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if *yybInfo != *yybInfo2 {
		t.FailNow()
	}
	fmt.Println(yybInfo2)
}

func TestCheckSend(t *testing.T) {
	defer logs.Close()
	config.LoadConfig("config.toml")

	config.LoadGiftData(config.GetDataPath() + "/yybgift.data")

	config.InitErrCode()

	Init()

	retCode := CheckSend("0e5ad65baa63ab717c3c6ef09790328754a1eb45@04211725151337@apple.com", 11477)
	if retCode != 0 {
		t.FailNow()
	}
}
