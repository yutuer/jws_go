package bag

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gamedata.DebugLoadLocalGamedata()
	driver.SetupRedis(":6379", 1, "", true)

	retCode := m.Run()

	os.Exit(retCode)
}

func DebugGetDBAccount() *db.Account {
	dbAccount := &db.Account{
		GameId:  0,
		ShardId: 0,
		UserId:  db.NewUserID(),
	}

	return dbAccount
}

func DebugGetNewStackBag() *StackBag {
	// 一个NewStackBag

	dbAccount := DebugGetDBAccount()
	bag := NewStackBag(*dbAccount)
	bag.unfixedItemCount = make(map[string]uint32)

	return bag
}

func (bag *StackBag) DebugAddBagItem(amount int) {
	// 加1个物品
	for i := 0; i < amount; i++ {
		bag.items[bag.nextId] = BagItem{ID: bag.nextId, Count: 1}
		bag.nextId++
	}
}

func (bag *StackBag) DebugAddBagItemWithoutAmount(amount int) {
	// 只加物品ID
	for i := 0; i < amount; i++ {
		bag.items[bag.nextId] = BagItem{ID: bag.nextId}
		bag.nextId++
	}
}

func (bag *StackBag) DebugFastAddItem(itemID string, amount uint32) {
	// 只改物品ID和数量
	bagItemData := gamedata.NewBagItemData()
	acid := uuid.NewV4().String()
	now_time := time.Now().UnixNano()
	rd := rand.New(rand.NewSource(now_time))

	bag.Add(*bagItemData, itemID, amount, acid, rd, now_time)
}

func TestNewStackBag(t *testing.T) {
	dbAccount := DebugGetDBAccount()
	bag := NewStackBag(*dbAccount)

	assert.NotNil(t, bag.dbkey)
	assert.NotNil(t, bag.nextId)
	assert.NotNil(t, bag.CreateTime)
	assert.NotNil(t, bag.dirtyCheck)
	assert.Empty(t, bag.items)
}

func TestStackBag_Items(t *testing.T) {
	// 需要 nextId
	bg := DebugGetNewStackBag()
	bg.DebugAddBagItem(50)
	bg.DebugAddBagItemWithoutAmount(50)

	bis := bg.Items()
	assert.Equal(t, len(bis), 50)
}

func TestStackBag_ItemToClients(t *testing.T) {
	// 需要 nextId
	bg := DebugGetNewStackBag()
	bg.DebugAddBagItem(50)

	bic := bg.ItemToClients()
	assert.Equal(t, len(bic), 50)
}

func TestStackBag_GetItem(t *testing.T) {
	bg := DebugGetNewStackBag()
	uId := bg.nextId + uint32(32)

	bg.DebugAddBagItem(32)
	bg.DebugAddBagItemWithoutAmount(32)

	randInt := uint32(rand.Intn(31))

	// 不存在
	assert.Nil(t, bg.GetItem(uId>>1))
	// 正常物品
	assert.Equal(t, bg.items[uId-randInt], *bg.GetItem(uId - randInt))
	// CanDelete
	assert.Nil(t, bg.GetItem(uId+randInt))
}

func TestStackBag_UpdateItem(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.UpdateItem(nil)

	bg.items[bg.nextId] = BagItem{ID: bg.nextId, Count: 233}

	bi := &BagItem{ID: bg.nextId, Count: 666}
	bg.UpdateItem(bi)

	assert.Equal(t, bg.items[bg.nextId].Count, int64(666))
}

func TestStackBag_GetItems(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugAddBagItemWithoutAmount(32)
	bg.DebugAddBagItem(32)

	var uIds []uint32
	for uId := range bg.items {
		uIds = append(uIds, uId)
	}

	// 不存在
	assert.Empty(t, bg.GetItems([]uint32{bg.nextId >> 1}))
	// 存在&全部
	assert.Equal(t, bg.Items(), bg.GetItems(uIds))
}

