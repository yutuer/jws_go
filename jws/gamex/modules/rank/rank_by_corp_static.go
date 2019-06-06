package rank

import (
	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/platform/planx/util"
)

type rankByCorpStaticCommand struct {
	is_get   bool                  // 是否是要获取acid对应的排行榜信息
	acid     string                // 账号id
	TopN     []CorpDataInRank      // 更新Top数据
	res_chan chan RankByCorpGetRes // 获取用的回复channel
}

func newSetTopKRankByCorpStaticCommand(rank_data [RankTopSize]CorpDataInRank) rankByCorpStaticCommand {
	return rankByCorpStaticCommand{
		is_get: false,
		TopN:   rank_data[:],
	}
}

func newGetRankByCorpStaticCommand(acid string, res_chan chan RankByCorpGetRes) rankByCorpStaticCommand {
	return rankByCorpStaticCommand{
		is_get:   true,
		acid:     acid,
		res_chan: res_chan,
	}
}

type RankByCorpStatic struct {
	topN     CorpTopN
	getScore getScoreFunc // 通过account获取分数

	res_pos_chan  chan rankByCorpStaticCommand
	res_topN_chan chan rankByCorpStaticCommand

	waitter   sync.WaitGroup
	rankdb    *rankDB
	rank_name string
	db_name   string

	scorePowBase int64
}

func (r *RankByCorpStatic) loadTopN() {
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

func (r *RankByCorpStatic) saveTopN() {
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

func (r *RankByCorpStatic) getTopN() [RankTopSize]CorpDataInRank {
	// 注意外部调用前需要复制
	return r.topN.TopN
}

func (r *RankByCorpStatic) getPos(acid string) int {
	return r.rankdb.getPos(r.rank_name, r.db_name, acid)
}

func (r *RankByCorpStatic) GetCorpInfo(a *helper.AccountSimpleInfo) *CorpDataInRank {
	c := &CorpDataInRank{}
	c.SetDataFromAccount(a, r.getScore(a))
	return c
}

func (r *RankByCorpStatic) setTopN(data [RankTopSize]CorpDataInRank) {
	r.topN.setTopN(data)
	r.saveTopN()
}

func (r *RankByCorpStatic) setRank(acid2PosScore map[string]PairPosScore) {
	params := make([]interface{}, 0, MAX_PARAMS_TO_REDIS*2+1)
	params = append(params, "") // For First db name
	for acid, ps := range acid2PosScore {
		params = append(params, ps.Score)
		params = append(params, acid)
		if len(params) >= MAX_PARAMS_TO_REDIS*2+1 {
			r.rankdb.adds(r.rank_name, r.db_name, params)
			params = make([]interface{}, 0, MAX_PARAMS_TO_REDIS*2+1)
			params = append(params, "") // For First db name
		}
	}
}

func (r *RankByCorpStatic) Start(rank_id int64, rank_name string, rankdb *rankDB, db_name string, getScore getScoreFunc) {
	r.scorePowBase = RankByCorpDelayPowBase
	r.getScore = getScore
	r.rankdb = rankdb
	r.rank_name = rank_name
	r.db_name = db_name

	r.loadTopN()

	r.res_pos_chan = make(chan rankByCorpStaticCommand, GetChannelSize)
	r.res_topN_chan = make(chan rankByCorpStaticCommand, AddChannelSize)

	r.waitter.Add(2)
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

	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.res_topN_chan

			if !ok {
				logs.Warn("res_pos_chan close")
				return
			}

			r.waitter.Add(1)
			func() {
				defer r.waitter.Done()
				if command.is_get && command.res_chan != nil {
					topN := r.getTopN()
					command.res_chan <- RankByCorpGetRes{
						TopN: topN[:],
					}
					return
				}

				if !command.is_get {
					r.setTopN(s2a(command.TopN))
					return
				}
				logs.Warn("res_topN_chan command error")
			}()

		}
	}()
}

func (r *RankByCorpStatic) Stop() {
	close(r.res_topN_chan)
	close(r.res_pos_chan)

	r.saveTopN()

	r.waitter.Wait()
}

func (r *RankByCorpStatic) Get(acid string) *RankByCorpGetRes {
	res_chan := make(chan RankByCorpGetRes)
	getCommand := newGetRankByCorpStaticCommand(acid, res_chan)
	res_1 := r.execCmd(getCommand)
	res_2 := r.execTopNCmd(getCommand)

	var res *RankByCorpGetRes
	if res_2.TopN[0].Name != "" {
		res_2.Pos = res_1.Pos
		res = res_2
	} else {
		res_1.Pos = res_2.Pos
		res = res_1
	}

	for i := 0; i < len(res.TopN); i++ {
		res.TopN[i].Score = res.TopN[i].Score / r.scorePowBase
	}
	res.Score = res.Score / r.scorePowBase
	return res
}

func (r *RankByCorpStatic) SetTopN(data [RankTopSize]CorpDataInRank) {
	r.execTopNCmdASync(newSetTopKRankByCorpStaticCommand(data))
	return
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpStatic) execCmdASync(cmd rankByCorpStaticCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_pos_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpStatic execCmd put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpStatic) execCmd(cmd rankByCorpStaticCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_pos_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpStatic execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpStatic <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpStatic) execTopNCmdASync(cmd rankByCorpStaticCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topN_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpStatic topN execCmd put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpStatic) execTopNCmd(cmd rankByCorpStaticCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topN_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpStatic topN execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpStatic topN <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}
