package pay

import (
	"time"

	"fmt"

	"errors"

	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/tipay"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

const (
	key_account_id       = "uid"
	key_account_name     = "account_name"
	key_role_name        = "role_name"
	key_order_no         = "order_no"
	key_good_idx         = "good_idx"
	key_good_name        = "good_name"
	key_platform         = "platform"
	key_money_amount     = "money_amount"
	key_hc_buy           = "hc_buy"
	key_hc_give          = "hc_give"
	key_sn               = "sn"
	key_pay_time_s       = "pay_time_s"
	key_tistatus         = "tistatus"
	key_delivered        = "delivered"
	key_receiveTimestamp = "receiveTimestamp"
)

var DBPayMgrAndroid pay.PayDB
var DBPayMgrIOS pay.PayDB

//XXX: YZH虽然支付Android(QuickSDK)和Apple原生支付SDK都同时配置。
//但是实际上不会同时在同一个生产大区服务器上出现。两种混合同时出现的情况只会是在开发和QA环境
type PayDBConfig struct {
	PayDBDriver string

	PayMongoUrl string

	PayAndroidDBName string
	PayIOSDBName     string

	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
}

func InitPayDB(pdbc PayDBConfig) error {
	var err error
	if pdbc.PayAndroidDBName != "" {
		DBPayMgrAndroid, err = tipay.NewPayDriver(pay.PayDBConfig{
			AWSRegion:    pdbc.AWSRegion,
			DBName:       pdbc.PayAndroidDBName,
			AWSAccessKey: pdbc.AWSAccessKey,
			AWSSecretKey: pdbc.AWSSecretKey,
			MongoDBUrl:   pdbc.PayMongoUrl,
			DBDriver:     pdbc.PayDBDriver,
		})
		if err != nil {
			logs.Error("initDB NewPayDB Android Err by %s", err.Error())
			return err
		}
	}

	if pdbc.PayIOSDBName != "" {
		DBPayMgrIOS, err = tipay.NewPayDriver(pay.PayDBConfig{
			AWSRegion:    pdbc.AWSRegion,
			DBName:       pdbc.PayIOSDBName,
			AWSAccessKey: pdbc.AWSAccessKey,
			AWSSecretKey: pdbc.AWSSecretKey,
			MongoDBUrl:   pdbc.PayMongoUrl,
			DBDriver:     pdbc.PayDBDriver,
		})
		if err != nil {
			logs.Error("initDB NewPayDB IOS Err by %s", err.Error())
			return err
		}
	}
	return nil
}

func LogIAP2DB(accountId, accountName, name string, avatar int,
	goodIdx uint32, goodName, order string, money uint32,
	platform, channelId, payTime string, hcBuy, hcGive uint32) error {
	if accountName == "" {
		accountName = fmt.Sprintf("nil-accountname-channel-%s-%s", channelId, accountId)
	}
	if name == "" {
		name = fmt.Sprintf("nil-name-channel-%s-%s", channelId, accountId)
	}
	t := time.Now()
	sn := t.UnixNano()
	switch platform {
	case uutil.Android_Platform:
		if DBPayMgrAndroid != nil {
			values := make(map[string]interface{}, 10)
			values[key_account_name] = accountName
			values[key_role_name] = name
			values[key_sn] = sn
			if goodName != "" {
				values[key_good_name] = goodName
			}
			values[key_platform] = platform
			values[key_hc_buy] = hcBuy
			values[key_hc_give] = hcGive
			values[key_pay_time_s] = payTime
			values[key_tistatus] = key_delivered
			values[key_receiveTimestamp] = t.Unix()
			if err := DBPayMgrAndroid.UpdateByHash(order, values); err != nil {
				logs.Error("[LogIAP2DB] %s %s %s %s %s %d err:%s", accountId, accountName, name, order, platform, goodIdx, err.Error())
				return err
			}
		} else {
			return errors.New("Android Pay DB is not configured.")
		}
	case uutil.IOS_Platform:
		if DBPayMgrIOS != nil {
			values := make(map[string]interface{}, 12)
			values[key_account_id] = accountId
			values[key_account_name] = accountName
			values[key_role_name] = name
			values[key_order_no] = order
			values[key_sn] = sn
			values[key_good_idx] = fmt.Sprintf("%d", goodIdx)
			values[key_good_name] = goodName
			values[key_platform] = platform
			values[key_money_amount] = fmt.Sprintf("%d", money)
			values[key_hc_buy] = hcBuy
			values[key_hc_give] = hcGive
			values[key_pay_time_s] = payTime
			values[key_tistatus] = key_delivered
			values[key_receiveTimestamp] = t.Unix()
			if err := DBPayMgrIOS.SetByHashM(order, values); err != nil {
				logs.Error("[LogIAP2DB] %s %s %s %s %s %d err:%s", accountId, accountName, name, order, platform, goodIdx, err.Error())
				return err
			}
		} else {
			return errors.New("IOS Pay DB is not configured.")
		}
	}

	return nil
}

func IsIOSOrderRepeat(order string) (bool, error) {
	if DBPayMgrIOS == nil {
		return false, errors.New("IOS Pay DB is not configured.")
	}
	return DBPayMgrIOS.IsExist(order)
	//res, err := DBPayMgrIOS.GetByHash(order)
	//if err != nil {
	//	return false, err
	//}
	//return res != nil, nil
}