func TestStackBag_GetItemsToClient(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugAddBagItem(16)
	bg.DebugAddBagItemWithoutAmount(12)

	var uIds []uint32
	for uId := range bg.items {
		uIds = append(uIds, uId)
	}

	bics := bg.GetItemsToClient(uIds)

	assert.Equal(t, len(bics), 16)
}

func TestStackBag_Add(t *testing.T) {
	bg := DebugGetNewStackBag()

	bagItemData := gamedata.NewBagItemData()
	acid := uuid.NewV4().String()
	now_time := time.Now().UnixNano()
	rd := rand.New(rand.NewSource(now_time))

	fixedItemID := "MAT_StarStone"
	fixeduId, _ := gamedata.GetFixedBagID(fixedItemID)

	// 可堆叠物品，之前数量为0
	e, idx20oldCount := bg.Add(*bagItemData, fixedItemID, 111111, acid, rd, now_time)

	assert.Equal(t, 0, e)
	assert.Equal(t, int64(0), idx20oldCount[fixeduId])
	assert.Equal(t, bg.items[fixeduId].Count, int64(111111))

	// 可堆叠物品，超过上限
	e, idx20oldCount = bg.Add(*bagItemData, fixedItemID, 999999, acid, rd, now_time)

	assert.Equal(t, 0, e)
	assert.Equal(t, int64(111111), idx20oldCount[fixeduId])
	assert.Equal(t, bg.items[fixeduId].Count, int64(999999))

	// 不可堆叠物品
	flexItemID := "WP_ALL_1_1"
	FlexUID := bg.nextId + 1

	e, idx20oldCount = bg.Add(*bagItemData, flexItemID, 10, acid, rd, now_time)

	assert.Equal(t, 0, e)
	assert.Equal(t, int64(0), idx20oldCount[FlexUID])
	assert.Equal(t, len(idx20oldCount), 10)
	assert.Equal(t, bg.items[FlexUID].Count, int64(1))
	assert.Equal(t, bg.unfixedItemCount[flexItemID], uint32(10))
}

func TestStackBag_DBName(t *testing.T) {
	bg := DebugGetNewStackBag()
	assert.NotNil(t, bg.DBName())
}

func TestStackBag_DBSave(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	bg.DebugFastAddItem("WP_ALL_1_1", 5)
	cb := redis.NewCmdBuffer()

	assert.NoError(t, bg.DBSave(cb, false))
}

func TestStackBag_DBLoad(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	bg.DebugFastAddItem("WP_ALL_1_1", 5)
	cb := redis.NewCmdBuffer()

	// 直接Load会认为是新的没有存档
	bg.DBSave(cb, false)
	assert.NoError(t, bg.DBLoad(false))
}

func TestStackBag_RemoveByID(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	bg.RemoveByID(bg.nextId)
	bg.RemoveByID(uID)

	assert.Equal(t, int64(0), bg.items[uID].Count)
}

func TestStackBag_RemoveWithFixedID(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	bg.RemoveWithFixedID("MAT_StarStone")

	assert.Equal(t, int64(0), bg.items[uID].Count)
}

func TestStackBag_GetCount(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)

	assert.Equal(t, bg.GetCount("MAT_StarStone"), uint32(999))
}

func TestStackBag_GetCountByBagId(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	bg.GetCountByBagId(uID)

	assert.Equal(t, int64(999), bg.items[uID].Count)
}

func TestStackBag_Has(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	bg.DebugFastAddItem("WP_ALL_1_1", 2)

	assert.True(t, bg.Has("MAT_StarStone", 0))
	assert.True(t, bg.Has("MAT_StarStone", 999))
	assert.False(t, bg.Has("MAT_StarStone", 999999))
	assert.False(t, bg.Has("WP_ALL_1_1", 1))
}

func TestStackBag_IsHasBagId(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	assert.False(t, bg.IsHasBagId(bg.nextId))
	assert.True(t, bg.IsHasBagId(uID))
}

