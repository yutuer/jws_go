package iap

import (
	"errors"
	"strings"

	"fmt"
	"reflect"

	"encoding/json"

	"time"

	"strconv"

	"gopkg.in/olivere/elastic.v2"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/planx/util/tipay"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
)

const (
	typ           = "fluentd"
	LogicTag_IAP  = "IAP"
	time_format   = "2006-1-2 15:04:05"
	currency_cost = "CostCurrency"
	currency_give = "GiveCurrency"
	currency_trim = "Currency"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("queryIAPInfoMail", queryIAPInfoMail)
	gm_command.AddGmCommandHandle("queryGateWayInfo", queryGateWayInfo) // dynamo
	gm_command.AddGmCommandHandle("queryIAPInfo", queryIAPInfo)         // elastic log
	gm_command.AddGmCommandHandle("queryHCInfo", queryHCInfo)           // elastic log
}

func queryGateWayInfo(c *gm_command.Context, server, accountid string, params []string) error {
	var st, et int64
	if len(params) >= 2 {
		if _st, err := strconv.ParseInt(params[0], 10, 64); err == nil {
			st = _st / 1000
		}
		if _et, err := strconv.ParseInt(params[1], 10, 64); err == nil {
			et = _et / 1000
		}
	}
	logs.Debug("IAP queryGateWayInfo %v s %d e %d", params, st, et)
	var ret res

	//TODO by YZH queryGateWayInfo only Andoird(QuickSDK) only
	db, err := tipay.NewPayDriver(pay.PayDBConfig{
		AWSRegion:    gmConfig.Cfg.AWS_Region,
		DBName:       gmConfig.Cfg.PayAndroidDBName,
		AWSAccessKey: gmConfig.Cfg.AWS_AccessKey,
		AWSSecretKey: gmConfig.Cfg.AWS_SecretKey,
		MongoDBUrl:   gmConfig.Cfg.PayMongoUrl,
		DBDriver:     gmConfig.Cfg.PayDBDriver,
	})

	if err != nil {
		retErr(c, ret, fmt.Sprintf("NewPayDynamoDB err %s", err.Error()))
		return err
	}

	ret.QueriesGateWay = make([]pay.GateWay_Pay_Info, 0, 128)
	pay_android_info, err := db.QueryByUid(accountid)
	if err != nil {
		retErr(c, ret, fmt.Sprintf("DynamoDB Query err %s", err.Error()))
		return err
	}
	ret.QueriesGateWay = _setDynamoData(pay_android_info, ret.QueriesGateWay, st, et)

	pay_ios_info, err := db.QueryByUid(accountid)
	if err != nil {
		retErr(c, ret, fmt.Sprintf("DynamoDB Query err %s", err.Error()))
		return err
	}
	ret.QueriesGateWay = _setDynamoData(pay_ios_info, ret.QueriesGateWay, st, et)

	ret.Ret = "ok"
	b, _ := json.Marshal(&ret)
	c.SetData(string(b))
	return nil
}

func queryIAPInfo(c *gm_command.Context, server, accountid string, params []string) error {
	var ret res
	if len(params) < 2 {
		retErr(c, ret, "params not enough")
		return errors.New("[queryIAPInfo] params not enough")
	}

	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("[queryIAPInfo]  %s %s --> %v", accountid, server, cfg)

	serv := strings.Split(cfg.ServerName, ":")
	indx := fmt.Sprintf(config.Cfg.ElasticIndex, serv[1])

	acids := strings.Split(accountid, ":")
	if len(acids) != 3 {
		retErr(c, ret, "acid err")
		return errors.New("[queryIAPInfo] acid err")
	}

	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(config.Cfg.ElasticUrl))
	if err != nil {
		retErr(c, ret, fmt.Sprintf("elastic.NewClient err: %s", err.Error()))
		return errors.New(fmt.Sprintf("[queryIAPInfo] elastic.NewClient err: %s", err.Error()))
	}
	start_time := _formatDate(params[0], false)
	end_time := _formatDate(params[1], true)
	logs.Info("[queryIAPInfo] %s %s", start_time, end_time)
	typ := elastic.NewMatchQuery("type_name", LogicTag_IAP)
	date := elastic.NewRangeQuery("@timestamp").
		Gte(start_time).
		Lte(end_time).
		TimeZone("+08:00")
	ac := elastic.NewMatchQuery("accountid", accountid)
	bl := elastic.NewBoolQuery().Must(typ, date, ac)

	logs.Info("[queryIAPInfo] query %s %s %s %s", indx, start_time, end_time, accountid)
	searchResult, err := client.Search().
		Index(indx).
		Query(bl).
		Sort("@timestamp", true). // sort by "time" field, ascending
		From(0).Size(10000).
		Pretty(true).
		Do()
	if err != nil {
		retErr(c, ret, fmt.Sprintf("search err: %s", err.Error()))
		return errors.New(fmt.Sprintf("[queryIAPInfo] search err: %s", err.Error()))
	}

	logs.Info("[queryIAPInfo] Query took %d ms, found %d", searchResult.TookInMillis, searchResult.TotalHits())

	ret.QueriesLog = make([]res_a_query, 0, searchResult.TotalHits())
	var ttyp QueryRes
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(QueryRes); ok {
			pay_time, err := strconv.ParseInt(t.Info.PayTime, 10, 64)
			if err != nil {
				logs.Error("[queryIAPInfo] PayTime strconv err: %s", err.Error())
				pay_time = 0
			}
			ret.QueriesLog = append(ret.QueriesLog, res_a_query{
				Time:        t.LogTime / int64(time.Millisecond),
				AccountId:   t.AccountId,
				AccountName: t.Info.AccountName,
				Name:        t.Info.Name,
				GoodIdx:     t.Info.GoodIdx,
				GoodName:    t.Info.GoodName,
				Money:       t.Info.Money,
				Order:       t.Info.Order,
				Platform:    t.Info.Platform,
				ChannelId:   t.Info.ChannelId,
				PayTime:     pay_time,
			})
		}
	}
	ret.Ret = "ok"
	b, _ := json.Marshal(&ret)
	c.SetData(string(b))
	return nil
}

