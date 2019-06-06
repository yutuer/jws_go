package redeemCodeModule

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//"github.com/cenkalti/backoff"
//"time"

//RedeemCodeDynamoDB 礼品码记录DynamoDB存取接口
type RedeemCodeDynamoDB struct {
	db        *dynamodb.DynamoDB
	region    string
	dbName    string
	accessKey string
	secretKey string
	isInited  bool
}

func NewRedeemCodeDynamoDB(region, dbName, accessKey, secretKey string) *RedeemCodeDynamoDB {
	db := &dynamodb.DynamoDB{}

	return &RedeemCodeDynamoDB{
		db:        db,
		region:    region,
		dbName:    dbName,
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

func (s *RedeemCodeDynamoDB) Clone() RedeemCodeDynamoDB {
	db := &dynamodb.DynamoDB{}
	return RedeemCodeDynamoDB{
		db:        db,
		region:    s.region,
		dbName:    s.dbName,
		accessKey: s.accessKey,
		secretKey: s.secretKey,
	}
}

func (s *RedeemCodeDynamoDB) Open() error {
	db, err := dynamodb.DynamoConnectInitFromPool(s.db,
		s.region,
		[]string{s.dbName},
		s.accessKey,
		s.secretKey,
		"")
	if err != nil {
		return err
	}
	s.db = db
	s.isInited = true
	return nil
}

func (s *RedeemCodeDynamoDB) Close() error {
	return nil
}

func (s *RedeemCodeDynamoDB) IsHasInited() bool {
	return s.isInited
}

func (s *RedeemCodeDynamoDB) SetRedeemCodeUsed(acID, code string) error {
	values := make(map[string]interface{}, 3)

	values["DoneBy"] = acID
	values["State"] = RedeemCodeStateDone

	return s.db.UpdateByHash(s.dbName, code, values)
}

/*
	DoneBy string  用户gid:sid:uuid
	Bind   string  gid, gid:sid,
	State  int     New|Done|OutdatedDone
	Begin  int     开始时间
	End    int     结束时间
	Value  string  Json内容 {
		BatchID string   `json:"bid"`
		GroupID string   `json:"gid"`
		ItemIDs []string `json:"items"`
		Counts  []uint32 `json:"counts"`
		Title string   `json:"title"`
	}
*/
func (s *RedeemCodeDynamoDB) GetRedeemCodeData(code string) (*RedeemCodeExchange, error) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		logs.Error("[%s]SyncMail Panic Err %v", user_id, err)
	//	}
	//}()

	data, err := s.db.GetByHashM(s.dbName, code)
	if err != nil {
		return nil, err
	}

	r := &RedeemCodeExchange{}

	var ok bool = false

	values, ok := dynamodb.GetFromAnys2String("RedeemCodeData", "Value", data)
	if !ok {
		return nil, fmt.Errorf("Value no ok by %s", code)
	}

	err = json.Unmarshal([]byte(values), r)
	if err != nil {
		logs.Error("Value Err by Value %s in %s", err.Error(), "")
		return nil, fmt.Errorf("Value Unmarshal no ok by %s", code)
	}

	r.DoneBy, ok = dynamodb.GetFromAnys2String("RedeemCodeData", "DoneBy", data)
	if !ok {
		r.DoneBy = ""
	}

	r.Bind, ok = dynamodb.GetFromAnys2String("RedeemCodeData", "Bind", data)
	if !ok {
		r.Bind = ""
	}

	r.State, ok = dynamodb.GetFromAnys2int64("RedeemCodeData", "State", data)
	if !ok {
		r.State = RedeemCodeStateNew
	}

	r.Begin, ok = dynamodb.GetFromAnys2int64("RedeemCodeData", "Begin", data)
	if !ok {
		return nil, fmt.Errorf("Begin no ok by %s", code)
	}

	r.End, ok = dynamodb.GetFromAnys2int64("RedeemCodeData", "End", data)
	if !ok {
		return nil, fmt.Errorf("End no ok by %s", code)
	}

	r.Code = code

	logs.Trace("GetRedeemCodeData %v", r)

	return r, nil
}
