package account_info

//import (
//	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
//	"vcs.taiyouxi.net/platform/planx/util/logs"
//)

//XXX by YZH 这个模块禁用了,可以删除

//var (
//	AccountInfoDynamo dynamodb.AccountInfoDynamoDB
//)

////DONE by YZH 利用存档中的DeviceToken替换掉,直接从redis拿到pushtoken
//func InitAccountInfoDynamoDB(region, db_name, accessKey, secretKey string) error {
//	AccountInfoDynamo = dynamodb.NewAccountInfoDynamoDB(
//		region, db_name, accessKey, secretKey)
//
//	err := AccountInfoDynamo.Open()
//	if err != nil {
//		logs.Error("initDB NewAccountInfoDynamoDB Err by %s",
//			err.Error())
//		return err
//	}
//
//	return nil
//}
