package csrob

import (
	"encoding/json"
	"fmt"

	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	maxLoadCarList = 100
)

const (
	tableGuildPrefix = "guild"

	suffixGuildList  = "list"
	suffixGuildInfo  = "info"
	suffixGuildEnemy = "enemy"
	suffixGuildTeam  = "team"
	suffixGuildCar   = "car"

	suffixGuildRobCount = "robs"
	suffixGuildRobTime  = "last"
)

func (db *GuildDB) tableGuildStatus() string {
	return fmt.Sprintf("csrob:%d:%s:%s%d", db.groupID, "common", "status", db.sid)
}
func (db *GuildDB) tableGuildRobRank(batch string) string {
	return fmt.Sprintf("csrob:%d:%s:%s:%s", db.groupID, "common", "robrank", batch)
}
func (db *GuildDB) tableGuildRobTimes(batch string) string {
	return fmt.Sprintf("csrob:%d:%s:%s:%s", db.groupID, "common", "robtimes", batch)
}
func (db *GuildDB) tableGuildName(guid string, suffix string) string {
	return fmt.Sprintf("csrob:%d:%s:%s:%s", db.groupID, tableGuildPrefix, guid, suffix)
}
func (db *GuildDB) fieldGuildRobTimes(guildID, suffixGuildRobCount string) string {
	return fmt.Sprintf("%s:%s", suffixGuildRobCount, guildID)
}

//GuildDB ..
type GuildDB struct {
	groupID uint32
	sid     uint32
}

func initGuildDB(res *resources) *GuildDB {
	return &GuildDB{
		groupID: res.groupID,
		sid:     uint32(res.sid),
	}
}

//-- guild:guid:info

func (db *GuildDB) getInfo(guid string) (*GuildInfo, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildInfo)
	bs, err := redis.Bytes(conn.Do("HGET", tableName, suffixGuildInfo))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getInfo guid [%s], HGET redis failed, %v", guid, err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	info := &GuildInfo{}
	err = json.Unmarshal(bs, info)
	if nil != err {
		return nil, makeError("GuildDB getInfo guid [%s], Unmarshal failed, %v, ...{%v}", guid, err, string(bs))
	}

	return info, nil
}

func (db *GuildDB) setInfo(info *GuildInfo) error {
	bs, err := json.Marshal(info)
	if nil != err {
		return makeError("GuildDB setInfo guid [%s], Marshal failed, %v, ...{%v}", info.GuildID, err, info)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(info.GuildID, suffixGuildInfo)
	_, err = conn.Do("HSET", tableName, suffixGuildInfo, bs)
	if nil != err {
		return makeError("GuildDB setInfo guid [%s], SET redis failed, %v", info.GuildID, err)
	}

	return nil
}

func (db *GuildDB) incrRobTimes(guid string, incr int, now int64, batch string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildRobTimes(batch)
	fieldRobCount := db.fieldGuildRobTimes(guid, suffixGuildRobCount)
	fieldRobTime := db.fieldGuildRobTimes(guid, suffixGuildRobTime)
	_, err := conn.Do("HINCRBY", tableName, fieldRobCount, incr)
	if nil != err {
		return makeError("GuildDB incrRobTimes guid [%s], HINCRBY redis failed, %v", guid, err)
	}

	_, err = conn.Do("HSET", tableName, fieldRobTime, now)
	if nil != err {
		return makeError("GuildDB incrRobTimes guid [%s], HSET redis failed, %v", guid, err)
	}

	return nil
}

func (db *GuildDB) getRobTimes(guid string, batch string) (uint32, int64, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return 0, 0, makeError("getDBConn Err")
	}

	tableName := db.tableGuildRobTimes(batch)
	fieldRobCount := db.fieldGuildRobTimes(guid, suffixGuildRobCount)
	fieldRobTime := db.fieldGuildRobTimes(guid, suffixGuildRobTime)
	res, err := redis.Values(conn.Do("HMGET", tableName, fieldRobCount, fieldRobTime))
	if nil != err {
		return 0, 0, makeError("GuildDB getRobTimes guid [%s], HMGET redis failed, %v", guid, err)
	}
	if 2 != len(res) {
		return 0, 0, nil
	}

	count, err := redis.Int64(res[0], nil)
	if nil != err && redis.ErrNil != err {
		return 0, 0, makeError("GuildDB getRobTimes guid [%s], parse redis failed, %v", guid, err)
	}

	robTime, err := redis.Int64(res[1], nil)
	if nil != err && redis.ErrNil != err {
		return 0, 0, makeError("GuildDB getRobTimes guid [%s], parse redis failed, %v", guid, err)
	}

	return uint32(count), robTime, nil
}

