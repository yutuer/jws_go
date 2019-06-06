package merge

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func updateGank(acid string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	accountID, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("updateGank ParseAccount %s err %s", acid, err.Error())
		return err
	}

	// profile
	a := new(account.Account)
	a.AccountID = accountID
	a.Profile = account.NewProfile(accountID)

	p := &a.Profile
	key := p.DBName()

	sgank, err := redis.String(resDB.Do("HGET", key, "gank"))
	if err != nil {
		return err
	}
	gank := &account.PlayerGank{}
	if err = json.Unmarshal([]byte(sgank), gank); err != nil {
		return err
	}

	gank.GankLastReviewLogTS = 0
	gank.GankNewestLogTS = 0

	rgank, err := json.Marshal(gank)
	if err != nil {
		return err
	}
	if err = cb.Send("HSET", key, "gank", string(rgank)); err != nil {
		return err
	}

	//err = driver.RestoreFromHashDB(resDB.RawConn(), key, p, false, false)
	//if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
	//	logs.Error("updateGank loadDB err %s %s", key, err.Error())
	//	return err
	//}
	//
	//p.Gank.GankLastReviewLogTS = 0
	//p.Gank.GankNewestLogTS = 0
	//
	//err = driver.DumpToHashDBCmcBuffer(cb, key, p)
	//if err != nil {
	//	logs.Error("updateGank saveDB err %s", err.Error())
	//	return err
	//}

	return nil
}
