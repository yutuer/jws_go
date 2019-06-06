package rank

import (
	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	rankByCorpCommandTypNull = iota
	rankByCorpCommandTypGet
	rankByCorpCommandTypAdd
	rankByCorpCommandTypAddInInit
	rankByCorpCommandTypSetTopN
	rankByCorpCommandTypMkTopN
	rankByCorpCommandTypBalance
	rankByCorpCommandTypeRename // 排行榜上有人改名，重新生成topN
)

type rankByCorpDynamicCommand struct {
	typ      int                   // 是否是要获取acid对应的排行榜信息
	acid     string                // 账号id
	data     CorpDataInRank        // 更新数据
	res_chan chan RankByCorpGetRes // 获取用的回复channel
}

type rankByCorpDynamicTopNCommand struct {
	typ      int                   // 是否是要获取acid对应的排行榜信息
	acid     string                // 账号id
	topN     []CorpDataInRank      //
	data     CorpDataInRank        // 更新数据
	res_chan chan RankByCorpGetRes // 获取用的回复channel
}

type RankByCorpDynamic struct {
	topN        CorpTopNRedis
	topNInMake  CorpTopNRedis
	topNChanged bool

	res_mkTopN_chan chan rankByCorpDynamicCommand
	res_posIO_chan  chan rankByCorpDynamicCommand
	res_topNIO_chan chan rankByCorpDynamicTopNCommand

	waitter  sync.WaitGroup
	rankDB   *rankDB
	rankName string
	dbName   string

	// for balance
	isNeedBalance bool
	isClean       bool
	// 0 是日结算, 1是周结算
	balanceFunc rankByCorpBalanceBatchFunc
	balanceChan chan bool

	scorePowBase int64

	rankId int64
	// for balance

	initScore int64
}

func (r *RankByCorpDynamic) delRank(acID string) error {
	logs.Debug("del rank, acid: %v", acID)
	r.rankDB.del(r.rankName, r.dbName, acID)
	r.res_mkTopN_chan <- rankByCorpDynamicCommand{
		typ:  rankByCorpCommandTypMkTopN,
		acid: acID,
		data: CorpDataInRank{
			scoreDeta: -1,
		},
	}
	return nil
}

// for balance TODO By Fanyang 整理Rank通用接口, 将balance的实现提出去
func (r *RankByCorpDynamic) setNeedBalanceTopN(is_clean bool, bf rankByCorpBalanceBatchFunc) chan<- bool {
	r.isClean = is_clean
	r.isNeedBalance = true
	r.balanceFunc = bf
	return r.balanceChan
}

func (r *RankByCorpDynamic) transBalance() {
	// 为了保持接口一致
	r.balanceChan = make(chan bool, 1)
	go func() {
		for {
			is_need_balance := <-r.balanceChan
			if is_need_balance && r.isNeedBalance {
				logs.Warn("RankByCorpDynamic  <-r.balance_chan[0] %s blance", r.rankName)
				r.res_posIO_chan <- rankByCorpDynamicCommand{
					typ: rankByCorpCommandTypBalance,
				}
			}
		}
	}()
}

func (r *RankByCorpDynamic) balance() {
	logs.Warn("Rank %s blance", r.rankName)
	top_uids, err := r.rankDB.getTopFromRedis(r.rankName, r.dbName)

	if err == nil {
		r.balanceFunc(top_uids)
	} else {
		logs.Error("getTopFromRedis Err By %s", err.Error())
	}
}

// for balance

func (r *RankByCorpDynamic) loadTopN() {
	res, err := r.rankDB.loadTopN(r.rankName, r.dbName)
	if err != nil && err != redis.ErrNil {
		logs.Error("loadTopN %s Err by %s", r.rankName, err.Error())
		return
	}
	if err == redis.ErrNil {
		return
	}
	r.topN.fromDB(res)
	r.topNInMake.fromDB(res)
}

func (r *RankByCorpDynamic) saveTopN() {
	data, err := r.topN.toDB()
	if err != nil {
		logs.Error("saveTopN %s Err by %s", r.rankName, err.Error())
		return
	}

	err = r.rankDB.saveTopN(r.rankName, r.dbName, data)

	if err != nil {
		logs.Error("rankdb saveTopN %s Err by %s", r.rankName, err.Error())
		return
	}

	r.topNChanged = false
}

func (r *RankByCorpDynamic) getTopN() [RankTopSize]CorpDataInRank {
	// 注意外部调用前需要复制
	re := r.topN.TopN
	return re
}

func (r *RankByCorpDynamic) setTopN(top [RankTopSize]CorpDataInRank) {
	// 注意外部调用前需要复制
	r.topN.setTopN(top)
}

