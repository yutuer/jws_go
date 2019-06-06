package merge

import (
	"strconv"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func MergeRank() error {
	BRankDB := GetBRankDBConn()
	defer BRankDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()
	ResRankDB := GetResRankDBConn()
	defer ResRankDB.Close()

	cb := redis.NewCmdBuffer()
	// corp gs
	if err := _corpGs(BRankDB, cb); err != nil {
		return err
	}

	// simple pvp
	if err := _simplePvp(BRankDB, cb); err != nil {
		return err
	}

	// guild gs
	if err := _guildGS(BRankDB, cb); err != nil {
		return err
	}

	// GuildGateEnemy
	//if err := _guildGateEnemy(BRankDB, cb); err != nil {
	//	return err
	//}

	// CorpTrial
	if err := _corpTrial(BRankDB, cb); err != nil {
		return err
	}
	// hero star
	if err := _heroStar(BRankDB, cb); err != nil {
		return err
	}
	// RankCorpHeroDiff
	//if err := _RankCorpHeroDiff(BRankDB, cb); err != nil {
	//	return err
	//}

	if _, err := ResRankDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeRank save res err %s", err.Error())
		return err
	}

	// 处理topN
	//cb = redis.NewCmdBuffer()

	//if err := _guildGateEnemyTopN(BRankDB, ResRankDB, ResDB, cb); err != nil {
	//	return err
	//}

	//if err := _corpTrialTopN(BRankDB, ResRankDB, cb); err != nil {
	//	return err
	//}

	//if _, err := ResRankDB.DoCmdBuffer(cb, true); err != nil {
	//	logs.Error("MergeRank topN save res err %s", err.Error())
	//	return err
	//}
	//
	//if err := mergeRankCorpHeroDiff(ResRankDB, BRankDB); err != nil {
	//	return err
	//}
	return nil
}

func _corpGs(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BCorpGS, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankCorpGs(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B corp gs err %v", err)
		return err
	}
	for i := 0; i < len(BCorpGS); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankCorpGs(Cfg.ResShardId), BCorpGS[i+1], BCorpGS[i]); err != nil {
			logs.Error("MergeRank ZADD B corp gs err %v", err)
			return err
		}
	}
	// 删除玩家
	for acid, _ := range delAcids {
		if err := cb.Send("ZREM", rank.TableRankCorpGs(Cfg.ResShardId), acid); err != nil {
			logs.Error("MergeRank _corpGs ZREM corp gs err %v", err)
			return err
		}
	}
	return nil
}

func _simplePvp(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BSimplePvp, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankSimplePvp(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B simple pvp err %v", err)
		return err
	}
	for i := 0; i < len(BSimplePvp); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankSimplePvp(Cfg.ResShardId), BSimplePvp[i+1], BSimplePvp[i]); err != nil {
			logs.Error("MergeRank ZADD B simple pvp err %v", err)
			return err
		}
	}
	// 删除玩家
	for acid, _ := range delAcids {
		if err := cb.Send("ZREM", rank.TableRankSimplePvp(Cfg.ResShardId), acid); err != nil {
			logs.Error("MergeRank _simplePvp ZREM simple pvp err %v", err)
			return err
		}
	}
	return nil
}

func _guildGS(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BGuildGs, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankGuildGS(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B guild gs err %v", err)
		return err
	}
	for i := 0; i < len(BGuildGs); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankGuildGS(Cfg.ResShardId), BGuildGs[i+1], BGuildGs[i]); err != nil {
			logs.Error("MergeRank ZADD B guild gs err %v", err)
			return err
		}
	}
	// 改变gs
	for guuid, score := range guild_score_chg {
		if err := cb.Send("ZADD", rank.TableRankGuildGS(Cfg.ResShardId), score, guuid); err != nil {
			logs.Error("MergeRank _guildGS ZADD guild gs err %v", err)
			return err
		}
	}
	// 删除工会
	for guuid, _ := range guild_del {
		if err := cb.Send("ZREM", rank.TableRankGuildGS(Cfg.ResShardId), guuid); err != nil {
			logs.Error("MergeRank _guildGS ZREM guild gs err %v", err)
			return err
		}
	}
	return nil
}

