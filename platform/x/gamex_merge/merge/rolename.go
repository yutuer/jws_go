package merge

import (
	"fmt"
	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

/*
	角色重名的修改玩家存档和names表
*/
const (
	name_split_sym = "."
)

var (
	account_name_old2new = make(map[string]string, 1)
	account_num          int
)

func MergeRoleName() error {
	ADB := GetADBConn()
	if ADB.IsNil() {
		return fmt.Errorf("MergeRoleName GetDBConn nil")
	}
	defer ADB.Close()
	BDB := GetBDBConn()
	if BDB.IsNil() {
		return fmt.Errorf("MergeRoleName GetDBConn nil")
	}
	defer BDB.Close()
	ResDB := GetResDBConn()
	if ResDB.IsNil() {
		return fmt.Errorf("MergeRoleName GetDBConn nil")
	}
	defer ResDB.Close()

	ANames, err := redis.StringMap(ADB.Do("HGETALL", driver.TableChangeName(Cfg.AShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("MergeRoleName A HGETALL err %s", err.Error())
		return err
	}
	BNames, err := redis.StringMap(BDB.Do("HGETALL", driver.TableChangeName(Cfg.BShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("MergeRoleName B HGETALL err %s", err.Error())
		return err
	}
	// 删除了的角色，删除对于的名字
	for n, id := range ANames {
		if isAcidDel(id) {
			logs.Debug("del name %s %s", n, id)
			delete(ANames, n)
		}
	}
	for n, id := range BNames {
		if isAcidDel(id) {
			logs.Debug("del name %s %s", n, id)
			delete(BNames, n)
		}
	}

	BNameDup := make(map[string]string, 64)
	for an, _ := range ANames {
		if bacid, ok := BNames[an]; ok {
			if isNameDel(an) {
				continue
			}
			BNameDup[an] = bacid
		}
	}
	logs.Info("MergeRoleName BNameDup %d", len(BNameDup))

	cb := redis.NewCmdBuffer()
	if len(BNameDup) > 0 {
		account_name_old2new = make(map[string]string, len(BNameDup))

		for oldName, acid := range BNameDup {
			newName := nameMerge(oldName, Cfg.BShardId)
			delete(BNames, oldName)
			BNames[newName] = acid
			account_name_old2new[oldName] = newName
			logs.Debug("rolename %s -> %s", oldName, newName)
			if isRobotAcid(acid) {
				continue
			}
			// account
			if err := accountNameChg(acid, newName, ResDB, cb); err != nil {
				return err
			}
		}
	}
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeRoleName save res names err %s", err.Error())
		return err
	}
	// names
	for n, id := range ANames {
		cb.Send("HSET", driver.TableChangeName(Cfg.ResShardId), n, id)
		if !isRobotAcid(id) {
			updateAccountFriend(id, ResDB, cb)
			updateGank(id, ResDB, cb)
			account_num++
		}
	}
	for n, id := range BNames {
		cb.Send("HSET", driver.TableChangeName(Cfg.ResShardId), n, id)
		if !isRobotAcid(id) {
			updateAccountFriend(id, ResDB, cb)
			updateGank(id, ResDB, cb)
			account_num++
		}
	}
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeRoleName save res names err %s", err.Error())
		return err
	}
	logs.Info("account sum num %d", account_num)
	return nil
}

func nameMerge(name string, shard uint) string {
	order := Cfg.Shard2Order[shard]
	ss := strings.Split(name, name_split_sym)
	if len(ss) > 1 {
		ss[1] = fmt.Sprintf("%d", order)
		return strings.Join(ss, name_split_sym)
	} else {
		return strings.Join([]string{name, fmt.Sprintf("%d", order)},
			name_split_sym)
	}
}

func accountNameChg(acid, name string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	accountID, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("MergeRoleName ParseAccount %s err %s", acid, err.Error())
		return err
	}

	// profile
	a := new(account.Account)
	a.AccountID = accountID
	a.Profile = account.NewProfile(accountID)

	p := &a.Profile
	profile_key := p.DBName()

	if err := cb.Send("HSET", profile_key, "name", name); err != nil {
		return err
	}
	if err := cb.Send("HSET", profile_key, "rename_count", 0); err != nil {
		return err
	}

	// simpleinfo
	_simpleInfo := simple_info.NewSimpleInfoProfile(accountID)
	simpleInfo := &_simpleInfo
	key := simpleInfo.DBName()
	if err := cb.Send("HSET", key, "Name", name); err != nil {
		return err
	}

	return nil
}

func isRobotAcid(acid string) bool {
	return team_pvp.IsTPvpRobotId(acid) || ws_pvp.IsRobotId(acid)
}
