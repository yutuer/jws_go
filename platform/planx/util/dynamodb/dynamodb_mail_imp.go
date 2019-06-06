package dynamodb

import (
	"errors"
	"time"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	. "vcs.taiyouxi.net/platform/planx/util/timail"
)

func (d *DynamoDB) QueryMail(
	table_name, user_id string, time_min, time_max, limit int64) ([]MailRes, error) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("QueryMail useRes Panic, Err %v", err)
		}
	}()

	params := &DDB.QueryInput{
		TableName:      aws.String(table_name),
		ConsistentRead: aws.Bool(false),
		ExpressionAttributeValues: map[string]*DDB.AttributeValue{
			":u":    CreateAttributeValue(user_id),
			":tmin": CreateAttributeValue(time_min),
			":tmax": CreateAttributeValue(time_max),
		},
		KeyConditionExpression: aws.String("uid = :u AND t BETWEEN :tmin AND :tmax"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int64(limit),
		ProjectionExpression:   aws.String("t,mail,begint,endt,isget,isread,acc2send,ctb2send,cte2send"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
	}

	resp, err := d.client.Query(params)

	if err != nil {
		logs.Error("QueryMail Err by %s", err.Error())
		return []MailRes{}, err
	}

	time_now := time.Now().Unix()

	res := make([]MailRes, 0, len(resp.Items))

	for _, i := range resp.Items {
		t, tok := i["t"]
		mail_json, mailok := i["mail"]
		time_begin, tbok := i["begint"]
		time_end, teok := i["endt"]
		is_get, is_getok := i["isget"]
		is_read, is_readok := i["isread"]

		//
		//acc2send string // 发送的账号,只有玩家id在此列表中才可以领取
		//ctb2send int64  // 在此时间开始之后注册的账号可以领取 为0表示不限开始时间
		//cte2send int64  // 在此时间之前注册的账号可以领取 为0表示不限结束时间
		//

		acc2send, acc2sendOk := i["acc2send"]
		ctb2send, ctb2sendOk := i["ctb2send"]
		cte2send, cte2sendOk := i["cte2send"]

		if !tok || !mailok {
			logs.Error("mail info error by %s!", user_id)
			continue
		}

		ti, tok := GetItemValue(t).(int64)
		mail_json_s, mailok := GetItemValue(mail_json).(string)

		var t_begin int64
		var t_end int64

		if tbok {
			tb, ok := GetItemValue(time_begin).(int64)
			if ok {
				t_begin = tb
			}
			if ok && tb >= time_now {
				continue
			}
		}

		if teok {
			te, ok := GetItemValue(time_end).(int64)
			if ok {
				t_end = te
			}
			if ok && te <= time_now {
				continue
			}
		}

		mail_from_db := MailRes{}
		mail_from_db.SetBeginEnd(t_begin, t_end)

		if acc2sendOk {
			s, _ := GetItemValue(acc2send).(string)
			mail_from_db.SetAcc2send(s)
		}
		if ctb2sendOk {
			s, _ := GetItemValue(ctb2send).(int64)
			mail_from_db.SetCtb2send(s)
		}
		if cte2sendOk {
			s, _ := GetItemValue(cte2send).(int64)
			mail_from_db.SetCte2send(s)
		}

		//logs.Trace("i %v", i)
		if is_getok {
			//logs.Trace("is_getok %v", is_getok)
			is_get_s, is_get_s_ok := GetItemValue(is_get).(string)
			//logs.Trace("is_get_s %s %v", is_get_s, is_get_s_ok)
			if is_get_s_ok && is_get_s == GETTED_STR {
				mail_from_db.SetIsGet(true)
			}
		}
		if is_readok {
			is_read_s, is_read_s_ok := GetItemValue(is_read).(string)
			if is_read_s_ok && is_read_s == READ_STR {
				mail_from_db.SetIsRead(true)
			}
		}

		if !tok || !mailok {
			logs.Error("mail info typ by %s for %v %v!", user_id, t, mail_json)
			continue
		}
		mail_from_db.SetID(ti)
		mail_from_db.SetMail(mail_json_s)

		res = append(res, mail_from_db)
	}
	if resp.ConsumedCapacity.CapacityUnits != nil {
		logs.Trace("Mail Query CapacityUnits %v", *resp.ConsumedCapacity.CapacityUnits)
	}

	//logs.Trace("Mail Res %v  %v", res, resp.Items)

	return res, nil

}

func genMailItem(user_id string, mail MailRes) map[string]*DDB.AttributeValue {
	item := map[string]*DDB.AttributeValue{ // Required
		"uid":      CreateAttributeValue(user_id),
		"t":        CreateAttributeValue(mail.Id),
		"ctb2send": CreateAttributeValue(mail.Ctb2send),
		"cte2send": CreateAttributeValue(mail.Cte2send),
		"mail":     CreateAttributeValue(mail.Mail),
		"begint":   CreateAttributeValue(mail.Begin),
		"endt":     CreateAttributeValue(mail.End),
	}

	if mail.Acc2send != "" {
		item["acc2send"] = CreateAttributeValue(mail.Acc2send)
	}
	if mail.IsGet {
		item["isget"] = CreateAttributeValue(GETTED_STR)
	}
	if mail.IsRead {
		item["isread"] = CreateAttributeValue(READ_STR)
	}
	return item
}

func (d *DynamoDB) SendMail(table_name, user_id string, mail MailRes) error {
	item := genMailItem(user_id, mail)

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item:      item,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ReturnValues:           aws.String("ALL_OLD"),
	}
	putOutItem, err := d.client.PutItem(put_input)
	if err == nil {
		if putOutItem.ConsumedCapacity.CapacityUnits != nil {
			logs.Trace("Mail Query CapacityUnits %v", *putOutItem.ConsumedCapacity.CapacityUnits)
		}
		if putOutItem.Attributes != nil && len(putOutItem.Attributes) > 0 {
			logs.Trace("Mail Overide writed, %v", putOutItem.Attributes)
		}
	}
	return err
}

