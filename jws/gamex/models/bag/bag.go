package bag

import (
	"encoding/json"
	"time"

	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dirtycheck"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

const (
	bagInitSize     = 64
	bagLogicMaxSize = 250
	bagMaxSize      = 300
)

/*//////////////////////////////////////
实现一个专门用来堆叠的背包
//////////////////////////////////////*/

//StackBag
//根据索引ID的行为决定如何处理堆叠。
//如果没有索引ID则生成索引ID，默认堆叠数量上限为1，不丢弃
type StackBag struct {
	dbkey      db.ProfileDBKey
	dirtyCheck dirtycheck.DirtyChecker
	Ver        int64 `redis:"version"`

	nextId           uint32 // XXX 0 is reserved
	items            map[uint32]BagItem
	unfixedItemCount map[string]uint32 // 用来存储非固定id物品（装备）的数量

	//LastTime   int64 `redis:"lasttime"`
	CreateTime int64 `redis:"createtime"`
}

func NewStackBag(account db.Account) *StackBag {
	now_t := time.Now().Unix()
	re := &StackBag{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "bag",
		},
		dirtyCheck: dirtycheck.NewDirtyChecker(),
		//Ver:        helper.CurrDBVersion,
		nextId: helper.AUTO_START_ID, //For Weapon,其他都能够固定ID
		items:  make(map[uint32]BagItem),
		//LastTime:   now_t,
		CreateTime: now_t,
	}
	return re
}

//Items
func (b *StackBag) Items() map[uint32]BagItem {
	items := make(map[uint32]BagItem, len(b.items))
	for k, bi := range b.items {
		if !bi.CanDelete() {
			items[k] = bi
		}
	}
	return items
}

func (b *StackBag) ItemToClients() map[uint32]helper.BagItemToClient {
	items := make(map[uint32]helper.BagItemToClient, len(b.items))
	for k, bi := range b.items {
		if !bi.CanDelete() {
			bic := helper.BagItemToClient{}
			FromBagItem2Client(&bic, &bi)
			items[k] = bic
		}
	}
	return items
}

func (b *StackBag) GetItemsToClient(uIds []uint32) map[uint32]helper.BagItemToClient {
	items := make(map[uint32]helper.BagItemToClient, len(uIds))
	for _, uid := range uIds {
		item := b.GetItem(uid)
		if item == nil {
			continue
		}
		bic := helper.BagItemToClient{}
		FromBagItem2Client(&bic, item)
		items[uid] = bic
	}
	return items
}

func (b *StackBag) GetItem(id uint32) *BagItem {
	bi, ok := b.items[id]
	if ok && !bi.CanDelete() {
		return &bi
	} else {
		return nil
	}
}

func (b *StackBag) UpdateItem(item *BagItem) {
	if item == nil {
		return
	}

	b.items[item.ID] = *item
}

func (b *StackBag) GetItems(uIds []uint32) map[uint32]BagItem {
	items := make(map[uint32]BagItem)
	for _, uid := range uIds {
		item := b.GetItem(uid)
		if item == nil {
			continue
		}
		items[uid] = *item
	}
	return items
}

func (b *StackBag) DBName() string {
	return b.dbkey.String()
}

func (b *StackBag) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	//b.LastTime = time.Now().Unix()
	key := b.DBName()
	cmds := redis.Args{}.Add(key)
	del := false
	for k, bi := range b.items {
		if bi.CanDelete() {
			delete(b.items, k)
			cmds = cmds.Add(k)
			del = true
		}
	}

	if del {
		logs.Trace("HDEL %v", cmds)
		if err := cb.Send("HDEL", cmds...); err != nil {
			return err
		}
	}

	items := b.Items() //MUST BE Items()
	items[0] = BagItem{ID: b.nextId}
	//TODO Bag会因为map key的不确定性导致dirtyCheck失效 T2248
	dirty := false
	if forceDirty {
		dirty = true
	} else {
		_, dirty = b.dirtyCheck.Check(BagSerialize(items))
	}
	if dirty {
		return driver.DumpToHashDBCmcBuffer(cb, key, items)
	} else {
		logs.Trace("StackBag DBSave is clean.")
		return nil
	}
}

