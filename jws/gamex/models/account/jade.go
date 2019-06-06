package account

import (
	"container/heap"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/jade"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Jade2Client struct {
	Id uint32 `codec:"id"`
}

const (
	Jade_Equip_ObjTyp = iota
	Jade_DestinyGen_ObjTyp
)

const (
	_ = iota
	Err_Obj_Not_Lock
	Err_Jade_Slot_Cfg_Not_Found
	Err_Corp_Lvl_Not_Enough
	Err_Jade_Already_Equip
	Err_Jade_Not_Equip // 5
	Err_Jade_Slot_Not_Match
	Err_Jade_Cfg_Not_Found
	Err_Jade_AddBag
	Err_Obj_Jade_Slot_Not_Unlock
	Err_Jade_Count_Not_Enough // 10
	Err_Jade_Cost
	Err_Jade_Give
)

func GetJadeImp(a *Account, objTyp int) IJade {
	switch objTyp {
	case Jade_Equip_ObjTyp:
		return a.Profile.GetEquipJades()
	case Jade_DestinyGen_ObjTyp:
		return a.Profile.GetDestGeneralJades()
	default:
		logs.Error("[GetJade] objTyp %d not found", objTyp)
	}
	return nil
}

type IJade interface {
	IsObjUnlock(a *Account, objId int) (ok bool, errcode int, errmsg string)
	IsObjSlotUnlock(a *Account, objId, slot_in_obj int) (ok bool, errcode int, errmsg string)
	EquipJade(a *Account, objId int, slot_in_obj int, jadeItemId uint32,
		jadeItemCfgId string, sync helper.ISyncRsp) (ok bool, errcode int, errmsg string, warnCode int)
	UnEquipJade(a *Account, objId int, slot_in_obj int, sync helper.ISyncRsp)
	CurrAll2Client() ([]int, []Jade2Client)
	AutoEquip(a *Account, objId int, sync helper.ISyncRsp) (isUpdate bool)
	JadeLvlUp(a *Account, objId int, slot_in_obj int, jadeItemId uint32, costJadeIds []uint32, JadeCostCount []int,
		sync helper.ISyncRsp) (ok bool, resJadeId uint32, errcode int, errmsg string, warnCode int)
	SyncUpdate(sync helper.ISyncRsp)
	resetJadeCap(slot int)
	getByIdx(idx int) uint32
	setByIdx(idx int, jadeId uint32)
}

func autoEquip(a *Account, objId int, jadeImp IJade, start, end int,
	sync helper.ISyncRsp) (isUpdate bool) {
	jadeBag := a.Profile.GetJadeBag()
	jades := jadeBag.GetFullJade2Client()
	// 先将背包里的宝石按槽位分组
	leftJade := make(map[int]jade.JadeHeap, gamedata.JadePartCount)
	for _, item := range jades {
		if ok, cfg := gamedata.IsJade(item.TableID); ok {
			if jadeBag.GetJadeInBagCount(item.ID) > 0 {
				slot := gamedata.GetJadeSlot(cfg.GetPart())
				if _, ok := leftJade[slot]; !ok {
					leftJade[slot] = make(jade.JadeHeap, 0, 10)
				}
				heap := leftJade[slot]
				heap = append(heap, &jade.JadeItem{
					Id:      item.ID,
					Cfg:     cfg,
					JadeLvl: cfg.GetJadeLevel(),
					Exp:     item.JadeExp,
				})
				leftJade[slot] = heap
			}
		}
	}

	// 当前包裹里没有宝石，直接退出
	if len(leftJade) <= 0 {
		return
	}

	for _, jadeHeap := range leftJade {
		heap.Init(&jadeHeap)
	}

	jadeImp.resetJadeCap(end - 1)
	not_empty_slots := make([]int, 0, gamedata.JadePartCount)
	// 优先找空位
	slot := 0
	for i := start; i < end; i++ {
		// 是否空解锁了
		ok, _, _ := jadeImp.IsObjSlotUnlock(a, objId, slot)
		if !ok {
			continue
		}
		if jadeImp.getByIdx(i) <= 0 {
			jadeHeap, ok := leftJade[slot]
			if ok {
				j := heap.Pop(&jadeHeap).(*jade.JadeItem)
				jadeImp.setByIdx(i, j.Id)
				jadeBag.TakeOutFromBag(j.Id)
				sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, j.Id,
					jadeBag.GetJade(j.Id).Count, "ChgJade")
				isUpdate = true
			}
		} else {
			not_empty_slots = append(not_empty_slots, i)
		}
		slot++
	}
	// 在找提高的
	for _, i := range not_empty_slots {
		oldJade := jadeImp.getByIdx(i)
		jadeIm := a.Profile.GetJadeBag().GetJade(oldJade)
		_, cfg := gamedata.IsJade(jadeIm.TableID)
		slot := gamedata.GetJadeSlot(cfg.GetPart())
		jadeHeap, ok := leftJade[slot]
		if ok {
			j := heap.Pop(&jadeHeap).(*jade.JadeItem)
			if j.JadeLvl > cfg.GetJadeLevel() ||
				j.Exp > jadeIm.JadeExp {
				// 脱下旧的
				jadeBag.PutInToBag(oldJade)
				sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, oldJade,
					jadeBag.GetJade(oldJade).Count, "ChgJade")
				// 穿上新的
				jadeImp.setByIdx(i, j.Id)
				jadeBag.TakeOutFromBag(j.Id)
				sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, j.Id,
					jadeBag.GetJade(j.Id).Count, "ChgJade")

				isUpdate = true
			} else {
				heap.Push(&jadeHeap, j)
			}
		}
	}
	return isUpdate
}

