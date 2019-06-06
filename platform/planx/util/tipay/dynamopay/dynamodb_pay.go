package dynamopay

import (
	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

type DbPayDynamoDB struct {
	*dynamodb.DynamoDB
	dbname    string
	indexname string
	region    string
	accessKey string
	secretKey string
}

func NewPayDynamoDB(dbname, region, accessKey, secretKey string) *DbPayDynamoDB {
	db := &dynamodb.DynamoDB{}

	return &DbPayDynamoDB{
		DynamoDB: db,
		dbname:   dbname,
		//TODO by YZH NewPayDynamoDB create index by accountid when needed.
		indexname: "uid-index",
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

func (s *DbPayDynamoDB) Open() error {
	db, err := dynamodb.DynamoConnectInitFromPool(s.DynamoDB,
		s.region,
		[]string{s.dbname},
		s.accessKey,
		s.secretKey,
		"")
	if err != nil {
		return err
	}
	s.DynamoDB = db
	return nil
}

func (s *DbPayDynamoDB) SetByHashM(hash interface{}, values map[string]interface{}) error {
	return s.DynamoDB.SetByHashM(s.dbname, hash, values)
}

func (s *DbPayDynamoDB) SetByHashM_IfNoExist(hash interface{}, values map[string]interface{}) error {
	return s.DynamoDB.SetByHashM_IfNoExist(s.dbname, hash, values)
}

func (s *DbPayDynamoDB) GetByHashM(hash interface{}) (map[string]interface{}, error) {
	kv, err := s.DynamoDB.GetByHashM(s.dbname, hash)
	return kv, err
}

func (s *DbPayDynamoDB) UpdateByHash(hash interface{}, values map[string]interface{}) error {
	return s.DynamoDB.UpdateByHash(s.dbname, hash, values)
}

func (s *DbPayDynamoDB) IsExist(hash interface{}) (bool, error) {
	res, err := s.DynamoDB.GetByHash(s.dbname, hash)
	if err != nil {
		return false, err
	}
	return res != nil, nil
}

const (
	query_limit = 1000
)

//QueryByUid 只有GMTools在使用这个工呢功能
func (s *DbPayDynamoDB) QueryByUid(ucid string) ([]pay.GateWay_Pay_Info, error) {
	params := &DDB.QueryInput{
		TableName:      aws.String(s.dbname),
		IndexName:      aws.String(s.indexname),
		ConsistentRead: aws.Bool(false),
		ExpressionAttributeValues: map[string]*DDB.AttributeValue{
			":u": dynamodb.CreateAttributeValue(ucid),
		},
		KeyConditionExpression: aws.String("uid = :u"),
		Limit:            aws.Int64(query_limit),
		ScanIndexForward: aws.Bool(true),
	}

	resp, err := s.Client().Query(params)

	if err != nil {
		logs.Error("[QueryByUid] Err by %s", err.Error())
		return nil, err
	}
	if len(resp.LastEvaluatedKey) > 0 {
		logs.Info("[QueryByUid] has LastEvaluatedKey %s", ucid)
	}

	res := make([]pay.GateWay_Pay_Info, 0, len(resp.Items))

	for _, i := range resp.Items {
		var info pay.GateWay_Pay_Info

		order_no, ok := i["order_no"]
		if ok {
			str_order_no, ok := dynamodb.GetItemValue(order_no).(string)
			if ok {
				info.Order = str_order_no
			}
		}
		sn, ok := i["sn"]
		if ok {
			i_sn, ok := dynamodb.GetItemValue(sn).(int64)
			if ok {
				info.SN = i_sn
			}
		}
		uid, ok := i["uid"]
		if ok {
			str_uid, ok := dynamodb.GetItemValue(uid).(string)
			if ok {
				info.Uid = str_uid
			}
		}
		account_name, ok := i["account_name"]
		if ok {
			str, ok := dynamodb.GetItemValue(account_name).(string)
			if ok {
				info.AccountName = str
			}
		}
		role_name, ok := i["role_name"]
		if ok {
			str, ok := dynamodb.GetItemValue(role_name).(string)
			if ok {
				info.RoleName = str
			}
		}
		good_idx, ok := i["good_idx"]
		if ok {
			str, ok := dynamodb.GetItemValue(good_idx).(string)
			if ok {
				info.GoodIdx = str
			}
		}
		good_name, ok := i["good_name"]
		if ok {
			str, ok := dynamodb.GetItemValue(good_name).(string)
			if ok {
				info.GoodName = str
			}
		}
		money_amount, ok := i["money_amount"]
		if ok {
			str, ok := dynamodb.GetItemValue(money_amount).(string)
			if ok {
				info.MoneyAmount = str
			}
		}
		platform, ok := i["platform"]
		if ok {
			str, ok := dynamodb.GetItemValue(platform).(string)
			if ok {
				info.Platform = str
			}
		}
		channel, ok := i["channel"]
		if ok {
			str, ok := dynamodb.GetItemValue(channel).(string)
			if ok {
				info.Channel = str
			}
		}
		status, ok := i["status"]
		if ok {
			str, ok := dynamodb.GetItemValue(status).(string)
			if ok {
				info.Status = str
			}
		}
		tistatus, ok := i["tistatus"]
		if ok {
			str, ok := dynamodb.GetItemValue(tistatus).(string)
			if ok {
				info.TiStatus = str
			}
		}
		hc_buy, ok := i["hc_buy"]
		if ok {
			i_h, ok := dynamodb.GetItemValue(hc_buy).(uint32)
			if ok {
				info.HcBuy = i_h
			}
		}
		hc_give, ok := i["hc_give"]
		if ok {
			i_h, ok := dynamodb.GetItemValue(hc_give).(uint32)
			if ok {
				info.HcGive = i_h
			}
		}
		pay_time_s, ok := i["pay_time_s"]
		if ok {
			str, ok := dynamodb.GetItemValue(pay_time_s).(string)
			if ok {
				info.PayTimeS = str
			}
		}
		pay_time, ok := i["pay_time"]
		if ok {
			str, ok := dynamodb.GetItemValue(pay_time).(string)
			if ok {
				info.PayTime = str
			}
		}
		receiveTimestamp, ok := i["receiveTimestamp"]
		if ok {
			i_t, ok := dynamodb.GetItemValue(receiveTimestamp).(int64)
			if ok {
				info.ReceiveTimeStamp = i_t
			}
		}
		str_is_test := ""
		is_test, ok := i["is_test"]
		if ok {
			str, ok := dynamodb.GetItemValue(is_test).(string)
			if ok {
				str_is_test = str
			}
		}
		str_note := ""
		note, ok := i["note"]
		if ok {
			str, ok := dynamodb.GetItemValue(note).(string)
			if ok {
				str_note = str
			}
		}
		if str_is_test == "1" || str_note == "sandbox" {
			info.IsTest = "Yes"
		} else {
			info.IsTest = "No"
		}

		res = append(res, info)
	}

	logs.Trace("[QueryByUid] Res %v", res)

	return res, nil
}

//type AndroidPayQuery struct {
//	Order_no   string
//	Game_order string
//	Amount     string
//}
//
//func (d *DynamoDB) QueryOrder(table_name, indexname, ucid, tistatus string, limit int64,
//	AWS_InitialInterval int64, AWS_Multiplier float64) ([]AndroidPayQuery, error) {
//	params := &DDB.QueryInput{
//		TableName:      aws.String(table_name),
//		IndexName:      aws.String(indexname),
//		ConsistentRead: aws.Bool(false),
//		ExpressionAttributeValues: map[string]*DDB.AttributeValue{
//			":u":  dynamodb.CreateAttributeValue(ucid),
//			":ts": dynamodb.CreateAttributeValue(tistatus),
//		},
//		KeyConditionExpression: aws.String("uid = :u AND tistatus = :ts"),
//		Limit:                aws.Int64(limit),
//		ProjectionExpression: aws.String("order_no, game_order, amount"),
//		//		ReturnConsumedCapacity: aws.String("TOTAL"),
//		ScanIndexForward: aws.Bool(true),
//	}
//
//	//	b := backoff.NewExponentialBackOff()
//	//	b.InitialInterval = time.Duration(AWS_InitialInterval) * time.Millisecond
//	//	b.MaxElapsedTime = time.Duration(AWS_InitialInterval) * time.Second
//	//	b.Multiplier = AWS_Multiplier
//
//	resp, err := d.Client().Query(params)
//	//	if err != nil {
//	//		ticker := backoff.NewTicker(b)
//	//		resp, err = func(ticker *backoff.Ticker, p *DDB.QueryInput) (*DDB.QueryOutput, error) {
//	//			for _ = range ticker.C {
//	//				output, err := d.client.Query(p)
//	//				if err == nil {
//	//					ticker.Stop()
//	//					return output, err
//	//				} else {
//	//					logs.Error("QueryOrder again Err by %s", err.Error())
//	//				}
//	//			}
//	//			ticker.Stop()
//	//			return nil, errors.New("TimeOut")
//	//		}(ticker, params)
//	//	}
//
//	if err != nil {
//		logs.Error("QueryOrder Err by %s", err.Error())
//		return nil, err
//	}
//	if len(resp.LastEvaluatedKey) > 0 {
//		logs.Info("[QueryOrder] has LastEvaluatedKey %s", ucid)
//	}
//
//	res := make([]AndroidPayQuery, 0, len(resp.Items))
//
//	for _, i := range resp.Items {
//		order_no, ook := i["order_no"]
//		game_order, gook := i["game_order"]
//		amount, aok := i["amount"]
//		if !ook || !gook || !aok {
//			logs.Error("QueryOrder error by %s!", ucid)
//			continue
//		}
//		str_order_no, stok := dynamodb.GetItemValue(order_no).(string)
//		str_game_order, sgook := dynamodb.GetItemValue(game_order).(string)
//		str_amount, saok := dynamodb.GetItemValue(amount).(string)
//		if !stok || !sgook || !saok {
//			logs.Error("QueryOrder getItemValue error by %s!", ucid)
//			continue
//		}
//
//		queryInfo := AndroidPayQuery{
//			Order_no:   str_order_no,
//			Game_order: str_game_order,
//			Amount:     str_amount,
//		}
//		res = append(res, queryInfo)
//	}
//
//	logs.Trace("QueryOrder Res %v", res)
//
//	return res, nil
//}
