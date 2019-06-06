package herogacharace

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// UpdateScore 返回最新Rank或者错误码,
// 没有排名, 返回errorcode WARN_ACTIVITY_NO_RANK
// AddScore 添加更新积分并获取当前排名, 更新当前最小积分, 小于最小积分的不需要进入排行榜
// ZADD myzset score member
// 获取我当前的排名  ZREVRANK key member , 如果不存在则返回 nil
func (hgr *HeroGachaRace) UpdateScore(activity HGRActivity, newScore uint64, member HGRankMember) (uint64, errorcode.ErrorCode) {
	hgr.locker.Lock()
	defer hgr.locker.Unlock()

	if hgr.redis == nil || hgr.curActivity == nil {
		if err := hgr.initScores(activity); err != nil {
			logs.Error("HeroGachaRace.UpdateScore initScores err:%s", err.Error())
			return 0, WARN_ACTIVITY_NOT_READY
		}
	}

	if hgr.NumItems >= MAXRANK {
		if hgr.MinScore > newScore {
			fmt.Println("a", hgr.MinScore, newScore)
			return 0, nil
		}
	}

	//score 先来的占据高分位 100.(1.0/201610101212600) 这里只需要取得 整数部分,全面舍去小数
	//XXX:  如果精度不够,可以尝试使用纳秒
	passedTime := time.Now().Unix() - hgr.curActivity.StartTime

	if passedTime <= 0 {
		return 0, WARN_ACTIVITY_ACTIVITYISNOTSTARTED
	}

	if err := hgr.redisPreCheck(*hgr.curActivity); err != nil {
		return 0, WARN_ACTIVITY_DB_FAILED
	}

	newScoreF := float64(newScore) + 1.0/float64(passedTime)
	conn := hgr.redis
	key := hgr.curActivity.GetRedisKey()
	memberKey := member.String()
	var replies []interface{}
	err := backoff.Retry(
		func() error {
			conn.Send("MULTI")
			//conn.Send("ZREVRANK", key, memberKey)
			conn.Send("ZADD", key, newScoreF, memberKey)
			conn.Send("ZREVRANK", key, memberKey)
			conn.Send("EXPIREAT", key, hgr.curActivity.GetRedisExpireUnix())
			conn.Send("ZCARD", key)
			conn.Send("ZREVRANGE", key, 0, MAXRANK-1, "WITHSCORES")
			r, err := redis.Values(conn.Do("EXEC"))
			if err == nil {
				replies = r
				return nil
			} else {
				return err
			}
		},
		New2SecBackOff(),
	)

	if err != nil {
		logs.Error("HeroGachaRace.UpdateScore failed %s", err.Error())
		return 0, WARN_ACTIVITY_DB_FAILED
	}
	//logs.Trace("******: %v", replies)
	num := len(replies)
	if num < 3 {
		logs.Error("HeroGachaRace.UpdateScore the number of repies should bigger than 3.")
		return 0, WARN_ACTIVITY_DB_FAILED
	}

	//var oldRank uint64
	//if replies[0] == nil {
	//	oldRank = 0
	//} else {
	//	if oRank, err := redis.Uint64(replies[0], nil); err != nil {
	//
	//	} else {
	//		oldRank = oRank + 1
	//	}
	//}
	if replies[1] == nil {
		return 0, WARN_ACTIVITY_NO_RANK
	}
	if newRank, err := redis.Uint64(replies[1], nil); err != nil {
		logs.Error("HeroGachaRace.UpdateScore rank parse failed.")
		return 0, WARN_ACTIVITY_DB_FAILED
	} else {
		newRank = newRank + 1

		hgr.updateRedisAllScores(replies[3:])

		if newRank > MAXRANK {
			return 0, WARN_ACTIVITY_NO_RANK
		} else {
			return newRank, nil
		}
	}

}

