package tipay

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/tipay/dynamopay"
	"vcs.taiyouxi.net/platform/planx/util/tipay/mongodbpay"
	. "vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

func NewPayDriver(mc PayDBConfig) (PayDB, error) {
	switch mc.DBDriver {
	case "MongoDB":
		mdb, err := mongodbpay.NewMongoDBPay(mc)
		if err != nil {
			return nil, err
		}
		return mdb, nil
	case "DynamoDB":
		fallthrough
	default:
		dbPayMgr := dynamopay.NewPayDynamoDB(mc.DBName, mc.AWSRegion, mc.AWSAccessKey, mc.AWSSecretKey)
		err := dbPayMgr.Open()
		if err != nil {
			logs.Error("initDB NewPayDynamoDB Err by %s", err.Error())
			return nil, err
		}
		return dbPayMgr, nil
	}
	return nil, nil
}
