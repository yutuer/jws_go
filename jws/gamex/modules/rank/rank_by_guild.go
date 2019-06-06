package rank

import (
	"sync"

	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 用于确定排序原则 同分先进居上
const RankByGuildPowBase = 10000

type getGuildScoreFunc func(a *guild_info.GuildSimpleInfo) int64
type rankByGuildBalanceFunc func(rank int, id string)
type rankByGuildFromCacheBalanceFunc func([RankTopSize]GuildDataInRank)

type RankByGuildGetRes struct {
	TopN  [RankTopSize]GuildDataInRank
	Pos   int
	Score int64
}

const (
	rankByGuildCommandTyp_Get = iota
	rankByGuildCommandTyp_Add
	rankByGuildCommandTyp_Adds
	rankByGuildCommandTyp_Del
	rankByGuildCommandTyp_Rename
)

type rankByGuildCommand struct {
	typ      int             // 类型
	acid     string          // 账号id
	data     GuildDataInRank // 更新数据
	acids    []string
	datas    []*GuildDataInRank
	res_chan chan<- RankByGuildGetRes // 获取用的回复channel
}

func newAddRankByGuildsCommand(acids []string, rank_datas []*GuildDataInRank) rankByGuildCommand {
	return rankByGuildCommand{
		typ:   rankByGuildCommandTyp_Adds,
		acids: acids,
		datas: rank_datas,
	}
}

func newAddRankByGuildCommand(acid string, rank_data *GuildDataInRank) rankByGuildCommand {
	return rankByGuildCommand{
		typ:  rankByGuildCommandTyp_Add,
		acid: acid,
		data: *rank_data,
	}
}

func newGetRankByGuildCommand(acid string, res_chan chan<- RankByGuildGetRes) rankByGuildCommand {
	return rankByGuildCommand{
		typ:      rankByGuildCommandTyp_Get,
		acid:     acid,
		res_chan: res_chan,
	}
}

func newDelRankByGuildCommand(acid string) rankByGuildCommand {
	return rankByGuildCommand{
		typ:  rankByGuildCommandTyp_Del,
		acid: acid,
	}
}

func newRenameGuildCommand() rankByGuildCommand {
	return rankByGuildCommand{
		typ: rankByGuildCommandTyp_Rename,
	}
}

type RankByGuild struct {
	topN     GuildTopN
	getScore getGuildScoreFunc // 通过account获取分数

	res_pos_chan  chan rankByGuildCommand
	res_topN_chan chan rankByGuildCommand

	waitter   sync.WaitGroup
	rankdb    *rankDB
	rank_name string
	db_name   string

	yesterday_rank *RankByGuild

	isClean      bool
	balanceFunc  rankByGuildBalanceFunc
	balance_chan chan bool

	scorePowBase int64

	rankId int64
}

func (r *RankByGuild) loadTopN() {
	res, err := r.rankdb.loadTopN(r.rank_name, r.db_name)
	if err != nil && err != redis.ErrNil {
		logs.Error("loadTopN %s Err by %s", r.rank_name, err.Error())
		return
	}
	if err != redis.ErrNil {
		r.topN.fromDB(res)
	}
	r.topN.db_name = r.db_name
	r.topN.rank_name = r.rank_name
	r.topN.rankdb = r.rankdb
}

func (r *RankByGuild) saveTopN() {
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

func (r *RankByGuild) setTopN(data [RankTopSize]GuildDataInRank) {
	r.topN.setTopN(data)
	r.saveTopN()
}

func (r *RankByGuild) getTopN() [RankTopSize]GuildDataInRank {
	// 注意外部调用前需要复制
	return r.topN.TopN
}

func (r *RankByGuild) getPos(acid string) int {
	return r.rankdb.getPos(r.rank_name, r.db_name, acid)
}

func (r *RankByGuild) add(acid string, data GuildDataInRank) {
	// 先存Redis, 计算topN以Redis为准
	// 用于确定排序原则 同分先进居上
	data.Score *= r.scorePowBase
	data.Score += r.scorePowBase - r.topN.CurrPowValue - 1
	r.topN.CurrPowValue += 1
	if r.topN.CurrPowValue >= r.scorePowBase {
		r.topN.CurrPowValue = 0
	}
	r.rankdb.add(r.rank_name, r.db_name, acid, data.getScore())
	r.topN.Add(acid, data)
	r.saveTopN()
}

func (r *RankByGuild) adds(acids []string, datas []*GuildDataInRank) {
	// 先存Redis, 计算topN以Redis为准
	// 用于确定排序原则 同分先进居上
	if len(acids) != len(datas) {
		logs.Error("RankByGuild len(acids) != len(datas)")
		return
	}
	for i, data := range datas {
		data.Score *= r.scorePowBase
		data.Score += r.scorePowBase - r.topN.CurrPowValue - 1
		r.topN.CurrPowValue += 1
		if r.topN.CurrPowValue >= r.scorePowBase {
			r.topN.CurrPowValue = 0
		}
		r.rankdb.add(r.rank_name, r.db_name, acids[i], data.getScore())
	}
	r.topN.Update()
	r.saveTopN()
}

func (r *RankByGuild) del(acid string) {
	r.rankdb.del(r.rank_name, r.db_name, acid)
	if r.topN.isInTopN(acid) {
		r.topN.Update()
		r.saveTopN()
	}
}

func (r *RankByGuild) rename() {

}

func (r *RankByGuild) setYesterdayRank(rank *RankByGuild) {
	r.yesterday_rank = rank
}

func (r *RankByGuild) setNeedBalanceTopN(is_clean bool, bf rankByGuildBalanceFunc) chan<- bool {
	r.isClean = is_clean
	r.balanceFunc = bf
	return r.balance_chan
}

func (r *RankByGuild) Start(rank_id int64, rank_name string, rankdb *rankDB, db_name string, getScore getGuildScoreFunc) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByGuild %s Panic, Err %v", rank_name, err)
		}
	}()
	r.getScore = getScore
	r.rankdb = rankdb
	r.rank_name = rank_name
	r.db_name = db_name
	r.rankId = rank_id
	r.scorePowBase = RankByGuildPowBase

	r.loadTopN()

	r.res_pos_chan = make(chan rankByGuildCommand, GetChannelSize)
	r.res_topN_chan = make(chan rankByGuildCommand, AddChannelSize)
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
			if command.res_chan == nil {
				logs.Warn("res_pos_chan command error")
				continue
			}

			command.res_chan <- RankByGuildGetRes{
				Pos: r.getPos(command.acid),
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
						switch command.typ {
						case rankByGuildCommandTyp_Adds:
							r.adds(command.acids, command.datas)
						case rankByGuildCommandTyp_Add:
							r.add(command.acid, command.data)
						case rankByGuildCommandTyp_Get:
							if command.res_chan != nil {
								logs.Trace("command %v", command)
								command.res_chan <- RankByGuildGetRes{
									TopN: r.getTopN(),
								}

							}
						case rankByGuildCommandTyp_Del:
							r.del(command.acid)
						case rankByGuildCommandTyp_Rename:
							r.rename()
						}
					}()
				}

			case is_need_balance := <-r.balance_chan:
				{
					//
					// 排行榜结算过程
					// 注意考虑到清榜之后要保留排名位置而清空积分
					// 所以榜中积分*1000做成定点数 后面的小数部分用来在清榜状态下保持排位
					// 发给客户端时/1000只保留整数
					// 清榜时由于不能将所有清零 所以只保留前1000名公会,
					// 先将现在的榜移到上周榜 再用现有的1000名填充至当前榜
					logs.Warn("Rank %s blance %v", r.rank_name, is_need_balance)

					top_uids, err := r.rankdb.getTopFromRedis(r.rank_name, r.db_name)

					if err == nil {
						for rank_1 := 0; (rank_1 < len(top_uids)) && (rank_1 < RankBalanceSize); rank_1++ {
							r.balanceFunc(rank_1+1, top_uids[rank_1])
						}

						if r.yesterday_rank != nil {
							r.yesterday_rank.setTopN(r.topN.TopN)
							logs.Warn("Rank %s blance %v Copy Start!", r.rank_name, is_need_balance)
							err := r.rankdb.copy(r.rank_name, r.yesterday_rank.rank_name, r.db_name)
							if err != nil {
								logs.Error("Rank %s blance %v Error by %s!",
									r.rank_name, is_need_balance, err.Error())
							}
							logs.Warn("Rank %s blance %v Clean Start!", r.rank_name, is_need_balance)
							r.rankdb.reName(r.rank_name, r.yesterday_rank.rank_name, r.db_name)
						} else {
							r.rankdb.reName(r.rank_name, r.rank_name+"Last", r.db_name)
						}

						for rank_1 := 0; (rank_1 < len(top_uids)) && (rank_1 < int(r.scorePowBase)); rank_1++ {
							r.rankdb.add(r.rank_name, r.db_name, top_uids[rank_1], (r.scorePowBase)-1-int64(rank_1))
						}

						r.topN.Update()
						r.saveTopN()

					} else {
						logs.Error("getTopFromRedis Err By %s", err.Error())
					}
				}
			}

		}
	}()
}