func (d *DynamoDB) DeleteMail(table_name, user_id string, mail_id int64) error {
	delInput := &DDB.DeleteItemInput{
		TableName: aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{ // Required
			"uid": CreateAttributeValue(user_id),
			"t":   CreateAttributeValue(mail_id),
		},
	}
	_, err := d.client.DeleteItem(delInput)
	return err
}

func (d *DynamoDB) SyncMail(
	table_name, user_id string,
	mail_add []MailRes,
	mails_del []int64) (errfinal error) {

	err := d._syncMail(table_name, user_id, mail_add, mails_del)
	if err != nil {
		b := NewExponentialBackOffSleepFirst()
		err = backoff.RetryNotify(func() error {
			nerr := d._syncMail(table_name, user_id, mail_add, mails_del)
			if nerr != nil {
				logs.Error("DynamoDB SyncMail Err by %s info %v %v %v %v", err.Error(),
					table_name, user_id, mail_add, mails_del)
				return nerr
			}
			return nil
		}, b, func(e error, d time.Duration) {
			logs.Warn("[%s]SyncMail Err by %s will retry in %d", user_id, e.Error(), d)
		})
	}
	return nil
}

func (d *DynamoDB) _syncMail(
	table_name, user_id string,
	mail_add []MailRes,
	mails_del []int64) (errfinal error) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("QueryMail useRes Panic, Err %v", r)
			// find out exactly what the error was and set err
			switch x := r.(type) {
			case string:
				errfinal = errors.New(x)
			case error:
				errfinal = x
			default:
				errfinal = errors.New("Unknown panic")
			}
		}
	}()

	requestNum := len(mail_add) + len(mails_del)

	switch requestNum {
	case 1:
		if len(mail_add) == 1 {
			errfinal = d.SendMail(table_name, user_id, mail_add[0])
			mail_add = mail_add[0:0]
		}

		if len(mails_del) == 1 {
			errfinal = d.DeleteMail(table_name, user_id, mails_del[0])
			mails_del = mails_del[0:0]
		}
		return
	default:
	}

	requestNum = len(mail_add) + len(mails_del)
	if requestNum <= 0 {
		errfinal = nil
		return
	}

	PutRequests := make([]*DDB.WriteRequest, 0, requestNum)

	for _, mail := range mail_add {
		item := genMailItem(user_id, mail)
		logs.Trace("mail save %v", item)

		PutRequests = append(PutRequests, &DDB.WriteRequest{
			PutRequest: &DDB.PutRequest{
				Item: item,
			},
		})
	}

	for _, m := range mails_del {
		PutRequests = append(PutRequests, &DDB.WriteRequest{
			DeleteRequest: &DDB.DeleteRequest{
				Key: map[string]*DDB.AttributeValue{ // Required
					"uid": CreateAttributeValue(user_id),
					"t":   CreateAttributeValue(m),
				},
			},
		})
	}

	params := &DDB.BatchWriteItemInput{
		RequestItems: map[string][]*DDB.WriteRequest{ // Required
			table_name: PutRequests,
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
	}

	/*
		AWS_InitialInterval=500 # 以time.Millisecond为单位
		AWS_Multiplier=1.5
		AWS_MaxElapsedTime= 300 # 以time.Second为单位
	*/
	//logs.Trace("BatchWriteItem start")
	resp, err := d.client.BatchWriteItem(params)

	printCapacity := func(bwio *DDB.BatchWriteItemOutput) {
		printDynamoDBBatchCapacity("SyncMail", bwio)
	}
	if err != nil {
		logs.Error("[%s]_syncMail Err by %s info %v %v", user_id, err.Error(), mail_add, mails_del)
		return err
	}

	printCapacity(resp)

	if len(resp.UnprocessedItems) > 0 {
		logs.Warn("[%s]_syncMail No Error But UnprocessedItems found %s will retry", user_id)
		b := NewExponentialBackOffSleepFirst()
		err = backoff.RetryNotify(func() error {
			newPutRequests := PutRequests[0:0]
			for _, v := range resp.UnprocessedItems {
				for _, failed := range v {
					if failed.PutRequest != nil || failed.DeleteRequest != nil {
						newPutRequests = append(newPutRequests, failed)
					}
				}
			}
			nparams := &DDB.BatchWriteItemInput{
				RequestItems: map[string][]*DDB.WriteRequest{ // Required
					table_name: newPutRequests,
				},
				ReturnConsumedCapacity: aws.String("TOTAL"),
			}
			nresp, nerr := d.client.BatchWriteItem(nparams)
			resp = nresp
			if nerr != nil {
				logs.Error("[%s]_syncMail Err by %s info %v %v", user_id, err.Error(), mail_add, mails_del)
				return nerr
			}

			printCapacity(resp)

			if len(resp.UnprocessedItems) > 0 {
				return fmt.Errorf("UnprocessedItems")
			}
			return nil
		}, b, func(e error, d time.Duration) {
			logs.Warn("[%s]_syncMail Err by %s will retry in %d", user_id, e.Error(), d)
		})

	}
	return err
}

