package rank

import (
	"errors"
	"strconv"

	"strings"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

type rankDB struct {
}

func (r *rankDB) add(rank_name, db_name string, id string, score int64) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s %d",
			rank_name, id, score)
		return
	}
	rank_name_in_db := rank_name
	_, err := _do(rank_name, conn, "ZADD", rank_name_in_db, score, id)
	if err != nil {
		logs.Error("Do Err %s by %s For %s %d",
			err.Error(), rank_name, id, score)
	}
}

// score 是 doulbe 类型的
func (r *rankDB) addWithDoubleScore(rank_name, db_name string, id string, score float64) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s %d",
			rank_name, id, score)
		return
	}
	rank_name_in_db := rank_name
	_, err := _do(rank_name, conn, "ZADD", rank_name_in_db, score, id)
	if err != nil {
		logs.Error("Do Err %s by %s For %s %d",
			err.Error(), rank_name, id, score)
	}
}

func (r *rankDB) del(rank_name, db_name string, id string) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s",
			rank_name, id)
		return
	}
	rank_name_in_db := rank_name
	_, err := _do(rank_name, conn, "ZREM", rank_name_in_db, id)
	if err != nil {
		logs.Error("Do Err %s by %s For %s",
			err.Error(), rank_name, id)
	}
}

func (r *rankDB) adds(rank_name, db_name string, data []interface{}) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %v",
			rank_name, data)
		return
	}
	rank_name_in_db := rank_name
	data[0] = rank_name_in_db
	_, err := _do(rank_name, conn, "ZADD", data...)
	if err != nil {
		logs.Error("Do Err %s by %s For %v",
			err.Error(), rank_name, data)
	}
}

func (r *rankDB) getTopNplusOne(rank_name, db_name string, id string) (string, int64) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s",
			rank_name, id)
		return "", 0
	}
	rank_name_in_db := rank_name
	res, err := redis.Strings(_do(rank_name, conn,
		"ZREVRANGE",
		rank_name_in_db,
		RankBalanceSize,
		RankBalanceSize,
		"WithScores"))
	if err != nil {
		logs.Error("Do Err %s by %s For %s",
			err.Error(), rank_name, id)
		return "", 0
	}

	if len(res) < 2 {
		logs.Error("Do Err Res len by %s For %s In %v",
			rank_name, id, res)
		return "", 0
	}

	logs.Trace("getPos res %v", res)
	res_score, err := strconv.ParseInt(res[1], 10, 64)
	if err != nil {
		logs.Error("Do Err %s by %s For %s In %v",
			err.Error(), rank_name, id, res)
		return "", 0
	}
	return res[0], res_score
}

func (r *rankDB) getPos(rank_name, db_name string, id string) int {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s",
			rank_name, id)
		return 0
	}
	rank_name_in_db := rank_name
	res, err := redis.Int(_do(rank_name, conn, "ZREVRANK", rank_name_in_db, id))
	if err != nil && err != redis.ErrNil {
		logs.Error("Do Err %s by %s For %s",
			err.Error(), rank_name, id)
		return 0
	}

	if err == redis.ErrNil {
		return 0
	}

	logs.Trace("getPos res %d", res)
	return res + 1
}

func (r *rankDB) getScore(rank_name, db_name string, id string) (int64, error) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s",
			rank_name, id)
		return 0, errors.New("DBConnErr")
	}
	rank_name_in_db := rank_name
	res, err := redis.Int64(_do(rank_name, conn, "ZSCORE", rank_name_in_db, id))
	if err != nil && err != redis.ErrNil {
		logs.Error("Do Err %s by %s For %s",
			err.Error(), rank_name, id)

		return 0, err
	}

	if err == redis.ErrNil {
		return 0, nil // 空值
	}

	logs.Trace("ZSCORE res %d", res)
	return res, nil
}

func (r *rankDB) getFloatScore(rank_name, db_name string, id string) (float64, error) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s For %s",
			rank_name, id)
		return 0, errors.New("DBConnErr")
	}
	rank_name_in_db := rank_name
	res, err := redis.Float64(_do(rank_name, conn, "ZSCORE", rank_name_in_db, id))
	if err != nil && err != redis.ErrNil {
		logs.Error("Do Err %s by %s For %s",
			err.Error(), rank_name, id)

		return 0, err
	}

	if err == redis.ErrNil {
		return 0, nil // 空值
	}

	logs.Trace("ZSCORE res %d", res)
	return res, nil
}

func (r *rankDB) loadTopN(rank_name, db_name string) ([]byte, error) {
	rank_name_in_db := rank_name + ":topN"

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name)
		return []byte{}, errors.New("GetDBConnNil")
	}

	return redis.Bytes(_do(rank_name, conn, "GET", rank_name_in_db))
}

func (r *rankDB) saveTopN(rank_name, db_name string, data []byte) error {
	rank_name_in_db := rank_name + ":topN"

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name)
		return errors.New("GetDBConnNil")
	}

	//logs.Trace("saveTopN %s by %v", rank_name, string(data))
	_, err := _do(rank_name, conn, "SET", rank_name_in_db, data)
	return err
}

func (r *rankDB) reName(rank_name, new_rank_name, db_name string) error {
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name)
		return errors.New("GetDBConnNil")
	}

	_, err := _do(rank_name, conn, "RENAME", rank_name_in_db, new_rank_name)
	return err
}

