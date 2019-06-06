package store

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/db"

	"github.com/stretchr/testify/assert"
)

func DebugGetDBAccount() *db.Account {
	dbAccount := &db.Account{
		GameId:  0,
		ShardId: 0,
		UserId:  db.NewUserID(),
	}

	return dbAccount
}

func (ps *PlayerStores) DebugGetStoresFromGameData() {
	market := newMarket()
	storeIds := gamedata.GetStoreIDs()

	for _, storeId := range storeIds {
		market.addStore(storeId)
	}

	market.afterLogin()
	ps.Market = market
}

func DebugGetNewPlayerStores() *PlayerStores {
	dbAccount := DebugGetDBAccount()
	ps := NewPlayerStores(*dbAccount)
	ps.onAfterLogin()

	return ps
}

func TestNewPlayerStores(t *testing.T) {
	dbAccount := DebugGetDBAccount()
	ps := NewPlayerStores(*dbAccount)

	assert.Equal(t, ps.dbKey.Prefix, "store")
	assert.NotNil(t, ps.CreateTime)
}

func TestPlayerStores_DBName(t *testing.T) {
	ps := DebugGetNewPlayerStores()

	assert.Contains(t, ps.DBName(), "store:")
}

func TestPlayerStores_Update(t *testing.T) {
	// 策划和程序童靴们你们就不能统一下index么

	ps := DebugGetNewPlayerStores()

	acid := ""
	now := time.Now().UnixNano()
	rd := rand.New(rand.NewSource(now))
	level := uint32(100)

	ps.DebugGetStoresFromGameData()

	// 不需要清空次数/刷新商店，注意这其实是1号商店
	ps.Market.Stores[0].ManualRefreshCount = 5
	ps.Market.Stores[0].LastRefreshTime = now

	// 需要清空次数/刷新商店，对了这个是2号
	ps.Market.Stores[1].ManualRefreshCount = 255

	results := ps.Update(acid, now, level, rd)

	assert.NotNil(t, results)
	assert.Equal(t, len(gamedata.GetStoreIDs()), len(results)+1)
	assert.False(t, results[1])
	assert.True(t, results[2])

	assert.Equal(t, ps.Market.Stores[0].ManualRefreshCount, uint32(5))
	assert.Equal(t, ps.Market.Stores[1].ManualRefreshCount, uint32(0))
}

func TestPlayerStores_GetStores(t *testing.T) {
	ps := DebugGetNewPlayerStores()
	ps.DebugGetStoresFromGameData()
	srcIds := gamedata.GetStoreIDs()

	stores := ps.GetStores()
	iDs := []uint32{}

	for _, store := range stores {
		iDs = append(iDs, store.StoreID)
	}

	// 1.8 新特性，实在懒得写interface，1.6跑不通就注释掉吧……
	sort.Slice(srcIds, func(i, j int) bool { return srcIds[i] < srcIds[j] })
	assert.Equal(t, srcIds, iDs)
}

func TestPlayerStores_GetStore(t *testing.T) {
	ps := DebugGetNewPlayerStores()
	ps.DebugGetStoresFromGameData()

	srcStore := ps.GetStore(1)

	assert.Equal(t, srcStore.StoreID, uint32(1))
}

func TestPlayerStores_GetShop(t *testing.T) {
	ps := DebugGetNewPlayerStores()
	shop := &Shop{ShopTyp: 0, Goods: []Good{Good{GoodId: "JD_1_12", UseTimes: 5}}}
	ps.Shops[2] = *shop

	assert.NotNil(t, ps.GetShop(0))
	assert.Nil(t, ps.GetShop(4))
	assert.Equal(t, ps.GetShop(2), shop)
}
