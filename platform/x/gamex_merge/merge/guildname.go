package merge

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

/*
	工会重名修改
	并删除工会里已经删除的角色，若工会已经空了，则删工会
*/
var (
	guild_name_old2new = make(map[string]string, 1)
	guild_score_chg    = make(map[string]int64, 128)
	guild_del          = make(map[string]struct{}, 128)
)

func MergeGuildName() error {
	ADB := GetADBConn()
	defer ADB.Close()
	BDB := GetBDBConn()
	defer BDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()

	ANames, err := redis.StringMap(ADB.Do("HGETALL", guild.TableGuildName(Cfg.AShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("MergeGuildName A HGETALL err %s", err.Error())
		return err
	}
	BNames, err := redis.StringMap(BDB.Do("HGETALL", guild.TableGuildName(Cfg.BShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("MergeGuildName B HGETALL err %s", err.Error())
		return err
	}

	// 删工会成员
	cb := redis.NewCmdBuffer()
	for n, id := range ANames {
		err, isdel := delGuildMem(id, ResDB, cb)
		if err != nil {
			logs.Error("MergeGuildName A delGuildMem err %s", err.Error())
			return err
		}
		if isdel {
			delete(ANames, n)
		}
	}
	for n, id := range BNames {
		err, isdel := delGuildMem(id, ResDB, cb)
		if err != nil {
			logs.Error("MergeGuildName B delGuildMem err %s", err.Error())
			return err
		}
		if isdel {
			delete(BNames, n)
		}
	}
	// 这里先存一次，因为后面还会在从db里加载工会
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeGuildName save res guildnames err %s", err.Error())
		return err
	}

	// 找重名
	cb = redis.NewCmdBuffer()
	BNameDup := make(map[string]string, 64)
	for an, _ := range ANames {
		if bacid, ok := BNames[an]; ok {
			BNameDup[an] = bacid
		}
	}
	logs.Info("MergeGuildName BNameDup %d", len(BNameDup))
	if len(BNameDup) > 0 {
		guild_name_old2new = make(map[string]string, len(BNameDup))
		for oldName, guilduuid := range BNameDup {
			newName := nameMerge(oldName, Cfg.BShardId)
			delete(BNames, oldName)
			BNames[newName] = guilduuid
			guild_name_old2new[oldName] = newName
			logs.Debug("guildname %s -> %s", oldName, newName)
			// guild name
			if err = guildNameChg(guilduuid, newName, ResDB, cb); err != nil {
				return err
			}
		}
	}
	// guild names
	for n, id := range ANames {
		if err := cb.Send("HSET", guild.TableGuildName(Cfg.ResShardId), n, id); err != nil {
			logs.Error("MergeGuildName HSET guildnames err %s", err.Error())
			return err
		}
	}
	for n, id := range BNames {
		if err := cb.Send("HSET", guild.TableGuildName(Cfg.ResShardId), n, id); err != nil {
			logs.Error("MergeGuildName HSET guildnames err %s", err.Error())
			return err
		}
	}
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeGuildName save res guildnames err %s", err.Error())
		return err
	}

	// guild mem name
	cb = redis.NewCmdBuffer()
	for _, guilduuid := range BNames {
		if err = guildMemNameChg(guilduuid, ResDB, cb); err != nil {
			return err
		}
	}
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeGuildName save res guildnames err %s", err.Error())
		return err
	}

	logs.Info("del guild num %d", len(guild_del))
	logs.Info("guild sum num %d", len(ANames)+len(BNames))
	return nil
}

func guildNameChg(guilduuid, name string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	g := &guild.GuildInfo{
		GuildInfoBase: guild_info.GuildInfoBase{
			Base: guild_info.GuildSimpleInfo{
				GuildUUID: guilduuid,
			},
		},
	}

	key := g.DBName()

	err := driver.RestoreFromHashDB(resDB.RawConn(), key, g, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("MergeGuildName guildNameChg RestoreFromHashDB err %s", err.Error())
		return err
	}

	g.Base.Name = name
	g.Base.RenameTimes = 0

	err = driver.DumpToHashDBCmcBuffer(cb, key, g)
	if err != nil {
		logs.Error("MergeGuildName guildNameChg saveDB err %s", err.Error())
		return err
	}

	// 成员的simpleinfo中的名称
	for i := 0; i < int(g.Base.MemNum); i++ {
		mem := &g.Members[i]
		memAccount, _ := db.ParseAccount(mem.AccountID)
		if err := menSimpleInfoGuildNameChg(memAccount, name, resDB, cb); err != nil {
			logs.Error("MergeGuildName guildNameChg account simpleinfo guildname chg err %s", err.Error())
			return err
		}
	}

	return nil
}

func guildMemNameChg(guilduuid string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	g := &guild.GuildInfo{
		GuildInfoBase: guild_info.GuildInfoBase{
			Base: guild_info.GuildSimpleInfo{
				GuildUUID: guilduuid,
			},
		},
	}

	key := g.DBName()

	err := driver.RestoreFromHashDB(resDB.RawConn(), key, g, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("MergeGuildName guildMemNameChg RestoreFromHashDB err %s", err.Error())
		return err
	}
	// 如果成员名字也被修改了，这里顺便修改玩家名称
	for i := 0; i < int(g.Base.MemNum); i++ {
		mem := &g.Members[i]
		newName, ok := account_name_old2new[mem.Name]
		if ok {
			mem.Name = newName
			logs.Debug("guild %s mem name -> %s", g.Base.Name, mem.Name)
		}
	}
	err = driver.DumpToHashDBCmcBuffer(cb, key, g)
	if err != nil {
		logs.Error("MergeGuildName guildMemNameChg saveDB err %s", err.Error())
		return err
	}

	return nil
}

func menSimpleInfoGuildNameChg(accountID db.Account, newGuildName string,
	resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	// account simpleinfo中公会名称修改
	_simpleInfo := simple_info.NewSimpleInfoProfile(accountID)
	simpleInfo := &_simpleInfo
	key := simpleInfo.DBName()
	err := driver.RestoreFromHashDB(resDB.RawConn(), key, simpleInfo, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("MergeRoleName accountNameChg simpleInfo loadDB err %s", err.Error())
		return err
	}

	simpleInfo.GuildName = newGuildName

	err = driver.DumpToHashDBCmcBuffer(cb, key, simpleInfo)
	if err != nil {
		logs.Error("MergeRoleName accountNameChg simpleInfo saveDB err %s", err.Error())
		return err
	}
	return nil
}

func delGuildMem(guilduuid string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) (
	err error, isDel bool) {
	g := &guild.GuildInfo{
		GuildInfoBase: guild_info.GuildInfoBase{
			Base: guild_info.GuildSimpleInfo{
				GuildUUID: guilduuid,
			},
		},
	}

	key := g.DBName()
	sid := Cfg.ResShardId

	err = driver.RestoreFromHashDB(resDB.RawConn(), key, g, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("MergeGuildName delGuildMem RestoreFromHashDB err %s", err.Error())
		return err, false
	}

	leftMms := make([]helper.AccountSimpleInfo, 0, 8)
	num := g.Base.MemNum
	for i := 0; i < num; i++ {
		mem := g.Members[i]
		if mem.AccountID == "" {
			break
		}
		if isAcidDel(mem.AccountID) {
			gs := mem.CurrCorpGs
			g.Base.MemNum -= 1
			g.Base.GuildGSSum -= int64(gs)
			key := guild.TableAccount2Guild(sid)
			err := cb.Send("HDEL", key, mem.AccountID)
			if err != nil {
				logs.Error("MergeGuildName delGuildMem HDEL err %s", err.Error())
				return err, false
			}
			guild_score_chg[g.Base.GuildUUID] = g.Base.GuildGSSum * rank.RankByGuildPowBase
			logs.Debug("delGuildMem %s from guild %s", mem.AccountID, g.Base.GuildUUID)
		} else {
			leftMms = append(leftMms, mem)
		}
	}

	if len(leftMms) > 0 {
		g.Members = [guild_info.MaxGuildMember]helper.AccountSimpleInfo{}
		for i, m := range leftMms {
			g.Members[i] = m
		}
	}
	logs.Debug("delGuildMem memnum %s %d", g.Base.GuildUUID, g.Base.MemNum)
	if g.Base.MemNum > 0 {
		err = driver.DumpToHashDBCmcBuffer(cb, key, g)
		if err != nil {
			logs.Error("MergeGuildName guildNameChg saveDB err %s", err.Error())
			return err, false
		}
	} else {
		// 删工会
		delete(guild_score_chg, g.Base.GuildUUID)
		guild_del[g.Base.GuildUUID] = struct{}{}

		guildName := guild.TableGuildName(sid)
		if err := cb.Send("HDEL", guildName, g.Base.Name); err != nil {
			logs.Error("MergeGuildName.delGuild HDEL %v %v db err: %v", guildName, g.Base.Name, err)
			return err, false
		}

		guildIdName := guild.TableGuildId2Uuid(sid)
		if err := cb.Send("HDEL", guildIdName, g.Base.GuildID); err != nil {
			logs.Error("MergeGuildName.delGuild HDEL %v %v db err: %v", guildIdName, g.Base.GuildID, err)
			return err, false
		}

		if err := cb.Send("DEL", g.DBName()); err != nil {
			logs.Error("MergeGuildName.delGuild DEL %v db err: %v", g.DBName(), err)
			return err, false
		}
		isDel = true
		logs.Debug("del guild %s %s", g.Base.GuildUUID, g.Base.Name)
	}

	return nil, isDel
}
