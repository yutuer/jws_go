package mailhelper

import (
	"time"

	"gopkg.in/mgo.v2"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/planx/util/timongodb"
)

type MailConfig struct {
	AWSRegion    string
	DBName       string
	AWSAccessKey string
	AWSSecretKey string

	MongoDBUrl string
	DBDriver   string
}

func NewMailDriver(mc MailConfig) (timail.Timail, error) {
	switch mc.DBDriver {
	case "MongoDB":
		mdb, err := initMongoDB(mc.MongoDBUrl, mc.DBName)
		if err != nil {
			return nil, err
		}
		return mdb, nil
	case "DynamoDB":
		fallthrough
	default:
		dydb, err := initDynamoDB(
			mc.AWSRegion, mc.DBName,
			mc.AWSAccessKey, mc.AWSSecretKey,
		)
		if err != nil {
			return nil, err
		}
		logs.Debug("new mail driver db name: %v", mc.DBName)
		return dydb, nil
	}
	return nil, nil
}

func initDynamoDB(region, db_name, accessKey, secretKey string) (timail.Timail, error) {
	dydb := dynamodb.NewMailDynamoDB(region, db_name, accessKey, secretKey)

	err := dydb.Open()
	if err != nil {
		logs.Error("initDB NewMailDynamoDB Err by %s", err.Error())
		return nil, err
	}

	return &dydb, nil
}

func initMongoDB(url, dbname string) (timail.Timail, error) {
	session, err := mgo.DialWithTimeout(url, 20*time.Second)
	if err != nil {
		return nil, err
	}

	var mdb timongodb.DBByMongoDB
	mdb.Session = session
	mdb.DBName = dbname
	session.SetMode(mgo.Monotonic, true)

	return &mdb, nil
}