// func (db *GuildDB) clearRobTimes(guid string) error {
// 	conn := getDBConn()
// 	defer conn.Close()
// 	if conn.IsNil() {
// 		return makeError("getDBConn Err")
// 	}

// 	tableName := db.tableGuildName(guid, suffixGuildInfo)
// 	ret, err := redis.Int(conn.Do("HDEL", tableName, suffixGuildRobCount, suffixGuildRobTime))
// 	if nil != err && redis.ErrNil != err {
// 		return makeError("GuildDB clearRobTimes guid [%s], HDEL redis failed, %v", guid, err)
// 	}

// 	if redis.ErrNil == err || 0 == ret {
// 		logs.Warn("[CSRob] GuildDB clearRobTimes, but guild [%s] have no rob times", guid)
// 		return nil
// 	}

// 	return nil
// }

//-- guild:guid:enemy

func (db *GuildDB) getEnemy(guid string) ([]GuildEnemy, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildEnemy)
	res, err := redis.Int64Map(conn.Do("HGETALL", tableName))
	if nil != err {
		return nil, makeError("GuildDB getEnemies guid [%s], HGETALL redis failed, %v", guid, err)
	}

	enemies := make([]GuildEnemy, 0, len(res)/2)
	for eid, c := range res {
		enemy := GuildEnemy{
			GuildID: eid,
			Count:   uint32(c),
		}

		enemies = append(enemies, enemy)
	}

	return enemies, nil
}

func (db *GuildDB) pushEnemy(guid string, enemy string, inc int) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildEnemy)
	_, err := conn.Do("HINCRBY", tableName, enemy, inc)
	if nil != err {
		return makeError("GuildDB pushEnemy guid [%s], HINCRBY redis failed, %v", guid, err)
	}

	return nil
}

func (db *GuildDB) removeEnemy(guid string, enemy string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildEnemy)
	ret, err := redis.Int(conn.Do("HDEL", tableName, enemy))
	if nil != err && redis.ErrNil != err {
		return makeError("GuildDB removeEnemy guid [%s], HDEL redis failed, %v", guid, err)
	}
	if redis.ErrNil == err {
		logs.Warn("[CSRob] GuildDB removeEnemy, guild [%s] remove enemy, but it's enemy list is not exist", guid)
		return nil
	}

	if 0 == ret {
		logs.Warn("[CSRob] GuildDB removeEnemy, guild [%s] remove enemy [%s], but it is not exist", guid, enemy)
	}

	return nil
}

//-- guild:guid:team

func (db *GuildDB) getTeams(guid string, nat uint32) ([]GuildTeam, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildTeam) + fmt.Sprintf(":%d", nat)
	res, err := redis.Strings(conn.Do("HVALS", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getTeams guid [%s], HVALS redis failed, %v", guid, err)
	}
	if redis.ErrNil == err {
		return []GuildTeam{}, nil
	}

	enemies := make([]GuildTeam, 0, len(res))
	for _, bs := range res {
		enemy := GuildTeam{}
		err := json.Unmarshal([]byte(bs), &enemy)
		if nil != err {
			logs.Warn("[CSRob] GuildDB getTeams guid [%s], Unmarshal failed, %v, ...{%v}", guid, err, bs)
			continue
		}

		enemies = append(enemies, enemy)
	}

	return enemies, nil
}

