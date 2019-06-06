package account

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type DestGeneralJades struct {
	DestinyGeneralIds  []int    `json:"dgeneral_ids"`
	DestinyGeneralJade []uint32 `json:"dgeneral_jad"`
}

func (dgj *DestGeneralJades) IsObjUnlock(a *Account, objId int) (ok bool, errcode int, errmsg string) {
	if nil == a.Profile.GetDestinyGeneral().GetGeneral(objId) {
		return false, Err_Obj_Not_Lock, fmt.Sprintf("Err_Avatar_Not_Lock destGeneral %d ", objId)
	}
	return true, 0, ""
}

func (dgj *DestGeneralJades) IsObjSlotUnlock(a *Account, objId, slot_in_obj int) (ok bool, errcode int, errmsg string) {
	if !a.Profile.GetDestinyGeneral().IsJadeUnlock(objId, slot_in_obj) {
		return false, Err_Obj_Jade_Slot_Not_Unlock, fmt.Sprintf("Err_Obj_Jade_Slot_Not_Unlock destGeneral %d ", objId)
	}
	return true, 0, ""
}

func (dgj *DestGeneralJades) EquipJade(a *Account, objId int, slot_in_avatar int,
	jadeItemId uint32, jadeItemCfgId string, sync helper.ISyncRsp) (ok bool, errcode int, errmsg string, warnCode int) {
	// 角色是否解锁
	if ok, errcode, errmsg := dgj.IsObjUnlock(a, objId); !ok {
		return false, errcode, errmsg, 0
	}
	// 槽位是否开
	ok, errcode, errmsg = dgj.IsObjSlotUnlock(a, objId, slot_in_avatar)
	if !ok {
		return
	}
	objIdx := dgj.getObjIndex(objId)
	_slot := objIdx*JADE_SLOT_MAX + slot_in_avatar
	// 是否此位置已经是这个
	if dgj.getByIdx(_slot) == jadeItemId {
		return true, 0, "", 0
	}
	// 是否已经装备了
	jadeBag := a.Profile.GetJadeBag()
	if jadeBag.GetJadeInBagCount(jadeItemId) <= 0 {
		return false, Err_Jade_Already_Equip,
			fmt.Sprintf("Err_Avatar_Jade_Already_Equip  avatar %d slot %d item %s", objId, slot_in_avatar, jadeItemCfgId),
			errCode.ClickTooQuickly
	}
	// 先脱下
	dgj.UnEquipJade(a, objId, slot_in_avatar, sync)
	// 穿上新的
	dgj.setByIdx(_slot, jadeItemId)
	jadeBag.TakeOutFromBag(jadeItemId)
	sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, jadeItemId,
		jadeBag.GetJade(jadeItemId).Count, "ChgJade")
	return true, 0, "", 0

}
func (dgj *DestGeneralJades) UnEquipJade(a *Account, objId int, slot_in_avatar int,
	sync helper.ISyncRsp) {
	objIdx := dgj.getObjIndex(objId)
	_slot := objIdx*JADE_SLOT_MAX + slot_in_avatar
	if _slot >= len(dgj.DestinyGeneralJade) {
		return
	}
	oldJade := dgj.getByIdx(_slot)
	if oldJade > 0 {
		dgj.setByIdx(_slot, 0)
		jadeBag := a.Profile.GetJadeBag()
		jadeBag.PutInToBag(oldJade)
		sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, oldJade,
			jadeBag.GetJade(oldJade).Count, "ChgJade")
	}
}
func (dgj *DestGeneralJades) CurrAll2Client() ([]int, []Jade2Client) {
	objCount := len(dgj.DestinyGeneralIds)
	length := objCount * gamedata.JadePartCount
	ids := make([]int, 0, objCount)
	res := make([]Jade2Client, 0, length)
	for i := 0; i < objCount; i++ {
		ids = append(ids, dgj.DestinyGeneralIds[i])
		for j := 0; j < gamedata.JadePartCount; j++ {
			idx := i*JADE_SLOT_MAX + j
			if idx >= len(dgj.DestinyGeneralJade) {
				return ids, res
			}
			res = append(res, Jade2Client{dgj.DestinyGeneralJade[idx]})
		}
	}
	return ids, res
}

func (dgj *DestGeneralJades) AutoEquip(a *Account, objId int, sync helper.ISyncRsp) (isUpdate bool) {
	objIdx := dgj.getObjIndex(objId)
	start := objIdx * JADE_SLOT_MAX
	end := start + JADE_SLOT_MAX
	return autoEquip(a, objId, dgj, start, end, sync)
}

func (dgj *DestGeneralJades) JadeLvlUp(a *Account, objId int, slot_in_obj int,
	jadeItemId uint32, costJadeIds []uint32, JadeCostCount []int, sync helper.ISyncRsp) (
	ok bool, resJadeId uint32, errcode int, errmsg string, warnCode int) {
	objIdx := dgj.getObjIndex(objId)
	start := objIdx * JADE_SLOT_MAX
	end := start + JADE_SLOT_MAX
	ok, resJadeId, errcode, errmsg, warnCode = jadeLvlUp(a, dgj, objId, slot_in_obj, start, end, jadeItemId, costJadeIds, JadeCostCount, sync)
	return
}

func (dgj *DestGeneralJades) SyncUpdate(sync helper.ISyncRsp) {
	sync.OnChangeDestinyGenJade()
}

func (dgj *DestGeneralJades) resetJadeCap(slot int) {
	new_len := ((slot / JADE_SLOT_MAX) + 3) * JADE_SLOT_MAX

	now_len := len(dgj.DestinyGeneralJade)
	if new_len > now_len {
		new_ce := make([]uint32, new_len, new_len)
		copy(new_ce, dgj.DestinyGeneralJade)
		dgj.DestinyGeneralJade = new_ce
	}
}

func (dgj *DestGeneralJades) getByIdx(idx int) uint32 {
	if idx >= len(dgj.DestinyGeneralJade) {
		return 0
	}
	return dgj.DestinyGeneralJade[idx]
}
func (dgj *DestGeneralJades) setByIdx(idx int, jadeId uint32) {
	dgj.resetJadeCap(idx)
	dgj.DestinyGeneralJade[idx] = jadeId
}

func (dgj *DestGeneralJades) getObjIndex(objId int) int {
	for i, id := range dgj.DestinyGeneralIds {
		if id == objId {
			return i
		}
	}
	dgj.DestinyGeneralIds = append(dgj.DestinyGeneralIds, objId)
	return len(dgj.DestinyGeneralIds) - 1
}
