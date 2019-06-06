package worldboss

import (
	"fmt"

	"encoding/json"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

const (
	fieldPrefixPlayerName      = "name"
	fieldPrefixPlayerVip       = "vip"
	fieldPrefixPlayerSid       = "sid"
	fieldPrefixPlayerLevel     = "level"
	fieldPrefixPlayerGs        = "gs"
	fieldPrefixPlayerGuildName = "guildname"
	fieldPrefixPlayerTeam      = "team"
)

//PlayerDB ..
type PlayerDB struct {
	group uint32
}

func newPlayerDB(res *resources) *PlayerDB {
	db := &PlayerDB{
		group: res.group,
	}
	return db
}

func (db *PlayerDB) tableNamePlayer(tag string) string {
	return fmt.Sprintf("worldboss:%d:player:%s", db.group, tag)
}

func (db *PlayerDB) fieldNamePlayer(acid string, prefix string) string {
	return fmt.Sprintf("%s:%s", prefix, acid)
}

func (db *PlayerDB) getPlayerInfo(acid string, tag string) (*PlayerInfo, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNamePlayer(tag)
	fieldName := db.fieldNamePlayer(acid, fieldPrefixPlayerName)
	fieldVip := db.fieldNamePlayer(acid, fieldPrefixPlayerVip)
	fieldSid := db.fieldNamePlayer(acid, fieldPrefixPlayerSid)
	fieldLevel := db.fieldNamePlayer(acid, fieldPrefixPlayerLevel)
	fieldGs := db.fieldNamePlayer(acid, fieldPrefixPlayerGs)
	fieldGuildName := db.fieldNamePlayer(acid, fieldPrefixPlayerGuildName)
	res, err := redis.Values(conn.Do(
		"HMGET", tableName,
		fieldName,
		fieldVip,
		fieldSid,
		fieldLevel,
		fieldGs,
		fieldGuildName,
	))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Do getPlayerInfo HMGET failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	if len(res) != 6 {
		return nil, fmt.Errorf("Redis Do getPlayerInfo HMGET, return length is not 3, is (%d)", len(res))
	}

	ret := &PlayerInfo{}

	name, err := redis.String(res[0], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Name failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Name = name

	vip, err := redis.Int(res[1], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Vip failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Vip = uint32(vip)

	sid, err := redis.Int(res[2], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Sid failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Sid = uint32(sid)

	lv, err := redis.Int(res[3], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Level failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Level = uint32(lv)

	gs, err := redis.Int64(res[4], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Gs failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Gs = int64(gs)

	guildname, err := redis.String(res[5], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse GuildName failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.GuildName = guildname

	return ret, nil
}

func (db *PlayerDB) setPlayerInfo(acid string, info PlayerInfo, tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNamePlayer(tag)
	fieldName := db.fieldNamePlayer(acid, fieldPrefixPlayerName)
	fieldVip := db.fieldNamePlayer(acid, fieldPrefixPlayerVip)
	fieldSid := db.fieldNamePlayer(acid, fieldPrefixPlayerSid)
	fieldLevel := db.fieldNamePlayer(acid, fieldPrefixPlayerLevel)
	fieldGs := db.fieldNamePlayer(acid, fieldPrefixPlayerGs)
	fieldGuildName := db.fieldNamePlayer(acid, fieldPrefixPlayerGuildName)
	_, err := conn.Do(
		"HMSET", tableName,
		fieldName, info.Name,
		fieldVip, info.Vip,
		fieldSid, info.Sid,
		fieldLevel, info.Level,
		fieldGs, info.Gs,
		fieldGuildName, info.GuildName,
	)
	if nil != err {
		return fmt.Errorf("Redis Do setPlayerInfo HMSET failed, %v", err)
	}

	return nil
}

func (db *PlayerDB) getPlayerTeam(acid string, tag string) (*TeamInfoDetail, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNamePlayer(tag)
	fieldTeam := db.fieldNamePlayer(acid, fieldPrefixPlayerTeam)
	res, err := redis.String(conn.Do(
		"HGET", tableName,
		fieldTeam,
	))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Do setPlayerInfo HGET failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	ret := &TeamInfoDetail{}
	if err := json.Unmarshal([]byte(res), ret); nil != err {
		return nil, fmt.Errorf("Unmarshal TeamInfo failed, %v", err)
	}

	return ret, nil
}

func (db *PlayerDB) setPlayerTeam(acid string, team TeamInfoDetail, tag string) error {
	bs, err := json.Marshal(team)
	if nil != err {
		return fmt.Errorf("Marshal TeamInfo failed, %v", err)
	}
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNamePlayer(tag)
	fieldTeam := db.fieldNamePlayer(acid, fieldPrefixPlayerTeam)
	_, err = conn.Do(
		"HSET", tableName,
		fieldTeam, string(bs),
	)
	if nil != err {
		return fmt.Errorf("Redis Do setPlayerInfo HSET failed, %v", err)
	}

	return nil
}

func (db *PlayerDB) removePlayerBatch(tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNamePlayer(tag)
	_, err := conn.Do(
		"DEL", tableName,
	)
	if nil != err {
		return fmt.Errorf("Redis Do removePlayerBatch DEL failed, %v", err)
	}
	return nil
}