func _guildGateEnemy(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BGuildGateEnemy, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankGuildGateEnemy(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B GuildGateEnemy err %v", err)
		return err
	}
	for i := 0; i < len(BGuildGateEnemy); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankGuildGateEnemy(Cfg.ResShardId), BGuildGateEnemy[i+1], BGuildGateEnemy[i]); err != nil {
			logs.Error("MergeRank ZADD B GuildGateEnemy err %v", err)
			return err
		}
	}
	// 删除工会
	for guuid, _ := range guild_del {
		if err := cb.Send("ZREM", rank.TableRankGuildGateEnemy(Cfg.ResShardId), guuid); err != nil {
			logs.Error("MergeRank _guildGateEnemy ZREM err %v", err)
			return err
		}
	}
	return nil
}

func _guildGateEnemyTopN(BRankDB, ResRankDB, ResDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	rank_name := rank.TableRankGuildGateEnemy(Cfg.BShardId)
	topNIDs, topNScores, err := _getTopWithScoreFromRedis(rank_name, BRankDB, ResRankDB)
	if err != nil {
		logs.Error("_guildGateEnemyTopN _getTopWithScoreFromRedis err %s", err.Error())
		return err
	}

	newTopN := [rank.RankTopSize]rank.GuildDataInRank{}

	for idx, id := range topNIDs {
		if idx < 0 || idx >= len(newTopN) {
			logs.Trace("GuildTopN Err by no newTopN %d", idx)
			continue
		}
		if idx >= len(topNScores) {
			continue
		}

		if _, ok := guild_del[id]; ok {
			continue
		}

		guildData := guild_info.LoadGuildInfo(id, ResDB)
		if guildData == nil {
			logs.Trace("_guildGateEnemyTopN GuildTopN Err by no data %s", id)
			continue
		}
		newTopN[idx].SetDataGuild(guildData, topNScores[idx])
	}

	c := &rank.GuildTopN{}
	c.TopN = newTopN
	c.MinScoreToTopN = c.TopN[len(c.TopN)-1].Score

	// to db
	data, err := json.Marshal(*c)
	if err != nil {
		logs.Error("_guildGateEnemyTopN json.Marshal Err by %s", err.Error())
		return err
	}

	res_rank_name := rank.TableRankGuildGateEnemy(Cfg.AShardId)
	rank_name_in_db := res_rank_name + ":topN"
	if err := cb.Send("SET", rank_name_in_db, data); err != nil {
		logs.Error("_guildGateEnemyTopN SET err %s", err.Error())
		return err
	}
	return nil
}

func _corpTrial(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BCorpTrial, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankCorpTrial(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B CorpTrial err %v", err)
		return err
	}
	for i := 0; i < len(BCorpTrial); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankCorpTrial(Cfg.ResShardId), BCorpTrial[i+1], BCorpTrial[i]); err != nil {
			logs.Error("MergeRank ZADD B CorpTrial err %v", err)
			return err
		}
	}
	// 删除玩家
	for acid, _ := range delAcids {
		if err := cb.Send("ZREM", rank.TableRankCorpTrial(Cfg.ResShardId), acid); err != nil {
			logs.Error("MergeRank _corpTrial ZREM err %v", err)
			return err
		}
	}
	return nil
}

func _heroStar(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BHeroStar, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		rank.TableRankCorpHeroStar(Cfg.BShardId), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B _heroStarTrial err %v", err)
		return err
	}
	for i := 0; i < len(BHeroStar); i += 2 {
		if err := cb.Send("ZADD", rank.TableRankCorpHeroStar(Cfg.ResShardId), BHeroStar[i+1], BHeroStar[i]); err != nil {
			logs.Error("MergeRank ZADD B _heroStarTrial err %v", err)
			return err
		}
	}
	// 删除玩家
	for acid, _ := range delAcids {
		if err := cb.Send("ZREM", rank.TableRankCorpHeroStar(Cfg.ResShardId), acid); err != nil {
			logs.Error("MergeRank _heroStarTrial ZREM err %v", err)
			return err
		}
	}
	return nil
}

