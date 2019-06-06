package rank

import (
	"sync"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RankByGuildStatic struct {
	topN      GuildTopN
	rankdb    *rankDB
	rank_name string
	db_name   string
	rankId    int64

	getScoreF    getGuildScoreFunc // 通过account获取分数
	scorePowBase int64
	mutex        sync.RWMutex
}

func (r *RankByGuildStatic) setTopN(data [RankTopSize]GuildDataInRank) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.topN.setTopN(data)
	r.saveTopN()
}

func (r *RankByGuildStatic) Get(acid string) *RankByGuildGetRes {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	res := &RankByGuildGetRes{
		TopN: r.topN.TopN,
	}
	for i := 0; i < len(res.TopN); i++ {
		res.TopN[i].Score = res.TopN[i].Score / r.scorePowBase
	}
	res.Score = res.Score / r.scorePowBase
	return res
}

func (r *RankByGuildStatic) Start(rank_id int64, rank_name string, rankdb *rankDB, db_name string, getScore getGuildScoreFunc) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByGuildStatic Panic, Err %v", err)
		}
	}()

	r.rankdb = rankdb
	r.rank_name = rank_name
	r.db_name = db_name
	r.rankId = rank_id

	r.topN.rankdb = rankdb
	r.topN.rank_name = rank_name
	r.topN.db_name = db_name

	r.getScoreF = getScore
	r.scorePowBase = RankByGuildPowBase

	r.loadTopN()

}

func (r *RankByGuildStatic) loadTopN() {
	res, err := r.rankdb.loadTopN(r.rank_name, r.db_name)
	if err != nil && err != redis.ErrNil {
		logs.Error("loadTopN %s Err by %s", r.rank_name, err.Error())
		return
	}
	if err == redis.ErrNil {
		return
	}
	r.topN.fromDB(res)
}

func (r *RankByGuildStatic) saveTopN() {
	data, err := r.topN.toDB()
	if err != nil {
		logs.Error("saveTopN %s Err by %s", r.rank_name, err.Error())
		return
	}

	err = r.rankdb.saveTopN(r.rank_name, r.db_name, data)

	if err != nil {
		logs.Error("rankdb saveTopN %s Err by %s", r.rank_name, err.Error())
		return
	}
}