func (db *GuildDB) getTeam(guid string, acid string, nat uint32) (*GuildTeam, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildTeam) + fmt.Sprintf(":%d", nat)
	res, err := redis.String(conn.Do("HGET", tableName, acid))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getTeam guid [%s], HGET redis failed, %v", guid, err)
	}
	if redis.ErrNil == err {
		return nil, nil
	}

	team := &GuildTeam{}
	if err := json.Unmarshal([]byte(res), team); nil != err {
		return nil, makeError("GuildDB getTeam guid [%s], Unmarshal failed, %v, ...{%v}", guid, err, res)
	}
	return team, nil
}

func (db *GuildDB) pushTeam(guid string, team *GuildTeam, nat uint32) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	bs, err := json.Marshal(team)
	if nil != err {
		return makeError("GuildDB pushTeam guid [%s], Marshal failed, %v, ...{%v}", guid, err, team)
	}

	tableName := db.tableGuildName(guid, suffixGuildTeam) + fmt.Sprintf(":%d", nat)
	_, err = conn.Do("HSET", tableName, team.Acid, bs)
	if nil != err {
		return makeError("GuildDB pushTeam guid [%s], HSET redis failed, %v", guid, err)
	}

	return nil
}

func (db *GuildDB) removeTeam(guid string, acid string, nat uint32) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildTeam) + fmt.Sprintf(":%d", nat)
	_, err := conn.Do("HDEL", tableName, acid)
	if nil != err {
		return makeError("GuildDB pushTeam guid [%s], HSET redis failed, %v", guid, err)
	}

	return nil
}

//- guild:guid:car

func (db *GuildDB) getCars(guid string, nat uint32) ([]GuildRobElem, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	now := time.Now().Unix()
	tableName := db.tableGuildName(guid, suffixGuildCar) + fmt.Sprintf(":%d", nat)
	res, err := redis.Strings(conn.Do("ZRANGEBYSCORE", tableName, now, "+inf", "LIMIT", 0, maxLoadCarList))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getCars guid [%s], LRANGE redis failed, %v", guid, err)
	}
	if redis.ErrNil == err {
		return []GuildRobElem{}, nil
	}

	list := make([]GuildRobElem, 0, len(res))
	for _, bs := range res {
		info := GuildRobElem{}
		err := json.Unmarshal([]byte(bs), &info)
		if nil != err {
			logs.Warn("[CSRob] GuildDB getCars guid [%s], Unmarshal failed, %v, ...{%v}", guid, err, bs)
			continue
		}

		list = append(list, info)
	}

	return list, nil
}

func (db *GuildDB) pushCar(guid string, nat uint32, elem GuildRobElem) error {
	bs, err := json.Marshal(elem)
	if nil != err {
		return makeError("GuildDB pushCars guid [%s], Marshal failed, %v, ...{%v}", guid, err, elem)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildCar) + fmt.Sprintf(":%d", nat)
	_, err = conn.Do("ZADD", tableName, elem.EndStamp, string(bs))
	if nil != err {
		return makeError("GuildDB pushCars guid [%s], RPUSH redis failed, %v", guid, err)
	}

	return nil
}

func (db *GuildDB) removeCar(guid string, nat uint32, edge int64) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName(guid, suffixGuildCar) + fmt.Sprintf(":%d", nat)
	if _, err := conn.Do("ZREMRANGEBYSCORE", tableName, 0, edge); nil != err && redis.ErrNil != err {
		return makeError("GuildDB removeCar guid [%s]:[%d], ZREMRANGEBYSCORE redis failed, %v", guid, nat, err)
	}

	return nil
}

//- guild:guid:list