func (r *rankDB) copy(rank_name, new_rank_name, db_name string) error {
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name_in_db)
		return errors.New("GetDBConnNil")
	}

	ok := driver.RedisSaveDataToOther(conn, rank_name_in_db, new_rank_name)

	if ok {
		return nil
	} else {
		return errors.New("RedisSaveDataToOtherErr")
	}
}

func (r *rankDB) getTopFromRedis(rank_name, db_name string) ([]string, error) {
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name_in_db)
		return []string{}, errors.New("GetDBConnNil")
	}
	return redis.Strings(_do(rank_name, conn, "ZREVRANGE", rank_name_in_db, 0, RankBalanceSize-1))
}

func (r *rankDB) getTopWithScoreFromRedis(rank_name, db_name string) ([]string, []int64, error) {
	return r._getTopWithScoreFromRedis(rank_name, db_name, RankTopSize)
}

func (r *rankDB) getTopWithFloatScoreFromRedis(rank_name, db_name string) ([]string, []float64, error) {
	return r._getTopWithFloatScoreFromRedis(rank_name, db_name, RankTopSize)
}

func (r *rankDB) getWithScoreFromRedis(rank_name, db_name string) ([]string, []int64, error) {
	return r._getTopWithScoreFromRedis(rank_name, db_name, RankBalanceSize)
}

func (r *rankDB) _getTopWithScoreFromRedis(rank_name, db_name string, size int) ([]string, []int64, error) {
	logs.Trace("getTopWithScoreFromRedis %s %s", rank_name, db_name)
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name_in_db)
		return []string{}, []int64{}, errors.New("GetDBConnNil")
	}
	resOld, err := conn.Do("ZREVRANGE", rank_name_in_db, 0, size, "WITHSCORES")
	//logs.Trace("resOld %v", resOld)
	res, err := redis.Strings(resOld, err)
	if err != nil {
		logs.Error("redis.Strings Err by %s", err.Error())
		return []string{}, []int64{}, err
	}
	ids := make([]string, 0, len(res))
	scores := make([]int64, 0, len(res))

	for i := 0; i+1 < len(res); i += 2 {
		ids = append(ids, res[i])
		s, err := strconv.ParseInt(res[i+1], 10, 64)
		if err != nil {
			logs.Error("strconv.Atoi %v Err by %s", res[i+1], err.Error())
			return []string{}, []int64{}, err
		}
		scores = append(scores, s)
	}

	//logs.Warn("getTopWithScoreFromRedis %v", res)
	//logs.Warn("getTopWithScoreFromRedis %v", ids)
	//logs.Warn("getTopWithScoreFromRedis %v", scores)

	return ids, scores, nil
}

// 新score采用的是float64
func (r *rankDB) _getTopWithFloatScoreFromRedis(rank_name, db_name string, size int) ([]string, []float64, error) {
	logs.Trace("getTopWithScoreFromRedis %s %s", rank_name, db_name)
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name_in_db)
		return []string{}, []float64{}, errors.New("GetDBConnNil")
	}
	resOld, err := conn.Do("ZREVRANGE", rank_name_in_db, 0, size, "WITHSCORES")
	//logs.Trace("resOld %v", resOld)
	res, err := redis.Strings(resOld, err)
	if err != nil {
		logs.Error("redis.Strings Err by %s", err.Error())
		return []string{}, []float64{}, err
	}
	ids := make([]string, 0, len(res))
	scores := make([]float64, 0, len(res))

	for i := 0; i+1 < len(res); i += 2 {
		ids = append(ids, res[i])
		s, err := strconv.ParseFloat(res[i+1], 64)
		if err != nil {
			logs.Error("strconv.Atoi %v Err by %s", res[i+1], err.Error())
			return []string{}, []float64{}, err
		}
		scores = append(scores, s)
	}

	return ids, scores, nil
}

func (r *rankDB) GetTopNFromRedis(rank_name, db_name string, n int) ([]string, error) {
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name_in_db)
		return []string{}, errors.New("GetDBConnNil")
	}
	return redis.Strings(_do(rank_name, conn, "ZREVRANGE", rank_name_in_db, 0, n-1))
}

func _do(rank_name string, db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	ss := strings.SplitN(rank_name, ":", 2)
	key := rank_name
	if len(ss) > 1 {
		key = ss[1]
	}
	return metricsModules.DoWraper("rank_db_"+key, db, commandName, args...)
}

func (r *rankDB) delKey(rank_name, db_name string) error {
	rank_name_in_db := rank_name

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s",
			rank_name)
		return errors.New("GetDBConnNil")
	}

	_, err := _do(rank_name, conn, "DEL", rank_name_in_db)
	return err
}

func loadRankAccountInfo(acid string) (name, platformId, deviceToken string) {
	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("loadRankAccountInfo GetDBConn nils")
		return
	}

	t := fmt.Sprintf("profile:%s", acid)
	ss, err := redis.Strings(_db.Do("HMGET", t, "name", "PlatformId", "DeviceToken"))
	if err != nil {
		logs.Error("loadRankAccountInfo HMGET err %s", err.Error())
		return
	}

	if len(ss) < 3 {
		logs.Error("loadRankAccountInfo res len err %d", len(ss))
		return
	}
	return ss[0], ss[1], ss[2]
}
