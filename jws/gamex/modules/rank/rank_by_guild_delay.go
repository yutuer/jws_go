package rank

import (
	"sync"

	"time"

	"fmt"
	"runtime"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	rankByGuildDelayCommandTypNull = iota
	rankByGuildDelayCommandTypGet
	rankByGuildDelayCommandTypAdd
	rankByGuildDelayCommandTypAddInInit
	rankByGuildDelayCommandTypSetTopN
	rankByGuildDelayCommandTypMkTopN
	rankByGuildDelayCommandTypBalance
	rankByGuildDelayCommandTypDel
	rankByGuildDelayCommandTypRename
)

type rankByGuildDelayCommand struct {
	typ            int                    // 是否是要获取acid对应的排行榜信息
	acid           string                 // 账号id
	data           GuildDataInRank        // 更新数据
	datas          []GuildDataInRank      // 更新数据
	needUpdateTopN bool                   // 强制更新TopN
	res_chan       chan RankByGuildGetRes // 获取用的回复channel
}

type rankByGuildDelayTopNCommand struct {
	typ      int                          // 是否是要获取acid对应的排行榜信息
	acid     string                       // 账号id
	topN     [RankTopSize]GuildDataInRank //
	data     GuildDataInRank              // 更新数据
	res_chan chan RankByGuildGetRes       // 获取用的回复channel
}

type RankByGuildDelay struct {
	sid         uint
	topN        GuildTopN
	topNInMake  GuildTopN
	topNChanged bool

	getScoreF getGuildScoreFunc // 通过account获取分数

	res_mkTopN_chan chan rankByGuildDelayCommand
	res_posIO_chan  chan rankByGuildDelayCommand
	res_getPos_chan chan rankByGuildDelayCommand
	res_topNIO_chan chan rankByGuildDelayTopNCommand

	scoreChangeReqs map[string]GuildDataInRank
	lastSendTime    int64

	waitter   sync.WaitGroup
	rankdb    *rankDB
	rank_name string
	db_name   string

	// for balance
	isNeedBalance bool
	isClean       bool
	balanceFunc   rankByGuildFromCacheBalanceFunc
	balance_chan  chan bool

	scorePowBase int64

	rankId int64
	// for balance
}

// for balance TODO By Fanyang 整理Rank通用接口, 将balance的实现提出去
func (r *RankByGuildDelay) setNeedBalanceTopN(is_clean bool, bf rankByGuildFromCacheBalanceFunc) chan<- bool {
	r.isClean = is_clean
	r.isNeedBalance = true
	r.balanceFunc = bf
	return r.balance_chan
}

func (r *RankByGuildDelay) transBalance() {
	// 为了保持接口一致
	r.balance_chan = make(chan bool, 1)
	go func() {
		for {
			is_need_balance := <-r.balance_chan
			if is_need_balance && r.isNeedBalance {
				logs.Warn("RankByGuildDelay  <-r.balance_chan %s blance", r.rank_name)
				r.res_topNIO_chan <- rankByGuildDelayTopNCommand{
					typ: rankByGuildDelayCommandTypBalance,
				}
			}
		}
	}()
}

func (r *RankByGuildDelay) balance() {
	logs.Warn("RankByGuildDelay %s blance", r.rank_name)
	r.balanceFunc(r.topN.TopN)
}

func (r *RankByGuildDelay) rename(data GuildDataInRank) {
	for i, ranker := range r.topN.TopN {
		if ranker.UUID == data.UUID {
			r.topN.TopN[i].Name = data.Name
			r.topN.TopN[i].ChiefName = data.ChiefName
		}
	}
}

// for balance