func TestStackBag_IsCanAddItem(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	bg.DebugFastAddItem("WP_ALL_1_1", 99)

	assert.True(t, bg.IsCanAddItem("WP_ALL_1_1", 10))
	assert.True(t, bg.IsCanAddItem("WP_ALL_1_1", 999))
	assert.True(t, bg.IsCanAddItem("MAT_StarStone", 999))
	assert.False(t, bg.IsCanAddItem("MAT_StarStone", 999999))
	assert.False(t, bg.IsCanAddItem("MAT_StarStone", 999999999))
}

func TestStackBag_GetEquipCount(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("WP_ALL_1_1", 2)
	bg.DebugFastAddItem("MAT_StarStone", 999)

	assert.Equal(t, bg.GetEquipCount(), uint32(2))
}

func TestStackBag_GetItemData(t *testing.T) {
	bg := DebugGetNewStackBag()
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	item, ok := bg.GetItemData(uID)
	assert.True(t, ok)
	assert.Equal(t, *item.ID, "MAT_StarStone")
}

func TestStackBag_GetItemIDByItem(t *testing.T) {
	bg := DebugGetNewStackBag()
	bagItem := &BagItem{ID: bg.nextId, TableID: "Test", Count: 1}
	bg.items[bg.nextId] = *bagItem

	assert.Equal(t, bg.GetItemIDByItem(bagItem), "Test")
}

func TestStackBag_UseByID(t *testing.T) {
	// 这个功能的源码画风突变
	bg := DebugGetNewStackBag()

	// 固定ID
	bg.DebugFastAddItem("MAT_StarStone", 999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	// 没有
	isSuccess, isRemoved, itemId, oldCount := bg.UseByID("", bg.nextId, 999)
	assert.True(t, !isSuccess && !isRemoved && itemId == "" && oldCount == 0)

	// 可以
	isSuccess, isRemoved, itemId, oldCount = bg.UseByID("", uID, 100)
	assert.True(t, isSuccess && !isRemoved && itemId == "MAT_StarStone" && oldCount == 999)

	// 数量不够
	isSuccess, isRemoved, itemId, oldCount = bg.UseByID("", uID, 999)
	assert.True(t, !isSuccess && !isRemoved && itemId == "" && oldCount == 0)

	// 用完
	isSuccess, isRemoved, itemId, oldCount = bg.UseByID("", uID, 899)
	assert.True(t, isSuccess && isRemoved && itemId == "MAT_StarStone" && oldCount == 899)

	// 可变ID
	bg.DebugFastAddItem("WP_ALL_1_1", 2)
	id := bg.nextId - 1

	// 数量不够
	isSuccess, isRemoved, itemId, oldCount = bg.UseByID("", id, 2)
	assert.True(t, !isSuccess && !isRemoved && itemId == "" && oldCount == 0)

	// 用完
	isSuccess, isRemoved, _, oldCount = bg.UseByID("", id, 1)
	assert.True(t, isSuccess && isRemoved && oldCount == 1)
}

func BenchmarkStackBag_Add(b *testing.B) {
	bg := DebugGetNewStackBag()

	// 模拟背包接近满时
	bg.DebugFastAddItem("WP_ALL_1_1", 250)

	bagItemData := gamedata.NewBagItemData()
	acid := uuid.NewV4().String()
	now_time := time.Now().UnixNano()
	rd := rand.New(rand.NewSource(now_time))

	// log的开销比本身操作高，关掉
	logs.Close()

	for i := 0; i < b.N; i++ {
		bg.Add(*bagItemData, "MAT_StarStone", 1, acid, rd, now_time)
	}
}

func BenchmarkStackBag_UseByID(b *testing.B) {
	bg := DebugGetNewStackBag()

	// 模拟背包接近满时
	bg.DebugFastAddItem("WP_ALL_1_1", 250)
	bg.DebugFastAddItem("MAT_StarStone", 999999)
	uID, _ := gamedata.GetFixedBagID("MAT_StarStone")

	// log的开销比本身操作高，关掉
	logs.Close()

	for i := 0; i < b.N; i++ {
		if bg.items[uID].Count == 0 {
			bg.DebugFastAddItem("MAT_StarStone", 999999)
		}
		bg.UseByID("", uID, 1)
	}
}
