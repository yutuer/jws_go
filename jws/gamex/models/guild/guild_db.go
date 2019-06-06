package guild

import (
	"time"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func NewPlayerGuild(account db.Account) *Guild {
	now_t := time.Now().Unix()
	return &Guild{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "pguild",
		},
		//Ver:        helper.CurrDBVersion,
		//LastTime:   now_t,
		CreateTime: now_t,
	}
}

func (p *Guild) DBName() string {
	return p.dbkey.String()
}

func (p *Guild) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := p.DBName()

	if forceDirty {
		p.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save PlayerGuild %s %v", p.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save PlayerGuild clean %s", p.dbkey.Account.String())
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *Guild) DBLoad(logInfo bool) error {
	key := p.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	p.dirtiesCheck = driver.GenDirtyHash(p)
	return err
}
