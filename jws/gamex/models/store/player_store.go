package store

import (
	"math/rand"
	"time"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//PlayerStores ..
type PlayerStores struct {
	dbKey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	Ver          int64 `redis:"version"`

	Market *Market `redis:"market"`
	Stores []interface{}
	// Stores [gamedata.StoreMaxSize]Store
	Shops [gamedata.MaxShopNum]Shop

	//LastTime   int64 `redis:"lasttime"`
	CreateTime int64            `redis:"createtime"`
	LimitShop  LimitShopBuyInfo `redis:"limitship"`
}

func (p *PlayerStores) onAfterLogin() {
	// 初始化ID
	if nil == p.Market {
		p.Market = newMarket()
	}
	p.Market.afterLogin()
	for i := 0; i < len(p.Shops); i++ {
		p.Shops[i].ShopTyp = uint32(i)
	}

	//陈旧字段遗弃
	p.Stores = []interface{}{}
}

//GetStore ..
func (p *PlayerStores) GetStore(idx uint32) *Store {
	return p.Market.getStore(idx)
}

//GetStores ..
func (p *PlayerStores) GetStores() []*Store {
	return p.Market.Stores
}

//GetShop ..
func (p *PlayerStores) GetShop(shopID uint32) *Shop {
	if shopID < 0 || shopID >= uint32(len(p.Shops)) {
		return nil
	}
	return &p.Shops[shopID]
}

//Update ..
func (p *PlayerStores) Update(acid string, now int64, lv uint32, rd *rand.Rand) map[uint32]bool {
	logs.Debug("[Store] PlayerStores Update, acid[%s]", acid)
	return p.Market.update(acid, now, lv, rd)
}

// DB ------------------------------------------------

//NewPlayerStores ..
func NewPlayerStores(account db.Account) *PlayerStores {
	now := time.Now().Unix()
	return &PlayerStores{
		dbKey: db.ProfileDBKey{
			Account: account,
			Prefix:  "store",
		},
		//Ver:        helper.CurrDBVersion,
		//LastTime:   now,
		CreateTime: now,
	}
}

//DBName ..
func (p *PlayerStores) DBName() string {
	return p.dbKey.String()
}

//DBSave ..
func (p *PlayerStores) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := p.DBName()

	if forceDirty {
		p.dirtiesCheck = nil
	}
	err, newDirtyCheck, chg := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save PlayerStores %s %v", p.dbKey.Account.String(), chg)
		} else {
			logs.Trace("Save PlayerStores clean %s", p.dbKey.Account.String())
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

//DBLoad ..
func (p *PlayerStores) DBLoad(logInfo bool) error {
	key := p.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	p.dirtiesCheck = driver.GenDirtyHash(p)
	p.onAfterLogin()

	return err
}