func jadeLvlUp(a *Account, jadeImp IJade, objId int, slot_in_obj, start, end int,
	jadeItemId uint32, costJadeIds []uint32, JadeCostCount []int, sync helper.ISyncRsp) (
	ok bool, resJadeId uint32, errcode int, errmsg string, warnCode int) {

	// 目标宝石是否在身上
	jadeIdx := -1
	s := 0
	for i := start; i < end; i++ {
		if jadeImp.getByIdx(i) == jadeItemId {
			jadeIdx = i
			break
		}
		s++
	}
	if jadeIdx < 0 {
		logs.Warn("jadeLvlUp Err_Avatar_Jade_Not_Equip avatar %d jadeitem %d", objId, jadeItemId)
		return false, 0, 0, "", errCode.ClickTooQuickly
	}
	if s != slot_in_obj {
		return false, 0, Err_Jade_Slot_Not_Match,
			fmt.Sprintf("Err_Avatar_Jade_Slot_Not_Match avatar %d slot %d %d", objId, s, slot_in_obj), 0
	}
	// 消耗的宝石检查
	jadeBag := a.Profile.GetJadeBag()
	for i, jId := range costJadeIds {
		// 消耗的宝石是否已经装备了
		if jadeBag.GetJadeInBagCount(jId) < int64(JadeCostCount[i]) {
			return false, 0, Err_Jade_Already_Equip,
				fmt.Sprintf("Err_Avatar_Jade_Already_Equip avatar %d jadeitem %d", objId, jadeItemId),
				errCode.ClickTooQuickly
		}
	}
	// 加exp
	jadeItem := jadeBag.GetJade(jadeItemId)
	newExp := jadeItem.JadeExp
	for i, jId := range costJadeIds {
		jade := jadeBag.GetJade(jId)
		if jade == nil {
			continue
		}
		c := JadeCostCount[i]
		newExp += jade.JadeExp * uint32(c)
	}
	if newExp > gamedata.JadeMaxExp {
		newExp = gamedata.JadeMaxExp
	}

	newLvl := gamedata.GetJadeLvlByExp(newExp)
	cost := CostGroup{}
	// 宝石升级
	ok, newJadeId := gamedata.GetJadeBySlotLvl(slot_in_obj, newLvl)
	if !ok {
		return false, 0, Err_Jade_Cfg_Not_Found,
			fmt.Sprintf("Err_Avatar_Jade_Cfg_Not_Found avatar %d slot %d lvl %d", objId, slot_in_obj, newLvl), 0
	}
	// 删除原有宝石
	jadeBag.PutInToBag(jadeItemId)
	cost.AddJadeByBagId(a, jadeItemId, 1)
	// 换上新宝石
	_, newJadeCfg := gamedata.IsJade(newJadeId)
	errCode, item_inner_type, idx2OldCount := a.Profile.GetJadeBag().AddJadeByTableId(newJadeId, 1, newExp, newJadeCfg)
	if errCode != helper.RES_AddToBag_Success {
		return false, 0, Err_Jade_AddBag,
			fmt.Sprintf("Err_Avatar_Jade_AddBag avatar %d oldjade %d newjade %s", objId, jadeItemId, newJadeId), 0
	}
	for newJadeBagId, oldCount := range idx2OldCount {
		// 穿上新宝石
		jadeImp.setByIdx(jadeIdx, newJadeBagId)
		jadeBag.TakeOutFromBag(newJadeBagId)
		sync.OnChangeUpdateItems(item_inner_type, newJadeBagId, oldCount, "JadeLvlUp")
		resJadeId = newJadeBagId
		break
	}

	// 删除消耗的宝石
	for i, jId := range costJadeIds {
		cost.AddJadeByBagId(a, jId, uint32(JadeCostCount[i]))
	}
	cost.CostBySync(a, sync, "JadeLvlUp")
	return true, resJadeId, 0, "", 0
}