func (r *RankByCorpDynamic) getPos(acid string) int {
	return r.rankDB.getPos(r.rankName, r.dbName, acid)
}

func (r *RankByCorpDynamic) getScore(acid string) (int64, error) {
	return r.rankDB.getScore(r.rankName, r.dbName, acid)
}

func (r *RankByCorpDynamic) add(acid string, data CorpDataInRank) {
	// 注意这里由于会减少是变动值
	id_in_db := data.getId()

	// 用于确定排序原则 同分先进居上
	data.scoreDeta += r.scorePowBase - r.topN.CurrPowValue - 1
	r.topN.CurrPowValue += 1
	if r.topN.CurrPowValue >= r.scorePowBase {
		r.topN.CurrPowValue = 0
	}

	// 先要获取当前redis里面的分数
	score, err := r.rankDB.getScore(
		r.rankName,
		r.dbName,
		id_in_db)

	if err != nil {
		logs.Error("[%s]RankByCorpDynamic getScore Err %s In %v",
			acid, err.Error(), data)
		return
	}

	new_score := score + data.scoreDeta
	data.setScore(new_score)
	r.rankDB.add(r.rankName, r.dbName, id_in_db, new_score)
	r.res_mkTopN_chan <- rankByCorpDynamicCommand{
		typ:  rankByCorpCommandTypMkTopN,
		acid: acid,
		data: data,
	}
}

func (r *RankByCorpDynamic) mkTopN(acid string, data CorpDataInRank) bool {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByCorpDynamic mkTopN Panic, Err %v", err)
		}
	}()

	// 注意外部调用前需要复制
	if data.scoreDeta >= 0 {
		return r.topNInMake.isTopN(data.getScore())
	} else if data.scoreDeta < 0 {
		return r.topNInMake.isHasInTopN(acid) > 0 || r.topNInMake.isTopN(data.getScore())
	}

	return true
}

func (r *RankByCorpDynamic) Start(rank_name string, rankdb *rankDB, db_name string) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByCorpDynamic Panic, Err %v", err)
		}
	}()

	r.rankDB = rankdb
	r.rankName = rank_name
	r.dbName = db_name

	r.topN.rankdb = rankdb
	r.topN.rank_name = rank_name
	r.topN.db_name = db_name

	r.topNInMake.rankdb = rankdb
	r.topNInMake.rank_name = rank_name
	r.topNInMake.db_name = db_name

	r.scorePowBase = RankByCorpDelayPowBase

	r.loadTopN()
	r.topN.Update()
	r.topNInMake.Update()

	r.res_posIO_chan = make(chan rankByCorpDynamicCommand, AddChannelSize)
	r.res_topNIO_chan = make(chan rankByCorpDynamicTopNCommand, AddChannelSize)
	r.res_mkTopN_chan = make(chan rankByCorpDynamicCommand, AddChannelSize)

	r.initScore = SimplePvpInitScore

	r.transBalance()

	// 分数更新协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.res_posIO_chan
			if !ok {
				logs.Warn("res_pos_chan close")
				return
			}
			switch command.typ {
			case rankByCorpCommandTypGet:
				if command.res_chan != nil {
					// rank只会发两个信息给这个chann
					s, _ := r.getScore(command.acid)
					command.res_chan <- RankByCorpGetRes{
						TopN:  make([]CorpDataInRank, RankTopSize),
						Pos:   r.getPos(command.acid),
						Score: s,
					}
				}
			case rankByCorpCommandTypAdd:
				r.add(command.acid, command.data)
			case rankByCorpCommandTypBalance:
				logs.Warn("RankByCorpDynamic rankByCorpCommandTypBalance %s blance",
					r.rankName)
				r.balance()
			default:
				logs.Warn("res_mkTopN_chan typ Err by %d", command.typ)
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
			case rankByCorpCommandTypMkTopN:
				is_need_update := r.mkTopN(command.acid, command.data)
				if is_need_update {
					r.topNInMake.Update()
					r.res_topNIO_chan <- rankByCorpDynamicTopNCommand{
						typ:  rankByCorpCommandTypSetTopN,
						topN: r.topNInMake.TopN[:],
					}
				}
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

			if !ok {
				logs.Warn("res_topNIO_chan close")
				return
			}

			r.waitter.Add(1)
			func() {
				defer r.waitter.Done()
				switch command.typ {
				case rankByCorpCommandTypGet:
					if command.res_chan != nil {
						// rank只会发两个信息给这个chann
						topN := r.getTopN()
						command.res_chan <- RankByCorpGetRes{
							TopN: topN[:],
						}
					}
				case rankByCorpCommandTypSetTopN:
					r.setTopN(s2a(command.topN))
				case rankByCorpCommandTypeRename:
					r.topN.Rename(command.acid, command.data.Name)
				default:
					logs.Warn("res_topN_chan command error")
				}
			}()

		}
	}()
}

