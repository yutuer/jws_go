package worship

import (
	"encoding/json"

	"errors"
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	keyNameInRedis = "topAccountWorship:%d:%d"
)

func getKeyNameInRedis(sid uint) string {
	return fmt.Sprintf(
		keyNameInRedis,
		game.Cfg.Gid,
		game.Cfg.GetShardIdByMerge(sid))
}

func (info *TopAccountWorship) loadDB() error {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("modules worship loaddb GetDBConn nil")
		return errors.New("worship loaddb GetDBConn nil")
	}

	data, err := redis.Bytes(db.Do("GET", info.dbKey))
	if err == redis.ErrNil {
		info.clean()
		return nil
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, info)
	if err != nil {
		return err
	}

	info.worshipMap = make(map[string]*worshipAccData, len(info.Worship))
	for i := 0; i < len(info.Worship); i++ {
		info.worshipMap[info.Worship[i].AccountID] =
			&info.Worship[i]
	}

	return nil
}

func (info *TopAccountWorship) saveDB() error {
	bb, err := json.Marshal(*info)
	if err != nil {
		logs.Error(
			"modules worship savedb marshal err %s",
			err.Error())
		return err
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("modules worship savedb GetDBConn nil")
		return errors.New("modules worship dbConn nil")
	}

	_, err = db.Do("SET", info.dbKey, string(bb))
	if err != nil {
		return err
	}

	return nil
}
