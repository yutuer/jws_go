package main

import (
	"fmt"
	"os"

	"strconv"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

func main() {
	// ./t {host} {port} {db} {user} {pass}
	if len(os.Args) < 6 {
		fmt.Println("args err, use by ./t {host} {port} {db} {user} {pass}")
		os.Exit(1)
		return
	}
	host, portStr, db, user, pass := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("args port err, use by ./t {host} {port} {db} {user} {pass}")
		os.Exit(2)
		return
	}

	store := storehelper.NewStorePostgreSQL(host, port, db, user, pass)

	store.Open()
	defer store.Close()

	storehelper.CreateAllTable(store.GetPool())
	logs.Close()
}