const jade_bag_lvlUp_count = 3

func JadeLvlUpInBag(a *Account, jadeInBag uint32, isAll bool, sync helper.ISyncRsp) (
	ok bool, resJadeId uint32, errcode int, errmsg string) {

	jadeBag := a.Profile.GetJadeBag()
	// 消耗的宝石是否已经装备了
	jc := jadeBag.GetJadeInBagCount(jadeInBag)
	if jc < int64(jade_bag_lvlUp_count) {
		return false, 0, Err_Jade_Count_Not_Enough,
			fmt.Sprintf("Err_Jade_Count_Not_Enough  jadeitem %d", jadeInBag)
	}

	count := 1
	costC := jade_bag_lvlUp_count
	if isAll {
		count = int(jc) / jade_bag_lvlUp_count
		costC = jade_bag_lvlUp_count * count
	}
	// 加exp
	jadeItem := jadeBag.GetJade(jadeInBag)
	newExp := jadeItem.JadeExp
	for i := 0; i < jade_bag_lvlUp_count-1; i++ {
		newExp += jadeItem.JadeExp
	}
	if newExp > gamedata.JadeMaxExp {
		newExp = gamedata.JadeMaxExp
	}

	_, cfg := gamedata.IsJade(jadeItem.TableID)
	newLvl := gamedata.GetJadeLvlByExp(newExp)
	cost := CostGroup{}
	cost.AddJadeByBagId(a, jadeInBag, uint32(costC))
	if !cost.CostBySync(a, sync, "JadeLvlUp") {
		return false, 0, Err_Jade_Cost, fmt.Sprintf("Err_Jade_Cost %d", jadeInBag)
	}
	// 宝石升级
	ok, newJadeId := gamedata.GetJadeBySlotLvl(gamedata.GetJadeSlot(cfg.GetPart()), newLvl)
	if !ok {
		return false, 0, Err_Jade_Cfg_Not_Found, fmt.Sprintf("Err_Avatar_Jade_Cfg_Not_Found lvl %d", newLvl)
	}

	// 换上新宝石
	_, newJadeCfg := gamedata.IsJade(newJadeId)
	errCode, item_inner_type, idx2OldCount := a.Profile.GetJadeBag().AddJadeByTableId(newJadeId, int64(count), newExp, newJadeCfg)
	if errCode != helper.RES_AddToBag_Success {
		return false, 0, Err_Jade_AddBag,
			fmt.Sprintf("Err_Avatar_Jade_AddBag avatar newjade %s", newJadeId)
	}
	for newJadeBagId, oldCount := range idx2OldCount {
		// 穿上新宝石
		sync.OnChangeUpdateItems(item_inner_type, newJadeBagId, oldCount, "JadeLvlUp")
		resJadeId = newJadeBagId
		break
	}

	return true, resJadeId, 0, ""
}