func (r *RankByGuild) Stop() {
	close(r.res_topN_chan)
	close(r.res_pos_chan)

	r.saveTopN()

	r.waitter.Wait()
}

func (r *RankByGuild) Get(acid string) *RankByGuildGetRes {
	res_chan := make(chan RankByGuildGetRes)
	getCommand := newGetRankByGuildCommand(acid, res_chan)
	r.res_pos_chan <- getCommand
	r.res_topN_chan <- getCommand

	logs.Trace("Rank Get Has Send")

	res_1 := <-res_chan
	logs.Trace("Rank Get res_1 %v", res_1)
	res_2 := <-res_chan
	logs.Trace("Rank Get res_2 %v", res_2)

	for i := 0; i < len(res_1.TopN); i++ {
		res_1.TopN[i].Score = res_1.TopN[i].Score / r.scorePowBase
	}

	for i := 0; i < len(res_2.TopN); i++ {
		res_2.TopN[i].Score = res_2.TopN[i].Score / r.scorePowBase
	}

	if res_2.TopN[0].Name != "" {
		res_2.Pos = res_1.Pos
		return &res_2
	} else {
		res_1.Pos = res_2.Pos
		return &res_1
	}
}

func (r *RankByGuild) Add(a *guild_info.GuildSimpleInfo) {
	logs.Trace("RankByGuild Add %v", *a)
	data := &GuildDataInRank{}
	data.SetDataGuild(a, r.getScore(a))
	r.res_topN_chan <- newAddRankByGuildCommand(a.GuildUUID, data)
	return
}

func (r *RankByGuild) Adds(as []*guild_info.GuildSimpleInfo) {
	logs.Trace("RankByGuild Adds %v", as)
	acids := make([]string, len(as))
	datas := make([]*GuildDataInRank, len(as))
	for i, a := range as {
		acids[i] = a.GuildUUID
		datas[i] = &GuildDataInRank{}
		datas[i].SetDataGuild(a, r.getScore(a))
	}
	r.res_topN_chan <- newAddRankByGuildsCommand(acids, datas)
	return
}

func (r *RankByGuild) GetInfoInRank(a *guild_info.GuildSimpleInfo) *GuildDataInRank {
	c := &GuildDataInRank{}
	if a != nil {
		c.SetDataGuild(a, r.getScore(a))
	}
	return c
}

func (r *RankByGuild) Del(guildID string) {
	logs.Trace("RankByGuild Del %v", guildID)
	data := &GuildDataInRank{}
	data.UUID = guildID
	r.res_topN_chan <- newDelRankByGuildCommand(guildID)
	return
}
