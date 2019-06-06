// 用于生成Account相关的测试结构, 使用方法：
//
// account.InitDebuger()
// var acc account.Account = Debuger.GetTestAccount()
// var acc *account.Account = Debuger.GetNewAccount()
// var prof profile.Account = Debuger.GetTestProfile()
// var prof *profile.Account = Debuger.GetNewProfile()
// ...
//
// 更多用法参考IAccountDebuger接口

package account

import (
	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/fashion"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/jws/gamex/models/astrology"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"reflect"
	"unsafe"
)

type AccountDebuger struct {
	account Account
}

type IAccountDebuger interface {
	GetTestAccount() Account
	GetNewAccount() *Account
	GetNewDBAccount(gameId uint, shardId uint) *db.Account
	GetTestProfile() Profile
	GetNewProfile() *Profile
	GetTestCorp() Corp
	GetNewCorp() *Corp
	UnlockAllHero()
	GetTestEnergy()PlayerEnergy
	GetNewEnergy() *PlayerEnergy
	GetTestGacha() PlayerGacha
	GetNewGacha() *PlayerGacha
	GetTestJadeBag() PlayerJadeBag
	GetNewJadeBag()*PlayerJadeBag
	GetTestStackBag() bag.StackBag
	GetNewStackBag() *bag.StackBag
}

var (
	Debuger *AccountDebuger
)

func InitDebuger() {
	Debuger = new(AccountDebuger)
	Debuger.account = *Debuger.getTestAccount()
}

func (d AccountDebuger) GetTestAccount() Account {
	return d.account
}

func (d AccountDebuger) GetNewAccount() *Account {
	acc := new(Account)
	*acc = *d.getTestAccount()
	return acc
}

func (d AccountDebuger) getTestAccount() *Account {
	account := new(Account)

	// 挂接dbaccount
	account.AccountID = d.getTestDBAccount()

	// 挂接profile
	profile := d.getTestProfile()
	account.Profile = profile

	// 挂接rander
	rander := rand.New(&util.Kiss64Rng{})
	rander.Seed(time.Now().UnixNano())
	account.rander = rander

	// 挂接StackBag
	bag := d.getTestStackBag()
	account.BagProfile.StackBag = bag

	return account
}

func (d AccountDebuger) getTestDBAccount() (dbAcc db.Account) {
	var (
		testGameId  uint      = 999999999 // 9个9
		testShardId uint      = 666666    // 6个6
		testUserId  db.UserID = db.UserIDFromStringOrNil("2ed48893-3182-47c7-8855-94a01ced6784")
	)

	dbAcc = db.Account{testGameId, testShardId, testUserId}
	return
}

func (d AccountDebuger) GetNewDBAccount(gameId uint, shardId uint) (dbAcc *db.Account) {
	dbAcc = &db.Account{gameId, shardId, db.NewUserIDWithName("Test")}
	return
}

func (d AccountDebuger) GetTestProfile() Profile {
	return d.account.Profile
}

func (d AccountDebuger) GetNewProfile() *Profile {
	p := new(Profile)
	*p = d.getTestProfile()
	return p
}

func (d AccountDebuger) getTestProfile() Profile {
	profile := new(Profile)

	// 挂接corp
	corp := Debuger.getTestCorp()
	corp.player = profile
	profile.CorpInf = corp

	// 挂接fashionbag
	bag := &fashion.PlayerFashionBag{}
	bag.Items = make(map[uint32]helper.FashionItem, 64)
	profile.fashionBag = *bag

	// 挂接market_activity
	ma := market_activity.PlayerMarketActivitys{}
	ma.RegHandler(nil)
	profile.MarketActivitys = ma

	// 挂接energy
	energy := Debuger.getTestEnergy()
	energy.player = profile
	profile.Energy = energy

	// 挂接JadeBag
	jb := Debuger.getTestJadeBag()
	profile.jadeBagInMem = jb

	// 挂接Astrology
	profile.Astrology = astrology.NewAstrology()

	return *profile
}

