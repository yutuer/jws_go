package csrob

import (
	"encoding/json"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	tablePlayerPrefix = "car"

	suffixPlayerInfo   = "info"
	suffixPlayerRecord = "record"
	suffixPlayerAppeal = "appeal"
	suffixPlayerEnemy  = "enemy"
	suffixPlayerRob    = "rob"
	suffixPlayerReward = "reward"
	suffixPlayerStatus = "status"
)

func (db *PlayerDB) tablePlayerName(acid string, suffix string) string {
	return fmt.Sprintf("csrob:%d:%s:%s:%s", db.groupID, tablePlayerPrefix, acid, suffix)
}

type PlayerDB struct {
	groupID uint32
}

func initPlayerDB(res *resources) *PlayerDB {
	return &PlayerDB{
		groupID: res.groupID,
	}
}

func (db *PlayerDB) testLink() error {
	info := &PlayerInfo{
		Acid: "testlink",
	}
	return db.setInfo(info)
}

//-- car:acid:info

func (db *PlayerDB) getInfo(acid string) (*PlayerInfo, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerInfo)
	bs, err := redis.Bytes(conn.Do("GET", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getInfo acid [%s], GET redis failed, %v", acid, err)
	}

	if redis.ErrNil == err {
		return nil, nil
	}

	info := &PlayerInfo{
		CarList:       []PlayerCarListElem{},
		CurrFormation: []int{},
	}
	err = json.Unmarshal(bs, info)
	if nil != err {
		return nil, makeError("PlayerDB getInfo acid [%s], Unmarshal failed, %v, ...{%v}", acid, err, string(bs))
	}

	return info, nil
}

func (db *PlayerDB) setInfo(info *PlayerInfo) error {
	bs, err := json.Marshal(info)
	if nil != err {
		return makeError("PlayerDB setInfo acid [%s], Marshal failed, %v, ...{%v}", info.Acid, err, info)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(info.Acid, suffixPlayerInfo)
	_, err = conn.Do("SET", tableName, bs)
	if nil != err {
		return makeError("PlayerDB setInfo acid [%s], SET redis failed, %v", info.Acid, err)
	}

	return nil
}

//-- car:acid:record

func (db *PlayerDB) getRecords(acid string, num int) ([]PlayerRecord, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRecord)
	res, err := redis.Strings(conn.Do("LRANGE", tableName, 0, num))
	if nil != err {
		return nil, makeError("PlayerDB getRecords acid [%s], LRANGE redis failed, %v", acid, err)
	}

	records := make([]PlayerRecord, 0, len(res))
	for _, single := range res {
		record := PlayerRecord{}
		err := json.Unmarshal([]byte(single), &record)
		if nil != err {
			logs.Warn("PlayerDB getRecords acid [%s], Unmarshal failed, %v, ...{%v}", acid, err, single)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

func (db *PlayerDB) pushRecord(acid string, record PlayerRecord) error {
	bs, err := json.Marshal(record)
	if nil != err {
		return makeError("PlayerDB pushRecord acid [%s], Marshal failed, %v, ...{%v}", acid, err, record)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRecord)
	_, err = conn.Do("LPUSH", tableName, string(bs))
	if nil != err {
		return makeError("PlayerDB pushRecord acid [%s], LPUSH redis failed, %v", acid, err)
	}

	return nil
}

//-- car:acid:appeal

func (db *PlayerDB) getAppeals(acid string, num int) ([]PlayerAppeal, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerAppeal)
	res, err := redis.Strings(conn.Do("LRANGE", tableName, 0, num))
	if nil != err {
		return nil, makeError("PlayerDB getAppeals acid [%s], LRANGE redis failed, %v", acid, err)
	}

	appeals := make([]PlayerAppeal, 0, len(res))
	for _, single := range res {
		appeal := PlayerAppeal{}
		err := json.Unmarshal([]byte(single), &appeal)
		if nil != err {
			logs.Warn("PlayerDB getAppeals acid [%s], Unmarshal failed, %v, ...{%v}", acid, err, single)
			continue
		}

		appeals = append(appeals, appeal)
	}

	return appeals, nil
}

func (db *PlayerDB) pushAppeal(acid string, appeal PlayerAppeal) error {
	bs, err := json.Marshal(appeal)
	if nil != err {
		return makeError("PlayerDB pushRecord acid [%s], Marshal failed, %v, ...{%v}", acid, err, appeal)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerAppeal)
	_, err = conn.Do("LPUSH", tableName, string(bs))
	if nil != err {
		return makeError("PlayerDB pushAppeal acid [%s], LPUSH redis failed, %v", acid, err)
	}

	return nil
}

//-- car:acid:enemy

func (db *PlayerDB) getEnemies(acid string) ([]PlayerEnemy, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerEnemy)
	res, err := redis.Int64Map(conn.Do("HGETALL", tableName))
	if nil != err {
		return nil, makeError("PlayerDB getEnemies acid [%s], HGETALL redis failed, %v", acid, err)
	}

	enemies := make([]PlayerEnemy, 0, len(res)/2)
	for eid, c := range res {
		enemy := PlayerEnemy{
			Acid:  eid,
			Count: uint32(c),
		}

		enemies = append(enemies, enemy)
	}

	return enemies, nil
}

func (db *PlayerDB) pushEnemy(acid string, enemy string, inc int) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerEnemy)
	_, err := conn.Do("HINCRBY", tableName, enemy, inc)
	if nil != err {
		return makeError("PlayerDB pushEnemy acid [%s], HINCRBY redis failed, %v", acid, err)
	}

	return nil
}

