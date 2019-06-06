package main

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	AWS_Region = "cn-north-1"
	//AWS_KeyId        = "AKIAOXC6JHPK7YYB6QBA"
	//AWS_AccessKey    = "PxjEikOdYwPhbtlEwyciAP8LmXi9ea+AV7AI7k+f"
	AWS_KeyId        = ""
	AWS_AccessKey    = ""
	AWS_SessionToken = ""
)

func main() {
	defer logs.Close()
	logs.Debug("DBByDynamoDB init")
	db := &dynamodb.DynamoDB{}
	db.Connect(AWS_Region, AWS_KeyId, AWS_AccessKey, AWS_SessionToken)

	db.InitTable()

	acidFromList := []string{
		"200:1054:7adcdc99-8ff4-4df0-a1c3-63d37ceb4e31",
		"200:1069:007c81a6-f05d-4693-8c2b-22b9a54232f9",
		"200:1045:3d0488b3-e607-4dd6-8505-643b498e34a0",
		"200:1039:4f64c1aa-d807-46e7-bf5f-a2fb21a5140a",
		"200:1065:a73738e7-1da6-4d4c-99d8-fded31814c1e",
		"200:1063:4aeef68f-4e18-411c-8fbd-4767a53e8521",
		"200:1058:ed2ba643-e2b8-4646-a6bb-91c0a46db55d",
		"200:1034:7c1889c6-05f4-4e06-b2e3-7a15e82872d4",
		"200:1031:896792e4-94be-438a-8ad6-a907ec55d0cc",
		"200:1057:3de5601a-dcd8-4bb1-b237-47513fc34b17",
	}
	acidToList := []string{
		"200:1054:75154dd2-a442-4c08-9603-f4e6afad36e8",
		"200:1069:77bf5148-87b2-4c19-a024-23c9aaf984d9",
		"200:1045:b45d1597-2654-4a09-b10d-3ead8173d78c",
		"200:1039:02b1766f-0c70-49bb-8030-425f03492565",
		"200:1065:08d91c74-14f0-4da3-b681-d99306f98e2b",
		"200:1063:8e4fa07a-331c-4705-9d44-1e7e6993ef8f",
		"200:1058:cad778b8-5386-4848-af10-28b936a127e6",
		"200:1034:f04b2984-7bed-409d-b4b6-330910bde73e",
		"200:1031:c2387183-f7bc-41e0-817e-a04e7901d470",
		"200:1057:338124c7-d40d-4840-a6e0-7b127a764286",
	}
	for i := range acidFromList {
		change(db, acidFromList[i], acidToList[i])
	}
}

func change(dyDb *dynamodb.DynamoDB, acid1, acid2 string) {
	account1, err := db.ParseAccount(acid1)
	if err != nil {
		logs.Error("query out error %v", err)
	}
	logs.Debug("account %v", account1.UserId)
	tableName := fmt.Sprintf("UserInfo_G%d", account1.GameId)
	logs.Debug("table name %s", tableName)

	account2, err := db.ParseAccount(acid2)

	uid1 := fmt.Sprintf("uid:%s", account1.UserId.String())
	uid2 := fmt.Sprintf("uid:%s", account2.UserId.String())

	device1 := GetDeviceId(dyDb, tableName, uid1)
	device2 := GetDeviceId(dyDb, tableName, uid2)
	logs.Debug("%s, %s", uid1, device1)
	logs.Debug("%s, %s", uid2, device2)

	//UpdateDeviceId(dyDb, tableName, uid1, device2)
	//UpdateDeviceId(dyDb, tableName, uid2, device1)
	//
	//deviceTableName := fmt.Sprintf("Device_G%d", account1.GameId)
	//updateUserInfo(dyDb, deviceTableName, account1.UserId.String(), device2)
	//updateUserInfo(dyDb, deviceTableName, account2.UserId.String(), device1)
}

func GetDeviceId(dyDb *dynamodb.DynamoDB, name string, uid string) string {
	data, err := dyDb.GetByHashM(name, uid)
	if err != nil {
		logs.Error("get device id %v", err)
		return ""
	}
	return data["device"].(string)
}

func UpdateDeviceId(dyDb *dynamodb.DynamoDB, tableName string, uid string, deviceId string) {
	dyDb.UpdateByHash(tableName, uid, map[string]interface{}{
		"device": deviceId,
	})
}

func updateUserInfo(dyDb *dynamodb.DynamoDB, tableName string, uid string, deviceId string) {
	dyDb.UpdateByHash(tableName, deviceId, map[string]interface{}{
		"user_id": uid,
	})
}