func (r *RankByGuildDelay) loadTopN() {
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

func (r *RankByGuildDelay) saveTopN() {
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

func (r *RankByGuildDelay) getTopN() [RankTopSize]GuildDataInRank {
	// 注意外部调用前需要复制
	re := r.topN.TopN
	return re
}

func (r *RankByGuildDelay) setTopN(top [RankTopSize]GuildDataInRank) {
	// 注意外部调用前需要复制
	r.topN.setTopN(top)
}

func (r *RankByGuildDelay) getPos(acid string) int {
	return r.rankdb.getPos(r.rank_name, r.db_name, acid)
}

func (r *RankByGuildDelay) getScore(acid string) (int64, error) {
	return r.rankdb.getScore(r.rank_name, r.db_name, acid)
}

/*
	r.rankdb.add(r.rank_name, r.db_name, acid, data.Score)
	r.res_mkTopN_chan <- rankByGuildDelayCommand{
		typ:  rankByGuildDelayCommandTypMkTopN,
		acid: acid,
		data: data,
	}
*/
func (r *RankByGuildDelay) tick() {
	//
	nowT := time.Now().Unix()
	if len(r.scoreChangeReqs) > 0 && (len(r.scoreChangeReqs) > 30 || (nowT-r.lastSendTime) > 30) {
		r.lastSendTime = nowT
		datas := make([]GuildDataInRank, 0, len(r.scoreChangeReqs))
		for _, c := range r.scoreChangeReqs {
			// 用于确定排序原则 同分先进居上
			c.Score += r.scorePowBase - r.topN.CurrPowValue - 1
			r.topN.CurrPowValue += 1
			if r.topN.CurrPowValue >= r.scorePowBase {
				r.topN.CurrPowValue = 0
			}
			datas = append(datas, c)
		}
		r.res_mkTopN_chan <- rankByGuildDelayCommand{
			typ:   rankByGuildDelayCommandTypMkTopN,
			datas: datas[:],
		}
		r.sendToRedis(datas)
		r.scoreChangeReqs = make(map[string]GuildDataInRank, 128)
	}
}

func (r *RankByGuildDelay) sendToRedis(data []GuildDataInRank) {
	logs.Trace("RankByGuildDelay sendToRedis %v", data)
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

func (r *RankByGuildDelay) add(acid string, data GuildDataInRank) {
	// 注意这里由于会减少是变动值
	od, ok := r.scoreChangeReqs[acid]
	if ok {
		// 合并两次请求
		ndata := data
		ndata.scoreDeta = od.scoreDeta + data.scoreDeta
		ndata.Score = data.Score
		r.scoreChangeReqs[acid] = ndata
	} else {
		r.scoreChangeReqs[acid] = data
	}
}

func (r *RankByGuildDelay) del(acid string) {
	delete(r.scoreChangeReqs, acid)
	r.rankdb.del(r.rank_name, r.db_name, acid)
	select {
	case r.res_mkTopN_chan <- rankByGuildDelayCommand{
		typ:            rankByGuildDelayCommandTypMkTopN,
		needUpdateTopN: true,
	}:
	default:
		logs.Warn("RankByGuildDelay del res_mkTopN_chan <- timeout")
	}
}

func (r *RankByGuildDelay) mkTopN(acid string, data []GuildDataInRank) bool {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByGuildDelay mkTopN Panic, Err %v", err)
		}
	}()

	// 注意外部调用前需要复制

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

func (r *RankByGuildDelay) Start(sid uint, rank_id int64, rank_name string, rankdb *rankDB, db_name string, getScore getGuildScoreFunc) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("RankByGuildDelay Panic, Err %v", err)
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			logs.Error("[RankByGuildDelay] Stack of %d bytes: %s\n", count, trace)
			panic(fmt.Errorf("RankByGuildDelay Start Err %v", err))
		}
	}()

	r.sid = sid
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

	r.scorePowBase = RankByGuildPowBase

	r.getScoreF = getScore
	r.scoreChangeReqs = make(map[string]GuildDataInRank, 128)

	r.loadTopN()
	r.topN.Update()
	r.topNInMake.Update()

	r.res_posIO_chan = make(chan rankByGuildDelayCommand, AddChannelSize)
	r.res_getPos_chan = make(chan rankByGuildDelayCommand, AddChannelSize)
	r.res_topNIO_chan = make(chan rankByGuildDelayTopNCommand, AddChannelSize)
	r.res_mkTopN_chan = make(chan rankByGuildDelayCommand, AddChannelSize)

	r.transBalance()

	// get 排行信息
	go func() {
		for {
			select {
			case command := <-r.res_getPos_chan:
				switch command.typ {
				case rankByGuildDelayCommandTypGet:
					func(cmd rankByGuildDelayCommand) {
						defer logs.PanicCatcherWithInfo("RankByGuildDelay1 Panic")
						if cmd.res_chan != nil {
							// rank只会发两个信息给这个chann
							bs := time.Now().UnixNano()
							s, _ := r.getScore(cmd.acid)
							uutil.MetricRedis(r.sid, "RankGuildDelay-getScore", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

							bs = time.Now().UnixNano()
							pos := r.getPos(cmd.acid)
							uutil.MetricRedis(r.sid, "RankGuildDelay-getPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

							cmd.res_chan <- RankByGuildGetRes{
								Pos:   pos,
								Score: s,
							}
						}
					}(command)
				}
			}
		}
	}()

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
				case rankByGuildDelayCommandTypAdd:
					func(cmd rankByGuildDelayCommand) {
						defer logs.PanicCatcherWithInfo("RankByGuildDelay2 Panic")
						r.add(cmd.acid, cmd.data)
					}(command)
				case rankByGuildDelayCommandTypDel:
					func(cmd rankByGuildDelayCommand) {
						defer logs.PanicCatcherWithInfo("RankByGuildDelay3 Panic")
						r.del(cmd.acid)
					}(command)
				default:
					logs.Warn("res_mkTopN_chan typ Err by %d", command.typ)
				}
			case <-timerChan:
				func() {
					defer logs.PanicCatcherWithInfo("RankByGuildDelay4 Panic")
					r.tick()
				}()
				timerChan = uutil.TimerMS.After(time.Second)
			}
		}
	}()

	// TopN数据生成协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			select {
			case command, ok := <-r.res_mkTopN_chan:
				if !ok {
					logs.Warn("res_mkTopN_chan close")
					return
				}

				switch command.typ {
				case rankByGuildDelayCommandTypMkTopN:
					func(cmd rankByGuildDelayCommand) {
						defer logs.PanicCatcherWithInfo("RankByGuildDelay5 Panic")
						is_need_update := r.mkTopN(cmd.acid, cmd.datas)
						if is_need_update || cmd.needUpdateTopN {
							r.topNInMake.Update()
							select {
							case r.res_topNIO_chan <- rankByGuildDelayTopNCommand{
								typ:  rankByGuildDelayCommandTypSetTopN,
								topN: r.topNInMake.TopN,
							}:
							default:
								logs.Warn("RankByGuildDelay5 res_topNIO_chan <- timeout")
							}
						}
					}(command)
				default:
					logs.Warn("res_mkTopN_chan typ Err by %d", command.typ)
				}
			}
		}
	}()

	// TopN IO 协程
	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			select {
			case command, ok := <-r.res_topNIO_chan:
				if !ok {
					logs.Warn("res_topNIO_chan close")
					return
				}

				r.waitter.Add(1)
				func() {
					defer r.waitter.Done()
					switch command.typ {
					case rankByGuildDelayCommandTypGet:
						func(cmd rankByGuildDelayTopNCommand) {
							defer logs.PanicCatcherWithInfo("RankByGuildDelay6 Panic")
							if cmd.res_chan != nil {
								// rank只会发两个信息给这个chann
								cmd.res_chan <- RankByGuildGetRes{
									TopN: r.getTopN(),
								}
							}
						}(command)
					case rankByGuildDelayCommandTypSetTopN:
						func(cmd rankByGuildDelayTopNCommand) {
							defer logs.PanicCatcherWithInfo("RankByGuildDelay7 Panic")
							r.setTopN(cmd.topN)
						}(command)
					case rankByGuildDelayCommandTypBalance:
						logs.Warn("RankByGuildDelay rankByGuildDelayCommandTypBalance %s blance", r.rank_name)
						func() {
							defer logs.PanicCatcherWithInfo("RankByGuildDelay8 Panic")
							r.balance()
						}()
					case rankByGuildDelayCommandTypRename:
						func() {
							defer logs.PanicCatcherWithInfo("RankByGuildDelayRename Panic")
							r.rename(command.data)
						}()
					default:
						logs.Warn("res_topN_chan command error")
					}
				}()
			}
		}
	}()
}

