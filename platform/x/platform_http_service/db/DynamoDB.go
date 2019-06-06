package db

import (
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/platform_http_service/config"
)

type DBByDynamoDB struct {
	ddb *dynamodb.DynamoDB
}

var DyDb *DBByDynamoDB

func (d *DBByDynamoDB) init(config config.DynamoDBConfig) error {
	logs.Debug("DBByDynamoDB init")
	d.ddb = &dynamodb.DynamoDB{}
	d.ddb.Connect(
		config.AWS_Region,
		config.AWS_KeyId,
		config.AWS_AccessKey,
		config.AWS_SessionToken)

	err := d.ddb.InitTable()
	if err != nil {
		logs.Error("init dynamodb table err %v", err)
		return err
	}
	return nil
}

func InitDynamoDB() {
	DyDb = new(DBByDynamoDB)
	DyDb.init(config.CommonConfig.DynamoConfig)
}

func (d *DBByDynamoDB) GetDb() *dynamodb.DynamoDB {
	return d.ddb
}
