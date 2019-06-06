package account_info

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	leave_guild_time     = "lastleavetime"
	db_player_guild_key  = "pguild"
	leave_guild_db       = "Leave_Guild_Time"
	next_join_guild_time = "leavetime"
	guild_assign_info    = "guild_assign_info"

	account_Error_Db = 1
)

type dbCmdBuffPrepare func(cb redis.CmdBuffer) error

// 直接存离开公会时间到DB, 相关描述@ models.guild.Guild  该逻辑已经作废
func SaveLeaveGuildTime(kickAcID string, leaveGuildTime int64) {
	kickDbAccount, err := db.ParseAccount(kickAcID)
	if err != nil {
		logs.Error("parse account error, %v", err)
		return
	}
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		return updateLeaveGuildTime(kickDbAccount, leaveGuildTime, cb)
	})
	if err != nil {
		logs.Error("update leave guild time err, %d", errCode)
	}
}

func dbCmdBuffExec(prepare dbCmdBuffPrepare) int {
	cb := redis.NewCmdBuffer()

	if err := prepare(cb); err != nil {
		logs.Error("leave guild dbCmdBuffExec err %v", err)
		return account_Error_Db
	}
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:Guild DB Save, cant get redis conn")
		return account_Error_Db
	}
	if _, err := modules.DoCmdBufferWrapper(leave_guild_db, db, cb, true); err != nil {
		logs.Error("DoCmdBuffer error %s", err.Error())
		return account_Error_Db
	}
	return 0
}

func updateLeaveGuildTime(dbAccount db.Account, leaveGuildTime int64, cb redis.CmdBuffer) error {
	logs.Debug("update leave guild time to redis, %s", dbAccount.String())
	tableName := db.ProfileDBKey{Account: dbAccount, Prefix: db_player_guild_key}.String()
	err := cb.Send("HSET", tableName, leave_guild_time, leaveGuildTime)
	if err != nil {
		logs.Error("update leave guild time HSET %s db err: %v", tableName, err)
		return err
	}
	return nil
}

func SaveInfoOnLeaveGuild(kickAcID string, nextJoinTime int64, lootID []string, times []int64, leaveGuildTime int64) {
	kickDbAccount, err := db.ParseAccount(kickAcID)
	if err != nil {
		logs.Error("parse account error, %v", err)
		return
	}
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		return updateGuildInfo(kickDbAccount, nextJoinTime, lootID, times, leaveGuildTime, cb)
	})
	if errCode != 0 {
		logs.Error("update next join guild time err, %d", errCode)
	}
}

func updateGuildInfo(dbAccount db.Account, nextJoinTime int64, lootID []string, times []int64, leaveGuildTime int64, cb redis.CmdBuffer) error {
	logs.Debug("update guild info to redis, %s", dbAccount.String())
	tableName := db.ProfileDBKey{Account: dbAccount, Prefix: db_player_guild_key}.String()
	err := cb.Send("HSET", tableName, next_join_guild_time, nextJoinTime)
	if err != nil {
		logs.Error("update next join guild time HSET %s db err: %v", tableName, err)
		return err
	}
	jsonInfo, err := buildGuildAssignStruct(lootID, times)
	if err != nil {
		logs.Error("update assign info HSET %s db err: %v", tableName, err)
		return err
	}
	logs.Debug("save guild info: %v", jsonInfo)
	err = cb.Send("HSET", tableName, guild_assign_info, jsonInfo)
	if err != nil {
		logs.Error("update guild assign info: %v", tableName, err)
		return err
	}
	err = cb.Send("HSET", tableName, leave_guild_time, leaveGuildTime)
	if err != nil {
		logs.Error("update leave guild time HSET %s db err: %v", tableName, err)
		return err
	}
	return nil
}

func buildGuildAssignStruct(lootID []string, times []int64) ([]byte, error) {
	info := struct {
		AssignID    []string `json:"assign_id"`
		AssignTimes []int64  `json:"assign_times"`
	}{
		AssignID:    append([]string{}, lootID[:]...),
		AssignTimes: append([]int64{}, times[:]...),
	}

	return json.Marshal(info)
}

// TODO error的处理
func BatchSaveNextJoinGuildTime(kickAcIDs []string, nextJoinTime int64, lootID []string, times []int64, leaveGuildTime int64) {
	dbAccounts := make([]db.Account, 0, len(kickAcIDs))
	for _, acid := range kickAcIDs {
		kickDbAccount, err := db.ParseAccount(acid)
		if err == nil {
			dbAccounts = append(dbAccounts, kickDbAccount)
		} else {
			logs.Error("parse account error, %v", err)
		}
	}
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, dbAccount := range dbAccounts {
			updateGuildInfo(dbAccount, nextJoinTime, lootID, times, leaveGuildTime, cb)
		}
		return nil
	})
}
