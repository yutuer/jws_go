package models

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/x/auth/config"
)

const AUTHTOKEN_TIMEOUT = 24 * 60 * 60 //min

type DBConfig struct {
	DynamoRegion          string
	DynamoAccessKeyID     string
	DynamoSecretAccessKey string
	DynamoSessionToken    string

	MongoDBUrl  string
	MongoDBName string
}

type DBByRedis struct {
}

var (
	loginRedisPool redispool.IPool
	authRedisPool  redispool.IPool
	db_config      DBConfig
)

func InitDynamo(config *config.CommonConfig) error {
	db_config.DynamoRegion = config.Dynamo_region
	db_config.DynamoAccessKeyID = config.Dynamo_accessKeyID
	db_config.DynamoSecretAccessKey = config.Dynamo_secretAccessKey
	db_config.DynamoSessionToken = config.Dynamo_sessionToken
	db_config.MongoDBUrl = config.MongoDBUrl
	db_config.MongoDBName = config.MongoDBName
	switch config.AuthDBDriver {
	case "MongoDB":
		logs.Info("Auth start with mongodb %v", db_config)
		db_interface = &DBByMongoDB{}
	case "DynamoDB":
		fallthrough
	default:
		db_interface = &DBByDynamoDB{}
		InitAuthRedis(config)
	}
	logs.Info("Auth start with %s %v", config.AuthDBDriver, db_config)
	return db_interface.Init(db_config)
}

func InitLoginRedis(config *config.CommonConfig) {
	loginRedisPool = newRedisPool("login.redis",
		config.LoginRedisAddr,
		config.LoginRedisDBPwd,
		config.LoginRedisDB)
}

func InitAuthRedis(config *config.CommonConfig) {
	authRedisPool = newRedisPool("auth.redis",
		config.AuthRedisAddr,
		config.AuthRedisDBPwd,
		config.AuthRedisDB)
}

func makeAuthTokenKey(authToken string) string {
	return fmt.Sprintf("at:%s", authToken)
}
