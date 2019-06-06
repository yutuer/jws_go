package rank

import (
	"sync"

	"time"

	"fmt"
	"runtime"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 用于确定排序原则 同分先进居上
const RankByCorpDelayPowBase = 10000

const (
	rankByCorpDelayCommandTypNull = iota
	rankByCorpDelayCommandTypGet
	rankByCorpDelayCommandTypAdd
	rankByCorpDelayCommandTypAddInInit
	rankByCorpDelayCommandTypSetTopN
	rankByCorpDelayCommandTypMkTopN
	rankByCorpDelayCommandTypBalance
)

type rankByCorpDelayCommand struct {
	typ        int                     // 是否是要获取acid对应的排行榜信息
	acid       string                  // 账号id
	data       CorpDataInRank          // 更新数据
	datas      []CorpDataInRank        // 更新数据
	acid2score map[string]PairPosScore // 排行榜全部数据镜像
	res_chan   chan RankByCorpGetRes   // 获取用的回复channel
}

type rankByCorpDelayTopNCommand struct {
	typ        int                     // 是否是要获取acid对应的排行榜信息
	acid       string                  // 账号id
	topNValid  bool                    //
	topN       []CorpDataInRank        //
	acid2score map[string]PairPosScore // 排行榜全部数据镜像
	data       CorpDataInRank          // 更新数据
	res_chan   chan RankByCorpGetRes   // 获取用的回复channel
}

type RankByCorpDelay struct {
	topN        CorpTopNRedis
	topNInMake  CorpTopNRedis
	topNChanged bool

	getScoreF getScoreFunc // 通过account获取分数

	res_mkTopN_chan chan rankByCorpDelayCommand
	res_posIO_chan  chan rankByCorpDelayCommand
	res_topNIO_chan chan rankByCorpDelayTopNCommand

	scoreChangeReqs map[string]CorpDataInRank
	lastSendTime    int64

	waitter   sync.WaitGroup
	rankdb    *rankDB
	rank_name string
	db_name   string

	// for balance
	isNeedBalance bool
	isClean       bool
	balanceFunc   rankByCorpFromCacheBalanceFunc
	balance_chan  chan bool

	scorePowBase int64

	rankId int64
	// for balance
}

func (r *RankByCorpDelay) delRank(acID string) error {
	logs.Debug("del rank, acid: %v", acID)
	r.rankdb.del(r.rank_name, r.db_name, acID)
	return nil
}

// for balance TODO By Fanyang 整理Rank通用接口, 将balance的实现提出去
func (r *RankByCorpDelay) setNeedBalanceTopN(is_clean bool, bf rankByCorpFromCacheBalanceFunc) chan<- bool {
	r.isClean = is_clean
	r.isNeedBalance = true
	r.balanceFunc = bf
	return r.balance_chan
}

func (r *RankByCorpDelay) transBalance() {
	// 为了保持接口一致
	r.balance_chan = make(chan bool, 1)
	go func() {
		for {
			is_need_balance := <-r.balance_chan
			if is_need_balance && r.isNeedBalance {
				logs.Warn("RankByCorpDelay  <-r.balance_chan %s blance", r.rank_name)
				r.res_topNIO_chan <- rankByCorpDelayTopNCommand{
					typ: rankByCorpDelayCommandTypBalance,
				}
			}
		}
	}()
}

func (r *RankByCorpDelay) balance() {
	logs.Warn("Rank %s blance", r.rank_name)
	r.balanceFunc(r.topN.TopN, r.topN.acid2PosScoreCache)
}

// for balance

func (r *RankByCorpDelay) loadTopN() {
	res, err := r.rankdb.loadTopN(r.rank_name, r.db_name)
	if err != nil && err != redis.ErrNil {
		logs.Error("loadTopN %s Err by %s", r.rank_name, err.Error())
		return
	}
	if err == redis.ErrNil {
		return
	}
	r.topN.fromDB(res)
	r.topNInMake.fromDB(res)
}

func (r *RankByCorpDelay) saveTopN() {
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

	r.topNChanged = false
}

func (r *RankByCorpDelay) getTopN() [RankTopSize]CorpDataInRank {
	// 注意外部调用前需要复制
	logs.Trace("getTopN %v", r.topN.TopN)
	re := r.topN.TopN
	return re
}

func (r *RankByCorpDelay) setTopN(top [RankTopSize]CorpDataInRank) {
	// 注意外部调用前需要复制
	//logs.Trace("setTopN %v", top)
	r.topN.setTopN(top)
}

func (r *RankByCorpDelay) setAcid2ScoreCache(acid2ScoreCache map[string]PairPosScore) {
	r.topN.setAcid2PosScoreCache(acid2ScoreCache)
}

func (r *RankByCorpDelay) getPosScoreFromCache(acid string) PairPosScore {
	return r.topN.acid2PosScoreCache[acid]
}

func (r *RankByCorpDelay) getPos(acid string) int {
	return r.rankdb.getPos(r.rank_name, r.db_name, acid)
}

func (r *RankByCorpDelay) getScore(acid string) (int64, error) {
	return r.rankdb.getScore(r.rank_name, r.db_name, acid)
}

/*
	score 计算方法
	score double 有效精度16位
	前8位表示战力， 后8未记录时间的倒数 (1/time, 保证时间越大， 值越低)
	限制: 战力最大9999万， 时间精度 22sec
*/
func (r *RankByCorpDelay) tick() {
	//
	nowT := time.Now().Unix()
	if len(r.scoreChangeReqs) > 0 && (len(r.scoreChangeReqs) > 30 || (nowT-r.lastSendTime) > 30) {
		r.lastSendTime = nowT
		datas := make([]CorpDataInRank, 0, len(r.scoreChangeReqs))
		for _, c := range r.scoreChangeReqs {
			// 用于确定排序原则 同分先进居上
			c.Score += r.scorePowBase - r.topN.CurrPowValue - 1
			r.topN.CurrPowValue += 1
			if r.topN.CurrPowValue >= r.scorePowBase {
				r.topN.CurrPowValue = 0
			}
			datas = append(datas, c)
		}

		c := rankByCorpDelayCommand{
			typ:   rankByCorpDelayCommandTypMkTopN,
			datas: datas[:],
		}

		a2s := r.mkAcid2PosScoreCache()
		if a2s != nil {
			c.acid2score = a2s
		}

		r.res_mkTopN_chan <- c
		r.sendToRedis(datas)
	}
}

const MAX_PARAMS_TO_REDIS = 64

func (r *RankByCorpDelay) sendToRedis(data []CorpDataInRank) {
	//logs.Trace("RankByCorpDelay sendToRedis %v", data)
	params := make([]interface{}, 0, MAX_PARAMS_TO_REDIS*2+1)
	params = append(params, "") // For First db name
	for i := 0; i < len(data); i++ {
		params = append(params, data[i].getScore())
		params = append(params, data[i].getId())
		if len(params) >= MAX_PARAMS_TO_REDIS*2+1 {
			r.rankdb.adds(r.rank_name, r.db_name, params)
			params = make([]interface{}, 0, MAX_PARAMS_TO_REDIS*2+1)
			params = append(params, "") // For First db name
			runtime.Gosched()
		}
	}

	if len(params) > 1 {
		bs := time.Now().UnixNano()
		r.rankdb.adds(r.rank_name, r.db_name, params)
		metricsSend(r.rank_name, fmt.Sprintf("%d", time.Now().UnixNano()-bs))
	}
}

func calcScore(scoreParam int64, baseScore int64) float64 {
	now := time.Now().Unix()
	return float64(scoreParam-scoreParam%baseScore) + 1e9*float64(baseScore)/float64(now)
}

func (r *RankByCorpDelay) add(acid string, data CorpDataInRank) {
	// 注意这里由于会减少是变动值
	logs.Trace("RankByCorpDelay add %v to %d", data, data.Score)
	od, ok := r.scoreChangeReqs[acid]
	if ok {
		// 合并两次请求

		ndata := data
		ndata.scoreDeta = od.scoreDeta + data.scoreDeta
		ndata.Score = data.Score
		logs.Trace("add RankByCorpDelay req %v + %v -> %v", od, data, ndata)
		r.scoreChangeReqs[acid] = ndata
	} else {
		r.scoreChangeReqs[acid] = data
	}
}

func (r *RankByCorpDelay) mkTopN(acid string, data []CorpDataInRank) bool {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByCorpDelay mkTopN Panic, Err %v", err)
		}
	}()

	// 注意外部调用前需要复制
	//logs.Trace("mkTopN %s %v %v", acid, r.topNInMake.TopN, data)

	res := false
	for i := 0; i < len(data); i++ {
		d := data[i]
		if d.scoreDeta >= 0 {
			res = res || r.topNInMake.isTopN(d.getScore()) // 注意如果topN中有人降分的话一定会触发更新,所以这里的标准是之前的miniscore
		}
		res = res || r.topNInMake.isHasInTopN(d.getId()) > 0
	}

	return res
}

