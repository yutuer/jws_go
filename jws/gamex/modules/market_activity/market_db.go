package market_activity

import (
	"fmt"

	"encoding/json"
	"errors"
	"strconv"

	time "time"

	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genTableName(sid uint, actType uint32) (string, error) {
	table_name, ok := tablename[actType]
	if false == ok {
		return "", errors.New(fmt.Sprintf("unkown activity id [%d]", actType))
	}
	return fmt.Sprintf("%d:%d:%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid), table_name), nil
}

type MarketRankDB struct {
	sid uint
}

func (mr *MarketRankDB) genTableName(actType uint32) (string, error) {
	return genTableName(mr.sid, actType)
}

func (mr *MarketRankDB) setSnapShoot(actType uint32, acid2score map[string]float64) error {
	//组织写入数据
	packet_index := 0
	packet_len := Redis_ZADD_Banch * 2
	packet_inner := 0
	packets := [][]interface{}{}
	for k, v := range acid2score {
		if 0 == packet_inner%packet_len {
			pack := []interface{}{}
			packets = append(packets, pack)
			packet_index += 1
		}

		packets[packet_index-1] = append(packets[packet_index-1], fmt.Sprintf("%f", v), k)

		packet_inner++
	}

	//写入redis
	table_name, err := mr.genTableName(actType)
	if nil != err {
		return err
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Err For activity %d", actType)
	}
	reply, err := redis.Int(conn.Do("exists", table_name))
	if err != nil {
		logs.Error("Check Exists redis Table err by %v, table_name: %s", err, table_name)
	} else {
		if reply == 1 {
			nowT := time.Now()
			_, err = conn.Do("rename", table_name, fmt.Sprintf("%s:%d%d%d", table_name, nowT.Year(), nowT.Month(), nowT.Day()))
			if err != nil {
				logs.Error("Rename redis Table err by %v, table_name: %s", err, table_name)
			}
		}
	}

	for i := 0; i < packet_index; i++ {
		_, err := conn.Do("ZADD", append([]interface{}{table_name}, packets[i]...)...)
		if nil != err {
			return fmt.Errorf("ZADD Redis Err For activity %d, err [%v], packet %v", actType, packets[i])
		}
	}

	return nil
}

func (mr *MarketRankDB) getSnapShoot(activity uint32) (map[string]float64, error) {
	//查询redis
	table_name, err := mr.genTableName(activity)
	if nil != err {
		return nil, err
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Err For activity %d", activity)
	}

	res, err := redis.Strings(conn.Do("ZRANGE", table_name, 0, -1, "WITHSCORES"))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("redis do range, %v", err)
	}

	if redis.ErrNil == err {
		return map[string]float64{}, nil
	}

	list := map[string]float64{}
	for i := 0; i < len(res); i += 2 {
		score, err := strconv.ParseFloat(res[i+1], 64)
		if nil != err {
			logs.Warn("[MarketActivityModule] getSnapShoot, Parse Redis Err For activity %d, %v", activity, err)
			continue
		}
		list[res[i]] = score
	}

	return list, nil
}

func (mr *MarketRankDB) setTopN(activity uint32, actID uint32, data *RankTopN) error {
	//制作下刷数据
	bs, err := json.Marshal(data)
	if nil != err {
		return fmt.Errorf("Marshal RankTopN Err For activity %d, %v", activity, err)
	}

	//写入redis
	table_name, err := mr.genTableName(activity)
	if nil != err {
		return err
	}
	table_name = table_name + fmt.Sprintf(":%d", actID) + ":topN"
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Err For activity %d", activity)
	}

	_, err = conn.Do("SET", table_name, bs)
	if nil != err {
		return fmt.Errorf("SET Redis RankTopN Err For activity %d, %v", activity, err)
	}

	return nil
}

func (mr *MarketRankDB) getTopN(activity uint32, actID uint32) (*RankTopN, error) {
	//写入redis
	table_name, err := mr.genTableName(activity)
	if nil != err {
		return nil, err
	}
	table_name = table_name + fmt.Sprintf(":%d", actID) + ":topN"
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Err For activity %d", activity)
	}

	bs, err := redis.Bytes(conn.Do("GET", table_name))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("GET Redis RankTopN Err For activity %d, %v", activity, err)
	}

	//如果为空
	if redis.ErrNil == err {
		return nil, nil
	}

	//解析数据
	data := &RankTopN{}
	err = json.Unmarshal(bs, data)
	if nil != err {
		return nil, fmt.Errorf("Unmarshal RankTopN Err For activity %d, %v", activity, err)
	}

	return data, nil
}

