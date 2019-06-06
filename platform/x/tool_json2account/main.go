package main

import (
	"io/ioutil"
	"os"
	"strconv"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/account_json"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/tool_json2account/json2account"
)

func readFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func main() {
	logs.Close() // 不打日志
	if len(os.Args) < 6 {
		json2account.PrintErrInfo("os.Args Err %v, Use By ./load_tool :6379 0 dbpassword {filename} {accountID}", os.Args)
		return
	}

	acID := os.Args[5]
	redisAddress := os.Args[1]
	redisNum, err := strconv.Atoi(os.Args[2])
	redispwd := os.Args[3]

	if err != nil {
		json2account.PrintErrInfo("os.Args Err %v By %s", os.Args, err.Error())
		return
	}

	pool := accountJson.NewRedisPool(redisAddress, redispwd, redisNum, 10)

	conn := pool.GetDBConn()
	if conn.IsNil() {
		json2account.PrintErrInfo("Redis Conn Nil %v", os.Args)
		return
	}

	bytes, err := readFile(os.Args[4])

	if err != nil {
		json2account.PrintErrInfo("LoadAllAccount Err %v By %s", os.Args, err.Error())
		return
	}

	if json2account.Imp(conn, acID, bytes) == "" {
		fmt.Println("{\"res\":\"ok\"}")
	}
}
