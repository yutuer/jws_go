package merge

import (
	"strings"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

var (
	delAcids       = make(map[string]struct{}, 1024)
	delNames2Acids = make(map[string]struct{}, 1024)
)

func checkRoleNeedDel(DB redispool.RedisPoolConn, key string) (err error, needDel bool) {
	hasPrefix := false
	var _prefix string
	for _, prefix := range account_prefix {
		if strings.HasPrefix(key, prefix) {
			hasPrefix = true
			_prefix = prefix
			break
		}
	}
	if !hasPrefix {
		return nil, false
	}
	acid := strings.TrimLeft(key, _prefix+":")
	if _, ok := delAcids[acid]; ok {
		return nil, true
	}

	accountID, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("checkRoleNeedDel ParseAccount %s err %s", acid, err.Error())
		return err, false
	}

	// profile
	a := new(account.Account)
	a.AccountID = accountID
	a.Profile = account.NewProfile(accountID)

	p := &a.Profile
	profile_key := p.DBName()

	corp, logoutTime, vip, err := getDelInfo(DB, profile_key)
	if err != nil {
		logs.Warn("get del info err, %v", err)
		delAcids[acid] = struct{}{}
		delNames2Acids[p.Name] = struct{}{}
		return nil, true
	}

	// n天没登录，等级小于m, 没付费过
	if corp.Level < uint32(Cfg.DelAccountCorpLvl) &&
		Cfg.checkDelAccountNotLoginTime(logoutTime+p.DebugAbsoluteTime) && // 使用DebugAbsoluteTime是为了方便qa测试
		vip.RmbPoint <= 0 {
		delAcids[acid] = struct{}{}
		delNames2Acids[p.Name] = struct{}{}
		return nil, true
	}
	return nil, false
}

func getDelInfo(DB redispool.RedisPoolConn, profile_key string) (*account.Corp, int64, *account.VIP, error) {
	scorp, err := redis.String(DB.Do("HGET", profile_key, "corp"))
	if err != nil {
		return nil, 0, nil, err
	}

	corp := &account.Corp{}
	if err = json.Unmarshal([]byte(scorp), corp); err != nil {
		return nil, 0, nil, err
	}

	logoutTime, err := redis.Int64(DB.Do("HGET", profile_key, "logouttime"))
	if err != nil {
		return nil, 0, nil, err
	}

	svip, err := redis.String(DB.Do("HGET", profile_key, "v"))
	if err != nil {
		return nil, 0, nil, err
	}

	vip := &account.VIP{}
	if err = json.Unmarshal([]byte(svip), vip); err != nil {
		return nil, 0, nil, err
	}
	return corp, logoutTime, vip, nil
}

func isNameDel(name string) bool {
	_, ok := delNames2Acids[name]
	return ok
}

func isAcidDel(acid string) bool {
	_, ok := delAcids[acid]
	return ok
}
