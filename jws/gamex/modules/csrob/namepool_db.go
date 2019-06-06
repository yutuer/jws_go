package csrob

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//NamePoolDB 人物缓存信息存储句柄
type NamePoolDB struct {
	groupID uint32
}

func initNamePoolDB(res *resources) *NamePoolDB {
	return &NamePoolDB{
		groupID: res.groupID,
	}
}

func (db *NamePoolDB) tablePlayerName() string {
	return fmt.Sprintf("csrob:%d:%s:%s", db.groupID, "namepool", "player")
}

func (db *NamePoolDB) tableGuildName() string {
	return fmt.Sprintf("csrob:%d:%s:%s", db.groupID, "namepool", "guild")
}

func (db *NamePoolDB) getPlayer(acid string) (*NamePoolPlayer, error) {
	logs.Debug("[CSRob] getPlayer {%v}", acid)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err, %v", conn.Err())
	}

	tableName := db.tablePlayerName()
	bs, err := redis.Bytes(conn.Do("HGET", tableName, acid))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("NamePoolDB getPlayer, HGET redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	name := &NamePoolPlayer{}
	err = json.Unmarshal(bs, name)
	if nil != err {
		return nil, makeError("NamePoolDB getPlayer, Marshal failed, %v, ...{%v}", err, bs)
	}

	return name, nil
}

func (db *NamePoolDB) pushPlayer(name *NamePoolPlayer) error {
	logs.Debug("[CSRob] pushPlayer {%v}", name)
	bs, err := json.Marshal(name)
	if nil != err {
		return makeError("NamePoolDB pushPlayer, Marshal failed, %v, ...{%v}", err, name)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err, %v", conn.Err())
	}

	tableName := db.tablePlayerName()
	_, err = conn.Do("HSET", tableName, name.Acid, bs)
	if nil != err {
		return makeError("NamePoolDB pushPlayer, HSET redis failed, %v", err)
	}

	return nil
}

func (db *NamePoolDB) getGuild(acid string) (*NamePoolGuild, error) {
	logs.Debug("[CSRob] getGuild {%v}", acid)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err, %v", conn.Err())
	}

	tableName := db.tableGuildName()
	bs, err := redis.Bytes(conn.Do("HGET", tableName, acid))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("NamePoolDB getGuild, HGET redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	name := &NamePoolGuild{}
	err = json.Unmarshal(bs, name)
	if nil != err {
		return nil, makeError("NamePoolDB getGuild, Marshal failed, %v, ...{%v}", err, bs)
	}

	return name, nil
}

func (db *NamePoolDB) pushGuild(name *NamePoolGuild) error {
	logs.Debug("[CSRob] pushGuild {%v}", name)
	bs, err := json.Marshal(name)
	if nil != err {
		return makeError("NamePoolDB pushGuild, Marshal failed, %v, ...{%v}", err, name)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err, %v", conn.Err())
	}

	tableName := db.tableGuildName()
	_, err = conn.Do("HSET", tableName, name.GuildID, bs)
	if nil != err {
		return makeError("NamePoolDB pushGuild, HSET redis failed, %v", err)
	}

	return nil
}