func _RankCorpHeroDiff(BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	if err := _ARankCorpHeroDiff(rank.TableRankCorpHeroDiffTU(Cfg.BShardId),
		rank.TableRankCorpHeroDiffTU(Cfg.ResShardId), BRankDB, cb); err != nil {
		return err
	}
	if err := _ARankCorpHeroDiff(rank.TableRankCorpHeroDiffZHAN(Cfg.BShardId),
		rank.TableRankCorpHeroDiffZHAN(Cfg.ResShardId), BRankDB, cb); err != nil {
		return err
	}
	if err := _ARankCorpHeroDiff(rank.TableRankCorpHeroDiffHU(Cfg.BShardId),
		rank.TableRankCorpHeroDiffHU(Cfg.ResShardId), BRankDB, cb); err != nil {
		return err
	}
	if err := _ARankCorpHeroDiff(rank.TableRankCorpHeroDiffSHI(Cfg.BShardId),
		rank.TableRankCorpHeroDiffSHI(Cfg.ResShardId), BRankDB, cb); err != nil {
		return err
	}
	return nil

}

func _ARankCorpHeroDiff(BTable, ResTable string, BRankDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	BCHD, err := redis.Strings(BRankDB.Do("ZREVRANGE",
		BTable, 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("MergeRank ZREVRANGE B _RankCorpHeroDiff err %v", err)
		return err
	}
	for i := 0; i < len(BCHD); i += 2 {
		if err := cb.Send("ZADD", ResTable, BCHD[i+1], BCHD[i]); err != nil {
			logs.Error("MergeRank ZADD B _RankCorpHeroDiff err %v", err)
			return err
		}
	}
	// 删除玩家
	for acid, _ := range delAcids {
		if err := cb.Send("ZREM", ResTable, acid); err != nil {
			logs.Error("MergeRank _RankCorpHeroDiff ZREM err %v", err)
			return err
		}
	}
	return nil
}

func _corpTrialTopN(BDB, ResDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	rank_name := rank.TableRankCorpTrial(Cfg.BShardId)
	topNIDs, topNScores, err := _getTopWithScoreFromRedis(rank_name, BDB, ResDB)
	if err != nil {
		logs.Error("_corpTrialTopN _getTopWithScoreFromRedis err %s", err.Error())
		return err
	}

	newTopN := [rank.RankTopSize]rank.CorpDataInRank{}

	for idx, id := range topNIDs {
		if idx < 0 || idx >= len(newTopN) {
			//logs.Error("ParseAccount %d", idx)
			continue
		}
		if idx >= len(topNScores) {
			//logs.Error("ParseAccount %d", idx)
			continue
		}
		accountDBID, err := db.ParseAccount(id)
		if err != nil {
			logs.Error("_corpTrialTopN ParseAccount %s Err By %s", accountDBID, err.Error())
			continue
		}
		accountData, err := simple_info.LoadAccountSimpleInfoProfile(accountDBID)
		if err != nil {
			logs.Trace("_corpTrialTopN LoadAccount %s Err By %s", accountDBID, err.Error())
			continue
		}
		newTopN[idx].SetDataFromAccount(accountData, topNScores[idx])
	}

	c := new(rank.CorpTopN)
	c.TopN = newTopN
	c.MinScoreToTopN = c.TopN[len(c.TopN)-1].Score

	// to db
	data, err := json.Marshal(*c)
	if err != nil {
		logs.Error("_corpTrialTopN _guildGateEnemyTopN json.Marshal Err by %s", err.Error())
		return err
	}

	res_rank_name := rank.TableRankCorpTrial(Cfg.AShardId)
	rank_name_in_db := res_rank_name + ":topN"
	if err := cb.Send("SET", rank_name_in_db, data); err != nil {
		logs.Error("_corpTrialTopN SET err %s", err.Error())
		return err
	}
	return nil
}

func _getTopWithScoreFromRedis(tableKey string, BDB, ResDB redispool.RedisPoolConn) ([]string, []int64, error) {
	resOld, err := BDB.Do("ZREVRANGE", tableKey, 0, rank.RankTopSize, "WITHSCORES")
	res, err := redis.Strings(resOld, err)
	if err != nil {
		logs.Error("_getTopWithScoreFromRedis ZREVRANGE Err by %s", err.Error())
		return []string{}, []int64{}, err
	}

	ids := make([]string, 0, len(res))
	scores := make([]int64, 0, len(res))
	for i := 0; i+1 < len(res); i += 2 {
		ids = append(ids, res[i])
		s, err := strconv.ParseInt(res[i+1], 10, 64)
		if err != nil {
			logs.Error("_getTopWithScoreFromRedis strconv.Atoi %v Err by %s", res[i+1], err.Error())
			return []string{}, []int64{}, err
		}
		scores = append(scores, s)
	}

	return ids, scores, nil
}

func mergeRankCorpHeroDiff(resDB, bDB redispool.RedisPoolConn) error {
	if err := _mergeRankCorpHeroDiff(rank.TableRankCorpHeroDiffTU(Cfg.ResShardId),
		rank.TableRankCorpHeroDiffTU(Cfg.BShardId), resDB, bDB); err != nil {
		return err
	}
	if err := _mergeRankCorpHeroDiff(rank.TableRankCorpHeroDiffZHAN(Cfg.ResShardId),
		rank.TableRankCorpHeroDiffZHAN(Cfg.BShardId), resDB, bDB); err != nil {
		return err
	}
	if err := _mergeRankCorpHeroDiff(rank.TableRankCorpHeroDiffHU(Cfg.ResShardId),
		rank.TableRankCorpHeroDiffHU(Cfg.BShardId), resDB, bDB); err != nil {
		return err
	}
	if err := _mergeRankCorpHeroDiff(rank.TableRankCorpHeroDiffSHI(Cfg.ResShardId),
		rank.TableRankCorpHeroDiffSHI(Cfg.BShardId), resDB, bDB); err != nil {
		return err
	}
	return nil
}

func _mergeRankCorpHeroDiff(rank_name_res, rank_name_b string, resDB, bDB redispool.RedisPoolConn) error {
	err, resTopN := loadCorpTopN(rank_name_res, resDB)
	if err != nil {
		return err
	}
	err, bTopN := loadCorpTopN(rank_name_b, bDB)
	if err != nil {
		return err
	}
	for _, t := range bTopN.TopN {
		resTopN.Add(t.ID, t)
	}
	resTopNb, err := json.Marshal(resTopN)
	if err != nil {
		logs.Error("_mergeRankCorpHeroDiff json.Marshal err %s", err.Error())
		return err
	}

	rank_name_in_db := rank_name_res + ":topN"
	_, err = redis.Bytes(resDB.Do("SET", rank_name_in_db, resTopNb))
	if err != nil {
		logs.Error("_mergeRankCorpHeroDiff SET %s", err.Error())
		return err
	}
	return nil
}

func loadCorpTopN(rank_name string, resDB redispool.RedisPoolConn) (error, *rank.CorpTopN) {
	rank_name_in_db := rank_name + ":topN"

	data, err := redis.Bytes(resDB.Do("GET", rank_name_in_db))
	if err != nil {
		logs.Error("loadCorpTopN GET err %s %s", rank_name_in_db, err.Error())
		return err, nil
	}

	topN := &rank.CorpTopN{}
	err = json.Unmarshal(data, topN)
	if err != nil {
		logs.Error("loadCorpTopN json.Unmarshal err %s", err.Error())
		return err, nil
	}
	return nil, topN
}
