package account

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type EquipmentJades struct {
	Jades []uint32 `json:"equip_jade"`
}

func (aj *EquipmentJades) CurrAll2Client() ([]int, []Jade2Client) {
	length := EQUIP_SLOT_CURR * gamedata.JadePartCount
	ids := make([]int, 0, EQUIP_SLOT_CURR)
	res := make([]Jade2Client, 0, length)
	for i := 0; i < EQUIP_SLOT_CURR; i++ {
		if i*gamedata.JadePartCount < len(aj.Jades) {
			ids = append(ids, i)
		}
		for j := 0; j < gamedata.JadePartCount; j++ {
			idx := i*JADE_SLOT_MAX + j
			if idx >= len(aj.Jades) {
				return ids, res
			}
			res = append(res, Jade2Client{aj.Jades[idx]})
		}
	}
	return ids, res
}

func (aj *EquipmentJades) IsObjUnlock(a *Account, equipSlot int) (ok bool, errcode int, errmsg string) {
	// 是否装备槽位有装备
	if a.Profile.GetEquips().GetEquip(equipSlot) <= 0 {
		return false, Err_Obj_Not_Lock,
			fmt.Sprintf("EquipmentJades IsObjUnlock equip slot %d not equipment", equipSlot)
	}
	return true, 0, ""
}

func (aj *EquipmentJades) IsObjSlotUnlock(a *Account, equipSlot, slot_in_equip int) (ok bool, errcode int, errmsg string) {
	ok, lvl := gamedata.GetEquipJadeUnlockLvl(uint32(equipSlot), slot_in_equip)
	if !ok {
		return false, Err_Jade_Slot_Cfg_Not_Found,
			fmt.Sprintf("Err_Avatar_Jade_Slot_Cfg_Not_Found equip %d slot %d", equipSlot, slot_in_equip)
	}
	l, _ := a.Profile.GetCorp().GetXpInfo()
	if l < lvl {
		return false, Err_Corp_Lvl_Not_Enough,
			fmt.Sprintf("Err_Avatar_Corp_Lvl_Not_Enough equip %d slot %d", equipSlot, slot_in_equip)
	}
	return true, 0, ""
}

// 并不作物品id是否在包裹检查
func (aj *EquipmentJades) EquipJade(a *Account, equipSlot int,
	slot_in_equip int, jadeItemId uint32, jadeItemCfgId string, sync helper.ISyncRsp) (
	ok bool, errcode int, errmsg string, warnCode int) {
	// 角色是否解锁
	if ok, errcode, errmsg := aj.IsObjUnlock(a, equipSlot); !ok {
		return false, errcode, errmsg, 0
	}
	// 槽位是否开
	ok, errcode, errmsg = aj.IsObjSlotUnlock(a, equipSlot, slot_in_equip)
	if !ok {
		return
	}
	_slot := equipSlot*JADE_SLOT_MAX + slot_in_equip
	// 是否此位置已经是这个
	if aj.getByIdx(_slot) == jadeItemId {
		return true, 0, "", 0
	}
	// 是否已经装备了
	jadeBag := a.Profile.GetJadeBag()
	if jadeBag.GetJadeInBagCount(jadeItemId) <= 0 {
		return false, Err_Jade_Already_Equip,
			fmt.Sprintf("Err_Avatar_Jade_Already_Equip  equip %d slot %d item %s", equipSlot, slot_in_equip, jadeItemCfgId),
			errCode.ClickTooQuickly
	}
	// 先脱下
	aj.UnEquipJade(a, equipSlot, slot_in_equip, sync)
	// 穿上新的
	aj.setByIdx(_slot, jadeItemId)
	jadeBag.TakeOutFromBag(jadeItemId)
	sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, jadeItemId,
		jadeBag.GetJade(jadeItemId).Count, "ChgJade")
	return true, 0, "", 0
}

func (aj *EquipmentJades) UnEquipJade(a *Account, equipSlot int, slot_in_equip int,
	sync helper.ISyncRsp) {
	_slot := equipSlot*JADE_SLOT_MAX + slot_in_equip
	if _slot >= len(aj.Jades) {
		return
	}
	oldJade := aj.getByIdx(_slot)
	if oldJade > 0 {
		aj.setByIdx(_slot, 0)
		jadeBag := a.Profile.GetJadeBag()
		jadeBag.PutInToBag(oldJade)
		sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, oldJade,
			jadeBag.GetJade(oldJade).Count, "ChgJade")
	}
}

// 需要外面检查对象已经解锁了
func (aj *EquipmentJades) AutoEquip(a *Account, equipSlot int, sync helper.ISyncRsp) (isUpdate bool) {
	start := equipSlot * JADE_SLOT_MAX
	end := start + JADE_SLOT_MAX
	return autoEquip(a, equipSlot, aj, start, end, sync)
}

// 需要外面检查物品类型是否是宝石，以及槽位是否相同，物品是否存在
func (aj *EquipmentJades) JadeLvlUp(a *Account, equipSlot int, slot_in_equip int,
	jadeItemId uint32, costJadeIds []uint32, JadeCostCount []int, sync helper.ISyncRsp) (
	ok bool, resJadeId uint32, errcode int, errmsg string, warnCode int) {
	start := equipSlot * JADE_SLOT_MAX
	end := start + JADE_SLOT_MAX
	ok, resJadeId, errcode, errmsg, warnCode = jadeLvlUp(a, aj, equipSlot, slot_in_equip, start, end, jadeItemId, costJadeIds, JadeCostCount, sync)
	return
}

func (aj *EquipmentJades) SyncUpdate(sync helper.ISyncRsp) {
	sync.OnChangeAvatarJade()
}

func (aj *EquipmentJades) getByIdx(idx int) uint32 {
	if idx >= len(aj.Jades) {
		return 0
	}
	return aj.Jades[idx]
}

func (aj *EquipmentJades) setByIdx(idx int, jadeId uint32) {
	aj.resetJadeCap(idx)
	aj.Jades[idx] = jadeId
}

func (aj *EquipmentJades) resetJadeCap(slot int) {
	new_len := ((slot / JADE_SLOT_MAX) + 3) * JADE_SLOT_MAX

	now_len := len(aj.Jades)
	if new_len > now_len {
		new_ce := make([]uint32, new_len, new_len)
		copy(new_ce, aj.Jades)
		aj.Jades = new_ce
	}
}

func (aj *EquipmentJades) GetSlotJadeForLog(slot int) []uint32 {
	aj.resetJadeCap(slot)
	res := make([]uint32, 0, 5)
	for j := 0; j < gamedata.JadePartCount; j++ {
		idx := slot*JADE_SLOT_MAX + j
		if idx >= len(aj.Jades) {
			return res
		}
		res = append(res, aj.Jades[idx])
	}
	return res
}