// GetAllScores 请比较自己最后一次记录的rank, 如果再UpdateScore时发生了变化则需要调用本函数
func (hgr *HeroGachaRace) GetAllScores() ([MAXRANK]HeroGachaRankItem, int, errorcode.ErrorCode) {
	hgr.locker.RLock()
	defer hgr.locker.RUnlock()

	if hgr.curActivity == nil {
		return hgr.Items, 0, WARN_ACTIVITY_NOT_READY
	}
	now_t := time.Now().Unix()
	if hgr.getScoreTS+RANK_REFRESH_TIME < now_t {
		hgr.getScoreTS = now_t
		hgr.pullAllScores(*hgr.curActivity)
	}

	for i := range hgr.Items {
		isid := hgr.Items[i].sid
		hgr.Items[i].ShardDisplayName = etcd.ParseDisplayShardName(etcd.GetSidDisplayName(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), isid))
	}

	return hgr.Items, hgr.NumItems, nil
}

func (hgr *HeroGachaRace) InitCurActivity(curAct *HGRActivity) {
	hgr.locker.Lock()
	defer hgr.locker.Unlock()

	hgr.pullAllScores(*curAct)
}

//// PullAllScores 获取前100位排名者
//// ZREVRANGE myzset 0 101 [WITHSCORES]
func (hgr *HeroGachaRace) ForcePullAllScores() {
	hgr.locker.Lock()
	defer hgr.locker.Unlock()

	if hgr.curActivity != nil {
		hgr.pullAllScores(*hgr.curActivity)
	}
}

func (hgr *HeroGachaRace) updateRedisAllScores(replies []interface{}) {
	zCount, err := redis.Uint64(replies[0], nil)
	if err != nil {
		logs.Error("HeroGachaRace.PullAllScores zCount values failed.")
		return
	}
	if !(zCount > uint64(0)) {
		return
	}

	replies, err = redis.Values(replies[1], nil)
	if err != nil {
		logs.Error("HeroGachaRace.updateRedisAllScores replies is nil, %s", err.Error())
		return
	}
	num := len(replies) / 2

	hgr.NumItems = num
	//logs.Trace("UUU %d, %v, %d", num, replies, hgr.NumItems)
	for i := 0; i < num; i++ {
		idx := i * 2

		member, err1 := redis.String(replies[idx], nil)
		if err1 != nil {
			logs.Error("HeroGachaRace.PullAllScores member parse failed2. %s", err1.Error())
			return
		}

		score, err2 := redis.Float64(replies[idx+1], nil)
		if err2 != nil {
			logs.Error("HeroGachaRace.PullAllScores score parse failed3. %s", err2.Error())
			return
		}

		err = hgr.Items[i].SetByRedisValue(uint64(i+1), member, score)
		if err != nil {
			logs.Error("HeroGachaRace.PullAllScores parse items failed %d, %s, %d, err:%s", i+1, member, score, err.Error())
			return
		}
	}

	hgr.MinScore = hgr.Items[hgr.NumItems-1].Score
	if zCount > 2*MAXRANK {
		hgr.redis.Do("ZREMRANGEBYRANK", hgr.curActivity.GetRedisKey(), 0, -(MAXRANK + 1))
	}
}

func (hgr *HeroGachaRace) redisPreCheck(activity HGRActivity) error {
	//确保链接还存在,如果不存在则重新创建一个新的链接
	if _, err := hgr.redis.Do("PING"); err != nil {
		logs.Warn("HeroGachaRace.pullAllScores reconnect to db.")
		hgr.redis = nil
		err := hgr.initRedis(activity)
		if err != nil {
			logs.Error("HeroGachaRace.UpdateScore initScores err2:%s", err.Error())
			return err
		}
	}
	return nil
}

func (hgr *HeroGachaRace) pullAllScores(activity HGRActivity) {
	if hgr.redis == nil || hgr.curActivity == nil {
		if err := hgr.initScores(activity); err != nil {
			logs.Error("HeroGachaRace.UpdateScore initScores err:%s", err.Error())
			return
		}
	}

	if err := hgr.redisPreCheck(activity); err != nil {
		return
	}

	conn := hgr.redis

	key := hgr.curActivity.GetRedisKey()
	var replies []interface{}
	err := backoff.Retry(
		func() error {
			conn.Send("MULTI")
			conn.Send("EXPIREAT", key, activity.GetRedisExpireUnix())
			conn.Send("ZCARD", key)
			conn.Send("ZREVRANGE", key, 0, MAXRANK-1, "WITHSCORES")
			r, err := redis.Values(conn.Do("EXEC"))
			if err == nil {
				replies = r
				return nil
			} else {
				return err
			}
		},
		New2SecBackOff(),
	)

	if err != nil {
		logs.Error("HeroGachaRace.PullAllScores failed %s,", err.Error())
		return
	}
	//logs.Trace("pullAllScores +++: %v, %v, %v",activity, replies[0], replies[1])

	hgr.updateRedisAllScores(replies[1:])
}

