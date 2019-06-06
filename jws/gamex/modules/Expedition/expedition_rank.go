package Expedition

import (
	"strings"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

/*
	1、1~9从低往高找
	2、每个根据策划配置的点找最接近的，第一个找1个，第二个找2个......第九个找9个
	3、如果第一个没找到，填补机器人，第二个里如果只有1个，并和第一个里的重复，则填补机器人.......
	4、开始拉存档
*/

func GetExpeditionEnemyId(acid string, sid uint, gs int64) (error, []string) {

	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetAccountByGsRange Err conn is nil")
		return fmt.Errorf("GetExpeditionEnemyId modules.GetDBSlaveConn nil"), nil
	}

	logs.Debug("GetExpeditionEnemyId %d", gs)

	cfgs := gamedata.GetExpeditionLvlCfgs()
	_acids := make(map[string]struct{}, len(cfgs))
	_acids[acid] = struct{}{}
	res := make([]string, len(cfgs))
	db_name := rank.TableRankCorpGs(sid)
	for i, cfg := range cfgs {
		gsn := int64(float64(gs)*float64(cfg.GetGS())) *
			rank.RankByCorpDelayPowBase
		m, err := redis.StringMap(_do(db_name, conn, "ZREVRANGEBYSCORE", db_name, gsn, "-inf",
			"WITHSCORES", "limit", "1", cfg.GetLevelID()))
		if err != nil {
			logs.Error("GetExpeditionEnemyId err %v", err)
			return err, nil
		}

		logs.Debug("GetExpeditionEnemyId ZREVRANGEBYSCORE %d %v", i, m)

		for oacid, _ := range m {
			if _, ok := _acids[oacid]; ok {
				continue
			}
			res[i] = oacid
			_acids[oacid] = struct{}{}
			break
		}
	}

	logs.Debug("GetExpeditionEnemyId res %v", res)
	return nil, res
}

func _do(rank_name string, db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	ss := strings.SplitN(rank_name, ":", 2)
	key := rank_name
	if len(ss) > 1 {
		key = ss[1]
	}
	return metricsModules.DoWraper("expedition_db_"+key, db, commandName, args...)
}
