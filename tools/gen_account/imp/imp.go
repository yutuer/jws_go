package imp

import (
	"fmt"

	"io/ioutil"
	"log"
	"os"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/x/tool_json2account/json2account"
)

var (
	dydb         *dynamodb.DynamoDB
	redisPL      redispool.IPool
	account_json []byte
)

func Init() error {
	db, err := dynamodb.DynamoConnectInitFromPool(&dynamodb.DynamoDB{},
		Cfg.DynamoRegion,
		[]string{Cfg.DynamoDBName},
		Cfg.DynamoAccessKeyID,
		Cfg.DynamoSecretAccessKey,
		"")
	if err != nil {
		return err
	}
	dydb = db

	if Cfg.WriteAccount {
		redisPL = redispool.NewSimpleRedisPool("gen_account",
			Cfg.Redis, Cfg.RedisDB, Cfg.RedisAuth, false, 10, true)

		f, err := os.OpenFile(Cfg.AccountJson, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		_account_json, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		account_json = _account_json
	}
	return nil
}

func GetUid(name string) (string, error) {
	_s, err := dydb.GetByHash(Cfg.DynamoDBName, fmt.Sprintf("un:%s", name))
	if err != nil {
		return "", err
	}

	return _s.(string), nil
}

func WriteAccount(acid string) bool {
	conn := redisPL.GetDBConn()
	res := json2account.Imp(conn, acid, account_json)
	if res != "" {
		log.Fatalf("write_account err: %s", res)
		return false
	}
	return true
}
