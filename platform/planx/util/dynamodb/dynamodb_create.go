package dynamodb

import (
	"sync"
)

// ---------------------------------------------------------------------------------------------------------------------
// dynamo 初始化连接池，相同的accessKeyID,secretAccessKey的只初始化一次；
// 当accessKeyID为空时连接池无效，每次都重新连接并初始化

var (
	dynamo_services map[string]*DynamoDB = make(map[string]*DynamoDB, 10)
	dynamo_mutex    sync.RWMutex
)

func DynamoConnectInitFromPool(db *DynamoDB,
	region string,
	table []string, //只是用来检查
	accessKeyID,
	secretAccessKey,
	sessionToken string) (*DynamoDB, error) {
	//TODO: YZH accessKeyID为空，则每次都重新初始化 in DynamoConnectInitFromPool
	if accessKeyID == "" || secretAccessKey == "" {
		if err := db.Connect(region, accessKeyID, secretAccessKey, sessionToken); err != nil {
			return db, err
		}
		if err := db.InitTableAndCheck(table); err != nil {
			return db, err
		}
		return db, nil
	}

	// 首次连接
	dynamo_mutex.Lock()
	defer dynamo_mutex.Unlock()
	_db, ok := dynamo_services[accessKeyID]
	if ok {
		db = _db
	}
	if err := db.Connect(region, accessKeyID, secretAccessKey, sessionToken); err != nil {
		return db, err
	}
	if err := db.InitTableAndCheck(table); err != nil {
		return db, err
	}
	dynamo_services[accessKeyID] = db
	return db, nil
}
