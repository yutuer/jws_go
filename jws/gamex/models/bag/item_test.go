package bag

import (
	"testing"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"github.com/stretchr/testify/assert"
)

func TestBagItem_CanDelete4(t *testing.T) {
	bic := make(chan *BagItem)
	defer close(bic)

	go func() {
		bic <- &BagItem{ID: 11, Count: 0}
		bic <- &BagItem{ID: 1, Count: 255}
		bic <- &BagItem{ID: 0, Count: 233}
	}()

	assert.True(t, (<-bic).CanDelete())
	assert.False(t, (<-bic).CanDelete())
	assert.True(t, (<-bic).CanDelete())
}

func TestBagItem_GetDataStr(t *testing.T) {
	bic := make(chan *BagItem)
	defer close(bic)

	itemData := gamedata.BagItemData{Id: 2333}

	go func() {
		bic <- &BagItem{ID: 11, Count: 1, ItemData: itemData}
		bic <- &BagItem{ID: 11, Count: 1}
		bic <- &BagItem{ID: 11, Count: 1, data: "Something"}
		bic <- &BagItem{ID: 11, Count: 1, data: "Something", ItemData: itemData}
	}()

	assert.Contains(t, (<-bic).GetDataStr(), "2333")
	assert.Equal(t, (<-bic).GetDataStr(), "")
	assert.Equal(t, (<-bic).GetDataStr(), "Something")
	assert.Equal(t, (<-bic).GetDataStr(), "Something")
}

func TestBagItem_ToClient(t *testing.T) {
	// 实在不知道要测什么了...
	assert.NotNil(t, (&BagItem{}).ToClient())
}

func TestBagItem_GetItemData(t *testing.T) {
	bic := make(chan *BagItem)
	defer close(bic)

	itemData := gamedata.BagItemData{Id: 2333}

	go func() {
		bic <- &BagItem{}
		bic <- &BagItem{ItemData: itemData}
	}()

	assert.NotNil(t, (<-bic).GetItemData())
	assert.NotNil(t, (<-bic).GetItemData())
}

func TestBagItem_SetItemData(t *testing.T) {
	bi := &BagItem{}
	itemData := &gamedata.BagItemData{Id: 2333}

	bi.SetItemData(itemData)
	assert.Equal(t, bi.ItemData, *itemData)

	itemData.Id = 32768
	assert.NotEqual(t, bi.ItemData, *itemData)
}

func TestBagItem_IsFixedID(t *testing.T) {
	bic := make(chan *BagItem)
	defer close(bic)

	go func() {
		bic <- &BagItem{}
		bic <- &BagItem{ID: 11}
		bic <- &BagItem{ID: uint32(0x8000000)}
	}()

	assert.False(t, (<-bic).IsFixedID())
	assert.True(t, (<-bic).IsFixedID())
	assert.False(t, (<-bic).IsFixedID())
}
