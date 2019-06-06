package dynamodb

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	. "vcs.taiyouxi.net/platform/planx/util/timail"
)

// MAX_Mail_ONE_GET * 10 不限制数量
func (s *MailDynamoDB) LoadAllMailByGM(user_id string) ([]MailReward, error) {
	res, err := s.db.QueryAllGmMail(s.db_name, user_id, 0, MkMailId(0, 0), MAX_Mail_ONE_GET*10)
	re := make([]MailReward, 0, len(res)+1)
	if err != nil {
		return re[:], err
	}
	for _, mail_from_db := range res {
		new_mail := MailReward{}
		new_mail.FormDB(&mail_from_db)
		re = append(re, new_mail)
	}

	return re, nil
}

func (d *DynamoDB) QueryAllGmMail(table_name,
	user_id string, time_min, time_max, limit int64) ([]MailRes, error) {
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
		Limit:                  aws.Int64(limit),
		ProjectionExpression:   aws.String("t, mail, begint, endt, isget, isread"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ScanIndexForward:       aws.Bool(true),
	}
	count := 0
	res := make([]MailRes, 0)
	// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/QueryAndScan.html#Pagination
	for {
		resp, err := d.client.Query(params)
		if err != nil {
			logs.Error("QueryMail Err by %s", err.Error())
			return []MailRes{}, err
		}
		res = d.parseMail(user_id, resp, res)
		if resp.LastEvaluatedKey == nil {
			break
		}
		count++
		params.ExclusiveStartKey = resp.LastEvaluatedKey
		logs.Info("Paginating the Results, %s, %d", user_id, count)
	}
	return res, nil
}

func (d *DynamoDB) parseMail(user_id string, resp *DDB.QueryOutput, res []MailRes) []MailRes {
	logs.Debug("dynamodb mail, size = %d", *resp.Count)
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
		var dat map[string]interface{}
		var dat1 map[string]interface{}
		var ch string
		json.Unmarshal([]byte(mail_json_s), &dat)
		if v, ok := dat["tag"]; ok {
			json.Unmarshal([]byte(v.(string)), &dat1)
			if d, ok := dat1["ch"]; ok {
				ch = d.(string)
			}
		}

		var t_begin int64
		var t_end int64
		if tbok {
			logs.Debug("dynamodb mail, %d", ti)
			if tType := GetMailSendTyp(ti); tType == Mail_Send_By_Sys || ch == "gm" {
				logs.Debug("dynamodb mail parse filter, not a gm mail, %d", ti)

				tb, ok := GetItemValue(time_begin).(int64)
				if ok {
					t_begin = tb
				}
			}
		}

		if teok {
			te, ok := GetItemValue(time_end).(int64)
			if ok {
				t_end = te
			}
		}

		mail_from_db := MailRes{}
		mail_from_db.SetBeginEnd(t_begin, t_end)

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

		if !tok || !mailok {
			logs.Error("mail info typ by %s for %v %v!", user_id, t, mail_json)
			continue
		}
		mail_from_db.SetID(ti)
		mail_from_db.SetMail(mail_json_s)

		res = append(res, mail_from_db)
	}
	return res
}
