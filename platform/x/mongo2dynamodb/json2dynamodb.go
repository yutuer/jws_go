package main

/*
	Convert json to dynamoDB
	It doesn't support nested object now
*/

import (
	"bytes"
	"io/ioutil"

	"fmt"

	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-simplejson"
	"gopkg.in/cheggaaa/pb.v1"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type convertV func(tableName string, item string, json *simplejson.Json) interface{}

type result struct {
	src     string
	dst     string
	all     int
	success int
	skip    int
	failed  int
}

var db dynamodb.DynamoDB
var cvtV convertV

func main() {
	defer logs.Close()
	logs.LoadLogConfig("conf/log.xml")
	logs.Info("convert json to dynamoDB")
	var cf FilterConfigs
	if _, err := toml.DecodeFile("conf/config.toml", &cf); err != nil {
		logs.Error("decode config file err by: %v", err)
		return
	}
	logs.Info("load config: %v", cf)

	db = dynamodb.DynamoDB{}
	cvtV = convertMongoKV

	err := db.Connect(
		cf.DynamoConfig.DynamoRegion,
		cf.DynamoConfig.DynamoAccessKeyID,
		cf.DynamoConfig.DynamoSecretAccessKey,
		cf.DynamoConfig.DynamoSessionToken,
	)
	if err != nil {
		logs.Error("DB connect err by %v", err)
	}
	logs.Info("connect dynamoDB success")
	err = db.InitTable()
	if err != nil {
		logs.Error("init table err by %v", err)
		return
	}
	logs.Info("DB table inited")
	// 确定Table都存在，如果不存在，直接退出吧
	for _, filter := range cf.FilterConfig {
		for _, dstTable := range filter.Output {
			if !db.IsTableHasCreate(dstTable.Table) {
				logs.Error("no table named %v in dynamoDB, break!", dstTable.Table)
				return
			}
		}
	}
	for _, filter := range cf.FilterConfig {
		for i, _ := range filter.Name {
			json := parseJson("json/" + filter.Name[i])
			for _, dstTable := range filter.Output {
				r := result{}
				r.src = filter.Name[i]
				r.dst = dstTable.Table
				writeDynamoDB(json, dstTable, db, &r)
			}
		}
	}
	logs.Info("convert success")
}

func parseJson(filepath string) *simplejson.Json {
	byteData, err := ioutil.ReadFile(filepath)
	if err != nil {
		logs.Error("read json file err by %v for file %v", err, filepath)
		return nil
	}
	json, err := simplejson.NewFromReader(bytes.NewBuffer(byteData))
	if err != nil {
		logs.Error("build json reader err by %v, for file %v", err, filepath)
		return nil
	}
	return json
}

func writeDynamoDB(json *simplejson.Json, output Output, db dynamodb.DynamoDB, res *result) {
	jsonArray, err := json.Array()
	if err != nil {
		logs.Error("json type assert array err by %v for output table %s", err, output.Table)
	}
	bar := pb.StartNew(len(jsonArray)).Prefix(output.Table)
	var batchUnit = 24
	writeBuffer := make([]map[string]interface{}, 0, batchUnit)
	for i, _ := range jsonArray {
		res.all += 1
		bar.Increment()
		subJson := json.GetIndex(i)
		outputMap := getOutputMapFromConfig(output, subJson)

		// TODO by ljz test interface{} == ""
		if len(outputMap) <= 0 || outputMap[output.Partition] == nil || outputMap[output.Partition] == "" {
			res.skip += 1
			logs.Warn("skip item json item: %v outputMap: %v, for table: %v partition: %v",
				*subJson, outputMap, output.Table, output.Partition)
		} else {
			writeBuffer = append(writeBuffer, outputMap)
		}
		if len(writeBuffer) >= batchUnit || i == len(jsonArray)-1 {
			if err, failCount := db.BatchSetByHashM(output.Table, output.Partition, writeBuffer); err != nil {
				res.failed += failCount
				res.success += len(writeBuffer) - failCount
				logs.Error("import dynamodb err by %v failed for output table: %v",
					err, output.Table)
			} else if failCount > 0 {
				res.failed += failCount
				res.success += len(writeBuffer) - failCount
				logs.Error("import dynamodb err by %v failed for output table: %v",
					err, output.Table)
			} else {
				res.success += len(writeBuffer)
			}
			writeBuffer = make([]map[string]interface{}, 0, batchUnit)
		}
	}
	bar.FinishPrint(fmt.Sprintf("convert %s to dynamoDB table %s result: all: %d, success: %d, failed: %d, skip: %d",
		res.src, res.dst, res.all, res.success, res.failed, res.skip))
}

func getOutputMapFromConfig(c Output, json *simplejson.Json) map[string]interface{} {
	temp := make(map[string]interface{})
	jsonRet, err := json.Map()
	if err != nil {
		logs.Error("json type assert err by %v for output table: %v", err, c.Table)
	}
	if len(c.Item) <= 0 {
		temp = jsonRet
		for _, item := range c.Ignore {
			delete(temp, item)
		}
	} else {
		for _, item := range c.Item {
			temp[item] = jsonRet[item]
		}
	}
	ret := make(map[string]interface{})
	for item, _ := range temp {
		value := cvtV(c.Table, item, json)
		if item == "" || value == nil || value == "" {
			//logs.Warn("skip json sub item: %v for table: %v", *json, c.Table)
			continue
		}
		key := cvtK(item, c.KeyMap)
		ret[key] = value
	}
	return ret
}

func cvtK(key string, keys []string) string {
	keyMap := make(map[string]string, 0)
	for _, item := range keys {
		keys := strings.Split(item, "->")
		if len(keys) < 2 {
			return key
		}
		keyMap[keys[0]] = keys[1]
	}
	newKey, ok := keyMap[key]
	if ok {
		return newKey
	}
	return key
}
