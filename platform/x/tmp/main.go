package main

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/tipay"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

const (
	key_tistatus       = "tistatus"
	tistatus_paid      = "Paid"
	tistatus_delivered = "Delivered"
	table              = "AndroidPay"
)

func main() {
	hash := "debug_order_no"
	db, err := tipay.NewPayDriver(pay.PayDBConfig{
		AWSRegion:    "cn-north-1",
		DBName:       "",
		AWSAccessKey: "",
		AWSSecretKey: "",

		MongoDBUrl: "",
		DBDriver:   "DynamoDB",
	})

	if err != nil {
		fmt.Printf("init db err %v", err)
		return
	}

	if err := f(db, hash); err != nil {
		fmt.Printf("err %v", err)
		return
	}
}

func f(db *pay.PayDB, hash string) error {
	oldValue, err := db.GetByHashM(hash)
	if err != nil {
		return err
	}
	if oldValue != nil && len(oldValue) > 0 {
		ret := true
		if v, ok := oldValue[key_tistatus]; ok {
			str, ok := v.(string)
			if ok {
				if str != tistatus_delivered {
					ret = false
				}
			}
		}
		if ret {
			fmt.Println("repeat !!")
			return nil
		}
	}
	values := make(map[string]dynamodb.Any, 11)
	values[key_tistatus] = tistatus_paid
	if err := db.SetByHashM(hash, values); err != nil {
		return err
	}

	values = map[string]interface{
		key_tistatus: tistatus_delivered,
	}
	if err := db.GetDB().UpdateByHash(table, hash, values); err != nil {
		return err
	}
	return nil
}