//-- car:acid:rob

const (
	fieldPrefixRob = "car"

	fieldRobInfo      = "info"
	fieldRobHelper    = "helper"
	fieldRobLock      = "lock"
	fieldRobLockOwner = "lockowner"
	fieldRobReward    = "reward"

	//redisSuccess = 1
	redisFail = 0
)

func (db *PlayerDB) fieldNameRob(id uint32, suffix string) string {
	return fmt.Sprintf("%s:%d:%s", fieldPrefixRob, id, suffix)
}

func (db *PlayerDB) tableNameRobList(acid string, id uint32) string {
	return fmt.Sprintf("%s:%d", db.tablePlayerName(acid, suffixPlayerRob), id)
}

//出车
func (db *PlayerDB) buildCar(acid string, info PlayerRobInfo) error {
	bs, err := json.Marshal(info)
	if nil != err {
		return makeError("PlayerDB buildRob acid [%s], Marshal failed, %v, ...{%v}", acid, err, info)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldInfo := db.fieldNameRob(info.CarID, fieldRobInfo)

	_, err = conn.Do("HSET", tableName, fieldInfo, bs)
	if nil != err {
		return makeError("PlayerDB buildRob acid [%s] car [%d], HSET redis failed, %v", acid, info.CarID, err)
	}

	return nil
}

func (db *PlayerDB) getRob(acid string, car uint32) (*PlayerRob, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fields := []interface{}{
		db.fieldNameRob(car, fieldRobInfo),
		db.fieldNameRob(car, fieldRobLock),
		db.fieldNameRob(car, fieldRobHelper),
		db.fieldNameRob(car, fieldRobReward),
	}
	values, err := redis.Values(conn.Do("HMGET", append([]interface{}{tableName}, fields...)...))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getRob acid [%s] car [%d], HMGET redis failed, %v", acid, car, err)
	}

	if redis.ErrNil == err {
		return nil, nil
	}

	if len(fields) != len(values) {
		return nil, makeError("PlayerDB getRob acid [%s] car [%d], HMGET redis result less, ...{%v}", acid, car, values)
	}

	//解析info, 如果连info都没有就返回nil
	ret := &PlayerRob{}
	if nil != values[0] {
		bs, err := redis.Bytes(values[0], nil)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Parse Bytes failed, %v, ...{%v}", acid, car, err, values[0])
		}
		err = json.Unmarshal(bs, &(ret.Info))
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Unmarshal failed, %v, ...{%v}", acid, car, err, string(bs))
		}
	} else {
		return nil, nil
	}
	ret.CarID = ret.Info.CarID
	ret.Acid = acid

	//解析是否锁定
	if nil != values[1] {
		expire, err := redis.Int64(values[1], nil)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Parse Int64 failed, %v, ...{%v}", acid, car, err, values[1])
		}
		if expire < time.Now().Unix() {
			ret.Robbing = false
		} else {
			ret.Robbing = true
		}
	} else {
		ret.Robbing = false
	}

	//解析Helper
	if nil != values[2] {
		bs, err := redis.Bytes(values[2], nil)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Parse Bytes failed, %v, ...{%v}", acid, car, err, values[2])
		}
		ret.Helper = &PlayerRobHelper{}
		err = json.Unmarshal(bs, ret.Helper)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Unmarshal failed, %v, ...{%v}", acid, car, err, string(bs))
		}
	}

	//解析Reward
	if nil != values[3] {
		bs, err := redis.Bytes(values[3], nil)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Parse Bytes failed, %v, ...{%v}", acid, car, err, values[3])
		}
		ret.Reward = &PlayerRobReward{}
		err = json.Unmarshal(bs, ret.Reward)
		if nil != err {
			return nil, makeError("PlayerDB getRob acid [%s] car [%d], Unmarshal failed, %v, ...{%v}", acid, car, err, string(bs))
		}
	}

	//读取被抢劫列表
	tableList := db.tableNameRobList(acid, car)
	list, err := redis.Strings(conn.Do("LRANGE", tableList, 0, -1))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getRob acid [%s] car [%d], LRANGE redis failed, %v", acid, car, err)
	}

	if redis.ErrNil == err {
		ret.Robbers = []string{}
	} else {
		ret.Robbers = list
	}

	return ret, nil
}