func (r *RankByGuildDelay) Stop() {
	close(r.res_posIO_chan)
	close(r.res_mkTopN_chan)
	close(r.res_topNIO_chan)

	r.saveTopN()

	r.waitter.Wait()
}

func (r *RankByGuildDelay) Get(acid string) *RankByGuildGetRes {
	res_chan := make(chan RankByGuildGetRes, 2) // rank只会发两个信息给这个chann
	getCommand := rankByGuildDelayCommand{
		typ:      rankByGuildDelayCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	getTopNCommand := rankByGuildDelayTopNCommand{
		typ:      rankByGuildDelayCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res_1 := r.execGetCmd(getCommand)
	res_2 := r.execTopNCmd(getTopNCommand)

	var res *RankByGuildGetRes

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

func (r *RankByGuildDelay) GetPos(acid string) (int, int64) {
	res_chan := make(chan RankByGuildGetRes, 1) // rank只会发两个信息给这个chann
	getCommand := rankByGuildDelayCommand{
		typ:      rankByGuildDelayCommandTypGet,
		acid:     acid,
		res_chan: res_chan,
	}
	res_1 := r.execGetCmd(getCommand)

	return res_1.Pos, res_1.Score / r.scorePowBase
}

func (r *RankByGuildDelay) Add(a *guild_info.GuildSimpleInfo, score, scoreOld int64, updateTopN bool) {
	data := GuildDataInRank{}
	data.SetDataGuild(a, score*r.scorePowBase)
	data.scoreDeta = (score - scoreOld) * r.scorePowBase

	r.execCmdASync(rankByGuildDelayCommand{
		typ:            rankByGuildDelayCommandTypAdd,
		acid:           a.GuildUUID,
		data:           data,
		needUpdateTopN: updateTopN,
	})
	return
}

func (r *RankByGuildDelay) GetTopNAccountID(n int) ([]string, error) {
	return r.rankdb.GetTopNFromRedis(r.rank_name, r.db_name, n)
}

func (r *RankByGuildDelay) GetInfoInRank(a *guild_info.GuildSimpleInfo) *GuildDataInRank {
	c := &GuildDataInRank{}
	if a != nil {
		c.SetDataGuild(a, r.getScoreF(a))
	}
	return c
}

func (r *RankByGuildDelay) Del(guildID string) {
	logs.Trace("RankByGuild Del %v", guildID)
	data := GuildDataInRank{}
	data.UUID = guildID
	r.execCmdASync(rankByGuildDelayCommand{
		typ:  rankByGuildDelayCommandTypDel,
		acid: guildID,
		data: data,
	})
	return
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execGetCmdASync(cmd rankByGuildDelayCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_getPos_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execGetCmdASync put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execGetCmd(cmd rankByGuildDelayCommand) *RankByGuildGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_getPos_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execGetCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execGetCmd <-res_chan timeout %v", cmd)
		return &RankByGuildGetRes{}
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execCmdASync(cmd rankByGuildDelayCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execCmdASync put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execCmd(cmd rankByGuildDelayCommand) *RankByGuildGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_posIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByGuildDelay execCmd <-res_chan timeout %v", cmd)
		return &RankByGuildGetRes{}
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execTopNCmdASync(cmd rankByGuildDelayTopNCommand) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topNIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay topN execCmd put timeout %v", cmd)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *RankByGuildDelay) execTopNCmd(cmd rankByGuildDelayTopNCommand) *RankByGuildGetRes {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case r.res_topNIO_chan <- cmd:
	case <-ctx.Done():
		logs.Error("RankByGuildDelay topN execCmd put timeout %v", cmd)
	}

	select {
	case res := <-cmd.res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("RankByGuildDelay topN <-res_chan timeout %v", cmd)
		return &RankByGuildGetRes{}
	}
}

func (r *RankByGuildDelay) OnGuildOrLeaderRename(a *guild_info.GuildSimpleInfo) {
	r.execTopNCmdASync(rankByGuildDelayTopNCommand{
		typ: rankByGuildDelayCommandTypRename,
		data: GuildDataInRank{
			UUID:      a.GuildUUID,
			Name:      a.Name,
			ChiefName: a.LeaderName,
		},
	})

}
