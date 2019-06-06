package csrob

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//RewardDB 奖励信息存储句柄
type RewardDB struct {
	groupID uint32
	sid     uint
}

func initRewardDB(res *resources) *RewardDB {
	return &RewardDB{
		groupID: res.groupID,
		sid:     res.sid,
	}
}

func (db *RewardDB) tableName() string {
	return fmt.Sprintf("csrob:%d:%d:%s", db.groupID, db.sid, "rewardbox")
}

func (db *RewardDB) getAllRewardFromBox() ([]RewardBoxElem, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableName()
	res, err := redis.Strings(conn.Do("LRANGE", tableName, 0, -1))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("RewardDB getAllRewardFromBox, LRANGE redis failed, %v", err)
	}

	if redis.ErrNil == err {
		return []RewardBoxElem{}, nil
	}

	rewards := make([]RewardBoxElem, 0, len(res))
	for _, single := range res {
		reward := RewardBoxElem{}
		err := json.Unmarshal([]byte(single), &reward)
		if nil != err {
			logs.Warn("RewardDB getAllRewardFromBox, Unmarshal failed, %v, ...{%v}", err, single)
			continue
		}

		rewards = append(rewards, reward)
	}
	return rewards, nil
}

func (db *RewardDB) pushRewardToBox(elem *RewardBoxElem) error {
	logs.Debug("[CSRob] pushRewardToBox {%v}", elem)
	bs, err := json.Marshal(elem)
	if nil != err {
		return makeError("RewardDB pushRewardToBox, Marshal failed, %v, ...{%v}", err, elem)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableName()
	_, err = conn.Do("RPUSH", tableName, string(bs))
	if nil != err {
		return makeError("RewardDB pushRewardToBox, RPUSH redis failed, %v", err)
	}

	return nil
}

func (db *RewardDB) removeRewardFromBox(elem *RewardBoxElem) error {
	logs.Debug("[CSRob] removeRewardFromBox {%v}", elem)
	bs, err := json.Marshal(elem)
	if nil != err {
		return makeError("RewardDB removeRewardFromBox, Marshal failed, %v, ...{%v}", err, elem)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableName()
	ret, err := redis.Int(conn.Do("LREM", tableName, 0, string(bs)))
	if nil != err {
		return makeError("RewardDB pushRewardToBox, LREM redis failed, %v", err)
	}

	if 1 != ret {
		logs.Warn("[CSRob] removeRewardFromBox but no exist")
		return nil
	}

	logs.Debug("[CSRob] removeRewardFromBox remove {%v}", elem)

	return nil
}

func (db *RewardDB) tableNameWeek() string {
	return fmt.Sprintf("csrob:%d:%s", db.groupID, "weekReward")
}

func (db *RewardDB) getRewardWeek() (*RewardWeek, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableNameWeek()
	bs, err := redis.Bytes(conn.Do("GET", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("RewardDB getRewardWeek, GET redis failed, %v", err)
	}

	if redis.ErrNil == err {
		return nil, nil
	}

	ret := &RewardWeek{}
	if err := json.Unmarshal(bs, ret); nil != err {
		return nil, makeError("RewardDB getRewardWeek, Unmarshal failed, %v, bs [%s]", err, string(bs))
	}

	return ret, nil
}

func (db *RewardDB) setRewardWeek(reward *RewardWeek) error {
	bs, err := json.Marshal(reward)
	if nil != err {
		return makeError("RewardDB setRewardWeek, Marshal failed, %v, reward [%v]", err, reward)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableNameWeek()
	_, err = conn.Do("SET", tableName, bs)
	if nil != err {
		return makeError("RewardDB setRewardWeek, SET redis failed, %v", err)
	}

	return nil
}