func (db *GuildDB) pushGuildToList(guid string, gs int64) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName("recommend", suffixGuildList)
	_, err := conn.Do("ZADD", tableName, gs, guid)
	if nil != err {
		return makeError("GuildDB pushGuildToList guid [%s], ZADD redis failed, %v", guid, err)
	}

	return nil
}

func (db *GuildDB) getGuildFromList(num int) ([]string, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildName("recommend", suffixGuildList)
	list, err := redis.Strings(conn.Do("ZREVRANGE", tableName, 0, num-1))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getGuildFromList, ZREVRANGE redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return []string{}, nil
	}

	return list, nil
}

func (db *GuildDB) removeGuildFromList(guid string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildName("recommend", suffixGuildList)
	ret, err := redis.Int(conn.Do("ZREM", tableName, guid))
	if nil != err && redis.ErrNil != err {
		return makeError("GuildDB removeGuildFromList, ZREM redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil
	}

	if 0 == ret {
		logs.Warn("[CSRob] GuildDB removeGuildFromList, remove guild [%s] but it is not exist", guid)
	}

	return nil
}

//- guild:common:status
func (db *GuildDB) getCommonStatus() (*GuildCommonStatus, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err")
	}

	tableName := db.tableGuildStatus()
	bs, err := redis.Bytes(conn.Do("GET", tableName))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB getGuildFromList, GET redis failed, %v", err)
	}

	if redis.ErrNil == err {
		return &GuildCommonStatus{}, nil
	}

	status := GuildCommonStatus{}
	err = json.Unmarshal([]byte(bs), &status)
	if nil != err {
		return nil, makeError("GuildDB getCommonStatus, Unmarshal failed, %v, ...{%v}", err, bs)
	}

	return &status, nil
}

func (db *GuildDB) setCommonStatus(status *GuildCommonStatus) error {
	logs.Debug("[CSRob] GuildDB setCommonStatus, %v", status)
	bs, err := json.Marshal(status)
	if nil != err {
		return makeError("GuildDB getCommonStatus, Unmarshal failed, %v, ...{%v}", err, bs)
	}

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildStatus()
	_, err = conn.Do("SET", tableName, bs)
	if nil != err {
		return makeError("GuildDB setCommonStatus, SET redis failed, %v", err)
	}

	return nil
}

//- guild:recommend:list
func (db *GuildDB) loadAllGuildIDs() ([]string, error) {
	tableName := db.tableGuildName("recommend", suffixGuildList)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return []string{}, makeError("getDBConn Err")
	}

	len, err := redis.Int(conn.Do("ZCOUNT", tableName, "-inf", "inf"))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("GuildDB loadAllGuildIDs, ZCOUNT redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return []string{}, nil
	}

	retList := []string{}
	patchNum := 50
	for i := 0; i < len; i += patchNum {
		subList, err := redis.Strings(conn.Do("ZREVRANGE", tableName, i, patchNum))
		if nil != err {
			return nil, makeError("GuildDB loadAllGuildIDs, ZREVRANGE redis failed, %v", err)
		}
		retList = append(retList, subList...)
	}

	return nil, nil
}

//- guild:common:robrank
func (db *GuildDB) pushGuildToRobRank(guid string, robs uint32, robtime int64, batch string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	score := db.combineRobScore(robs, robtime)

	tableName := db.tableGuildRobRank(batch)
	_, err := conn.Do("ZADD", tableName, score, guid)
	if nil != err {
		return makeError("GuildDB pushGuildToRobRank guid [%s], ZADD redis failed, %v", guid, err)
	}
	return nil
}

func (db *GuildDB) removeGuildFromRobRank(guid string, batch string) error {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err")
	}

	tableName := db.tableGuildRobRank(batch)
	ret, err := redis.Int(conn.Do("ZREM", tableName, guid))
	if nil != err && redis.ErrNil != err {
		return makeError("GuildDB removeGuildFromRobRank, ZREM redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return nil
	}

	if 0 == ret {
		logs.Warn("[CSRob] GuildDB removeGuildFromRobRank, remove guild [%s] but it is not exist", guid)
	}

	return nil
}

