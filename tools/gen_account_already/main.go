package main

import (
	"log"

	"io"
	"os"

	"encoding/csv"

	"io/ioutil"

	"fmt"

	"github.com/BurntSushi/toml"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/x/tool_json2account/json2account"
	"vcs.taiyouxi.net/tools/gen_account_already/imp"
)

var (
	account_jsons [][]byte
	redisPL       redispool.IPool
)

func main() {
	if _, err := toml.DecodeFile("conf/config.toml", &imp.Cfg); err != nil {
		log.Fatalf("toml err %v", err)
		return
	}
	log.Println("conf ", imp.Cfg)

	n := 0
	for _, i := range imp.Cfg.AccountSplitNum {
		n += i
	}
	if imp.Cfg.AccountNum != n {
		log.Fatalf("AccountNum not equal %d %d", imp.Cfg.AccountNum, n)
		return
	}

	err := read_json()
	if err != nil {
		log.Fatalf("read_json err %s", err.Error())
		return
	}

	redisPL = redispool.NewSimpleRedisPool("gen_account",
		imp.Cfg.Redis, imp.Cfg.RedisDB, imp.Cfg.RedisAuth, false, 10, true)

	f, err := os.OpenFile(imp.Cfg.AccountUidCsv, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open AccountUidCsv err %s", err.Error())
		return
	}
	defer f.Close()

	uids := make([]string, 0, n)
	reader := csv.NewReader(f)
	reader.Read() // skip header
	for {
		ss, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("read csv err %v", err)
			return
		}
		uids = append(uids, ss[len(ss)-1])
		//log.Println("uid ", ss[len(ss)-1])
	}

	if len(uids) != n {
		log.Println("uid num not equal n %d %d", len(uids), n)
		return
	}

	iuid := 0
	for ijson, n := range imp.Cfg.AccountSplitNum {
		for i := 0; i < n; i++ {
			acid := fmt.Sprintf("%s:%s", imp.Cfg.GidSid, uids[iuid])
			log.Println(ijson, i, acid)
			js := account_jsons[ijson]
			if !write_account(acid, js) {
				return
			}
			iuid += 1
		}
	}
}

func read_json() error {
	account_jsons = make([][]byte, 0, len(imp.Cfg.AccountJsons))
	for _, j := range imp.Cfg.AccountJsons {
		f, err := os.OpenFile(j, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		_account_json, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		account_jsons = append(account_jsons, _account_json)
	}
	return nil
}

func write_account(acid string, account_json []byte) bool {
	conn := redisPL.GetDBConn()
	res := json2account.Imp(conn, acid, account_json)
	if res != "" {
		log.Fatalf("write_account err: %s", res)
		return false
	}
	return true
}
