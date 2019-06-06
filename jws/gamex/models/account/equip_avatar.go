package account

import "vcs.taiyouxi.net/platform/planx/util/logs"

type AvatarEquips struct {
	Curr_equips []uint32 `json:"curr"`
}

func (ae *AvatarEquips) resetCurrEquip(slot int) {
	new_len := ((slot / AVATAR_EQUIP_SLOT_MAX) + 3) * AVATAR_EQUIP_SLOT_MAX

	now_len := len(ae.Curr_equips)
	if new_len > now_len {
		new_ce := make([]uint32, new_len, new_len)
		copy(new_ce, ae.Curr_equips)
		ae.Curr_equips = new_ce
	}
}

func (ae *AvatarEquips) Curr() ([]uint32, int) {
	ae.resetCurrEquip(AVATAR_NUM_CURR * AVATAR_EQUIP_SLOT_MAX)
	l := AVATAR_NUM_CURR * AVATAR_EQUIP_SLOT_MAX
	if len(ae.Curr_equips) < l {
		l = len(ae.Curr_equips)
	}
	return ae.Curr_equips[:l], AVATAR_EQUIP_SLOT_MAX
}

func (ae *AvatarEquips) CurrByAvatar(avatar_id int) []uint32 {
	ae.resetCurrEquip(AVATAR_NUM_CURR * AVATAR_EQUIP_SLOT_MAX)
	str := avatar_id * AVATAR_EQUIP_SLOT_MAX
	end := str + AVATAR_EQUIP_SLOT_MAX
	return ae.Curr_equips[str:end]

}

func (ae *AvatarEquips) GetEquip(avatar_id, slot_in_avatar int) uint32 {
	slot := avatar_id*AVATAR_EQUIP_SLOT_MAX + slot_in_avatar
	if slot >= ALL_AVATAR_EQUIP_SLOT_MAX {
		logs.Error("[GetEquip] Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(ae.Curr_equips) {
		return 0
	}
	return ae.Curr_equips[slot]
}

func (ae *AvatarEquips) EquipImp(avatar_id, slot_in_avatar int, id uint32) {
	slot := avatar_id*AVATAR_EQUIP_SLOT_MAX + slot_in_avatar
	if slot >= ALL_AVATAR_EQUIP_SLOT_MAX {
		logs.Error("[EquipImp] Slot Too Large %d", slot)
	}

	if slot >= len(ae.Curr_equips) {
		ae.resetCurrEquip(slot)
	}
	ae.Curr_equips[slot] = id
}

func (ae *AvatarEquips) UnEquipImp(avatar_id, slot_in_avatar int) {
	slot := avatar_id*AVATAR_EQUIP_SLOT_MAX + slot_in_avatar
	if slot >= len(ae.Curr_equips) {
		return
	}
	ae.Curr_equips[slot] = 0
}

func (ae *AvatarEquips) IsHasEquip(bagid uint32) bool {
	// TODO by FanYang in [AvatarEquip) IsHasEquip 优化性能]
	for _, e := range ae.Curr_equips {
		if e == bagid {
			return true
		}
	}
	return false
}

// Debug用 卸载所有装备
func (ae *AvatarEquips) Reset(id string) {
	logs.Error("[%s]Reset All avatar Equip!", id)
	for idx, _ := range ae.Curr_equips {
		ae.Curr_equips[idx] = 0
	}
}
