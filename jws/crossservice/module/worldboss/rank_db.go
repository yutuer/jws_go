package worldboss

import (
	"fmt"

	"encoding/json"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//RankDB ..
type RankDB struct {
	group uint32
}

func newRankDB(res *resources) *RankDB {
	db := &RankDB{
		group: res.group,
	}
	return db
}

func (db *RankDB) tableNameRank(tag string) string {
	return fmt.Sprintf("worldboss:%d:rank:%s", db.group, tag)
}

func (db *RankDB) tableNameFormationRank(tag string) string {
	return fmt.Sprintf("worldboss:%d:formationrank:%s", db.group, tag)
}

func (db *RankDB) getAllRankMember(tag string) (map[string]DamageRankElem, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameRank(tag)
	res, err := redis.Values(conn.Do("HGETALL", tableName))
	if nil != err {
		return nil, fmt.Errorf("Redis Do getAllRankMember HGETALL failed, %v", err)
	}

	ret := make(map[string]DamageRankElem)
	for i := 0; i < len(res)-1; i += 2 {
		acid, err := redis.String(res[i], nil)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB getAllRankMember, Parse Acid failed, %v", err)
			continue
		}
		bs, err := redis.String(res[i+1], nil)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB getAllRankMember, Parse DamageRankElem failed, %v", err)
			continue
		}
		elem := DamageRankElem{}
		if err := json.Unmarshal([]byte(bs), &elem); nil != err {
			logs.Warn("[WorldBoss] RankDB getAllRankMember, Unmarshal DamageRankElem failed, %v", err)
			continue
		}
		ret[acid] = elem
	}

	return ret, nil
}

func (db *RankDB) pushRankMember(members []DamageRankElem, tag string) error {
	val := []interface{}{}
	tableName := db.tableNameRank(tag)
	val = append(val, tableName)
	for _, elem := range members {
		bs, err := json.Marshal(elem)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB pushRankMember, Marshal DamageRankElem failed, %v", err)
			continue
		}
		val = append(val, elem.Acid)
		val = append(val, string(bs))
	}

	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	_, err := conn.Do("HMSET", val...)
	if nil != err {
		return fmt.Errorf("Redis Do pushRankMember HMSET failed, %v", err)
	}

	return nil
}

func (db *RankDB) removeRank(tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameRank(tag)
	_, err := conn.Do("DEL", tableName)
	if nil != err {
		return fmt.Errorf("Redis Do removeRank DEL failed, %v", err)
	}
	return nil
}

func (db *RankDB) getAllFormationRankMember(tag string) (map[string]FormationRankElem, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameFormationRank(tag)
	res, err := redis.Values(conn.Do("HGETALL", tableName))
	if nil != err {
		return nil, fmt.Errorf("Redis Do getAllFormationRankMember HGETALL failed, %v", err)
	}

	ret := make(map[string]FormationRankElem)
	for i := 0; i < len(res)-1; i += 2 {
		acid, err := redis.String(res[i], nil)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB getAllFormationRankMember, Parse Acid failed, %v", err)
			continue
		}
		bs, err := redis.String(res[i+1], nil)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB getAllFormationRankMember, Parse FormationRankElem failed, %v", err)
			continue
		}
		elem := FormationRankElem{}
		if err := json.Unmarshal([]byte(bs), &elem); nil != err {
			logs.Warn("[WorldBoss] RankDB getAllFormationRankMember, Unmarshal FormationRankElem failed, %v", err)
			continue
		}
		ret[acid] = elem
	}

	return ret, nil
}

func (db *RankDB) pushFormationRankMember(members []FormationRankElem, tag string) error {
	val := []interface{}{}
	tableName := db.tableNameFormationRank(tag)
	val = append(val, tableName)
	for _, elem := range members {
		bs, err := json.Marshal(elem)
		if nil != err {
			logs.Warn("[WorldBoss] RankDB pushFormationRankMember, Marshal FormationRankElem failed, %v", err)
			continue
		}
		val = append(val, elem.Acid)
		val = append(val, string(bs))
	}

	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	_, err := conn.Do("HMSET", val...)
	if nil != err {
		return fmt.Errorf("Redis Do pushFormationRankMember HMSET failed, %v", err)
	}

	return nil
}

func (db *RankDB) removeFormationRank(tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameFormationRank(tag)
	_, err := conn.Do("DEL", tableName)
	if nil != err {
		return fmt.Errorf("Redis Do removeFormationRank DEL failed, %v", err)
	}
	return nil
}
