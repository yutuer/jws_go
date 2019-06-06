package guild

import (
	"fmt"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func checkPlayerInGuild(accountId string) guildCommandRes {
	key := TableAccount2GuildByAccount(accountId)

	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return genErrRes(Err_DB)
	}

	res, err := redis.Int64(_do(_db, "HEXISTS", key, accountId))
	if err != nil {
		logs.Error("guild checkPlayerInGuild db HEXISTS err: %v", err)
		return genErrRes(Err_DB)
	}
	if res != 0 { // 存在
		return genWarnRes(errCode.GuildPlayerAlreadyInOther)
	}
	return guildCommandRes{}
}

func GetPlayerGuild(accountId string) (guildUuid string) {
	key := TableAccount2GuildByAccount(accountId)

	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return ""
	}

	guid, err := redis.String(_do(_db, "HGET", key, accountId))
	if err != nil && err != redis.ErrNil {
		logs.Error("guild getPlayerGuild db HGET err: %v", err)
		return ""
	}
	return guid
}

func checkGuildName(name, acid string) (error, bool) {
	g, err := db.ParseAccount(acid)
	if err != nil {
		return err, false
	}
	guildName := TableGuildName(g.ShardId)

	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return err, false
	}

	res, err := redis.Int64(_do(_db, "HEXISTS", guildName, name))
	if err != nil {
		logs.Error("checkAndSaveGuildName HEXISTS %s db err: %v", guildName, err)
		return err, false
	}
	// 已经存在
	if res == 1 {
		return nil, true
	}
	return nil, false
}

func renameGuildName(oldName, newName, guildName string, guildUUID string) (int, bool) {
	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return Err_Unknown_Err, false
	}

	// 名字检查
	if len(newName) > 64 {
		logs.Error("Guild name size to long %d", len(newName))
		return Err_CODE_ERR_Name_Len, true
	}

	// 检查敏感词
	if gamedata.CheckSymbol(newName) || gamedata.CheckSensitive(newName) {
		return errCode.GuildWordIllegal, false
	}

	res, err := redis.Int64(_do(_db, "HEXISTS", guildName, newName))
	if err != nil {
		logs.Error("renameGuildName HEXISTS %s db err: %v", guildName, err)
		return Err_Unknown_Err, true
	}
	if res == 1 {
		return errCode.GuildNameRepeat, false
	}
	_, err = redis.Int64(_do(_db, "HSET", guildName, newName, guildUUID))
	if err != nil {
		logs.Error("renameGuildName HSET %s db err: %v", guildName, err)
		return Err_Unknown_Err, true
	}
	_, err = redis.Int64(_do(_db, "HDEL", guildName, oldName))
	if err != nil {
		logs.Error("renameGuildName HDEL %s db err: %v", guildName, err)
		return Err_Unknown_Err, true
	}
	return 0, false
}

func genGuildId(sid uint) (err error, guildId int64) {
	guildIdSeedName := tableGuildSeed(sid)

	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return err, 0
	}

	id, err := redis.Int64(_do(_db, "INCR", guildIdSeedName))
	if err != nil {
		logs.Error("genGuildIdAndSave INCR db err: %v", err)
		return err, 0
	}
	return nil, genGuildIdByShard(sid, id)
}

func findGuildUuid(guid string, id int64) (error, string) {
	g, err := db.ParseAccount(guid)
	if err != nil {
		return err, ""
	}
	guildIdName := TableGuildId2Uuid(g.ShardId)

	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return err, ""
	}

	uuid, err := redis.String(_do(_db, "HGET", guildIdName, fmt.Sprintf("%d", id)))
	if err != nil && err != redis.ErrNil {
		logs.Error("findGuildUuid HGET %s %d db err: %v", guildIdName, id, err)
		return err, ""
	}
	return nil, uuid
}

func saveGuild(name, guildUuid string, guildId int64, cb redis.CmdBuffer) error {
	g, err := db.ParseAccount(guildUuid)
	if err != nil {
		return err
	}

	guildName := TableGuildName(g.ShardId)
	guildIdName := TableGuildId2Uuid(g.ShardId)

	// guildname
	err = cb.Send("HSET", guildName, name, guildUuid)
	if err != nil {
		logs.Error("checkAndSaveGuildName HSET %s db err: %v", guildName, err)
		return err
	}
	// guildid -> uuid用来用id查找公会
	err = cb.Send("HSET", guildIdName, guildId, guildUuid)
	if err != nil {
		logs.Error("genGuildIdAndSave HSET %s db err: %v", guildIdName, err)
		return err
	}

	return nil
}

