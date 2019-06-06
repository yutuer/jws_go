package worldboss

import (
	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

const (
	fieldNameBossLevel         = "lv"
	fieldNameBossHP            = "hp"
	fieldNameBossSeq           = "seq"
	fieldNameRoundLastReset    = "last_reset"
	fieldNameRoundLastNewBoss  = "last_new_boss"
	fieldNameRoundLastReward   = "last_reward"
	fieldNameCommonTotalDamage = "common_total_damage"
)

//BossDB ..
type BossDB struct {
	group uint32
}

func newBossDB(res *resources) *BossDB {
	db := &BossDB{
		group: res.group,
	}
	return db
}

func (db *BossDB) tableNameBoss(tag string) string {
	return fmt.Sprintf("worldboss:%d:boss:%s", db.group, tag)
}

func (db *BossDB) getBossStatus(tag string) (*BossStatus, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss(tag)
	res, err := redis.Values(conn.Do(
		"HMGET", tableName,
		fieldNameBossLevel,
		fieldNameBossHP,
		fieldNameBossSeq,
	))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Do getBossStatus HMGET failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	if len(res) != 3 {
		return nil, fmt.Errorf("Redis Do getBossStatus HMGET, return length is not 3, is (%d)", len(res))
	}

	ret := &BossStatus{}

	lv, err := redis.Int(res[0], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Level failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Level = uint32(lv)

	hp, err := redis.Int64(res[1], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse HP failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.HPCurr = uint64(hp)

	seq, err := redis.Int(res[2], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Seq failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.Seq = uint32(seq)

	return ret, nil
}

func (db *BossDB) setBossStatus(status BossStatus, tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss(tag)
	_, err := conn.Do(
		"HMSET", tableName,
		fieldNameBossLevel, status.Level,
		fieldNameBossHP, status.HPCurr,
		fieldNameBossSeq, status.Seq,
	)
	if nil != err {
		return fmt.Errorf("Redis Do setBossStatus HMSET failed, %v", err)
	}

	return nil
}

func (db *BossDB) getBossCommonStatus(tag string) (*BossCommonStatus, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss(tag)
	res, err := redis.Values(conn.Do(
		"HMGET", tableName,
		fieldNameCommonTotalDamage,
	))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Do getBossStatus HMGET failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("Redis Do getBossStatus HMGET, return length is not 3, is (%d)", len(res))
	}

	ret := &BossCommonStatus{}

	td, err := redis.Int64(res[0], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse Level failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.TotalDamage = uint64(td)

	return ret, nil
}

func (db *BossDB) setBossCommonStatus(status BossCommonStatus, tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss(tag)
	_, err := conn.Do(
		"HMSET", tableName,
		fieldNameCommonTotalDamage, status.TotalDamage,
	)
	if nil != err {
		return fmt.Errorf("Redis Do setBossCommonStatus HMSET failed, %v", err)
	}

	return nil
}

func (db *BossDB) removeBossBatch(tag string) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss(tag)
	_, err := conn.Do(
		"DEL", tableName,
	)
	if nil != err {
		return fmt.Errorf("Redis Do removeBossBatch DEL failed, %v", err)
	}
	return nil
}

func (db *BossDB) getRoundStatus() (*RoundStatus, error) {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss("0")
	res, err := redis.Values(conn.Do(
		"HMGET", tableName,
		fieldNameRoundLastReset,
		fieldNameRoundLastNewBoss,
		fieldNameRoundLastReward,
	))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Do getRoundStatus HMGET failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	if len(res) != 3 {
		return nil, fmt.Errorf("Redis Do getRoundStatus HMGET, return length is not 3, is (%d)", len(res))
	}

	ret := &RoundStatus{}

	resetTime, err := redis.Int64(res[0], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse LastResetTime failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.LastResetTime = resetTime

	newTime, err := redis.Int64(res[1], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse LastNewBossTime failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.LastNewBossTime = newTime

	rewardTime, err := redis.Int64(res[2], nil)
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("Redis Parse LastRewardTime failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}
	ret.LastRewardTime = rewardTime

	return ret, nil
}

func (db *BossDB) setRoundStatus(status RoundStatus) error {
	conn := csdb.GetDBConn(db.group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Failed")
	}

	tableName := db.tableNameBoss("0")
	_, err := conn.Do(
		"HMSET", tableName,
		fieldNameRoundLastReset, status.LastResetTime,
		fieldNameRoundLastNewBoss, status.LastNewBossTime,
		fieldNameRoundLastReward, status.LastRewardTime,
	)
	if nil != err {
		return fmt.Errorf("Redis Do setRoundStatus HMSET failed, %v", err)
	}

	return nil
}
