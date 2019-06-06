package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/jackc/pgx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

var pool *pgx.ConnPool

func readFile(filename string) (*simplejson.Json, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	js, err := simplejson.NewJson(bytes)

	if err != nil {
		return nil, err
	}

	return js, nil
}

func main() {
	data, err := readFile("./acIDInfo.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to readFile:", err)
		os.Exit(1)
	}

	profilestr, err := data.Get("profile").Encode()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to String:", err)
		os.Exit(2)
	}

	bagstr, err := data.Get("bag").Encode()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to String:", err)
		os.Exit(2)
	}

	//u := uuid.NewV4()

	// ./t {host} {port} {db} {user} {pass}
	if len(os.Args) < 6 {
		fmt.Println("args err, use by ./t {host} {port} {db} {user} {pass}")
		os.Exit(3)
		return
	}
	host, portStr, db, user, pass := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("args port err, use by ./t {host} {port} {db} {user} {pass}")
		os.Exit(4)
		return
	}

	store := storehelper.NewStorePostgreSQL(host, port, db, user, pass)

	store.Open()
	defer store.Close()

	rh := func(string) ([]byte, bool) { return []byte{}, false }

	for i := 0; i < 10000; i++ {
		u1 := uuid.NewV4()
		err := store.Put("profile:0:10:"+u1.String(), profilestr, rh)
		if err != nil {
			fmt.Println("put err by %s", err.Error())
			os.Exit(5)
			return
		}
	}

	for i := 0; i < 10000; i++ {
		u1 := uuid.NewV4()
		err := store.Put("bag:0:10:"+u1.String(), bagstr, rh)
		if err != nil {
			fmt.Println("put err by %s", err.Error())
			os.Exit(5)
			return
		}
	}
	logs.Close()
}