func (d *DynamoDB) BatchSendMail(table_name string, mail_add []MailRes) (error, []MailKey) {
	// 去重
	mmail := HelperBatchDupDrop(mail_add)

	logs.Debug("BatchSendMail begin mails len %d", len(mmail))
	PutRequests := make([]*DDB.WriteRequest, 0, len(mmail))
	for _, mail := range mmail {
		item := genMailItem(mail.Userid, mail)

		PutRequests = append(PutRequests, &DDB.WriteRequest{
			PutRequest: &DDB.PutRequest{
				Item: item,
			},
		})
	}

	params := &DDB.BatchWriteItemInput{
		RequestItems: map[string][]*DDB.WriteRequest{ // Required
			table_name: PutRequests,
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
	}

	printCapacity := func(bwio *DDB.BatchWriteItemOutput) {
		printDynamoDBBatchCapacity("BatchSendMail", bwio)
	}

	resp, err := d.client.BatchWriteItem(params)

	if err != nil {
		logs.Error("BatchSendMail BatchWriteItem Err by %s", err.Error())
		return err, nil
	}
	printCapacity(resp)
	if len(resp.UnprocessedItems) > 0 {
		b := NewExponentialBackOffSleepFirst()

		err := backoff.RetryNotify(func() error {
			newPutRequests := PutRequests[0:0]
			//PutRequests := make([]*DDB.WriteRequest, 0, len(resp.UnprocessedItems))
			for _, v := range resp.UnprocessedItems {
				for _, failed := range v {
					if failed.PutRequest != nil || failed.DeleteRequest != nil {
						newPutRequests = append(newPutRequests, failed)
					}
				}
			}

			nparams := &DDB.BatchWriteItemInput{
				RequestItems: map[string][]*DDB.WriteRequest{ // Required
					table_name: newPutRequests,
				},
				ReturnConsumedCapacity: aws.String("TOTAL"),
			}

			nresp, nerr := d.client.BatchWriteItem(nparams)
			resp = nresp
			printCapacity(nresp)

			if nerr != nil {
				logs.Error("BatchSendMail Err by %s info %v", nerr.Error(), mmail)
				return nerr
			}

			if len(resp.UnprocessedItems) > 0 {
				return UnprocessedItemsError{}
			}
			return nil
		}, b, func(e error, d time.Duration) {
			logs.Warn("BatchSendMail Err by %s will retry in %d", e.Error(), d)
		})

		if err != nil {
			if _, ok := err.(UnprocessedItemsError); ok {
				faileds := make([]MailKey, 0, len(resp.UnprocessedItems))
				for _, v := range resp.UnprocessedItems {
					for _, failed := range v {
						uid := failed.PutRequest.Item["uid"]
						id := failed.PutRequest.Item["t"]
						k := MailKey{GetItemValue(uid).(string), GetItemValue(id).(int64)}
						faileds = append(faileds, k)
					}
				}
				return nil, faileds
			}
			return err, nil
		}
	}
	return nil, nil
}

type UnprocessedItemsError struct {
}

func (e UnprocessedItemsError) Error() string {
	return "UnprocessedItems"
}
