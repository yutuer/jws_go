package gve_notify

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
)

const Robot_Count = 2

var robotRand *rand.Rand

func initRand() {
	var rng util.Kiss64Rng
	rng.Seed(time.Now().Unix())
	robotRand = rand.New(&rng)
}

// 找到两个战力相符离线队友或机器人
func GenGVERobotCompanion(sid uint, gs int64, acID string) ([]*helper.Avatar2ClientByJson, error) {
	// 从战力榜中拉取队友信息
	db_name := rank.TableRankCorpGs(sid)
	dur := Robot_Count + 1
	minGS, maxGS := gamedata.GetGVEGameBotGSFloat()

	tgtGS := gamedata.RandInt31(int32((minGS * float32(gs))), int32((maxGS * float32(gs))), robotRand)
	logs.Debug("gs info: %d, %d, %d", int32((minGS * float32(gs))), int32((maxGS * float32(gs))), tgtGS)
	gsn := int64(float64(tgtGS)) *
		rank.RankByCorpDelayPowBase
	acids, err := getInfoFromRank(db_name, gsn, gsn, dur)

	logs.Debug("GenRobotCompanion acid: %v", acids)
	if err != nil {

		logs.Debug("No Rank Available For GS: %d", gs)
		// 拉取机器人
		infos, err := loadTrueRobot(Robot_Count)
		if err != nil {
			// 理论上不可能执行到这里
			logs.Error("Fetal error(Can't get Robot) by %v", err)
			return nil, err
		}
		return infos, nil

	}
	tgtIDs := make([]string, 0, Robot_Count)
	for _, id := range acids {
		if id == acID {
			continue
		}
		tgtIDs = append(tgtIDs, id)
		if len(tgtIDs) >= Robot_Count {
			break
		}
	}
	logs.Debug("Final GenRobotCompanion acid: %v", tgtIDs)
	infos, err := loadRobotByID(tgtIDs)
	if err != nil {
		logs.Error("Fetal error(LoadRobotById failed) by", err)
		return nil, err
	}
	return infos, nil
}

// 或许通用
// 从某一服某一排行榜拉取一定信息,
// value1, value2为分数(分别为向上取和向下取得起始值)
// 将会从排行榜拉取玩家分数所在榜中位置的或前或后指定范围(dur)的玩家

func getInfoFromRank(dbName string, value1, value2 int64, dur int) ([]string, error) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBSlaveConn Err conn is nil")
		return nil, fmt.Errorf("GetRankInfo modules.GetDBSlaveConn nil")
	}

	logs.Debug("Redis Command: ZRANGEBYSCORE %s %d +inf WITHSCORES limit 0 %d", dbName, value1, dur)
	m1, err := redis.StringMap(metricsModules.DoWraper("GVE_"+dbName, conn, "ZRANGEBYSCORE", dbName, value1, "+inf",
		"WITHSCORES", "limit", "0", dur))
	if err != nil {
		logs.Error("_GetExpeditionEnemyId err %v", err)
		return nil, err
	}
	if len(m1) < dur {
		m2, err := redis.StringMap(metricsModules.DoWraper("GVE_"+dbName, conn, "ZREVRANGEBYSCORE", dbName, value2, "-inf",
			"WITHSCORES", "limit", "0", dur-len(m1)))
		if err != nil {
			logs.Error("ZREVRANGEBYSCORE err %v", err)
			return nil, err
		}
		for k, v := range m2 {
			m1[k] = v
		}
	}
	if len(m1) < dur {
		return nil, fmt.Errorf("No Available Player in Rank")
	}
	res := make([]string, 0, len(m1))
	for k := range m1 {
		res = append(res, k)
	}
	return res, nil
}

func loadRobotByID(acIDs []string) ([]*helper.Avatar2ClientByJson, error) {
	res := make([]*helper.Avatar2ClientByJson, len(acIDs))
	for i, acid := range acIDs {
		dbAccountID, err := db.ParseAccount(acid)
		if err != nil {
			logs.Error("loadRobotByID db.ParseAccount %d %v", i, err)
			continue
		}
		enemy_account, err := account.LoadPvPAccount(dbAccountID)

		a := &helper.Avatar2ClientByJson{}

		avatar := 0
		heroTm := enemy_account.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_TEAMBOSS)
		if heroTm != nil && len(heroTm) > 0 {
			avatar = heroTm[0]
		}
		err = account.FromAccount2Json(
			a,
			enemy_account,
			avatar)
		if err != nil {
			logs.Error("loadRobotByID account.FromAccount err %d %v", i, err)
			continue
		}
		res[i] = a
	}
	return res, nil
}

func loadTrueRobot(count int) ([]*helper.Avatar2ClientByJson, error) {
	ret := make([]*helper.Avatar2ClientByJson, 0, count)
	droid := gamedata.GetRandDroidForSimplePvp()
	for i := 0; i < count; i++ {
		a := &helper.Avatar2ClientByJson{}
		err := account.FromDroidAccount2Json(a, droid, -1)
		if err != nil {
			return ret, err
		}
		a.SimplePvpScore = rank.SimplePvpInitScoreReal
		a.SimplePvpRank = 9999
		a.AcID = fmt.Sprintf("a.AcID%d", i+1)
		ret = append(ret, a)
	}
	return ret, nil
}
