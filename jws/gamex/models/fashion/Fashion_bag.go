package fashion

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerFashionBagDB struct {
	Items  FashionItemSlice
	LastId uint32
}

type PlayerFashionBag struct {
	Items  map[uint32]helper.FashionItem
	lastId uint32
}

func (fb *PlayerFashionBag) GetFashionAll() []helper.FashionItem {
	res := make([]helper.FashionItem, 0, len(fb.Items))
	for _, f := range fb.Items {
		res = append(res, f)
	}
	return res
}

func (fb *PlayerFashionBag) GetFashion2String() []string {
	aEquips := fb.GetFashionAll()
	res := make([]string, 0, 10)
	for _, idData := range aEquips {
		res = append(res, idData.TableID)
	}

	return res
}

func (fb *PlayerFashionBag) GetFashions2Client(ids []uint32) []helper.FashionItem {
	res := make([]helper.FashionItem, 0, len(ids))
	for _, id := range ids {
		if item, ok := fb.Items[id]; ok {
			res = append(res, item)
		}
	}
	return res
}

func (fb *PlayerFashionBag) HasFashionByBagId(id uint32) bool {
	_, ok := fb.Items[id]
	return ok
}

func (fb *PlayerFashionBag) GetFashionInfo(id uint32) (bool, helper.FashionItem) {
	f, ok := fb.Items[id]
	return ok, f
}

func (fb *PlayerFashionBag) HasFashionByTableId(tableId string, cfg *ProtobufGen.Item,
	now_time int64) (bool, uint32) {
	for _, f := range fb.Items {
		if f.TableID == tableId {
			if !gamedata.IsFashionPerm(cfg, f.ExpireTimeStamp) &&
				now_time > f.ExpireTimeStamp {
				return false, 0
			} else {
				return true, f.ID
			}
		}
	}
	return false, 0
}

func (fb *PlayerFashionBag) AddFashionByTableId(tableId string, cfg *ProtobufGen.Item,
	now_time int64) (errCode int, item_inner_type int, idx2OldCount map[uint32]int64) {

	idx2OldCount = make(map[uint32]int64, 1)
	if has, _ := fb.HasFashionByTableId(tableId, cfg, now_time); has {
		// 时装加超过一个，不报错了，默默的过去  TDB by zhangzhen
		logs.Info("add fashion %s, but already have !!", tableId)
		//		return helper.RES_AddToBag_MaxCount, helper.Item_Inner_Type_Fashion, idx2OldCount
		return helper.RES_AddToBag_Success, helper.Item_Inner_Type_Fashion, idx2OldCount
	}

	fb.lastId++
	newId := fb.lastId
	fb.Items[newId] = helper.FashionItem{
		ID:              newId,
		TableID:         tableId,
		ExpireTimeStamp: gamedata.CalcFashionTime(cfg, now_time),
	}
	idx2OldCount[newId] = 0
	return helper.RES_AddToBag_Success, helper.Item_Inner_Type_Fashion, idx2OldCount
}

func (fb *PlayerFashionBag) RemoveFashion(bagId uint32) (bool, string) {
	oldItem, ok := fb.Items[bagId]
	if !ok {
		return false, ""
	}
	delete(fb.Items, bagId)
	return true, oldItem.TableID
}

func (fb *PlayerFashionBag) ToDB() PlayerFashionBagDB {
	db_info := PlayerFashionBagDB{
		Items:  make(FashionItemSlice, 0, len(fb.Items)),
		LastId: fb.lastId,
	}
	for _, f := range fb.Items {
		db_info.Items = append(db_info.Items, f)
	}
	sort.Sort(db_info.Items)
	return db_info
}

func (fb *PlayerFashionBag) FromDB(db *PlayerFashionBagDB) error {
	fb.Items = make(map[uint32]helper.FashionItem, len(db.Items))
	for _, item := range db.Items {
		fb.Items[item.ID] = item
	}
	fb.lastId = db.LastId
	return nil
}

type FashionItemSlice []helper.FashionItem

func (fs FashionItemSlice) Len() int {
	return len(fs)
}

func (fs FashionItemSlice) Less(i, j int) bool {
	return fs[i].ID < fs[j].ID
}

func (fs FashionItemSlice) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}
