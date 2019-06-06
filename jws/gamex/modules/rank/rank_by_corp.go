package rank

import (
	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type getScoreFunc func(a *helper.AccountSimpleInfo) int64
type rankByCorpBalanceFunc func(rank int, id string)
type rankByCorpBalanceBatchFunc func(randUids []string)
type rankByCorpFromCacheBalanceFunc func(topN [RankTopSize]CorpDataInRank, acid2PosScore map[string]PairPosScore)

type RankByCorpGetRes struct {
	TopN  []CorpDataInRank
	Pos   int
	Score int64
}

type rankByCorpCommand struct {
	is_update bool                    // 是否是更新
	is_get    bool                    // 是否是要获取acid对应的排行榜信息
	is_reload bool                    // 是否需要重新加载topN
	acid      string                  // 账号id
	data      CorpDataInRank          // 更新数据
	res_chan  chan<- RankByCorpGetRes // 获取用的回复channel
}

func newAddRankByCorpCommand(acid string, rank_data *CorpDataInRank) rankByCorpCommand {
	return rankByCorpCommand{
		is_get: false,
		acid:   acid,
		data:   *rank_data,
	}
}

func newUpdateRankByCorpCommand(acid string, rank_data *CorpDataInRank) rankByCorpCommand {
	return rankByCorpCommand{
		is_update: true,
		acid:      acid,
		data:      *rank_data,
	}
}

func newGetRankByCorpCommand(acid string, res_chan chan<- RankByCorpGetRes) rankByCorpCommand {
	return rankByCorpCommand{
		is_get:   true,
		acid:     acid,
		res_chan: res_chan,
	}
}

type RankByCorp struct {
	topN     CorpTopN
	getScore getScoreFunc // 通过account获取分数

	res_pos_chan  chan rankByCorpCommand
	res_topN_chan chan rankByCorpCommand

	yesterday_rank *RankByCorpStatic

	waitter   sync.WaitGroup
	rankdb    *rankDB
	rank_name string
	db_name   string

	isClean      bool
	balanceFunc  rankByCorpBalanceFunc
	balance_chan chan bool

	rankId int64
}

func (r *RankByCorp) delRank(acID string) error {
	logs.Debug("del rank, acid: %v", acID)
	r.rankdb.del(r.rank_name, r.db_name, acID)
	r.ReloadTopN()
	return nil
}

func (r *RankByCorp) loadTopN() {
	res, err := r.rankdb.loadTopN(r.rank_name, r.db_name)
	if err != nil && err != redis.ErrNil {
		logs.Error("loadTopN %s Err by %s", r.rank_name, err.Error())
		return
	}
	if err != redis.ErrNil {
		r.topN.fromDB(res)
	}
	//logs.Trace("getTopWithScoreFromRedis 3 %v", r.topN)
}

func (r *RankByCorp) saveTopN() {
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

func (r *RankByCorp) getTopN() [RankTopSize]CorpDataInRank {
	// 注意外部调用前需要复制
	return r.topN.TopN
}

func (r *RankByCorp) getPos(acid string) int {
	return r.rankdb.getPos(r.rank_name, r.db_name, acid)
}

func (r *RankByCorp) GetCorpInfo(a *helper.AccountSimpleInfo) *CorpDataInRank {
	c := &CorpDataInRank{}
	c.SetDataFromAccount(a, r.getScore(a))
	return c
}

// 内存中对topN进行排序， redis对所有的排序
func (r *RankByCorp) add(acid string, data CorpDataInRank) {
	oldScore, err := r.rankdb.getFloatScore(r.rank_name, r.db_name, acid)
	if err != nil {
		logs.Error("get score err by %v", err)
		return
	}
	if int64(int64(oldScore)/r.topN.ScorePow) >= data.getScore() {
		return
	}
	// 先存Redis, 计算topN以Redis为准
	isTopNChange, scoreWithSpeck := r.topN.Add(acid, data)
	r.rankdb.addWithDoubleScore(r.rank_name, r.db_name, acid, scoreWithSpeck)
	if isTopNChange {
		r.saveTopN()
	}
}

func (r *RankByCorp) update(acid string, data CorpDataInRank) {
	// 先存Redis, 计算topN以Redis为准
	isInTopN := r.topN.updateInfoWithoutScore(acid, data)
	if isInTopN {
		r.saveTopN()
	}
}

func (r *RankByCorp) setYesterdayRank(rank *RankByCorpStatic) {
	r.yesterday_rank = rank
}

func (r *RankByCorp) setNeedBalanceTopN(
	timeToBalance util.TimeToBalance,
	is_clean bool,
	bf rankByCorpBalanceFunc) (string, chan<- bool, util.TimeToBalance) {
	r.isClean = is_clean
	r.balanceFunc = bf
	return r.rank_name, r.balance_chan, timeToBalance
}

func (r *RankByCorp) Start(
	rank_id int64,
	rank_name string,
	power int64,
	rankdb *rankDB,
	db_name string,
	getScore getScoreFunc) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByCorp %s Panic, Err %v", rank_name, err)
		}
	}()
	r.getScore = getScore
	r.rankdb = rankdb
	r.rank_name = rank_name
	r.db_name = db_name
	r.rankId = rank_id

	r.topN.rankdb = rankdb
	r.topN.rank_name = rank_name
	r.topN.db_name = db_name

	r.loadTopN()
	r.topN.Reload()
	r.topN.ScorePow = power

	r.res_pos_chan = make(chan rankByCorpCommand, GetChannelSize)
	r.res_topN_chan = make(chan rankByCorpCommand, AddChannelSize)
	r.balance_chan = make(chan bool, 1)

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.res_pos_chan
			if !ok {
				logs.Warn("res_pos_chan close")
				return
			}
			if !command.is_get || command.res_chan == nil {
				logs.Warn("res_pos_chan command error")
				continue
			}

			command.res_chan <- RankByCorpGetRes{
				TopN: make([]CorpDataInRank, RankTopSize),
				Pos:  r.getPos(command.acid),
			}
		}
	}()

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			select {
			case command, ok := <-r.res_topN_chan:
				{
					if !ok {
						logs.Warn("res_pos_chan close")
						return
					}

					r.waitter.Add(1)
					func() {
						defer r.waitter.Done()
						if command.is_reload {
							r.topN.Reload()
							logs.Warn("RankByCorp %s Reload", r.db_name)
							return
						}

						if command.is_update {
							r.update(command.acid, command.data)
							return
						}

						if command.is_get && command.res_chan != nil {
							logs.Trace("command %v", command)
							topN := r.getTopN()
							command.res_chan <- RankByCorpGetRes{
								TopN: topN[:],
							}
							return
						}

						if !command.is_get {
							r.add(command.acid, command.data)
							return
						}
						logs.Warn("res_topN_chan command error")
					}()
				}
			case is_need_balance := <-r.balance_chan:
				{
					logs.Warn("Rank %s blance %v", r.rank_name, is_need_balance)

					top_uids, err := r.rankdb.getTopFromRedis(r.rank_name, r.db_name)

					if err == nil {
						for rank_1 := 0; (rank_1 < len(top_uids)) && (rank_1 < RankBalanceSize); rank_1++ {
							r.balanceFunc(rank_1+1, top_uids[rank_1])
						}
						if r.yesterday_rank != nil {
							r.yesterday_rank.SetTopN(r.topN.TopN)
						}
						if r.isClean {
							logs.Warn("Rank %s blance %v Clean Start!", r.rank_name, is_need_balance)
							if r.yesterday_rank != nil {
								r.rankdb.reName(r.rank_name, r.yesterday_rank.rank_name, r.db_name)
							}
							r.topN.clean()
							r.saveTopN()
						} else {
							logs.Warn("Rank %s blance %v Copy Start!", r.rank_name, is_need_balance)
							if r.yesterday_rank != nil {
								err := r.rankdb.copy(r.rank_name, r.yesterday_rank.rank_name, r.db_name)
								if err != nil {
									logs.Error("Rank %s blance %v Error by %s!",
										r.rank_name, is_need_balance, err.Error())
								}
							}
						}
					} else {
						logs.Error("getTopFromRedis Err By %s", err.Error())
					}
				}
			}

		}
	}()
}