type QueryRes struct {
	Time      string `json:"utc8"`
	AccountId string `json:"accountid"`
	Type      string `json:"type_name"`
	Info      Info   `json:"info"`
	LogTime   int64  `json:"logtime"`
}

type Info struct {
	Platform    string
	AccountName string
	ChannelId   string
	Name        string
	GoodIdx     uint32
	GoodName    string
	Money       uint32
	Order       string
	PayTime     string
}

type res struct {
	Ret            string                 `json:"ret"`
	QueriesLog     []res_a_query          `json:"queries_log"`
	QueriesMail    []res_a_query          `json:"queries_mail"`
	QueriesHC      []res_hc_query         `json:"queries_hc"`
	QueriesGateWay []pay.GateWay_Pay_Info `json:"queries_android_pay"`
}

type res_a_query struct {
	Time        int64
	AccountId   string
	AccountName string
	Name        string
	GoodIdx     uint32
	GoodName    string
	Money       uint32
	Order       string
	Platform    string
	ChannelId   string
	PayTime     int64
	IsGot       bool
}

func queryIAPInfoMail(c *gm_command.Context, server, accountid string, params []string) error {
	var ret res
	acids := strings.Split(accountid, ":")
	if len(acids) != 3 {
		retErr(c, ret, "acid err")
		return errors.New(fmt.Sprintf("[queryIAPInfo] acid err: %s", accountid))
	}

	mail_profile_name := "profile:" + accountid
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)

	if cfg.RedisName == "" {
		retErr(c, ret, fmt.Sprintf("GetAllMail Err No Name %s", server))
		logs.Error("GetAllMail Err No Name %s", server)
		return errors.New("NoServerName")
	}
	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    gmConfig.Cfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: gmConfig.Cfg.AWS_AccessKey,
		AWSSecretKey: gmConfig.Cfg.AWS_SecretKey,
		MongoDBUrl:   cfg.MailMongoUrl,
		DBDriver:     cfg.MailDBDriver,
	})

	if err != nil {
		retErr(c, ret, fmt.Sprintf("NewMailDynamoDB err %s", err.Error()))
		return err
	}

	res, err := m.LoadAllMail(mail_profile_name)
	if err != nil {
		retErr(c, ret, fmt.Sprintf("LoadAllMail err %s", err.Error()))
		return err
	}

	ret.QueriesMail = make([]res_a_query, 0, len(res))
	for _, mail := range res {
		if timail.GetMailSendTyp(mail.Idx) == timail.Mail_Send_By_AndroidIAP ||
			timail.GetMailSendTyp(mail.Idx) == timail.Mail_Send_By_IOSIAP {
			var order timail.IAPOrder
			if err := json.Unmarshal([]byte(mail.Tag), &order); err != nil {
				retErr(c, ret, fmt.Sprintf("AndroidIAPOrder unmarshal err %s", err.Error()))
				return err
			}

			platform := uutil.Android_Platform
			if timail.GetMailSendTyp(mail.Idx) == timail.Mail_Send_By_IOSIAP {
				platform = uutil.IOS_Platform
			}
			money, err := strconv.ParseFloat(order.Amount, 32)
			if err != nil {
				retErr(c, ret, fmt.Sprintf("strconv.Atoi(money) err %s", err.Error()))
				return err
			}
			goodidx, err := strconv.Atoi(order.Game_order)
			if err != nil {
				retErr(c, ret, fmt.Sprintf("strconv.Atoi(goodidx) err %s", err.Error()))
				return err
			}
			pay_time, err := strconv.ParseInt(order.PayTime, 10, 64)
			if err != nil {
				logs.Error("[queryIAPInfoMail] PayTime strconv err: %s", err.Error())
				pay_time = 0
			}
			ret.QueriesMail = append(ret.QueriesMail, res_a_query{
				Time:      mail.TimeBegin,
				AccountId: accountid,
				GoodIdx:   uint32(goodidx),
				Money:     uint32(money),
				Order:     order.Order_no,
				Platform:  platform,
				ChannelId: order.Channel,
				PayTime:   pay_time,
				IsGot:     mail.IsGetted,
			})
		}
	}
	ret.Ret = "ok"
	b, _ := json.Marshal(&ret)
	c.SetData(string(b))
	return nil
}