func (d AccountDebuger) GetTestCorp() Corp {
	return d.account.Profile.CorpInf
}

func (d AccountDebuger) GetNewCorp() *Corp {
	corp := new(Corp)
	*corp = d.getTestCorp()
	return corp
}

func (d AccountDebuger) getTestCorp() Corp {
	corp := new(Corp)
	corp.OnAccountInit()
	corp.SetHandler(nil)
	return *corp
}

func (d *AccountDebuger) UnlockAllHero() {
	c := d.account.Profile.CorpInf
	p := d.account
	n := len(c.UnlockAvatars)
	for i := 0; i < n; i++ {
		c.UnlockAvatar(&p, i)
		d.account.Profile.Hero.HeroStarLevel[i] = 15
		d.account.Profile.Hero.SetWholeCharHasGot(i)
	}
}

func (d AccountDebuger) GetTestStackBag() bag.StackBag {
	return d.account.BagProfile.StackBag
}

func (d AccountDebuger) GetNewStackBag() *bag.StackBag {
	bg := new(bag.StackBag)
	*bg = d.account.BagProfile.StackBag
	return bg
}

func (d AccountDebuger) getTestStackBag() bag.StackBag {
	dbAccount := d.getTestDBAccount()
	bg := bag.NewStackBag(dbAccount)

	unfixedItemCount := make(map[string]uint32)
	var i interface{}
	i = unfixedItemCount
	unsafeModify("unfixedItemCount", i, bg)

	return *bg
}


func (d AccountDebuger) getTestFashionBag() fashion.PlayerFashionBag {
	bag := &fashion.PlayerFashionBag{}
	bag.Items = make(map[uint32]helper.FashionItem, 64)
	return *bag
}

func (d AccountDebuger) getTestMarketActivitys() market_activity.PlayerMarketActivitys {
	ma := &market_activity.PlayerMarketActivitys{}
	ma.RegHandler(nil)

	return *ma
}

func (d AccountDebuger) GetTestEnergy() PlayerEnergy {
	return d.account.Profile.Energy
}

func (d AccountDebuger) GetNewEnergy() *PlayerEnergy {
	e := new(PlayerEnergy)
	*e = d.getTestEnergy()
	return e
}

func (d AccountDebuger) getTestEnergy() PlayerEnergy {
	eng := &PlayerEnergy{}
	eng.SetHandler(nil)

	return *eng
}

func (d AccountDebuger) GetTestGacha() PlayerGacha {
	return d.account.Profile.Gacha
}

func (d AccountDebuger) GetNewGacha() *PlayerGacha {
	g := new(PlayerGacha)
	return g
}

func (d AccountDebuger) getTestJadeBag() PlayerJadeBag {
	jb := &PlayerJadeBag{}
	jb.JadesMap = make(map[uint32]JadeItem)

	return *jb
}

func (d AccountDebuger) GetTestJadeBag() PlayerJadeBag {
	return d.account.Profile.jadeBagInMem
}

func (d AccountDebuger) GetNewJadeBag() *PlayerJadeBag {
	jb := new(PlayerJadeBag)
	*jb = d.getTestJadeBag()

	return jb
}

func unsafeValueOf(val reflect.Value) reflect.Value {
	uptr := unsafe.Pointer(val.UnsafeAddr())
	return reflect.NewAt(val.Type(), uptr).Elem()
}

/*
	这个函数直接把i的值写入到j.<fieldName>中
	慎用！！！
*/
func unsafeModify(fieldName string, i, j interface{}) {
	// 直接弄到地址
	voj := reflect.ValueOf(j).Elem()
	foj := voj.FieldByName(fieldName)
	addr := unsafeValueOf(foj)

	// 直接赋值
	voi := reflect.ValueOf(i)
	addr.Set(voi)
}