func (b *StackBag) DBLoad(logInfo bool) error {
	key := b.DBName()

	var jsonitems map[uint32][]byte

	db := driver.GetDBConn()
	defer db.Close()
	err := driver.RestoreFromHashDB(db.RawConn(), key, &jsonitems, false, logInfo)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}

	// 先清空原有的内存数据，因为在登陆后可能强制Load
	b.items = make(map[uint32]BagItem, len(jsonitems))
	b.unfixedItemCount = make(map[string]uint32, len(jsonitems))

	//logs.Trace("[StackBag] DBLoad %v", jsonitems)
	for k, v := range jsonitems {

		var bi BagItem
		err := json.Unmarshal(v, &bi)
		if err != nil && bi.ID == k {
			panic(err)
		}

		if k == 0 {
			b.nextId = bi.ID
			b.items[0] = bi
			continue
		}

		//根据ItemID更新玩家背包中ID为服务器最新ID
		//武器相关的非固定ID不重新生成
		if bi.IsFixedID() {
			uID, ok := gamedata.GetFixedBagID(bi.ItemID)
			if !ok || uID == 0 {
				logs.Error("[DBLoad] for key %s, get unknow item(%s) in Bag. Removed from his bag. %v", key, bi.ItemID, bi)
				continue
			}
			bi.ID = uID
		} else {
			if _, ok := b.unfixedItemCount[bi.TableID]; ok {
				b.unfixedItemCount[bi.TableID] = b.unfixedItemCount[bi.TableID] + 1
			} else {
				b.unfixedItemCount[bi.TableID] = 1
			}
		}

		b.items[bi.ID] = bi
	}
	//
	// 目前有两种物品，非固定id和固定id
	// 对于材料等道具（固定id），生成的id会随着表变化而变化
	// 所以加载进内存时，需要重新分配id
	// 但是此时数据库中还存在老id的记录
	// 如果老id被复用那么存盘时会被覆盖，但是如果老id没有被利用，那么数据库中的哪一项就不会被删除
	// 所以这里添加一个标记老id的物品，count为0，如果这个id被利用，那么没有影响，如果没被利用
	// 数据库中的项就会被删去

	for k, _ := range jsonitems {
		_, ok := b.items[k]
		if !ok {
			b.items[k] = BagItem{
				ID: k,
			}
		}
	}

	b.dirtyCheck.Check(b.items)
	return nil
}

//Remove 从背包中直接去掉某种固定ID物品, 非固定ID物品则什么都不会发生
func (b *StackBag) RemoveWithFixedID(ItemID string) {
	if uID, ok := gamedata.GetFixedBagID(ItemID); ok {
		b.RemoveByID(uID)
	}
}

func (b *StackBag) RemoveByID(ID uint32) {
	if bi, ok := b.items[ID]; ok {
		bi.Count = 0
		b.items[ID] = bi
	}
}

// 获得装备数量
func (b *StackBag) GetEquipCount() uint32 {
	var count uint32
	for _, bi := range b.items {
		if !bi.CanDelete() {
			if gamedata.IsItemNeedCountToBag(bi.TableID) {
				count += uint32(bi.Count)
			}
		}
	}
	return count
}

// 同上 使用背包Id
func (b *StackBag) UseByID(acid string, ID uint32, count uint32) (
	isSuccess, isRemove bool, itemId string, oldCount uint32) {

	logs.Trace("UseByID:%d,%d", ID, count)
	uID := ID
	bi := b.GetItem(uID)
	if bi != nil {
		// 更新物品数量
		_, ok := gamedata.GetFixedBagID(bi.TableID)
		if !ok {
			n, ok := b.unfixedItemCount[bi.TableID]
			if ok {
				if n > 1 {
					b.unfixedItemCount[bi.TableID] = n - 1
				} else {
					delete(b.unfixedItemCount, bi.TableID)
				}
			}
		}
		// 减数量
		oldCount = uint32(bi.Count)
		if uint32(bi.Count) > count {
			bi.Count -= int64(count)
			b.items[uID] = *bi
			return true, false, bi.TableID, oldCount
		} else if uint32(bi.Count) == count {
			b.RemoveByID(uID)
			return true, true, bi.TableID, oldCount
		} else {
			return false, false, "", 0
		}
	}
	return false, false, "", 0
}

func (b *StackBag) GetItemIDByItem(i *BagItem) string {
	return i.TableID
}

func getStackBagItemData(ItemID string) (
	ok bool,
	NewItemID string,
	maxOwnNum uint32) {

	ok = false
	NewItemID = ItemID

	item, ok := gamedata.GetProtoItem(ItemID)
	if ok {
		//logs.Trace("[StackBag] Add %v, %v", item, ok)
		// 如果存在，则根据表格中要求更新堆叠上线数量，更新丢弃与否的行为
		if !gamedata.IsFixedIDItem(item.GetType()) {
			//Weapon数据类型ItemID是UUID每次都不同
			NewItemID = uuid.NewV4().String()
		}
		maxOwnNum = item.GetOwnMaxNum()
	}
	//logs.Trace("getStackBagItemData %v %d %v %s",
	//	ok, iStackUpLimit, bDrop, NewItemID)
	return
}