func (mr *MarketRankDB) getTopRange(activity uint32, num uint32) ([]pair, error) {
	//访问redis
	table_name, err := mr.genTableName(activity)
	if nil != err {
		return nil, err
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return nil, fmt.Errorf("GetDBConn Err For activity %d", activity)
	}

	res, err := redis.Strings(conn.Do("ZREVRANGE", table_name, 0, num, "WITHSCORES"))
	if nil != err && redis.ErrNil != err {
		return nil, fmt.Errorf("ZREVRANGE Redis getTopRange Err For activity %d, %v", activity, err)
	}

	if redis.ErrNil == err {
		return []pair{}, nil
	}

	list := []pair{}
	for i := 0; i < len(res); i += 2 {
		score, err := strconv.ParseFloat(res[i+1], 64)
		if nil != err {
			return nil, fmt.Errorf("Parse Redis getTopRange Err For activity %d, %v", activity, err)
		}
		list = append(
			list,
			pair{
				Acid:  res[i],
				Score: score,
			},
		)
	}

	return list, nil
}

func (mr *MarketRankDB) getPosAndRedisScore(activity uint32, id string) (int, float64) {
	//访问redis
	table_name, err := mr.genTableName(activity)
	if nil != err {
		return 0, 0
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err For activity %d", activity)
		return 0, 0
	}

	rank, err := redis.Int(conn.Do("ZREVRANK", table_name, id))
	if nil != err && redis.ErrNil != err {
		logs.Error("ZREVRANK Redis getPosAndRedisScore Err For activity %d, %v", activity, err)
		return 0, 0
	}

	if redis.ErrNil == err {
		return 0, 0
	}

	score, err := redis.Float64(conn.Do("ZSCORE", table_name, id))
	if nil != err && redis.ErrNil != err {
		logs.Error("ZSCORE Redis getPosAndRedisScore Err For activity %d, %v", activity, err)
		return 0, 0
	}

	if redis.ErrNil == err {
		logs.Warn("[MarketActivityModule] getPosAndRedisScore, got rank but can't get score")
		return rank + 1, 0
	}

	return rank + 1, score
}

func (mr *MarketRankDB) genRecordTableName() string {
	return fmt.Sprintf("%d:%d:%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(mr.sid), "SnapShootRecord")
}

func (mr *MarketRankDB) getRankRecord() *MarketRankRecord {
	//访问redis
	table_name := mr.genRecordTableName()
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err For getRankRecord")
		return nil
	}

	bs, err := redis.Bytes(conn.Do("GET", table_name))
	if nil != err && redis.ErrNil != err {
		logs.Error("GET Redis getRankRecord Err %v", err)
		return nil
	}

	//如果为空
	if redis.ErrNil == err {
		return nil
	}

	//解析数据
	data := &MarketRankRecord{}
	err = json.Unmarshal(bs, data)
	if nil != err {
		logs.Error("Unmarshal getRankRecord Err %v", err)
		return nil
	}

	if nil == data.RankBatch {
		data.RankBatch = map[string]uint32{}
	}
	//兼容性修改, by qiaozhu @20170516
	if _, exist := data.RankBatch[fmt.Sprintf("%d", Activity_Rank)]; false == exist {
		data.RankBatch[fmt.Sprintf("%d", Activity_Rank)] = data.RankParentID
	}

	return data
}

func (mr *MarketRankDB) setRankRecord(record *MarketRankRecord) error {
	//制作下刷数据
	bs, err := json.Marshal(record)
	if nil != err {
		return fmt.Errorf("Marshal setRankRecord Err %v", err)
	}

	//写入redis
	table_name := mr.genRecordTableName()
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("GetDBConn Err For setRankRecord")
	}

	_, err = conn.Do("SET", table_name, bs)
	if nil != err {
		return fmt.Errorf("SET Redis setRankRecord Err, %v", err)
	}

	return nil
}

func (mr *MarketRankDB) debugClearSnap(actType uint32, actID uint32) {
	table_name, err := mr.genTableName(actType)
	if nil != err {
		logs.Error("[MarketActivityModule] debugClearSnap failed when genTableName, [%d:%d], %v", actType, actID, err)
		return
	}
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("[MarketActivityModule] debugClearSnap failed when GetDBConn, [%d:%d]", actType, actID)
		return
	}

	_, err = conn.Do("DEL", table_name)
	if nil != err {
		logs.Error("[MarketActivityModule] debugClearSnap failed when DEL SnapShoot, [%d:%d], %v", actType, actID, err)
		return
	}

	topN_table_name := table_name + fmt.Sprintf(":%d", actID) + ":topN"

	_, err = conn.Do("DEL", topN_table_name)
	if nil != err {
		logs.Error("[MarketActivityModule] debugClearSnap failed when DEL TopN, [%d:%d], %v", actType, actID, err)
		return
	}
}
