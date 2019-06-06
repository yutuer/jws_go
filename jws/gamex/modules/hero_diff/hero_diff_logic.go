package hero_diff

import (
	"sync"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type heroDiffInfo struct {
	LastResetTime      int64 `json:"last_reset_time"`
	TodayStage         []int `json:"today_stage"`
	LastMaxStageIndex  int   `json:"last_max_stage_id"`
	LastResetStageTime int64 `json:"last_reset_stage_time"`
	lock               sync.RWMutex
}

func (hdi *heroDiffInfo) updateTodayStage(updateTime int64) []int {
	hdi.lock.Lock()
	defer hdi.lock.Unlock()
	if hdi.LastResetStageTime != updateTime {
		hdi.TodayStage, hdi.LastMaxStageIndex = GetStageIDSeq(hdi.LastMaxStageIndex)
		hdi.LastResetStageTime = updateTime
	}
	return hdi.TodayStage
}

func (hd *heroDiffModule) GetTodayStage(now_t int64) []int {
	updateTime := util.DailyBeginUnixByStartTime(now_t,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
	return hd.heroDiffInfo.updateTodayStage(updateTime)
}

func (hd *heroDiffModule) saveToDB() {
	cb := redis.NewCmdBuffer()
	if hd.heroDiffInfo.TodayStage == nil {
		hd.heroDiffInfo.TodayStage = make([]int, 0)
	}
	if err := driver.DumpToHashDBCmcBuffer(cb, TableHeroDiff(hd.sid), &hd.heroDiffInfo); err != nil {
		logs.Error("DumpToHashDBCmcBuffer err %v", err)
		return
	}
	conn := driver.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return
	}
	if _, err := metricsModules.DoCmdBufferWrapper(
		db_counter_key, conn, cb, true); err != nil {
		logs.Error("DoCmdBuffer error %s", err.Error())
		return
	}
	logs.Debug("herodiff save: %v", hd.heroDiffInfo)
}

func (hd *heroDiffModule) loadFromDB() {
	conn := driver.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return
	}
	err := driver.RestoreFromHashDB(conn.RawConn(),
		TableHeroDiff(hd.sid), &hd.heroDiffInfo, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("restorefromhashdb err by %v", err)
	}
	logs.Debug("herodiff loaded: %v", hd.heroDiffInfo)
}

func (hd *heroDiffModule) deleteRank() {
	logs.Debug("delete herodiff rank")
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return
	}
	cb := redis.NewCmdBuffer()
	cb.Send("DEL", rank.TableRankCorpHeroDiffHU(hd.sid))
	cb.Send("DEL", rank.TableRankCorpHeroDiffZHAN(hd.sid))
	cb.Send("DEL", rank.TableRankCorpHeroDiffTU(hd.sid))
	cb.Send("DEL", rank.TableRankCorpHeroDiffSHI(hd.sid))
	if _, err := conn.DoCmdBuffer(cb, true); err != nil && err != redis.ErrNil {
		logs.Error("del herodiff rank err by %v", err)
		return
	}
	for i := 0; i < len(rank.GetModule(hd.sid).RankByHeroDiff); i++ {
		rank.GetModule(hd.sid).RankByHeroDiff[i].ReloadTopN()
	}
	logs.Debug("delete herodiff rank over")
}