func (db *PlayerDB) pushRob(acid string, car uint32, robber string, limit uint32) (bool, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return false, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldLock := db.fieldNameRob(car, fieldRobLock)
	fieldOwner := db.fieldNameRob(car, fieldRobLockOwner)
	tableList := db.tableNameRobList(acid, car)

	script := `
		local tableName = KEYS[1]
		local tableList = KEYS[2]
		local fieldLock = KEYS[3]
		local fieldOwner = KEYS[4]
		local owner = KEYS[5]
		local lenLimit = KEYS[6]

		local older = redis.call("HGET", tableName, fieldOwner)
		if older == owner
		then
			local oldlen = redis.call("LLEN", tableList)
			if tonumber(oldlen) >= tonumber(lenLimit)
			then
				return 0
			else
				redis.call("LPUSH", tableList, owner)
				redis.call("HDEL", tableName, fieldLock)
				redis.call("HDEL", tableName, fieldOwner)
				return 1
			end
		else
			return 0
		end

		return 0
	`

	ok, err := redis.Int(conn.Do("EVAL", script, 6, tableName, tableList, fieldLock, fieldOwner, robber, limit))
	if nil != err {
		return false, makeError("PlayerDB pushRob acid [%s], EVAL redis failed, %v", acid, err)
	}

	if redisFail == ok {
		return false, nil
	}

	return true, nil
}

func (db *PlayerDB) touchRob(acid string, car uint32, robber string, timeout int64) (bool, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return false, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldLock := db.fieldNameRob(car, fieldRobLock)
	fieldOwner := db.fieldNameRob(car, fieldRobLockOwner)

	script := `
		local tableName = KEYS[1]
		local fieldLock = KEYS[2]
		local timeout = KEYS[3]
		local fieldOwner = KEYS[4]
		local owner = KEYS[5]
		local now = KEYS[6]

		local res = redis.call("HSETNX", tableName, fieldLock, timeout)
		if 1 == res
		then
			redis.call("HSET", tableName, fieldOwner, owner)
			return 1
		else
			local expire = redis.call("HGET", tableName, fieldLock)
			if tonumber(expire) < tonumber(now)
			then
				redis.call("HSET", tableName, fieldLock, timeout)
				redis.call("HSET", tableName, fieldOwner, owner)
				return 1
			else
				local older = redis.call("HGET", tableName, fieldOwner)
				if older == owner
				then
					redis.call("HSET", tableName, fieldLock, timeout)
					redis.call("HSET", tableName, fieldOwner, owner)
					return 1
				else
					return 0
				end
			end
		end
		return 0
	`

	ok, err := redis.Int(conn.Do("EVAL", script, 6, tableName, fieldLock, timeout, fieldOwner, robber, time.Now().Unix()))
	if nil != err {
		return false, makeError("PlayerDB touchRob acid [%s], EVAL redis failed, %v", acid, err)
	}

	if redisFail == ok {
		return false, nil
	}

	return true, nil
}

