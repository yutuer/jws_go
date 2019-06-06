package imp

import (
	"bufio"
	"io"
	"os"

	"log"

	"github.com/bitly/go-simplejson"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

var (
	A2id        map[string]info
	RankAcid    []string
	Device2Acid map[string]string
	RankPool    redispool.IPool
	gmdb        *dynamodb.DynamoDB
)

type info struct {
	uid    string
	device string
}

func Init() error {
	A2id = make(map[string]info, 16)
	Device2Acid = make(map[string]string, 16)
	RankPool = redispool.NewSimpleRedisPool("mirror",
		Cfg.RankRedis, Cfg.RankRedisDB, Cfg.RankRedisAuth, false, 10, true)
	db, err := dynamodb.DynamoConnectInitFromPool(&dynamodb.DynamoDB{},
		Cfg.DynamoRegion,
		[]string{Cfg.DynamoGMInfo},
		"",
		"",
		"")
	if err != nil {
		return err
	}
	gmdb = db
	return nil
}

func GetRank() error {
	db := RankPool.GetDBConn()
	res, err := redis.Strings(db.Do("zrevrange", Cfg.RankTable, 0, 100))
	if err != nil {
		return err
	}
	RankAcid = res
	return nil
}

func FindAccount() error {
	f, err := os.OpenFile(Cfg.GamexLogicLog, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		j, err := simplejson.NewJson([]byte(line))
		if err != nil {
			return err
		}
		typ, err := j.Get("type_name").String()
		if err != nil {
			return err
		}
		if typ == "Login" {
			an, err := j.Get("info").Get("AccountName").String()
			if err != nil {
				return err
			}
			if Cfg.FindAccountName(an) {
				uid, err := j.Get("userid").String()
				if err != nil {
					return err
				}
				A2id[an] = info{
					uid: uid,
				}
			}
		}
	}
	return nil
}

func FindDevice() error {
	f, err := os.OpenFile(Cfg.AuthLogicLog, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		j, err := simplejson.NewJson([]byte(line))
		if err != nil {
			return err
		}
		typ, err := j.Get("type_name").String()
		if err != nil {
			return err
		}
		if typ == "Login" {
			uid, err := j.Get("accountid").String()
			if err != nil {
				return err
			}
			for k, v := range A2id {
				if v.uid == uid {
					dev, err := j.Get("info").Get("DeviceID").String()
					if err != nil {
						return err
					}
					v.device = dev
					A2id[k] = v
					break
				}
			}
		}
	}
	return nil
}

func ShowRes() {
	i := 0
	for _, info := range A2id {
		ac, _ := db.ParseAccount(RankAcid[i])
		Device2Acid[info.device] = ac.UserId.String()
		log.Println(info.device, ac.UserId.String())
		i++
	}
}

func WriteDynamo() error {
	for k, v := range Device2Acid {
		err := gmdb.SetByHash(Cfg.DynamoGMInfo, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
