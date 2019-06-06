package main

import (
	"os"
	"strconv"

	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func printErrInfo(format string, v ...interface{}) {
	info := fmt.Sprintf(format, v...)
	infoJson := struct {
		ErrInfo string
	}{
		ErrInfo: info,
	}
	infoJsonStr, _ := json.Marshal(infoJson)
	fmt.Println(string(infoJsonStr))
	return
}

func main() {
	logs.Close() // 不打日志
	if len(os.Args) < 4 {
		printErrInfo("os.Args Err %v, Use By ./load_tool :6379 0 {accountID}", os.Args)
		return
	}

	acID, err := db.ParseAccount(os.Args[3])
	if err != nil {
		printErrInfo("os.Args Err %v By %s", os.Args, err.Error())
		return
	}

	redisAddress := os.Args[1]
	redisNum, err := strconv.Atoi(os.Args[2])

	if err != nil {
		printErrInfo("os.Args Err %v By %s", os.Args, err.Error())
		return
	}

	driver.SetupRedisForSimple(redisAddress, redisNum, "", false)
	defer driver.ShutdownRedis()

	acc, err := account.LoadFullAccount(acID, true)

	if err != nil {
		printErrInfo("LoadAllAccount Err %v By %s", os.Args, err.Error())
		return
	}

	accjson, err := acc.ToJSON()

	if err != nil {
		printErrInfo("ToJSON Err %v By %s", os.Args, err.Error())
		return
	}

	fmt.Println(string(accjson))
}