func (db *PlayerDB) unTouchRob(acid string, car uint32, robber string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldLock := db.fieldNameRob(car, fieldRobLock)
	fieldOwner := db.fieldNameRob(car, fieldRobLockOwner)

	script := `
		local tableName = KEYS[1]
		local fieldLock = KEYS[2]
		local fieldOwner = KEYS[3]
		local owner = KEYS[4]

		local older = redis.call("HGET", tableName, fieldOwner)
		if older == owner
		then
			redis.call("HDEL", tableName, fieldLock)
			redis.call("HDEL", tableName, fieldOwner)
		end
		return 1
	`

	_, err := redis.Int(conn.Do("EVAL", script, 4, tableName, fieldLock, fieldOwner, robber))
	if nil != err {
		return makeError("PlayerDB unTouchRob acid [%s], EVAL redis failed, %v", acid, err)
	}

	return nil
}

func (db *PlayerDB) getRobHelper(acid string, car uint32) (*PlayerRobHelper, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldHelper := db.fieldNameRob(car, fieldRobHelper)

	bs, err := redis.Bytes(conn.Do("HGET", tableName, fieldHelper))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getRobHelper acid [%s] car [%d], HGET redis failed, %v", acid, car, err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	helper := &PlayerRobHelper{}
	if err := json.Unmarshal(bs, helper); nil != err {
		return nil, makeError("PlayerDB getRobHelper acid [%s], Unmarshal failed, %v", acid, err)
	}

	return helper, nil
}

func (db *PlayerDB) setRobHelper(acid string, car uint32, helper *PlayerRobHelper) (bool, error) {
	bs, err := json.Marshal(helper)
	if nil != err {
		return false, makeError("PlayerDB setRobHelper acid [%s], Marshal failed, %v, ...{%v}", acid, err, helper)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return false, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldHelper := db.fieldNameRob(car, fieldRobHelper)

	ok, err := redis.Int(conn.Do("HSETNX", tableName, fieldHelper, bs))
	if nil != err {
		return false, makeError("PlayerDB setRobHelper acid [%s] car [%d], HSETNX redis failed, %v", acid, car, err)
	}

	if redisFail == ok {
		return false, nil
	}

	return true, nil
}

func (db *PlayerDB) setRobReward(acid string, car uint32, reward *PlayerRobReward) error {
	bs, err := json.Marshal(reward)
	if nil != err {
		return makeError("PlayerDB setRobReward acid [%s], Marshal failed, %v, ...{%v}", acid, err, reward)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	fieldReward := db.fieldNameRob(car, fieldRobReward)

	_, err = conn.Do("HSET", tableName, fieldReward, bs)
	if nil != err {
		return makeError("PlayerDB setRobReward acid [%s] car [%d], HSET redis failed, %v", acid, car, err)
	}

	return nil
}

//-- car:acid:reward

func (db *PlayerDB) getRewardInfo(acid string) (*PlayerRewardInfo, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerReward)
	bs, err := redis.Bytes(conn.Do("GET", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getRewardInfo acid [%s], GET redis failed, %v", acid, err)
	}

	if redis.ErrNil == err {
		return nil, nil
	}

	info := &PlayerRewardInfo{}
	err = json.Unmarshal(bs, info)
	if nil != err {
		return nil, makeError("PlayerDB getRewardInfo acid [%s], Unmarshal failed, %v, ...{%v}", acid, err, string(bs))
	}

	return info, nil
}

func (db *PlayerDB) setRewardInfo(acid string, info *PlayerRewardInfo) error {
	bs, err := json.Marshal(info)
	if nil != err {
		return makeError("PlayerDB setRewardInfo acid [%s], Marshal failed, %v, ...{%v}", acid, err, info)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerReward)
	_, err = conn.Do("SET", tableName, bs)
	if nil != err {
		return makeError("PlayerDB setRewardInfo acid [%s], SET redis failed, %v", acid, err)
	}

	return nil
}

func (db *PlayerDB) clearMyCars(acid string, carList []uint32) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRob)
	_, err := conn.Do("DEL", tableName)
	if nil != err && redis.ErrNil != err {
		return makeError("PlayerDB clearMyCars acid [%s], DEL redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return nil
	}

	for _, carID := range carList {
		robTableName := db.tableNameRobList(acid, carID)
		_, err := conn.Do("DEL", robTableName)
		if nil != err && redis.ErrNil != err {
			logs.Error("PlayerDB clearMyCars acid [%s], DEL redis failed, %v", acid, err)
			continue
		}
		if redis.ErrNil == err {
			continue
		}
	}

	return nil
}

func (db *PlayerDB) clearMyAppealBefore(acid string, before int64) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerAppeal)

	//取出所有现有的求援信息
	res, err := redis.Strings(conn.Do("LRANGE", tableName, 0, -1))
	if nil != err {
		return makeError("PlayerDB clearMyAppealBefore acid [%s], LRANGE redis failed, %v", acid, err)
	}

	//收集需要删除的求援
	delList := make([]string, 0, len(res))
	for _, single := range res {
		appeal := PlayerAppeal{}
		err := json.Unmarshal([]byte(single), &appeal)
		if nil != err {
			logs.Warn("PlayerDB clearMyAppealBefore acid [%s], Unmarshal failed, %v, ...{%v}", acid, err, single)
			continue
		}

		if appeal.EndStamp < before || appeal.AppealTime < before {
			delList = append(delList, single)
		}
	}

	//删除需要删除求援
	for _, single := range delList {
		_, err := conn.Do("LREM", tableName, 0, single)
		if nil != err && redis.ErrNil != err {
			logs.Error(fmt.Sprintf("%v", makeError("PlayerDB clearMyAppealBefore acid [%s], LREM redis failed, %v", acid, err)))
			continue
		}
		if redis.ErrNil == err {
			logs.Warn("PlayerDB clearMyAppealBefore acid [%s], LREM but it not exist, ...{%v}", acid, single)
			continue
		}
	}

	return nil
}

