package pay

//var (
//	DBPayRelease *dynamodb.DbPayDynamoDB
//	DBPayDebug   *dynamodb.DbPayDynamoDB
//)

//
//func GenPayDynamoDB(region,
//	accessKeyID,
//	secretAccessKey string) (*dynamodb.DbPayDynamoDB, error) {
//	dbPayMgr := dynamodb.NewPayDynamoDB(region, accessKeyID, secretAccessKey)
//TODO by YZH 这里传入要检查的数据库存在性的列表
//	err := dbPayMgr.Open(nil)
//	if err != nil {
//		logs.Error("initDB NewPayDynamoDB Err by %s", err.Error())
//		return nil, err
//	}
//
//	return dbPayMgr, nil
//}