func queryHCInfo(c *gm_command.Context, server, accountid string, params []string) error {
	var ret res
	if len(params) < 2 {
		retErr(c, ret, "params not enough")
		return errors.New("[queryHCInfo] params not enough")
	}

	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("[queryHCInfo]  %s %s --> %v", accountid, server, cfg)

	serv := strings.Split(cfg.ServerName, ":")
	indx := fmt.Sprintf(config.Cfg.ElasticIndex, serv[1])

	acids := strings.Split(accountid, ":")
	if len(acids) != 3 {
		retErr(c, ret, "acid err")
		return errors.New("[queryHCInfo] acid err")
	}

	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(config.Cfg.ElasticUrl))
	if err != nil {
		retErr(c, ret, fmt.Sprintf("elastic.NewClient err: %s", err.Error()))
		return errors.New(fmt.Sprintf("[queryHCInfo] elastic.NewClient err: %s", err.Error()))
	}
	start_time := _formatDate(params[0], false)
	end_time := _formatDate(params[1], true)

	typ_cond := elastic.NewBoolQuery().Should(elastic.NewMatchQuery("type_name", currency_cost),
		elastic.NewMatchQuery("type_name", currency_give))
	date_cond := elastic.NewRangeQuery("@timestamp").
		Gte(start_time).
		Lte(end_time).
		TimeZone("+08:00")
	ac_cond := elastic.NewMatchQuery("accountid", accountid)
	curr_cond := elastic.NewQueryStringQuery("VI_HC*")
	bl := elastic.NewBoolQuery().Must(typ_cond, date_cond, ac_cond, curr_cond)

	logs.Info("[queryHCInfo] query %s %s %s %s", indx, start_time, end_time, accountid)
	searchResult, err := client.Search().
		Index(indx).
		Query(bl).
		Sort("@timestamp", true). // sort by "time" field, ascending
		From(0).Size(10000).
		Pretty(true).
		Do()
	if err != nil {
		retErr(c, ret, fmt.Sprintf("search err: %s", err.Error()))
		return errors.New(fmt.Sprintf("[queryHCInfo] search err: %s", err.Error()))
	}

	logs.Info("[queryHCInfo] Query took %d ms, found %d", searchResult.TookInMillis, searchResult.TotalHits())

	ret.QueriesHC = make([]res_hc_query, 0, searchResult.TotalHits())
	var ttyp HCQueryRes
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(HCQueryRes); ok {
			ret.QueriesHC = append(ret.QueriesHC, res_hc_query{
				Time:      t.LogTime / int64(time.Second),
				AccountId: t.AccountId,
				TypCG:     strings.TrimSuffix(t.Type, currency_trim),
				CurTyp:    t.Info.Type,
				Chg:       t.Info.Value,
				BefV:      t.Info.BefValue,
				AftV:      t.Info.AftValue,
				Reason:    t.Info.Reason,
			})
		}
	}
	ret.Ret = "ok"
	b, _ := json.Marshal(&ret)
	c.SetData(string(b))
	return nil
}

type HCQueryRes struct {
	LogTime   int64  `json:"logtime"`
	AccountId string `json:"accountid"`
	Type      string `json:"type_name"`
	Info      HCInfo `json:"info"`
}

type HCInfo struct {
	Reason   string
	Type     string
	Value    int64 // change value
	BefValue int64
	AftValue int64
}

type res_hc_query struct {
	Time      int64
	AccountId string
	TypCG     string // cost / give
	CurTyp    string // currency type
	Chg       int64
	BefV      int64
	AftV      int64
	Reason    string
}

func retErr(c *gm_command.Context, ret res, msg string) {
	ret.Ret = msg
	b, _ := json.Marshal(&ret)
	c.SetData(string(b))
}

func _formatDate(strDate string, end bool) string {
	ss := strings.Split(strDate, "-")
	m, _ := strconv.Atoi(ss[1])
	d, _ := strconv.Atoi(ss[2])
	if end {
		return fmt.Sprintf("%s-%02d-%02dT23:59:59.999", ss[0], m, d)
	}
	return fmt.Sprintf("%s-%02d-%02dT00:00:00.000", ss[0], m, d)
}

func _setDynamoData(src []pay.GateWay_Pay_Info,
	dest []pay.GateWay_Pay_Info,
	st, et int64) []pay.GateWay_Pay_Info {
	if st > 0 && et > 0 {
		for _, info := range src {
			if info.ReceiveTimeStamp >= st && info.ReceiveTimeStamp <= et {
				dest = append(dest, info)
			}
		}
	} else {
		dest = append(dest, src...)
	}
	return dest
}