func (db *PlayerDB) clearMyEnemies(acid string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerEnemy)
	_, err := conn.Do("DEL", tableName)
	if nil != err && redis.ErrNil != err {
		return makeError("PlayerDB clearMyEnemies acid [%s], DEL redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return nil
	}
	return nil
}

func (db *PlayerDB) trimMyRecords(acid string, num int) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tablePlayerName(acid, suffixPlayerRecord)
	_, err := conn.Do("LTRIM", tableName, 0, num)
	if nil != err && redis.ErrNil != err {
		return makeError("PlayerDB trimMyRecords acid [%s], LTRIM redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return nil
	}
	return nil
}

//csrob:$group:car:$acid:status

var (
	fieldPlayerStatusVIP              = "vip"
	fieldPlayerStatusAppealCount      = "accept_appeal_count"
	fieldPlayerStatusAutoAcceptBottom = "auto_accept_bottom"
	fieldPlayerStatusLastUpdate       = "last_update"
)

func (db *PlayerDB) getPlayerStatus(acid string) (*PlayerStatus, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)
	res, err := redis.Values(conn.Do("HGETALL", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getPlayerStatus acid [%s], HGETALL redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return &PlayerStatus{}, nil
	}
	ret := &PlayerStatus{}
	if err := redis.ScanStruct(res, ret); nil != err {
		return nil, makeError("PlayerDB getPlayerStatus acid [%s], ScanStruct failed, %v", acid, err)
	}

	return ret, nil
}

