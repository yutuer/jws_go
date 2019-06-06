package imp

import (
	"fmt"
	"strconv"
	"time"

	"bufio"
	"os"

	"path"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

func Run() error {
	if Cfg.TimeBegin == "" {
		return fmt.Errorf("TimeBegin is empty")
	}
	if len(Cfg.GidShard) <= 0 {
		return fmt.Errorf("GidShard is empty")
	}
	l, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return err
	}

	scandb := storehelper.NewStoreDynamoDB(Cfg.AWS_Region, Cfg.PayTable,
		Cfg.AWS_AccessKey, Cfg.AWS_SecretKey, "", "")
	if err := scandb.Open(); err != nil {
		return err
	}

	db2, err := dynamodb.DynamoConnectInitFromPool(&dynamodb.DynamoDB{},
		Cfg.AWS_Region,
		[]string{Cfg.PayTable},
		Cfg.AWS_AccessKey,
		Cfg.AWS_SecretKey,
		"")
	if err != nil {
		return err
	}

	var ts_begin, ts_division int64
	_ts_begin, err := time.ParseInLocation("2006-1-2", Cfg.TimeBegin, l)
	if err != nil {
		return err
	}
	ts_begin = _ts_begin.Unix()
	if Cfg.TimeDivision != "" {
		_ts_division, err := time.ParseInLocation("2006-1-2", Cfg.TimeDivision, l)
		if err != nil {
			return err
		}
		ts_division = _ts_division.Unix()
	}

	payFeedBacks := make(map[string]game.PayFeedBackRec, 1024)
	err = scandb.ScanKV(func(idx int, key, data string) error {
		res, err := db2.GetByHashM(Cfg.PayTable, key)
		if err != nil {
			return err
		}
		//logs.Debug("res %v", res)
		// 过滤测试订单
		str_is_test := ""
		is_test, ok := res["is_test"]
		if ok {
			str, ok := is_test.(string)
			if ok {
				str_is_test = str
			}
		}
		str_note := ""
		note, ok := res["note"]
		if ok {
			str, ok := note.(string)
			if ok {
				str_note = str
			}
		}
		if str_is_test == "1" || str_note == "sandbox" {
			return nil
		}
		// 过滤android
		platform, ok := res["platform"]
		if ok {
			str, ok := platform.(string)
			if ok {
				if str != "android" {
					return nil
				}
			}
		}
		// 过滤时间
		var t int64
		pay_time_s, ok := res["pay_time_s"]
		if ok {
			str, ok := pay_time_s.(string)
			if ok {
				t, err = strconv.ParseInt(str, 10, 64)
				if err != nil {
					return err
				}
				if t < ts_begin {
					return nil
				}
			}
		}
		// 过滤服务器
		k := ""
		uid, ok := res["uid"]
		if ok {
			str_uid, ok := uid.(string)
			if ok {
				account, err := db.ParseAccount(str_uid)
				if err != nil {
					return err
				}
				gidsid := account.ServerString()
				for _, s := range Cfg.GidShard {
					if s == gidsid {
						k = account.UserId.String()
						break
					}
				}
			}
		}
		if k == "" {
			return nil
		}
		// 记录
		//var money_amount float64
		//_money_amount, ok := res["money_amount"]
		//if ok {
		//	str, ok := _money_amount.(string)
		//	if ok {
		//		money_amount, err = strconv.ParseFloat(str, 64)
		//		if err != nil {
		//			return err
		//		}
		//	}
		//}
		var hcbuy float64
		_hcbuy, ok := res["hc_buy"]
		if ok {
			ihcbuy, ok := _hcbuy.(int64)
			if ok {
				hcbuy = float64(ihcbuy)
			} else {
				logs.Error("hc_buy not int %v", res["hc_buy"])
			}
		}
		// money
		var money float64
		_money, ok := res["money_amount"]
		if ok {
			str, ok := _money.(string)
			if ok {
				money, err = strconv.ParseFloat(str, 64)
				if err != nil {
					return err
				}
			}
		}
		if money <= 0 {
			return nil
		}
		// account_name
		var account_name string
		_account_name, ok := res["account_name"]
		if ok {
			str, ok := _account_name.(string)
			if ok && str != "" {
				account_name = str
			}
		}
		//if account_name == "" {
		//	return nil
		//}
		//
		rec, ok := payFeedBacks[k]
		if !ok {
			rec = game.PayFeedBackRec{
				Uid:         k,
				AccountName: account_name,
			}
		}
		if ts_division <= 0 || t < ts_division {
			rec.FirstPay += hcbuy
			rec.Money += money
		} else {
			//rec.SecondPay += money_amount
		}
		payFeedBacks[k] = rec
		logs.Debug("get payFeedBacks %v", rec)
		return nil
	}, 100, 1, false)

	if err != nil {
		return err
	}

	resFile := path.Join(Cfg.ResPath, game.PayFeedBackFile)
	file, err := os.OpenFile(resFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	bf := bufio.NewWriterSize(file, 10240)
	logs.Info("%v", payFeedBacks)
	db := driver.GetDBConn()
	defer db.Close()
	for _, v := range payFeedBacks {
		// 临时代码 begin
		if Cfg.CorrectFromRedis {
			key := fmt.Sprintf("profile:206:6001:%s", v.Uid)
			_vip, err := redis.String(db.Do("hget", key, "v"))
			if err != nil {
				logs.Warn("load from redis err %s %s", key, err.Error())
				continue
			}
			ac_vip := &account.VIP{}
			if err := json.Unmarshal([]byte(_vip), ac_vip); err != nil {
				logs.Warn("unmarshal VIP %v %s %s", err, key, _vip)
				continue
			}
			v.FirstPay = float64(ac_vip.RmbPoint)
		}
		// 临时代码 end
		bb, err := json.Marshal(v)
		if err != nil {
			logs.Error("json.Marshal err %v", err)
			continue
		}
		bf.WriteString(fmt.Sprintf("%s\n", string(bb)))
	}
	bf.Flush()
	file.Close()
	return nil
}

func Run2() error {
	resFile := path.Join(Cfg.ResPath, game.PayFeedBackFile)
	file, err := os.OpenFile(resFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	bf := bufio.NewWriterSize(file, 10240)
	a := game.PayFeedBackRec{
		Uid:         "u1",
		AccountName: "an1",
		FirstPay:    1.1,
		SecondPay:   2.1,
	}
	ba, _ := json.Marshal(a)
	bf.WriteString(fmt.Sprintf("%s\n", string(ba)))
	b := game.PayFeedBackRec{
		Uid:         "u2",
		AccountName: "an2",
		FirstPay:    2.1,
		SecondPay:   3.1,
	}
	bb, _ := json.Marshal(b)
	bf.WriteString(fmt.Sprintf("%s\n", string(bb)))
	bf.Flush()
	file.Close()
	return nil
}
