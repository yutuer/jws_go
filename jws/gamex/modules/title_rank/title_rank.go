package title_rank

import (
	"sync"

	"math"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

type TitleRank struct {
	sid                    uint
	mutx                   sync.RWMutex
	rankSimplePvpYesterday titleAcids
	rankTeamPvpYesterday   titleAcids
	rankWuShuangYesterday  titleAcids
	rank7DayGs             titleAcids
	save_chan              chan saveCmd
	waitter                util.WaitGroupWrapper
}

type titleAcids struct {
	Acids []string `json:"acids"`
}

type saveCmd struct {
	dbname string
	save   titleAcids
}

func newTitleRank(sid uint) *TitleRank {
	return &TitleRank{
		sid: sid,
		rankSimplePvpYesterday: titleAcids{make([]string, 1)},
		rankTeamPvpYesterday:   titleAcids{make([]string, 1)},
		rankWuShuangYesterday:  titleAcids{make([]string, 1)},
		rank7DayGs:             titleAcids{make([]string, 1)},
		save_chan:              make(chan saveCmd, 128),
	}
}

func (r *TitleRank) Start() {
	//
	conn := modules.GetDBConn()
	defer conn.Close()
	str_simplePvp, err := redis.String(_do(conn, "GET", TableTitleSimplePvpRank(r.sid)))
	if err == nil && str_simplePvp != "" {
		acids := &titleAcids{}
		if err := json.Unmarshal([]byte(str_simplePvp), acids); err == nil {
			r.rankSimplePvpYesterday = *acids
			logs.Debug("TitleRank load rankSimplePvpYesterday %v", r.rankSimplePvpYesterday)
		}
	}
	str_teamPvp, err := redis.String(_do(conn, "GET", TableTitleTeamPvpRank(r.sid)))
	if err == nil && str_teamPvp != "" {
		acids := &titleAcids{}
		if err := json.Unmarshal([]byte(str_teamPvp), acids); err == nil {
			r.rankTeamPvpYesterday = *acids
			logs.Debug("TitleRank load rankTeamPvpYesterday %v", r.rankTeamPvpYesterday)
		}
	}
	str_7DayGs, err := redis.String(_do(conn, "GET", TableTitle7DayGsRank(r.sid)))
	if err == nil && str_7DayGs != "" {
		acids := &titleAcids{}
		if err := json.Unmarshal([]byte(str_7DayGs), acids); err == nil {
			r.rank7DayGs = *acids
			logs.Debug("TitleRank load rank7DayGs %v", r.rank7DayGs)
		}
	}
	str_WuShuang, err := redis.String(_do(conn, "GET", TableTitleWuShuangRank(r.sid)))
	if err == nil && str_WuShuang != "" {
		acids := &titleAcids{}
		if err := json.Unmarshal([]byte(str_WuShuang), acids); err == nil {
			r.rankWuShuangYesterday = *acids
			logs.Debug("TitleRank load rankWuShuangYesterday %v", r.rankWuShuangYesterday)
		}
	}
	r.waitter.Wrap(func() {
		for command := range r.save_chan {
			func(cmd saveCmd) {
				conn := modules.GetDBConn()
				defer conn.Close()
				bb, err := json.Marshal(cmd.save)
				if err != nil {
					logs.Error("TitleRank save json.Marshal err %s", err.Error())
					return
				}
				_, err = _do(conn, "SET", cmd.dbname, string(bb))
				if err != nil {
					logs.Error("TitleRank save do err %s", err.Error())
					return
				}
				logs.Debug("TitleRank save success %v %v", cmd.dbname, cmd.save)
			}(command)
		}
	})
}

func (r *TitleRank) AfterStart(g *gin.Engine) {

}

func (r *TitleRank) BeforeStop() {
}

func (r *TitleRank) Stop() {
	close(r.save_chan)
	r.waitter.Wait()
}

func (r *TitleRank) SetSimplePvpRank(rankUids []string) {
	l := int(math.Min(float64(len(rankUids)), float64(gamedata.TitleSimpePvpSum())))
	var old titleAcids
	_rankUids := r._prepareUids(rankUids, l)

	r.mutx.Lock()
	old = r.rankSimplePvpYesterday
	r.rankSimplePvpYesterday = titleAcids{_rankUids}
	logs.Debug("TitleRank SetSimplePvpRank rankSimplePvpYesterday %v", r.rankSimplePvpYesterday)
	r._exec(saveCmd{
		dbname: TableTitleSimplePvpRank(r.sid),
		save:   titleAcids{_rankUids[:]},
	})
	r.mutx.Unlock()

	r._sendMsg(_rankUids, old.Acids)
}

func (r *TitleRank) SetTeamPvpRank(rankUids []string) {
	l := int(math.Min(float64(len(rankUids)), float64(gamedata.TitleTeamPvpRankSum())))
	var old titleAcids
	_rankUids := rankUids[:l]

	r.mutx.Lock()
	old = r.rankTeamPvpYesterday
	r.rankTeamPvpYesterday = titleAcids{_rankUids}
	logs.Debug("TitleRank SetTeamPvpRank rankTeamPvpYesterday %v", r.rankTeamPvpYesterday)
	r._exec(saveCmd{
		dbname: TableTitleTeamPvpRank(r.sid),
		save:   titleAcids{_rankUids[:]},
	})
	r.mutx.Unlock()

	r._sendMsg(_rankUids, old.Acids)
}

func (r *TitleRank) Set7DayGsRank(rankUids []string) {
	l := int(math.Min(float64(len(rankUids)), float64(gamedata.Title7DayGsRankSum())))
	var old titleAcids
	_rankUids := r._prepareUids(rankUids, l)

	r.mutx.Lock()
	old = r.rank7DayGs
	r.rank7DayGs = titleAcids{_rankUids}
	logs.Debug("TitleRank Set7DayGsRank rank7DayGsYesterday %v", r.rank7DayGs)
	r._exec(saveCmd{
		dbname: TableTitle7DayGsRank(r.sid),
		save:   titleAcids{_rankUids[:]},
	})
	r.mutx.Unlock()

	r._sendMsg(_rankUids, old.Acids)
}

func (r *TitleRank) SetWuShuangRank(rankUids []string) {
	l := int(math.Min(float64(len(rankUids)), float64(gamedata.TitleWushuangRankSum())))
	var old titleAcids
	_rankUids := r._prepareUids(rankUids, l)

	r.mutx.Lock()
	old = r.rankWuShuangYesterday
	r.rankWuShuangYesterday = titleAcids{_rankUids}
	logs.Debug("TitleRank SetWuShuangRank rankWuShuangYesterday %v", r.rankWuShuangYesterday)
	r._exec(saveCmd{
		dbname: TableTitleWuShuangRank(r.sid),
		save:   titleAcids{_rankUids[:]},
	})
	r.mutx.Unlock()

	r._sendMsg(_rankUids, old.Acids)
}

func (r *TitleRank) _prepareUids(rankUids []string, length int) []string {
	_rankUids := rankUids[:length]
	_tmp := make([]string, 0, len(_rankUids))
	for _, ruid := range _rankUids {
		if ruid == "" {
			break
		}
		_tmp = append(_tmp, ruid)
	}
	return _tmp
}

func (r *TitleRank) _sendMsg(rankUids, old []string) {
	uids := make(map[string]struct{}, len(rankUids)+len(old))
	for _, uid := range rankUids {
		uids[uid] = struct{}{}
	}
	for _, uid := range old {
		uids[uid] = struct{}{}
	}
	for uid, _ := range uids {
		player_msg.Send(uid, player_msg.PlayerMsgTitleCode,
			player_msg.DefaultMsg{})
	}
}

func (r *TitleRank) GetSimplePvpRank(uid string) int {
	r.mutx.RLock()
	defer r.mutx.RUnlock()
	for rank, ruid := range r.rankSimplePvpYesterday.Acids {
		if uid == ruid {
			return rank + 1
		}
	}
	return 0
}

func (r *TitleRank) GetTeamPvpRank(uid string) int {
	r.mutx.RLock()
	defer r.mutx.RUnlock()
	for rank, ruid := range r.rankTeamPvpYesterday.Acids {
		if uid == ruid {
			return rank + 1
		}
	}
	return 0
}

func (r *TitleRank) GetWuShuangRank(uid string) int {
	r.mutx.RLock()
	defer r.mutx.RUnlock()
	for rank, ruid := range r.rankWuShuangYesterday.Acids {
		if uid == ruid {
			return rank + 1
		}
	}
	return 0
}

func (r *TitleRank) Get7DayGsRank(uid string) int {
	r.mutx.RLock()
	defer r.mutx.RUnlock()
	for rank, ruid := range r.rank7DayGs.Acids {
		if uid == ruid {
			return rank + 1
		}
	}
	return 0
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *TitleRank) _exec(cmd saveCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	chann := r.save_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("TitleRank CommandExec chann full, cmd put timeout")
	}
}

func _do(db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	return metricsModules.DoWraper("Title_Rank_DB", db, commandName, args...)
}