func (db *PlayerDB) resetPlayerStatus(acid string, old *PlayerStatus, now time.Time) (*PlayerStatus, error) {
	script := `
		local tableName = KEYS[1]
		local fieldLastUpdate = KEYS[2]
		local fieldAppealCount = KEYS[3]
		local oldUpdate = KEYS[4]
		local newUpdate = KEYS[5]

		local older = redis.call("HGET", tableName, fieldLastUpdate)
		if older == oldUpdate
		then
			redis.call("HSET", tableName, fieldLastUpdate, newUpdate)
			redis.call("HSET", tableName, fieldAppealCount, 0)
		end
		return 1
	`
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)
	ok, err := redis.Int(conn.Do("EVAL", script, 5, tableName, fieldPlayerStatusLastUpdate, fieldPlayerStatusAppealCount, old.LastUpdate, now.Unix()))
	if nil != err {
		return nil, makeError("PlayerDB resetPlayerStatus acid [%s], EVAL redis failed, %v", acid, err)
	}

	if redisFail == ok {
		return old, nil
	}

	res, err := redis.Values(conn.Do("HGETALL", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerDB getPlayerStatus acid [%s], HGETALL redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return &PlayerStatus{}, nil
	}
	ret := &PlayerStatus{}
	if err := redis.ScanStruct(res, ret); nil != err {
		return nil, makeError("PlayerDB getPlayerStatus acid [%s], ScanStruct failed, %v", acid, err)
	}
	return ret, nil
}

func (db *PlayerDB) pushPlayerStatusAutoAppeal(acid string, now time.Time, incr int) (uint32, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return 0, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	ret, err := redis.Int(conn.Do("HINCRBY", tableName, fieldPlayerStatusAppealCount, incr))
	if nil != err && redis.ErrNil != err {
		return 0, makeError("PlayerDB getPlayerStatus acid [%s], HINCRBY redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return 0, nil
	}

	if _, err := conn.Do("HSET", tableName, fieldPlayerStatusLastUpdate, now.Unix()); nil != err {
		return uint32(ret), makeError("PlayerDB getPlayerStatus acid [%s], HSET redis failed, %v", acid, err)
	}

	return uint32(ret), nil
}

func (db *PlayerDB) setPlayerStatusVIP(acid string, vip uint32) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	if _, err := conn.Do("HSET", tableName, fieldPlayerStatusVIP, vip); nil != err {
		return makeError("PlayerDB setPlayerStatusVIP acid [%s], HSET redis failed, %v", acid, err)
	}

	return nil
}

func (db *PlayerDB) getPlayerStatusVIP(acid string) (uint32, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return 0, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	ret, err := redis.Int(conn.Do("HGET", tableName, fieldPlayerStatusVIP))
	if nil != err && redis.ErrNil != err {
		return 0, makeError("PlayerDB getPlayerStatusVIP acid [%s], HGET redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return 0, nil
	}
	return uint32(ret), nil
}

func (db *PlayerDB) getPlayerStatusAppealCount(acid string) (uint32, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return 0, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	ret, err := redis.Int(conn.Do("HGET", tableName, fieldPlayerStatusAppealCount))
	if nil != err && redis.ErrNil != err {
		return 0, makeError("PlayerDB getPlayerStatusAppealCount acid [%s], HGET redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return 0, nil
	}
	return uint32(ret), nil
}

func (db *PlayerDB) setPlayerStatusAutoAcceptBottom(acid string, bottom []uint32) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	bs, err := json.Marshal(bottom)
	if nil != err {
		return makeError("PlayerDB setPlayerStatusAutoAcceptBottom acid [%s], Marshal failed, %v", acid, err)
	}

	if _, err := conn.Do("HSET", tableName, fieldPlayerStatusAutoAcceptBottom, string(bs)); nil != err {
		return makeError("PlayerDB setPlayerStatusAutoAcceptBottom acid [%s], HSET redis failed, %v", acid, err)
	}

	return nil
}

func (db *PlayerDB) getPlayerStatusAutoAcceptBottom(acid string) ([]uint32, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return []uint32{}, makeError("getDBConn Err")
	}
	tableName := db.tablePlayerName(acid, suffixPlayerStatus)

	bs, err := redis.String(conn.Do("HGET", tableName, fieldPlayerStatusAutoAcceptBottom))
	if nil != err && redis.ErrNil != err {
		return []uint32{}, makeError("PlayerDB getPlayerStatusAutoAcceptBottom acid [%s], HGET redis failed, %v", acid, err)
	}
	if redis.ErrNil == err {
		return []uint32{}, nil
	}

	ret := []uint32{}
	if err := json.Unmarshal([]byte(bs), &ret); nil != err {
		return []uint32{}, makeError("PlayerDB getPlayerStatusAutoAcceptBottom acid [%s], Unmarshal failed, %v", acid, err)
	}

	return ret, nil
}