func addGuildMem(accountID, guildUUID string, cb redis.CmdBuffer) error {
	key := TableAccount2GuildByAccount(accountID)
	err := cb.Send("HSET", key, accountID, guildUUID)
	if err != nil {
		logs.Error("guild addmem db HSET err: %v", err)
		return err
	}
	return nil
}

func delGuildMem(accountID string, cb redis.CmdBuffer) error {
	key := TableAccount2GuildByAccount(accountID)
	err := cb.Send("HDEL", key, accountID)
	if err != nil {
		logs.Error("guild delMem db HDEL err: %v", err)
		return err
	}
	return nil
}

func delGuild(guildInfo *GuildInfo, cb redis.CmdBuffer) error {
	g, err := db.ParseAccount(guildInfo.GuildInfoBase.Base.GuildUUID)
	if err != nil {
		return err
	}
	for i := 0; i < guildInfo.Base.MemNum && guildInfo.Members[i].AccountID != ""; i++ {
		mem := guildInfo.Members[i]
		key := TableAccount2GuildByAccount(mem.AccountID)

		err := cb.Send("HDEL", key, mem.AccountID)
		if err != nil {
			logs.Error("guild delMem db HDEL err: %v", err)
			return err
		}
	}

	guildName := TableGuildName(g.ShardId)
	if err := cb.Send("HDEL", guildName, guildInfo.Base.Name); err != nil {
		logs.Error("GuildWorker.delGuild HDEL %v %v db err: %v", guildName, guildInfo.Base.Name, err)
		return err
	}

	guildIdName := TableGuildId2Uuid(g.ShardId)
	if err := cb.Send("HDEL", guildIdName, guildInfo.Base.GuildID); err != nil {
		logs.Error("GuildWorker.delGuild HDEL %v %v db err: %v", guildIdName, guildInfo.Base.GuildID, err)
		return err
	}

	if err := cb.Send("DEL", guildInfo.DBName()); err != nil {
		logs.Error("GuildWorker.delGuild DEL %v db err: %v", guildInfo.DBName(), err)
		return err
	}
	return nil
}

func loadGuildSimple(guid string) (error, *guild_info.GuildSimpleInfo) {
	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return fmt.Errorf("guild GetDBConn nil"), nil
	}

	ss, err := redis.String(_do(_db, "HGET", guild_info.GuildDBName(guid), "Base"))
	if err != nil {
		logs.Error("loadGuildSimple HGET guid %s err %s", guid, err.Error())
		return err, nil
	}
	info := &guild_info.GuildSimpleInfo{}
	if err := json.Unmarshal([]byte(ss), info); err != nil {
		logs.Error("loadGuildSimple json.Unmarshal err %s", err.Error())
		return err, nil
	}
	return nil, info
}

func loadAllGuildUuid(sid uint) (error, []string) {
	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("guild GetDBConn nil")
		return fmt.Errorf("guild GetDBConn nil"), nil
	}
	res, err := redis.Strings(_do(_db, "HGETALL", TableGuildName(sid)))
	if err != nil {
		logs.Error("loadAllGuildUuid HGETALL err %s", err.Error())
		return err, nil
	}
	if len(res) < 2 {
		return nil, []string{}
	}
	ret := make([]string, 0, len(res)/2)
	for i := 1; i < len(res); i += 2 {
		ret = append(ret, res[i])
	}
	return nil, ret
}

type dbCmdBuffPrepare func(cb redis.CmdBuffer) error

func dbCmdBuffExec(prepare dbCmdBuffPrepare) int {
	cb := redis.NewCmdBuffer()

	if err := prepare(cb); err != nil {
		logs.Error("guild dbCmdBuffExec err %v", err)
		return Err_DB
	}
	if cb.GetCmdNumber() <= 0 {
		return 0
	}
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:Guild DB Save, cant get redis conn")
		return Err_DB
	}
	if _, err := modules.DoCmdBufferWrapper(Guild_DB_Counter_Name, db, cb, true); err != nil {
		logs.Error("DoCmdBuffer error %s", err.Error())
		return Err_DB
	}
	return 0
}

func _do(db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	return modules.DoWraper(Guild_DB_Counter_Name, db, commandName, args...)
}