func (r *RankByCorpDynamic) Stop() {
	close(r.res_mkTopN_chan)
	close(r.res_posIO_chan)
	close(r.res_topNIO_chan)

	r.saveTopN()

	r.waitter.Wait()
}

func (r *RankByCorpDynamic) Get(acid string) *RankByCorpGetRes {
	res_chan := make(chan RankByCorpGetRes, 2) // rank只会发两个信息给这个chann
	getCommand := rankByCorpDynamicCommand{
		typ:      rankByCorpCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	getTopNCommand := rankByCorpDynamicTopNCommand{
		typ:      rankByCorpCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res_1 := r.execCmd(getCommand)
	res_2 := r.execTopNCmd(getTopNCommand)

	var res *RankByCorpGetRes

	if len(res_2.TopN) > 0 && res_2.TopN[0].Name != "" {
		res_2.Pos = res_1.Pos
		res_2.Score = res_1.Score
		res = res_2
	} else {
		res_1.Pos = res_2.Pos
		res_1.Score = res_2.Score
		res = res_1
	}

	for i := 0; i < len(res.TopN); i++ {
		res.TopN[i].Score = r.initScore + res.TopN[i].Score/r.scorePowBase
	}
	res.Score = r.initScore + res.Score/r.scorePowBase

	return res
}

func (r *RankByCorpDynamic) GetPos(acid string) (int, int64) {
	res_chan := make(chan RankByCorpGetRes, 1) // rank只会发两个信息给这个chann
	getCommand := rankByCorpDynamicCommand{
		typ:      rankByCorpCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res_1 := r.execCmd(getCommand)

	return res_1.Pos, r.initScore + res_1.Score/r.scorePowBase
}

func (r *RankByCorpDynamic) AddDeta(a *helper.AccountSimpleInfo, score_d int64) {
	data := CorpDataInRank{}
	data.setDataFromAccountDeta(a, score_d*r.scorePowBase)

	r.execCmdASync(rankByCorpDynamicCommand{
		typ:  rankByCorpCommandTypAdd,
		acid: a.AccountID,
		data: data,
	})
	return
}

func (r *RankByCorpDynamic) AddEnemyDeta(info *helper.AccountSimpleInfo, score_d int64) {
	data := CorpDataInRank{}
	data.setDataFromAccountDeta(info, score_d*r.scorePowBase)

	r.execCmdASync(rankByCorpDynamicCommand{
		typ:  rankByCorpCommandTypAdd,
		acid: info.AccountID,
		data: data,
	})
	return
}

func (r *RankByCorpDynamic) GetTopNAccountID(n int) ([]string, error) {
	return r.rankDB.GetTopNFromRedis(r.rankName, r.dbName, n)
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDynamic) execCmdASync(cmd rankByCorpDynamicCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic execCmd put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDynamic) execCmd(cmd rankByCorpDynamicCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByCorpDynamic) execTopNCmd(cmd rankByCorpDynamicTopNCommand) *RankByCorpGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topNIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic topN execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic topN <-res_chan timeout %v", cmd)
		return &RankByCorpGetRes{}
	}
}

func (r *RankByCorpDynamic) execTopNCmdAsyc(cmd rankByCorpDynamicTopNCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topNIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByCorpDynamic topN execCmd put timeout %v", cmd)
	}
}

func (r *RankByCorpDynamic) Clean() {
	error := r.rankDB.delKey(r.rankName, r.dbName)
	if error != nil {
		logs.Error("RankByC	orpDynamic clean db failed")
	}
	error = r.rankDB.delKey(r.rankName+":topN", r.dbName)
	if error != nil {
		logs.Error("RankByCorpDynamic clean db failed")
	}
	r.topN.clean()
	r.topNInMake.clean()
	r.loadTopN()
}

func (r *RankByCorpDynamic) OnChangeName(simpleInfo *helper.AccountSimpleInfo) {
	// 这里只需要修改topN即可， 每当排行变化时， 会重新生成一下榜单
	r.execTopNCmdAsyc(rankByCorpDynamicTopNCommand{
		typ:  rankByCorpCommandTypeRename,
		acid: simpleInfo.AccountID,
		data: CorpDataInRank{Name: simpleInfo.Name},
	})
}
