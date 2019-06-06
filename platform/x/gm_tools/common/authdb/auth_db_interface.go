package authdb

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authCfg "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

var (
	authDB models.DBInterface
)

func Init(cfg config.CommonConfig) {
	authCfg.Cfg.Dynamo_NameDevice = cfg.AuthDynamoDBDevice
	authCfg.Cfg.Dynamo_NameName = cfg.AuthDynamoDBName
	authCfg.Cfg.Dynamo_NameUserInfo = cfg.AuthDynamoDBUserInfo
	logs.Info("gmtools init auth models %v", cfg)
	switch cfg.AuthDBDriver {
	case "MongoDB":
		authDB = &models.DBByMongoDB{}
		authDB.Init(models.DBConfig{
			MongoDBName: cfg.AuthMongoDBName,
			MongoDBUrl:  cfg.AuthMongoDBUrl,
		})
	case "DynamoDB":
		fallthrough
	default:
		authDB = &models.DBByDynamoDB{}
		authDB.Init(models.DBConfig{
			DynamoRegion:          cfg.AWS_Region,
			DynamoAccessKeyID:     cfg.AWS_AccessKey,
			DynamoSecretAccessKey: cfg.AWS_SecretKey,
			DynamoSessionToken:    "",
		})
	}

}

func GetDB() models.DBInterface {
	return authDB
}