func (r *RankByCorp) Stop() {
	close(r.res_topN_chan)
	close(r.res_pos_chan)

	r.saveTopN()

	r.waitter.Wait()
}

func (r *RankByCorp) Get(acid string) *RankByCorpGetRes {
	res_chan := make(chan RankByCorpGetRes)
	getCommand := newGetRankByCorpCommand(acid, res_chan)
	r.res_pos_chan <- getCommand
	r.res_topN_chan <- getCommand

	logs.Trace("Rank Get Has Send")

	res_1 := <-res_chan
	logs.Trace("Rank Get res_1 %v", res_1)
	res_2 := <-res_chan
	logs.Trace("Rank Get res_2 %v", res_2)

	for i := 0; i < len(res_1.TopN); i++ {
		res_1.TopN[i].Score = res_1.TopN[i].Score / r.topN.ScorePow
	}

	for i := 0; i < len(res_2.TopN); i++ {
		res_2.TopN[i].Score = res_2.TopN[i].Score / r.topN.ScorePow
	}

	if len(res_2.TopN) > 0 && res_2.TopN[0].Name != "" {
		res_2.Pos = res_1.Pos
		return &res_2
	} else {
		res_1.Pos = res_2.Pos
		return &res_1
	}
}

func (r *RankByCorp) Add(a *helper.AccountSimpleInfo) {
	if r.getScore(a) == 0 || a.Name == "" {
		logs.Debug("can't reach the min requirement for rank: %s, info: %v", r.db_name, *a)
		return
	}
	data := &CorpDataInRank{}
	data.SetDataFromAccount(a, r.getScore(a))
	data.SetExtraData(nil)
	r.execCmd(newAddRankByCorpCommand(a.AccountID, data))
	return
}

