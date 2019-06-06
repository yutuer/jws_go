package main

import (
	"fmt"
	"os"
	"strconv"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/platform/planx/util/account_json"
	"vcs.taiyouxi.net/platform/x/tool_account2json/account2json"
)

func main() {
	logs.Close() // 不打日志
	if len(os.Args) < 5 {
		account2json.PrintErrInfo("os.Args Err %v, Use By ./load_tool :6379 0 password {accountID}", os.Args)
		return
	}

	acID := os.Args[4]
	redisAddress := os.Args[1]
	redisNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		account2json.PrintErrInfo("os.Args Err %v By %s", os.Args, err.Error())
		return
	}
	pool := accountJson.NewRedisPool(redisAddress, os.Args[3], redisNum, 10)

	conn := pool.GetDBConn()
	if conn.IsNil() {
		account2json.PrintErrInfo("Redis Conn Nil %v", os.Args)
		return
	}

	fmt.Println(account2json.Imp(conn, acID))

}
