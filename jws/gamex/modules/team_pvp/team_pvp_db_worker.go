package team_pvp

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	_ = iota
	DB_Cmd_Save
	DB_Cmd_Balance
)

type dbCmd struct {
	typ int
	chg map[int]helper.AccountSimpleInfo
}

type dbWorker struct {
	waitter        util.WaitGroupWrapper
	cmd_chan       chan dbCmd
	failedCmdQueue []*dbCmd
}

func (w *dbWorker) start(sid uint) {
	w.cmd_chan = make(chan dbCmd, 2048)

	w.waitter.Wrap(func() {
		for cmd := range w.cmd_chan {
			func() {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("TeamPvp dbWorker Panic")
				w.processCommand(sid, &cmd)
			}()
		}
	})
}

func (w *dbWorker) stop() {
	close(w.cmd_chan)
	w.waitter.Wait()
}

func (w *dbWorker) processCommand(sid uint, cmd *dbCmd) {
	switch cmd.typ {
	case DB_Cmd_Save:
		w.saveChg(sid, cmd)
	case DB_Cmd_Balance:
		w.balance(sid)
	}
}

func (w *dbWorker) saveChg(sid uint, cmd *dbCmd) {
	if w.failedCmdQueue == nil {
		w.failedCmdQueue = make([]*dbCmd, 0, 1)
	}
	w.failedCmdQueue = append(w.failedCmdQueue, cmd)

	_db := modules.GetDBConn()
	defer _db.Close()

	cb := redis.NewCmdBuffer()
	for _, tempCmd := range w.failedCmdQueue {
		if tempCmd.chg != nil && len(tempCmd.chg) > 0 {
			for r, p := range tempCmd.chg {
				jInfo, err := json.Marshal(p)
				if err != nil {
					logs.Error("TeamPvp dbWorker saveChg Marshal err %s", err.Error())
				}
				cb.Send("HSET", TableTeamPvpRank(sid), r, string(jInfo))
			}
		}
	}
	if _, err := metricsModules.DoCmdBufferWrapper(TeamPvp_DB_Counter_Name, _db, cb, true); err != nil {
		logs.Error("TeamPvp dbWorker saveChg DoCmdBufferWrapper err %s", err.Error())
	} else {
		w.failedCmdQueue = w.failedCmdQueue[:0]
	}
}

func (w *dbWorker) balance(sid uint) {
	_db := modules.GetDBConn()
	defer _db.Close()

	_r := uint32(1)
	rankUids := make([]string, gamedata.TitleTeamPvpRankSum())
	for _r <= gamedata.GetTPvpRewardRankMax() {
		param := make([]interface{}, 10)
		for i := 0; i < 10; i++ {
			param[i] = _r
			_r++
		}
		logs.Debug("TeamPvp balance %v", param)
		args := make([]interface{}, 0, len(param)+1)
		args = append(args, TableTeamPvpRank(sid))
		args = append(args, param...)
		ss, err := redis.Strings(_do(_db, "HMGET", args...))
		if err != nil {
			logs.Error("TPvp balance HMGET err %s %v  %s",
				TableTeamPvpRank(sid), param, err.Error())
			continue
		}

		for i, info := range ss {
			if info == "" {
				continue
			}
			_rid := param[i]
			rid, _ := _rid.(uint32)
			sm := &helper.AccountSimpleInfo{}
			if err := json.Unmarshal([]byte(info), sm); err != nil {
				logs.Error("TeamPvp balance Unmarshal err %s %d %s", err.Error(), rid, info)
				continue
			}
			// 机器人不发奖
			if _, err := ParseTPvpRobotId(sm.AccountID); err == nil {
				continue
			}
			// 玩家发奖
			items, counts := gamedata.GetTPvpSectorReward(uint32(rid))
			if items == nil || counts == nil {
				continue
			}
			mail_sender.BatchSendTeamPvpReward(sid, sm.AccountID, int(rid), items, counts)
			// 为title记录
			if int(rid)-1 < gamedata.TitleTeamPvpRankSum() {
				rankUids[rid-1] = sm.AccountID
			}
			// sysnotice
			ac, _ := db.ParseAccount(sm.AccountID)
			cfg := gamedata.TeamPvpSysNotic(uint32(rid))
			if cfg != nil {
				sysnotice.NewSysRollNotice(ac.ServerString(), int32(cfg.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, sm.Name).Send()
			}
		}
	}
	title_rank.GetModule(sid).SetTeamPvpRank(rankUids)

	//m, err := redis.StringMap(_do(_db, "HGETALL", TableTeamPvpRank(sid)))
	//if err != nil {
	//	logs.Error("TeamPvp balance err %v", err)
	//	return
	//}
	//
	//rankUids := make([]string, gamedata.TitleTeamPvpRankSum())
	//for r, info := range m {
	//	rid, err := strconv.Atoi(r)
	//	if err != nil {
	//		logs.Error("TeamPvp balance Atoi err %s %s %s", err.Error(), r, info)
	//		continue
	//	}
	//
	//}
}
