package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"strconv"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	. "vcs.taiyouxi.net/platform/x/auth/models"
)

var mdb DBByMongoDB
var count int

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	item := v.(map[string]interface{})
	return item["S"].(string)
}

func getInt64(v interface{}) int64 {
	item := v.(map[string]interface{})
	if s, err := strconv.ParseInt(item["N"].(string), 10, 32); err != nil {
		logs.Error("parse %v error.", v)
		os.Exit(1)
	} else {
		return s
	}
	return 0
}

func visit(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}
	if b, err := ioutil.ReadFile(path); err != nil {
		logs.Error("ioutil.ReadFile: %s, %s", path, err.Error())
		return err
	} else {
		var all map[string]interface{}
		if err := json.Unmarshal(b, &all); err != nil {
			logs.Error("json.Unmarshal: %s, %s", path, err.Error())
			return err
		}
		items := all["Items"].([]interface{})
		for _, i := range items {
			item := i.(map[string]interface{})
			ma := MongoAuth{
				DevicdID:    getString(item["Id"]),
				DisplanName: getString(item["dn"]),
				//LastGidSid: getString(item["gid_sid"]),
				UserID:       getString(item["user_id"]),
				ChannelID:    getString(item["channelid"]),
				LastAuthTime: getInt64(item["lasttime"]),
				CreateTime:   getInt64(item["createtime"]),
			}

			//fmt.Println("Item %v", ma)
			if err := mdb.GetDB().C("Devices").Insert(ma); err != nil {
				return err
			}
			//fmt.Println("Item %t", item)
		}
		count += int(all["Count"].(float64))
		fmt.Printf("Visited: %s, %v\n", path, all["Count"])
	}

	return nil
}

func main() {
	flag.Parse()

	mdb.Init(DBConfig{
		MongoDBUrl:  flag.Arg(1),
		MongoDBName: flag.Arg(2),
	})

	root := flag.Arg(0)
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v %d\n", err, count)

}