func (r *RankByCorpDelay) mkAcid2PosScoreCache() map[string]PairPosScore {
	acids, scores, err := r.rankdb.getWithScoreFromRedis(r.rank_name, r.db_name)
	if err != nil {
		logs.Error("RankByCorpDelay tick getWithScoreFromRedis err %v", err)
		return nil
	}
	acid2score := make(map[string]PairPosScore, len(acids))
	for idx, acid := range acids {
		score := scores[idx]
		acid2score[acid] = PairPosScore{
			Pos:   idx + 1,
			Score: score,
		}
	}
	return acid2score
}

func (r *RankByCorpDelay) Start(rank_id int64, rank_name string, rankdb *rankDB, db_name string, getScore getScoreFunc) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByCorpDelay Panic, Err %v", err)
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			logs.Error("[GameCHANSerever] Stack of %d bytes: %s\n", count, trace)
			panic(fmt.Errorf("RankByCorpDelay Start Err %v", err))
		}
	}()

	r.rankdb = rankdb
	r.rank_name = rank_name
	r.db_name = db_name
	r.rankId = rank_id

	r.topN.rankdb = rankdb
	r.topN.rank_name = rank_name
	r.topN.db_name = db_name

	r.topNInMake.rankdb = rankdb
	r.topNInMake.rank_name = rank_name
	r.topNInMake.db_name = db_name

	r.scorePowBase = RankByCorpDelayPowBase

	r.getScoreF = getScore
	r.scoreChangeReqs = make(map[string]CorpDataInRank, 128)

	r.loadTopN()
	r.topN.Update()
	r.topN.setAcid2PosScoreCache(r.mkAcid2PosScoreCache())
	r.topNInMake.Update()

	r.res_posIO_chan = make(chan rankByCorpDelayCommand, AddChannelSize)
	r.res_topNIO_chan = make(chan rankByCorpDelayTopNCommand, AddChannelSize)
	r.res_mkTopN_chan = make(chan rankByCorpDelayCommand, AddChannelSize)

	r.transBalance()

	// 分数更新协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		timerChan := uutil.TimerMS.After(time.Second)
		for {
			select {
			case command, ok := <-r.res_posIO_chan:
				if !ok {
					logs.Warn("res_pos_chan close")
					return
				}

				switch command.typ {
				case rankByCorpDelayCommandTypGet:
					if command.res_chan != nil {
						// rank只会发两个信息给这个chann
						s, _ := r.getScore(command.acid)
						command.res_chan <- RankByCorpGetRes{
							TopN:  make([]CorpDataInRank, RankTopSize),
							Pos:   r.getPos(command.acid),
							Score: s,
						}
					}
				case rankByCorpDelayCommandTypAdd:
					r.add(command.acid, command.data)
				default:
					logs.Warn("res_mkTopN_chan typ Err by %d", command.typ)
				}
			case <-timerChan:
				r.tick()
				timerChan = uutil.TimerMS.After(time.Second)
			}
		}
	}()

	// TopN数据生成协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.res_mkTopN_chan
			if !ok {
				logs.Warn("res_mkTopN_chan close")
				return
			}

			switch command.typ {
			case rankByCorpDelayCommandTypMkTopN:
				c := rankByCorpDelayTopNCommand{
					typ:        rankByCorpDelayCommandTypSetTopN,
					acid2score: command.acid2score,
				}
				is_need_update := r.mkTopN(command.acid, command.datas)
				if is_need_update {
					r.topNInMake.Update()
					c.topNValid = true
					c.topN = r.topNInMake.TopN[:]
				}
				r.res_topNIO_chan <- c
			default:
				logs.Warn("res_mkTopN_chan typ Err by %d", command.typ)
			}
		}
	}()

	// TopN IO 协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.res_topNIO_chan
			//logs.Trace("command %v", command)

			if !ok {
				logs.Warn("res_topNIO_chan close")
				return
			}

			r.waitter.Add(1)
			func() {
				defer r.waitter.Done()
				switch command.typ {
				case rankByCorpDelayCommandTypGet:
					if command.res_chan != nil {
						// rank只会发两个信息给这个chann
						ps := r.getPosScoreFromCache(command.acid)
						topN := r.getTopN()
						command.res_chan <- RankByCorpGetRes{
							TopN:  topN[:],
							Pos:   ps.Pos,
							Score: ps.Score,
						}
					}
				case rankByCorpDelayCommandTypSetTopN:
					if command.topNValid {
						r.setTopN(s2a(command.topN))
					}
					r.setAcid2ScoreCache(command.acid2score)
				case rankByCorpDelayCommandTypBalance:
					logs.Warn("RankByCorpDelay rankByCorpDelayCommandTypBalance %s blance", r.rank_name)
					r.balance()
				default:
					logs.Warn("res_topN_chan command error")
				}
			}()

		}
	}()
}

