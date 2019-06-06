package csrob

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//PlayerRankDB ..
type PlayerRankDB struct {
	groupID uint32
}

func initPlayerRankDB(res *resources) *PlayerRankDB {
	return &PlayerRankDB{
		groupID: res.groupID,
	}
}

func (db *PlayerRankDB) tableRankName(nat uint32) string {
	return fmt.Sprintf("csrob:%d:formationrank:%d:rank", db.groupID, nat)
}

func (db *PlayerRankDB) tableFormationName(nat uint32) string {
	return fmt.Sprintf("csrob:%d:formationrank:%d:formation", db.groupID, nat)
}

func (db *PlayerRankDB) pushFormationAndRank(nat uint32, acid string, team *RankTeam, now int64) error {
	bs, err := json.Marshal(team)
	if nil != err {
		return makeError("PlayerRankDB pushFormation acid [%s], Marshal failed, %v, ...{%v}", acid, err, team)
	}

	gs := uint32(0)
	for _, h := range team.Heros {
		gs += uint32(h.Gs)
	}
	rankScore := db.combineRankFormationScore(gs, now)

	tableRankName := db.tableRankName(nat)
	tableFormationName := db.tableFormationName(nat)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err, %v", conn.Err())
	}

	_, err = conn.Do("MULTI")
	if nil != err {
		return makeError("PlayerRankDB pushFormationAndRank acid [%s], MULTI redis failed, %v", acid, err)
	}
	_, err = conn.Do("HSET", tableFormationName, acid, string(bs))
	if nil != err {
		return makeError("PlayerRankDB pushFormationAndRank acid [%s], HSET redis failed, %v", acid, err)
	}
	_, err = conn.Do("ZADD", tableRankName, rankScore, acid)
	if nil != err {
		return makeError("PlayerRankDB pushFormationAndRank acid [%s], ZADD redis failed, %v", acid, err)
	}
	_, err = conn.Do("EXEC")
	if nil != err {
		return makeError("PlayerRankDB pushFormationAndRank acid [%s], EXEC redis failed, %v", acid, err)
	}

	return nil
}

func (db *PlayerRankDB) rangeFormationByRank(nat, num uint32) ([]*RankTeam, error) {
	tableRankName := db.tableRankName(nat)
	tableFormationName := db.tableFormationName(nat)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, makeError("getDBConn Err, %v", conn.Err())
	}

	resIDs, err := redis.Strings(conn.Do("ZREVRANGE", tableRankName, 0, num))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerRankDB rangeFormationByRank, ZREVRANGE redis failed, %v", err)
	}

	if redis.ErrNil == err {
		return []*RankTeam{}, nil
	}

	args := []interface{}{}
	args = append(args, tableFormationName)
	for _, id := range resIDs {
		args = append(args, id)
	}
	resVal, err := redis.Strings(conn.Do("HMGET", args...))
	if nil != err && redis.ErrNil != err {
		return nil, makeError("PlayerRankDB rangeFormationByRank, HMGET redis failed, %v", err)
	}
	if redis.ErrNil == err {
		logs.Warn("[CSRob] PlayerRankDB rangeFormationByRank, got from ranks but can not get from formations")
		return []*RankTeam{}, nil
	}

	list := []*RankTeam{}
	for _, val := range resVal {
		team := &RankTeam{}
		err := json.Unmarshal([]byte(val), team)
		if nil != err {
			logs.Warn("[CSRob] PlayerRankDB rangeFormationByRank, Unmarshal failed, %v, ...{%v}", err, val)
			continue
		}
		list = append(list, team)
	}

	return list, nil
}

func (db *PlayerRankDB) getFormationAndPos(nat uint32, acid string) (*RankTeam, uint32, error) {
	tableRankName := db.tableRankName(nat)
	tableFormationName := db.tableFormationName(nat)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, 0, makeError("getDBConn Err, %v", conn.Err())
	}

	resPos, err := redis.Int(conn.Do("ZREVRANK", tableRankName, acid))
	if nil != err && redis.ErrNil != err {
		return nil, 0, makeError("PlayerRankDB getFormationAndPos, ZREVRANK redis failed, %v", err)
	}

	if redis.ErrNil == err {
		logs.Warn("[CSRob] PlayerRankDB getFormationAndPos, but can not get rank pos")
		return nil, 0, nil
	}

	resPos++

	resVal, err := redis.String(conn.Do("HGET", tableFormationName, acid))
	if nil != err && redis.ErrNil != err {
		return nil, 0, makeError("PlayerRankDB getFormationAndPos, HGET redis failed, %v", err)
	}
	if redis.ErrNil == err {
		logs.Warn("[CSRob] PlayerRankDB getFormationAndPos, got from ranks but can not get from formations")
		return nil, 0, nil
	}

	team := &RankTeam{}
	if err := json.Unmarshal([]byte(resVal), team); nil != err {
		logs.Warn("[CSRob] PlayerRankDB getFormationAndPos, Unmarshal failed, %v, ...{%v}", err, resVal)
		return nil, 0, nil
	}

	return team, uint32(resPos), nil
}

func (db *PlayerRankDB) removeFromFormationAndRank(nat uint32, acid string) error {
	tableRankName := db.tableRankName(nat)
	tableFormationName := db.tableFormationName(nat)

	conn := getDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return makeError("getDBConn Err, %v", conn.Err())
	}

	ret, err := redis.Int(conn.Do("ZREM", tableRankName, acid))
	if nil != err && redis.ErrNil != err {
		return makeError("PlayerRankDB removeFromFormationAndRank, ZREM redis failed, %v", err)
	}

	if redis.ErrNil == err {
		logs.Warn("[CSRob] PlayerRankDB removeFromFormationAndRank, but rank table [%d] is not exist", nat)
	}

	if 0 == ret {
		logs.Warn("[CSRob] PlayerRankDB removeFromFormationAndRank, but rank table [%d] have no player [%s]", nat, acid)
	}

	ret, err = redis.Int(conn.Do("HDEL", tableFormationName, acid))
	if nil != err && redis.ErrNil != err {
		return makeError("PlayerRankDB removeFromFormationAndRank, HDEL redis failed, %v", err)
	}
	if redis.ErrNil == err {
		logs.Warn("[CSRob] PlayerRankDB removeFromFormationAndRank, but formation table [%d] is not exist", nat)
		return nil
	}

	if 0 == ret {
		logs.Warn("[CSRob] PlayerRankDB removeFromFormationAndRank, but formation table [%d] have no player [%s]", nat, acid)
	}

	return nil
}

const baseRankFormationScore = 100000

func (db *PlayerRankDB) combineRankFormationScore(gs uint32, t int64) float64 {
	return float64(gs)*baseRankFormationScore + baseRankFormationScore - float64(t)/baseRankFormationScore
}

func (db *PlayerRankDB) parseRankFormationScore(score float64) (uint32, int64) {
	gs := score / baseRankFormationScore
	t := (baseRankFormationScore - score - (baseRankFormationScore * gs)) * baseRankFormationScore
	return uint32(gs), int64(t)
}