func (b *StackBag) Add(data gamedata.BagItemData, ItemID string, count uint32, acid string, rd *rand.Rand, now_time int64) (
	errCode int, idx2OldCount map[uint32]int64) {
	ok, NewItemID, maxOwnNum := getStackBagItemData(ItemID)
	if !ok {
		return helper.RES_AddToBag_Err, nil
	}

	errCode = helper.RES_AddToBag_Success
	newData := data
	uID, ok := gamedata.GetFixedBagID(ItemID)
	if ok {
		idx2OldCount = make(map[uint32]int64, 1)
		if data.IsNil() {
			newDataP := gamedata.MakeItemData(acid, rd, ItemID)
			if newDataP != nil {
				newData = *newDataP
			}
		}
		//物品是固定数字ID物品
		//查询ItemID是否已经在背包中存在
		nbi, had := b.items[uID]
		if had {
			newCount := nbi.Count + int64(count)
			if newCount > int64(maxOwnNum) {
				newCount = int64(maxOwnNum)
				//errCode = helper.RES_AddToBag_MaxCount
			}
			idx2OldCount[uID] = nbi.Count
			b.items[uID] = b._genItem(uID, ItemID, NewItemID, newData, newCount)
		} else {
			newCount := int64(count)
			if newCount > int64(maxOwnNum) {
				newCount = int64(maxOwnNum)
				//errCode = helper.RES_AddToBag_MaxCount
			}
			idx2OldCount[uID] = 0
			b.items[uID] = b._genItem(uID, ItemID, NewItemID, newData, newCount)
		}
	} else {
		idx2OldCount = make(map[uint32]int64, count)
		for i := 0; i < int(count); i++ {
			n, ok := b.unfixedItemCount[ItemID]
			if ok {
				if n >= maxOwnNum {
					return helper.RES_AddToBag_MaxCount, nil
				}
			}
			if data.IsNil() {
				newDataP := gamedata.MakeItemData(acid, rd, ItemID)
				if newDataP != nil {
					newData = *newDataP
				}
			}
			uID = b.getNewId()
			b.items[uID] = b._genItem(uID, ItemID, NewItemID, newData, 1)
			idx2OldCount[uID] = 0
			b.unfixedItemCount[ItemID] = n + 1
		}
	}
	return
}

func (b *StackBag) _genItem(uID uint32, ItemID, NewItemID string,
	data gamedata.BagItemData, count int64) BagItem {
	var bi BagItem
	bi.ID = uID
	bi.TableID = ItemID
	bi.ItemID = NewItemID
	bi.ItemData = data
	bi.Count = count
	return bi
}

// 判断是否可以添加某物品进背包，目前是判断是否超最大物品数量上限
func (b *StackBag) IsCanAddItem(itemId string, count uint32) bool {
	itemCfg, ok := gamedata.GetProtoItem(itemId)
	if !ok {
		return false
	}
	if count > itemCfg.GetOwnMaxNum() {
		logs.Debug("IsCanAddItem GetOwnMaxNum %s %d", itemId, count)
		return false
	}
	if uID, ok := gamedata.GetFixedBagID(itemId); ok {
		nbi, had := b.items[uID]
		if had && nbi.Count+int64(count) >= int64(itemCfg.GetOwnMaxNum()) {
			return false
		}
	} else {
		n, ok := b.unfixedItemCount[itemId]
		if ok && n+count >= itemCfg.GetOwnMaxNum() {
			return false
		}
	}
	return true
}

//Has 是否有某个固定数字ID的物品在背包中，满足数量count
//Weapon等非固定数字ID的物品，因为ItemID是UUID所以，永远返回false
func (b *StackBag) Has(ItemID string, count uint32) bool {
	uID, ok := gamedata.GetFixedBagID(ItemID)
	if !ok {
		return false
	}

	bi := b.items[uID]
	return uint32(bi.Count) >= count
}

//GetCount通常在合成前，或者进入关卡前，玩家需要某种物品的数量显示
//目前只支持固定ID的物品
func (b *StackBag) GetCount(ItemID string) uint32 {
	if uID, ok := gamedata.GetFixedBagID(ItemID); ok {
		return b.GetCountByBagId(uID)
	}
	return 0
}

func (b *StackBag) GetCountByBagId(bag_id uint32) uint32 {
	item := b.GetItem(bag_id)
	if item != nil {
		return uint32(item.Count)
	}
	return 0
}

// 获取一个物品的信息
func (b *StackBag) GetItemData(bag_id uint32) (*ProtobufGen.Item, bool) {
	//logs.Trace("All Bag %v", b.items)
	//logs.Trace("id %d -> %v", bag_id, b.items[bag_id])
	item := b.GetItem(bag_id)
	if item == nil {
		return nil, false
	}

	item_id := b.GetItemIDByItem(item)

	return gamedata.GetProtoItem(item_id)
}

// id对应的装备是否拥有
func (b *StackBag) IsHasBagId(bag_id uint32) bool {
	return b.GetItem(bag_id) != nil
}

// 获取一个新的Id
func (b *StackBag) getNewId() uint32 {
	b.nextId++
	return b.nextId
}

func IsFixedID(id uint32) bool {
	if id < helper.AUTO_START_ID && id != 0 {
		return true
	}
	return false
}