func (hgr *HeroGachaRace) balance() {
	if !game.Cfg.GetHotActValidData(hgr.sid, uutil.Hot_Value_Limit_Hero) {
		return
	}
	hgr.locker.Lock()
	defer hgr.locker.Unlock()
	hgr.pullAllScores(*hgr.curActivity)
	gachaRankId := make([]string, 0, 100)
	gachaRank := make([]string, 0, 100)
	gachaRankScore := make([]string, 0, 100)
	for i := 0; i < hgr.NumItems; i++ {
		item := hgr.Items[i]
		ac, _ := db.ParseAccount(item.Member.AccountID)
		if game.Cfg.GetShardIdByMerge(ac.ShardId) != hgr.sid { // 不是本服的不结算
			continue
		}
		cfg := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRRankConfig(hgr.curActivity.ActivityId,
			item.Rank, item.Score)
		// 发奖
		items := make([]string, 0, 4)
		counts := make([]uint32, 0, 4)
		for _, reward := range cfg.GetLoot_Table() {
			items = append(items, reward.GetItemID())
			counts = append(counts, reward.GetItemNum())
		}

		gachaRankId = append(gachaRankId, item.Member.AccountID)
		gachaRank = append(gachaRank, fmt.Sprintf("%d", item.Rank))
		gachaRankScore = append(gachaRankScore, fmt.Sprintf("%d", item.Score))
		if len(items) > 0 {
			mail_sender.BatchSendHeroGachaRaceRankMail(hgr.sid,
				item.Member.AccountID, int(item.Score), int(item.Rank), items, counts)
		}
		logs.Debug("HeroGachaRace balanec %s %d",
			item.Member.AccountID, item.Rank)
	}
	logiclog.LogGachaRank(gachaRankId, gachaRank, gachaRankScore)

}

func (hgr *HeroGachaRace) debugClear() {
	key := hgr.curActivity.GetRedisKey()

	if _, err := hgr.redis.Do("DEL", key); err != nil {
		logs.Error("HeroGachaRace debugClear err %v", err)
	}
}

func (hgr *HeroGachaRace) OnPlayerRename(acid, oldName, newName string, score uint64) {
	hgr.locker.Lock()
	defer hgr.locker.Unlock()

	if hgr.redis == nil || hgr.curActivity == nil {
		return
	}

	if hgr.NumItems >= MAXRANK {
		if hgr.MinScore > score {
			return
		}
	}

	passedTime := time.Now().Unix() - hgr.curActivity.StartTime

	if passedTime <= 0 {
		return
	}

	if err := hgr.redisPreCheck(*hgr.curActivity); err != nil {
		return
	}

	conn := hgr.redis
	key := hgr.curActivity.GetRedisKey()

	oldMember := HGRankMember{AccountID: acid, PlayerName: oldName}
	newMember := HGRankMember{AccountID: acid, PlayerName: newName}

	newScoreF := float64(score) + 1.0/float64(passedTime)

	// 修改redis
	err := backoff.Retry(
		func() error {
			conn.Send("MULTI")
			conn.Send("ZREM", key, oldMember.String())
			conn.Send("ZADD", key, newScoreF, newMember.String())
			_, err := redis.Values(conn.Do("EXEC"))
			if err == nil {
				return nil
			} else {
				return err
			}
		},
		New2SecBackOff(),
	)

	if err != nil {
		logs.Error("HeroGachaRace.OnPlayerName failed %s", err.Error())
		return
	}

	// 更新内存
	for i, ranker := range hgr.Items {
		if ranker.Member == oldMember {
			hgr.Items[i].Member = newMember
			hgr.Items[i].PlayerName = newName
		}
	}
}