func (r *RankByCorp) AddScoreCanZero(a *helper.AccountSimpleInfo) {
	if a.Name == "" {
		logs.Debug("can't reach the min requirement for rank: %s, info: %v", r.db_name, *a)
		return
	}
	data := &CorpDataInRank{}
	data.SetDataFromAccount(a, r.getScore(a))
	data.SetExtraData(nil)
	r.execCmd(newAddRankByCorpCommand(a.AccountID, data))
	return
}

func (r *RankByCorp) AddWithExtraData(a *helper.AccountSimpleInfo, extraData interface{}) {
	data := &CorpDataInRank{}
	data.SetDataFromAccount(a, r.getScore(a))
	data.SetExtraData(extraData)
	r.execCmd(newAddRankByCorpCommand(a.AccountID, data))
	return
}

func (r *RankByCorp) UpdateIfInTopN(a *helper.AccountSimpleInfo) {
	data := &CorpDataInRank{}
	data.SetDataFromAccount(a, r.getScore(a))
	r.execCmd(newUpdateRankByCorpCommand(a.AccountID, data))
	return
}

func (r *RankByCorp) UpdateIfInTopNWithExtraData(a *helper.AccountSimpleInfo, extraData interface{}) {
	data := &CorpDataInRank{}
	data.SetDataFromAccount(a, r.getScore(a))
	data.SetExtraData(extraData)
	r.execCmd(newUpdateRankByCorpCommand(a.AccountID, data))
	return
}

func (r *RankByCorp) ReloadTopN() {
	r.execCmd(rankByCorpCommand{
		is_reload: true,
	})
	return
}

func (r *RankByCorp) RebaseScore(rs float64) int64 {
	return int64(rs) / r.topN.ScorePow
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorp) execCmd(cmd rankByCorpCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topN_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorp execCmd put timeout %v", cmd)
	}
}
