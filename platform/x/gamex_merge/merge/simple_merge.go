package merge

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/dest_gen_first"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func GuildSimpleMerge() error {
	BDB := GetBDBConn()
	defer BDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()

	// guild:account2guild
	BAccount2Guild, err := redis.StringMap(BDB.Do("HGETALL", guild.TableAccount2Guild(Cfg.BShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("DestGenFirstMerge B guild:account2guild HGETALL err %s", err.Error())
		return err
	}

	cb := redis.NewCmdBuffer()
	for k, v := range BAccount2Guild {
		if err = cb.Send("HSET", guild.TableAccount2Guild(Cfg.ResShardId), k, v); err != nil {
			logs.Error("DestGenFirstMerge B guild:account2guild HSET err %s", err.Error())
			return err
		}
	}

	// guild:id2uuid
	BGuildId2uuid, err := redis.StringMap(BDB.Do("HGETALL", guild.TableGuildId2Uuid(Cfg.BShardId)))
	if err != nil && err != redis.ErrNil {
		logs.Error("DestGenFirstMerge B guild:id2uuid HGETALL err %s", err.Error())
		return err
	}
	for k, v := range BGuildId2uuid {
		if err = cb.Send("HSET", guild.TableGuildId2Uuid(Cfg.ResShardId), k, v); err != nil {
			logs.Error("DestGenFirstMerge B guild:id2uuid HSET err %s", err.Error())
			return err
		}
	}

	// save
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("DestGenFirstMerge save err %s", err.Error())
		return err
	}
	return nil
}
func DestGenFirstMerge() error {
	ADB := GetADBConn()
	defer ADB.Close()
	BDB := GetBDBConn()
	defer BDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()
	cb := redis.NewCmdBuffer()

	// destgenfirst
	ADGF := &dest_gen_first.DestGenFirstModule{}
	err := driver.RestoreFromHashDB(ADB.RawConn(),
		dest_gen_first.TableDestGenFirst(Cfg.AShardId), ADGF, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}

	if len(ADGF.FirstDests) <= 0 {
		ADGF.FirstDests = make([]dest_gen_first.DestGenFirst, helper.MaxDestingGeneralCount)
	} else {
		for i, d := range ADGF.FirstDests {
			if d.FirstPlayerName != "" && isNameDel(d.FirstPlayerName) {
				ADGF.FirstDests[i] = dest_gen_first.DestGenFirst{}
			}
		}
	}
	BDGF := &dest_gen_first.DestGenFirstModule{}
	err = driver.RestoreFromHashDB(BDB.RawConn(),
		dest_gen_first.TableDestGenFirst(Cfg.BShardId), BDGF, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	if len(BDGF.FirstDests) <= 0 {
		BDGF.FirstDests = make([]dest_gen_first.DestGenFirst, helper.MaxDestingGeneralCount)
	} else {
		for i, d := range BDGF.FirstDests {
			if d.FirstPlayerName != "" && isNameDel(d.FirstPlayerName) {
				BDGF.FirstDests[i] = dest_gen_first.DestGenFirst{}
			}
		}
	}
	for i := 0; i < helper.MaxDestingGeneralCount; i++ {
		ainfo := ADGF.FirstDests[i]
		if len(BDGF.FirstDests) <= i {
			continue
		}
		binfo := BDGF.FirstDests[i]
		if ainfo.FirstPlayerName != "" &&
			binfo.FirstPlayerName != "" {
			if binfo.FirstPlayerTimeStamp < ainfo.FirstPlayerTimeStamp {
				ADGF.FirstDests[i] = binfo
			}
		} else if ainfo.FirstPlayerName == "" &&
			binfo.FirstPlayerName != "" {
			ADGF.FirstDests[i] = binfo
		}
	}
	if err := driver.DumpToHashDBCmcBuffer(cb,
		dest_gen_first.TableDestGenFirst(Cfg.ResShardId), ADGF); err != nil {
		return fmt.Errorf("DestGenFirstMerge B destgenfirst err %v", err)
	}

	// save
	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("DestGenFirstMerge save err %s", err.Error())
		return err
	}
	return nil
}
