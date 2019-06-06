package dynamodb

//
//import (
//	"strings"
//
//	"strconv"
//
//	"github.com/aws/aws-sdk-go/aws"
//	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
//	"vcs.taiyouxi.net/platform/planx/util/logs"
//)
//
//const (
//	key_account_id   = "account_id"
//	key_platform_id  = "platform_id"
//	key_device_token = "device_token"
//	Key_device_info  = "device_info"
//)
//
////XXX by YZH 利用存档中的DeviceToken替换掉,直接从redis拿到pushtoken
//
//type AccountInfoDynamoDB struct {
//	db        *DynamoDB
//	region    string
//	db_name   string
//	accessKey string
//	secretKey string
//}
//
//func NewAccountInfoDynamoDB(region, db_name, accessKey, secretKey string) AccountInfoDynamoDB {
//	db := &DynamoDB{}
//
//	return AccountInfoDynamoDB{
//		db:        db,
//		region:    region,
//		db_name:   db_name,
//		accessKey: accessKey,
//		secretKey: secretKey,
//	}
//}
//
//func (s *AccountInfoDynamoDB) Open() error {
//	db, err := DynamoConnectInitFromPool(s.db,
//		s.region,
//		[]string{s.db_name},
//		s.accessKey,
//		s.secretKey,
//		"")
//	if err != nil {
//		return err
//	}
//	s.db = db
//	return nil
//}
//
//func (s *AccountInfoDynamoDB) SetAccountInfoData(account_id, platform_id, device_token, device_info string) {
//	value := make(map[string]Any, 2)
//	value[key_platform_id] = platform_id
//	value[key_device_token] = device_token
//	value[Key_device_info] = device_info
//	if err := s.db.SetByHashM(s.db_name, account_id, value); err != nil {
//		logs.Error("AccountInfoDynamoDB SetByHashM db err %s", err.Error())
//	}
//}
//
//func (s *AccountInfoDynamoDB) GetAccountDeviceInfo(acid string) *AccountDeviceInfo {
//	kv, err := s.db.GetByHashM(s.db_name, acid)
//	if err != nil {
//		logs.Error("GetAccountDeviceInfo GetByHashM Err by %s", err.Error())
//		return nil
//	}
//	if kv == nil || len(kv) <= 0 {
//		return nil
//	}
//	var res *AccountDeviceInfo
//	platform, ok := GetFromAnys2String("GetAccountDeviceInfo", key_platform_id, kv)
//	if !ok {
//		logs.Error("GetAccountDeviceInfo GetFromAnys2String by %s %s not found",
//			acid, key_platform_id)
//		return nil
//	}
//	device, ok := GetFromAnys2String("GetAccountDeviceInfo", key_device_token, kv)
//	if !ok {
//		logs.Error("GetAccountDeviceInfo GetFromAnys2String by %s %s not found",
//			acid, key_device_token)
//		return nil
//	}
//
//	res = &AccountDeviceInfo{
//		PlatformType: platform,
//		DeviceInfo:   device,
//	}
//	return res
//}
//
//func (s *AccountInfoDynamoDB) GetAccountDeviceInfos(account_ids []string) (map[string]AccountDeviceInfo, error) {
//	norepeat := make(map[string]string, len(account_ids))
//	for _, id := range account_ids {
//		norepeat[id] = ""
//	}
//	account_ids_norepeat := make([]string, 0, len(norepeat))
//	for k, _ := range norepeat {
//		account_ids_norepeat = append(account_ids_norepeat, k)
//	}
//	return s.db.BatchGetAccountDeviceInfo(s.db_name, key_account_id, account_ids_norepeat)
//}
//
//// 临时返现接口
//func (s *AccountInfoDynamoDB) QueryPayFeedBack(db_name, k string) (firstPay, secondPay float64, err error) {
//	m, err := s.db.GetByHashM(db_name, k)
//	if err != nil {
//		return 0, 0, err
//	}
//	if m == nil || len(m) <= 0 {
//		return 0, 0, nil
//	}
//	_got, ok := m["got"]
//	if ok {
//		str, ok := _got.(string)
//		if ok && str == "ok" {
//			return 0, 0, nil
//		}
//	}
//	_f, ok := m["first_pay"]
//	if ok {
//		f, ok := _f.(string)
//		if ok {
//			firstPay, err = strconv.ParseFloat(f, 64)
//			if err != nil {
//				return 0, 0, err
//			}
//		}
//	}
//	_f, ok = m["second_pay"]
//	if ok {
//		f, ok := _f.(string)
//		if ok {
//			secondPay, err = strconv.ParseFloat(f, 64)
//			if err != nil {
//				return 0, 0, err
//			}
//		}
//	}
//	return firstPay, secondPay, nil
//}
//
//func (s *AccountInfoDynamoDB) GotPayFeedBack(db_name, k string) error {
//	m, _ := s.db.GetByHashM(db_name, k)
//	m["got"] = "ok"
//	return s.db.SetByHashM(db_name, k, m)
//}
//
//type AccountDeviceInfo struct {
//	PlatformType string
//	DeviceInfo   string
//}
//
//func (d *DynamoDB) BatchGetAccountDeviceInfo(table_name string,
//	account_item string, account_values []string) (map[string]AccountDeviceInfo, error) {
//	query_item_names := []string{key_account_id, key_platform_id, key_device_token}
//	queryItemParam := strings.Join(query_item_names, ",")
//	batchInput := &DDB.BatchGetItemInput{
//		RequestItems:           make(map[string]*DDB.KeysAndAttributes, 1),
//		ReturnConsumedCapacity: aws.String("TOTAL"),
//	}
//	keys := &DDB.KeysAndAttributes{
//		Keys:                 make([]map[string]*DDB.AttributeValue, 0, len(account_values)),
//		ProjectionExpression: aws.String(queryItemParam),
//	}
//	for _, v := range account_values {
//		kv := make(map[string]*DDB.AttributeValue, 1)
//		dbv := &DDB.AttributeValue{
//			S: aws.String(v),
//		}
//		kv[account_item] = dbv
//		keys.Keys = append(keys.Keys, kv)
//	}
//	batchInput.RequestItems[table_name] = keys
//
//	//	b := backoff.NewExponentialBackOff()
//	//	b.InitialInterval = time.Duration(game.Cfg.AWS_InitialInterval) * time.Millisecond
//	//	b.MaxElapsedTime = time.Duration(game.Cfg.AWS_InitialInterval) * time.Second
//	//	b.Multiplier = game.Cfg.AWS_Multiplier
//
//	resp, err := d.client.BatchGetItem(batchInput)
//	//	if err != nil {
//	//		ticker := backoff.NewTicker(b)
//	//		resp, err = func(ticker *backoff.Ticker, p *DDB.BatchGetItemInput) (*DDB.BatchGetItemOutput, error) {
//	//			for _ = range ticker.C {
//	//				output, err := d.client.BatchGetItem(p)
//	//				if err == nil {
//	//					ticker.Stop()
//	//					return output, err
//	//				} else {
//	//					logs.Error("BatchGet again Err by %s %v", err.Error(), account_values)
//	//				}
//	//			}
//	//			ticker.Stop()
//	//			return nil, errors.New("TimeOut")
//	//		}(ticker, batchInput)
//	//	}
//	if err != nil {
//		logs.Error("BatchGet Err by %s", err.Error())
//		return nil, err
//	}
//
//	result := make(map[string]AccountDeviceInfo, len(account_values))
//	for tableName, res := range resp.Responses {
//		if tableName != table_name {
//			logs.Error("BatchGet response tablename err %s != %s", tableName, table_name)
//			continue
//		}
//		for _, kv := range res {
//			account, ok := kv[key_account_id]
//			if !ok {
//				break
//			}
//			platform, ok := kv[key_platform_id]
//			if !ok {
//				break
//			}
//			device, ok := kv[key_device_token]
//			if !ok {
//				break
//			}
//			result[GetItemValue(account).(string)] = AccountDeviceInfo{
//				PlatformType: GetItemValue(platform).(string),
//				DeviceInfo:   GetItemValue(device).(string),
//			}
//		}
//	}
//	return result, nil
//}
