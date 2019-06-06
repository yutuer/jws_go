package store

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"github.com/stretchr/testify/assert"
)

/*
	IsTreasure字段没配，商店4废弃
	每个格子只配了一个Group
	格子刷新机制只剩下最普通的了

*/

func TestMain(m *testing.M) {
	gamedata.DebugLoadLocalGamedata()
	retCode := m.Run()
	os.Exit(retCode)
}

func TestStore_GetBlank(t *testing.T) {
	store := &Store{}
	blank := &Blank{BlankID: 233}
	store.Blanks = append(store.Blanks, blank)

	assert.Nil(t, store.GetBlank(666))
	assert.Equal(t, store.GetBlank(233), blank)
}

func TestStore_ManualRefresh(t *testing.T) {
	store := &Store{StoreID: 9}

	acid := ""
	lv := uint32(100)
	now := time.Now().UnixNano()
	rd := rand.New(rand.NewSource(now))

	// 获得数据
	store.refresh(acid, lv, rd)

	store.ManualRefresh(acid, now, lv, rd)

	assert.Equal(t, store.ManualRefreshCount, uint32(1))
	assert.Equal(t, store.LastRefreshTime, now)

	// Refresh后格子随机排列，只能随机几次看结果了

	preBlanks := store.Blanks
	blankChanges := make([]bool, len(store.Blanks))

	for i := 0; i < 20; i++ {
		for j := range store.Blanks {
			// BlankID或者GoodIndex改变都算
			if preBlanks[j].GoodIndex != store.Blanks[j].GoodIndex ||
				preBlanks[j].BlankID != store.Blanks[j].BlankID {
				blankChanges[j] = true
			}
			preBlanks[j] = store.Blanks[j]
		}
		store.ManualRefresh(acid, now, lv, rd)
	}

	r := false
	for _, c := range blankChanges {
		r = r || c
	}

	assert.False(t, r)
}

/*	用来看所有Store的所有BlankID
func TestSomething (t *testing.T) {
	storeIds := gamedata.GetStoreIDs()

	for _, storeId := range storeIds {
		storeCfg := gamedata.GetStoreCfg(storeId)
		blankIDs := storeCfg.GetBlankIDs()
		t.Logf(":", storeId, blankIDs)
	}
}
*/
