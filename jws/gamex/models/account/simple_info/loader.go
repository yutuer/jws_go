package simple_info

import (
	"time"

	"fmt"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type AccountSimpleInfoProfile struct {
	helper.AccountSimpleInfo
	dirtiesCheck map[string]interface{}
	dbkey        db.ProfileDBKey
	Ver          int64 `redis:"version"`
	CreateTime   int64 `redis:"createtime"`
}

func NewSimpleInfoProfile(account db.Account) AccountSimpleInfoProfile {
	now_t := time.Now().Unix()
	return AccountSimpleInfoProfile{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "simpleinfo",
		},
		//LastTime:   now_t,
		CreateTime: now_t,
		//Ver:        helper.CurrDBVersion,
		AccountSimpleInfo: helper.AccountSimpleInfo{
			TeamPvpAvatar:   [helper.TeamPvpAvatarsCount]int{0, 1, 2},
			TeamPvpAvatarLv: [helper.TeamPvpAvatarsCount]int{1, 1, 1},
		},
	}
}

func (p *AccountSimpleInfoProfile) DBName() string {
	return p.dbkey.String()
}

func (p *AccountSimpleInfoProfile) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
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
			logs.Trace("Save AccountSimpleInfoProfile %s %v", p.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save AccountSimpleInfoProfile clean %s", p.dbkey.Account.String())
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *AccountSimpleInfoProfile) DBLoad(logInfo bool) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	key := p.DBName()

	if _db.IsNil() {
		return fmt.Errorf("AccountSimpleInfoProfile DBLoad db nil")
	}
	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	p.dirtiesCheck = driver.GenDirtyHash(p)
	return err
}

func (p *AccountSimpleInfoProfile) SetFromOther(info *helper.AccountSimpleInfo) {
	p.AccountSimpleInfo = *info
}

func LoadAccountSimpleInfoProfile(dbaccount db.Account) (*helper.AccountSimpleInfo, error) {
	simpleInfo := NewSimpleInfoProfile(dbaccount)
	err := simpleInfo.DBLoad(false)
	if err != nil {
		return nil, err
	}
	return &simpleInfo.AccountSimpleInfo, nil

}
