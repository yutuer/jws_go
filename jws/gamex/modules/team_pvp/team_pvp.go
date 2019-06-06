package team_pvp

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/balance_timer"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

const (
	TeamPvp_DB_Counter_Name = "TeamPvp_DB"
	TeamPvp_Rank_Show_Count = 100
)

type TPEnemy struct {
	Acid     string `json:"id" codec:"id"`
	Name     string `json:"n" codec:"n"`
	Gs       int    `json:"gs" codec:"gs"`
	Rank     int    `json:"rk" codec:"rk"`
	FAs      []int  `json:"fas" codec:"fas"`
	FAStarLv []int  `json:"fastarlv" codec:"fastarlv"`
}

type TeamPvp struct {
	sid uint
	w   worker
	dbw dbWorker
	rc  chan bool
}

func newTeamPvp(sid uint) *TeamPvp {
	return &TeamPvp{
		sid: sid,
		rc:  make(chan bool, 10),
	}
}

func (r *TeamPvp) Start() {
	r.dbw.start(r.sid)
	r.w.start(r.sid)
	go func() {
		for {
			select {
			case b := <-r.rc:
				if b {
					r.CommandSaveExec(dbCmd{
						typ: DB_Cmd_Balance,
					})
				}
			}
		}
	}()
	balance.GetModule(r.sid).RegBalanceNotifyChan("TeamPvp",
		r.rc, gamedata.GetTeamPVPBalanceBegin())
}

func (r *TeamPvp) AfterStart(g *gin.Engine) {

}

func (r *TeamPvp) BeforeStop() {
}

func (r *TeamPvp) Stop() {
	r.w.stop()
	r.dbw.stop()
	r.saveOnStop()
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *TeamPvp) CommandExec(cmd TeamPvpCmd) *TeamPvpRet {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	res_chan := make(chan TeamPvpRet, 1)
	cmd.resChan = res_chan
	errRet := &TeamPvpRet{}
	chann := r.w.cmd_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("TeamPvp CommandExec chann full, cmd put timeout")
		return errRet
	}

	select {
	case res := <-res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("TeamPvp CommandExec apply <-res_chan timeout")
		return errRet
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *TeamPvp) CommandSaveExec(cmd dbCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	chann := r.dbw.cmd_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("TeamPvp CommandSaveExec chann full, cmd put timeout")
	}
}

func _do(db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	return metricsModules.DoWraper(TeamPvp_DB_Counter_Name, db, commandName, args...)
}

func (r *TeamPvp) saveOnStop() {
	cb := redis.NewCmdBuffer()
	for _, playerInfo := range r.w.playerInfo {
		jInfo, err := json.Marshal(*playerInfo.info)
		if err != nil {
			logs.Error("TeamPvp dbWorker saveChg Marshal err %s", err.Error())
		}
		cb.Send("HSET", TableTeamPvpRank(r.sid), playerInfo.rank, string(jInfo))
	}

	_db := modules.GetDBConn()
	defer _db.Close()

	if _, err := metricsModules.DoCmdBufferWrapper(TeamPvp_DB_Counter_Name, _db, cb, true); err != nil {
		logs.Error("save team pvp info on stop err %s", err.Error())
	} else {
		logs.Warn("save team pvp info on stop, size = %d", len(r.w.playerInfo))
	}
}

func (r *TeamPvp) UpdateInfo(simpleInfo *helper.AccountSimpleInfo) {
	r.CommandExec(TeamPvpCmd{
		Typ:      TeamPvp_Cmd_Update,
		Acid:     simpleInfo.AccountID,
		AcidInfo: simpleInfo,
	})
}