func (db *GuildDB) rangeFromRobRank(num int, batch string) ([]GuildRankElem, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return []GuildRankElem{}, makeError("getDBConn Err")
	}

	tableName := db.tableGuildRobRank(batch)
	res, err := redis.Values(conn.Do("ZREVRANGE", tableName, 0, num, "withscores"))
	if nil != err && redis.ErrNil != err {
		return []GuildRankElem{}, makeError("GuildDB rangeFromRobRank, ZREVRANGE redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return []GuildRankElem{}, nil
	}

	retList := make([]GuildRankElem, 0, len(res)/2)
	for i := 0; i < len(res)-1; i += 2 {
		guid, err := redis.String(res[i], nil)
		if nil != err {
			logs.Error(fmt.Sprint(makeError("GuildDB rangeFromRobRank, parse redis failed, %v", err)))
			continue
		}
		score, err := redis.Float64(res[i+1], nil)
		if nil != err {
			logs.Error(fmt.Sprint(makeError("GuildDB rangeFromRobRank, parse redis failed, %v", err)))
			continue
		}

		if "" == guid {
			continue
		}
		count, robTime := db.parseRobScore(score)
		retList = append(retList,
			GuildRankElem{
				GuildID:  guid,
				RobCount: count,
				RobTime:  robTime,
				Rank:     (uint32(i) / 2) + 1,
			},
		)
	}

	return retList, nil
}

func (db *GuildDB) getRankFromRobRank(guid string, batch string) (uint32, error) {
	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return 0, makeError("getDBConn Err")
	}

	tableName := db.tableGuildRobRank(batch)
	ret, err := redis.Int(conn.Do("ZREVRANK", tableName, guid))
	if nil != err && redis.ErrNil != err {
		return 0, makeError("GuildDB getRankFromRobRank, ZREVRANK redis failed, %v", err)
	}
	if redis.ErrNil == err {
		return 0, nil
	}

	return uint32(ret + 1), nil
}

// func (db *GuildDB) clearRankRob(batch string) error {
// 	conn := getDBConn()
// 	defer conn.Close()
// 	if conn.IsNil() {
// 		return makeError("getDBConn Err")
// 	}

// 	tableName := db.tableGuildRobRank()
// 	now := time.Now()
// 	bakName := tableName + fmt.Sprintf(":%4d%02d%02d", now.Year(), now.Month(), now.Day())

// 	exist, err := redis.Int(conn.Do("EXISTS", tableName))
// 	if nil != err {
// 		return makeError("GuildDB clearRankRob, EXISTS redis failed, %v", err)
// 	}
// 	if 0 == exist {
// 		logs.Warn("[CSRob] clearRankRob, but redis key is not exist")
// 		return nil
// 	}

// 	ret, err := redis.String(conn.Do("RENAME", tableName, bakName))
// 	if nil != err && redis.ErrNil != err {
// 		return makeError("GuildDB clearRankRob, RENAME redis failed, %v", err)
// 	}
// 	if redis.ErrNil == err {
// 		logs.Warn("[CSRob] clearRankRob, but redis key is not exist")
// 		return nil
// 	}

// 	if "OK" != ret {
// 		logs.Warn("[CSRob] clearRankRob, but redis return is [%s]", ret)
// 		return nil
// 	}

// 	return nil
// }

const baseRobScore = 100000

func (db *GuildDB) combineRobScore(robs uint32, robtime int64) float64 {
	return float64(robs)*baseRobScore + baseRobScore - float64(robtime)/baseRobScore
}

func (db *GuildDB) parseRobScore(score float64) (uint32, int64) {
	count := score / baseRobScore
	t := (baseRobScore - score - (baseRobScore * count)) * baseRobScore
	return uint32(count), int64(t)
}
