package account

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	宝石物品脱离通用bag和item单独实现，因为宝石经验相同情况下需要堆叠，经验不同时不堆叠
	用slice实现，根据id升序排列，但根据tableid和jadeexp查找时就遍历吧，总量限制在200以内
*/
const all_jade_max_count = 200

type JadeItem struct {
	ID            uint32 `json:"id" codec:"id"`
	TableID       string `json:"tid" codec:"tid"`
	Count         int64  `json:"c" codec:"c"`
	CountNotInBag int64  `json:"cnb" codec:"cnb"` // 不在包裹里的数量
	JadeExp       uint32 `json:"exp" codec:exp"`  // 龙玉经验
}

func (j JadeItem) CountInBag() int64 {
	return j.Count - j.CountNotInBag
}

type PlayerJadeBagDB struct {
	Jades  jade_slice
	NextId uint32
}

type PlayerJadeBag struct {
	JadesMap map[uint32]JadeItem
	NextId   uint32
}

func (jb *PlayerJadeBag) GetFullJade2Client() []JadeItem {
	res := make([]JadeItem, 0, len(jb.JadesMap))
	for _, j := range jb.JadesMap {
		res = append(res, j)
	}
	return res
}

func (jb *PlayerJadeBag) GetJades2Client(ids []uint32) map[uint32]JadeItem {
	res := make(map[uint32]JadeItem, len(ids))
	for _, id := range ids {
		if j, ok := jb.JadesMap[id]; ok {
			res[j.ID] = j
		}
	}
	return res
}

func (jb *PlayerJadeBag) AddJadeByTableId(tableId string, count int64, exp uint32, jadeCfg *ProtobufGen.Item) (
	errCode int, item_inner_type int, idx2OldCount map[uint32]int64) {

	errCode = helper.RES_AddToBag_Success
	maxOwnNum := jadeCfg.GetOwnMaxNum()
	idx2OldCount = make(map[uint32]int64, 1)
	if exp <= 0 {
		exp = gamedata.GetJadeExpByLvl(jadeCfg.GetJadeLevel())
	}

	// 找已有的进行合并
	for k, j := range jb.JadesMap {
		if j.TableID == tableId && j.JadeExp == exp {
			idx2OldCount[j.ID] = j.Count
			newCount := j.Count + count
			if newCount > int64(maxOwnNum) {
				newCount = int64(maxOwnNum)
				errCode = helper.RES_AddToBag_MaxCount
			}
			j.Count = newCount
			jb.JadesMap[k] = j
			logs.Trace("AddJade merge res %v", j)
			return errCode, helper.Item_Inner_Type_Jade, idx2OldCount
		}
	}

	if count > int64(maxOwnNum) {
		count = int64(maxOwnNum)
		errCode = helper.RES_AddToBag_MaxCount
	}

	// 没有相同的, 宝石总数会在特定的获得地方检查，这里一定会加到包裹
	id := jb.nextId()
	jb.JadesMap[id] = JadeItem{
		ID:      id,
		TableID: tableId,
		Count:   count,
		JadeExp: exp,
	}
	idx2OldCount[id] = 0
	logs.Trace("AddJade new res %v", jb.JadesMap[id])
	return errCode, helper.Item_Inner_Type_Jade, idx2OldCount
}

func (jb *PlayerJadeBag) RemoveJade(id uint32, count int64) (
	isSuccess, isRemove bool, itemId string, oldCount uint32) {
	if j, ok := jb.JadesMap[id]; ok {
		countInBag := j.CountInBag()
		if countInBag > count || (countInBag == count && j.CountNotInBag > 0) {
			j.Count = j.Count - count
			jb.JadesMap[id] = j
			return true, false, j.TableID, uint32(countInBag)
		} else if j.CountNotInBag <= 0 && j.Count == count { // 删除
			delete(jb.JadesMap, id)
			return true, true, j.TableID, uint32(countInBag)
		}
	}

	return false, false, "", 0
}

func (jb *PlayerJadeBag) GetJadeSumCount() uint32 {
	return uint32(len(jb.JadesMap))
}

func (jb *PlayerJadeBag) GetJade(id uint32) *JadeItem {
	if j, ok := jb.JadesMap[id]; ok {
		return &j
	}
	return nil
}

func (jb *PlayerJadeBag) GetJadeInBagCount(id uint32) int64 {
	if j, ok := jb.JadesMap[id]; ok {
		return j.CountInBag()
	}
	return 0
}

func (jb *PlayerJadeBag) GetJadeData(id uint32) (*ProtobufGen.Item, bool) {
	if j, ok := jb.JadesMap[id]; ok {
		if ok, cfg := gamedata.IsJade(j.TableID); ok {
			return cfg, true
		}
	}
	return nil, false
}

func (jb *PlayerJadeBag) TakeOutFromBag(id uint32) {
	if j, ok := jb.JadesMap[id]; ok {
		if j.Count-j.CountNotInBag > 0 {
			j.CountNotInBag++
			jb.JadesMap[id] = j
		}
	}
}

func (jb *PlayerJadeBag) PutInToBag(id uint32) {
	if j, ok := jb.JadesMap[id]; ok {
		if j.CountNotInBag > 0 {
			j.CountNotInBag--
			jb.JadesMap[id] = j
		}
	}
}

func (jb *PlayerJadeBag) ToDB() PlayerJadeBagDB {
	db_info := PlayerJadeBagDB{
		Jades:  make(jade_slice, 0, len(jb.JadesMap)),
		NextId: jb.NextId,
	}
	for _, j := range jb.JadesMap {
		db_info.Jades = append(db_info.Jades, j)
	}
	sort.Sort(db_info.Jades)
	return db_info
}

func (p *PlayerJadeBag) FromDB(data *PlayerJadeBagDB) error {
	p.NextId = data.NextId
	p.JadesMap = make(map[uint32]JadeItem, len(data.Jades))
	for _, j := range data.Jades {
		p.JadesMap[j.ID] = j
	}
	return nil
}

func (jb *PlayerJadeBag) nextId() uint32 {
	jb.NextId++
	return jb.NextId
}

type jade_slice []JadeItem

func (js jade_slice) Len() int {
	return len(js)
}

func (js jade_slice) Less(i, j int) bool {
	return js[i].ID < js[j].ID
}

func (js jade_slice) Swap(i, j int) {
	js[i], js[j] = js[j], js[i]
}