func (r *RankByCorpDelay) Stop() {
	close(r.res_mkTopN_chan)
	close(r.res_posIO_chan)
	close(r.res_topNIO_chan)

	r.saveTopN()
	r.waitter.Wait()
}

func (r *RankByCorpDelay) Get(acid string) *RankByCorpGetRes {
	res_chan := make(chan RankByCorpGetRes, 1) // rank只会发两个信息给这个chann
	getTopNCommand := rankByCorpDelayTopNCommand{
		typ:      rankByCorpDelayCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res := r.execTopNCmd(getTopNCommand)

	for i := 0; i < len(res.TopN); i++ {
		res.TopN[i].Score = res.TopN[i].Score / r.scorePowBase
	}
	res.Score = res.Score / r.scorePowBase

	return res
}

func (r *RankByCorpDelay) GetPos(acid string) (int, int64) {
	res_chan := make(chan RankByCorpGetRes, 1) // rank只会发两个信息给这个chann
	getCommand := rankByCorpDelayCommand{
		typ:      rankByCorpDelayCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res_1 := r.execCmd(getCommand)

	logs.Trace("GetPos res_1 %v", res_1)
	return res_1.Pos, res_1.Score / r.scorePowBase
}

func (r *RankByCorpDelay) Add(a *helper.AccountSimpleInfo, score, scoreOld int64) {
	if len(a.GsHeroIds) < helper.CorpHeroGsNum {
		return
	}
	data := CorpDataInRank{}
	data.setDataFromAccountDeta(a, (score-scoreOld)*r.scorePowBase)
	data.Score = score * r.scorePowBase

	r.execCmdASync(rankByCorpDelayCommand{
		typ:  rankByCorpDelayCommandTypAdd,
		acid: a.AccountID,
		data: data,
	})
	return
}

func (r *RankByCorpDelay) GetTopNAccountID(n int) ([]string, error) {
	return r.rankdb.GetTopNFromRedis(r.rank_name, r.db_name, n)
}

func (r *RankByCorpDelay) GetCorpInfo(a *helper.AccountSimpleInfo) *CorpDataInRank {
	c := &CorpDataInRank{}
	c.SetDataFromAccount(a, r.getScoreF(a))
	return c
}

func (r *RankByCorpDelay) RebaseScore(redisScore int64) int64 {
	return redisScore / r.scorePowBase
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDelay) execCmdASync(cmd rankByCorpDelayCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDelay execCmd put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDelay) execCmd(cmd rankByCorpDelayCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDelay execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpDelay <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDelay) execTopNCmd(cmd rankByCorpDelayTopNCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topNIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDelay TopN execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpDelay TopN <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}
