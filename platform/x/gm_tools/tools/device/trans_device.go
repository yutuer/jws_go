package device

import (
	"errors"
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

var dyDb *dynamodb.DynamoDB

func RegCommands() {
	gm_command.AddGmCommandHandle("transDevice", transDevice)
}

func InitDb(cfg config.CommonConfig) {
	dyDb = &dynamodb.DynamoDB{}
	dyDb.Connect(cfg.AWS_Region, cfg.AWS_AccessKey, cfg.AWS_SecretKey, "")
	dyDb.InitTable()
}

func transDevice(c *gm_command.Context, server, accountid string, params []string) error {
	if len(params) != 3 {
		return errors.New(fmt.Sprintf("param length error, %v", params))
	}

	acid1 := params[0]
	acid2 := params[1]
	account1, err := db.ParseAccount(acid1)
	if err != nil {
		logs.Error("account error %v", err)
		return errors.New(fmt.Sprintf("account error, %v", err))
	}
	tableName := fmt.Sprintf("UserInfo_G%d", account1.GameId)
	logs.Debug("table name %s", tableName)

	account2, err := db.ParseAccount(acid2)
	if err != nil {
		logs.Error("account error %v", err)
		return errors.New(fmt.Sprintf("account error, %v", err))
	}

	uid1 := fmt.Sprintf("uid:%s", account1.UserId.String())
	uid2 := fmt.Sprintf("uid:%s", account2.UserId.String())

	device1 := GetDeviceId(dyDb, tableName, uid1)
	device2 := GetDeviceId(dyDb, tableName, uid2)

	if device1 == "" || device2 == "" {
		return errors.New(fmt.Sprintf("device nil, %s, %s, %s", tableName, device1, device2))
	}
	logs.Debug("%s, %s", uid1, device1)
	logs.Debug("%s, %s", uid2, device2)

	if err := UpdateDeviceId(dyDb, tableName, uid1, device2); err != nil {
		return errors.New(fmt.Sprintf("update device err, %s, %v", tableName, err))
	}
	if err := UpdateDeviceId(dyDb, tableName, uid2, device1); err != nil {
		return errors.New(fmt.Sprintf("update device err, %s, %v", tableName, err))
	}
	deviceTableName := fmt.Sprintf("Devices_G%d", account1.GameId)
	if err := updateUserInfo(dyDb, deviceTableName, account1.UserId.String(), device2); err != nil {
		return errors.New(fmt.Sprintf("update user info err, %s, %v", deviceTableName, err))
	}
	if err := updateUserInfo(dyDb, deviceTableName, account2.UserId.String(), device1); err != nil {
		return errors.New(fmt.Sprintf("update user info err, %s, %v", deviceTableName, err))
	}
	return nil
}

func GetDeviceId(dyDb *dynamodb.DynamoDB, name string, uid string) string {
	data, err := dyDb.GetByHashM(name, uid)
	if err != nil {
		logs.Error("get device id %v", err)
		return ""
	}
	return data["device"].(string)
}

func UpdateDeviceId(dyDb *dynamodb.DynamoDB, tableName string, uid string, deviceId string) error {
	return dyDb.UpdateByHash(tableName, uid, map[string]interface{}{
		"device": deviceId,
	})
}

func updateUserInfo(dyDb *dynamodb.DynamoDB, tableName string, uid string, deviceId string) error {
	return dyDb.UpdateByHash(tableName, deviceId, map[string]interface{}{
		"user_id": uid,
	})
}
