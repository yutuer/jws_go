package account

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/uutil/count"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 战队装备，和avatar无关

type Equips struct {
	CurrEquips        []uint32             `json:"curr"`
	LvUpgrade         []uint32             `json:"up"`        // 装备强化等级
	LvEvolution       []uint32             `json:"ev"`        // 精炼等级
	LvStar            []uint32             `json:"star"`      // 装备星级
	LvStarXp          []uint32             `json:"starXp"`    // 星级经验
	LvMaterialEnhance []uint32             `json:"lme"`       // 材料强化等级
	MaterialEnhance   [][]bool             `json:"me"`        // 材料强化材料装备情况
	StarHcUpCount     count.CountData      `json:"starCount"` // 升星**已使用**次数, 每日清零
	StarSysNotice     []gamedata.Condition `json:"star_sys"`  // 记录升星跑过的跑马灯,防止重复跑,因为升星升一星需要点很多次,只有第一次到某星级需要跑马灯
}

func (p *Equips) OnAccountInit() {
	p.StarHcUpCount = count.NewDailyClear(
		gamedata.DailyStartTypCommon)
}

func (p *Equips) Init() {
}

// 装备穿脱相关接口 只负责数据的更新，如果没有数据就新增，相关逻辑在logics/equip.go
func (b *Equips) resetCurrEquip(slot int) {
	new_len := slot + 5

	// AvatarEquip.resetCurrEquip 效率优化
	// 因为角色数量很固定 所以不需要优化
	// 这个逻辑很难调用到
	now_len := len(b.CurrEquips)
	if new_len > now_len {
		new_ce := make([]uint32, new_len, new_len)
		copy(new_ce, b.CurrEquips)
		b.CurrEquips = new_ce
	}

	now_len = len(b.LvUpgrade)
	if new_len > now_len {
		new_ulv := make([]uint32, new_len, new_len)
		copy(new_ulv, b.LvUpgrade)
		b.LvUpgrade = new_ulv
	}

	now_len = len(b.LvEvolution)
	if new_len > now_len {
		new_elv := make([]uint32, new_len, new_len)
		copy(new_elv, b.LvEvolution)
		b.LvEvolution = new_elv
	}

	now_len = len(b.LvStar)
	if new_len > now_len {
		new_elv := make([]uint32, new_len, new_len)
		copy(new_elv, b.LvStar)
		b.LvStar = new_elv

		new_starsys := make([]gamedata.Condition, new_len, new_len)
		copy(new_starsys, b.StarSysNotice)
		b.StarSysNotice = new_starsys
	}

	now_len = len(b.LvStarXp)
	if new_len > now_len {
		new_elv := make([]uint32, new_len, new_len)
		copy(new_elv, b.LvStarXp)
		b.LvStarXp = new_elv
	}

	now_len = len(b.LvMaterialEnhance)
	if new_len > now_len {
		new_elv := make([]uint32, new_len, new_len)
		copy(new_elv, b.LvMaterialEnhance)
		b.LvMaterialEnhance = new_elv
	}

	now_len = len(b.MaterialEnhance)
	if new_len > now_len {
		new_elv := make([][]bool, new_len, new_len)
		for i := 0; i < new_len; i++ {
			if i >= len(b.MaterialEnhance) || len(b.MaterialEnhance[i]) <= 0 {
				new_elv[i] = make([]bool, EQUIP_MAT_ENHANCE_MAT)
			} else {
				new_elv[i] = b.MaterialEnhance[i]
			}
		}
		b.MaterialEnhance = new_elv
	}

	b.Init()
}

// Debug用 卸载所有装备
func (b *Equips) Reset(id string) {
	logs.Error("[%s]Reset All Equip!", id)
	for idx, _ := range b.CurrEquips {
		b.CurrEquips[idx] = 0
		b.LvEvolution[idx] = 0
		b.LvUpgrade[idx] = 0
		b.LvStar[idx] = 0
		b.LvStarXp[idx] = 0
		b.LvMaterialEnhance[idx] = 0
		b.MaterialEnhance[idx] = make([]bool, EQUIP_MAT_ENHANCE_MAT)
	}
}

func (b *Equips) Curr() ([]uint32, []uint32, []uint32, []uint32, []uint32, []uint32, [][]bool, int) {
	b.resetCurrEquip(EQUIP_SLOT_MAX)
	return b.CurrEquips[:],
		b.LvUpgrade[:],
		b.LvEvolution[:],
		b.LvStar[:],
		b.LvStarXp[:],
		b.LvMaterialEnhance[:],
		b.MaterialEnhance[:],
		EQUIP_SLOT_MAX
}

func (b *Equips) CurrInfo() ([]uint32, []uint32, []uint32, []uint32, []uint32, [][]bool) {
	b.resetCurrEquip(EQUIP_SLOT_MAX)
	str := 0
	end := str + EQUIP_SLOT_MAX
	return b.CurrEquips[str:end],
		b.LvUpgrade[str:end],
		b.LvEvolution[str:end],
		b.LvStar[str:end],
		b.LvMaterialEnhance[str:end],
		b.MaterialEnhance[str:end]
}

func (b *Equips) GetEquip(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("[GetEquip] Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.CurrEquips) {
		return 0
	}
	return b.CurrEquips[slot]
}

func (b *Equips) EquipImp(slot int, id uint32) {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("[EquipImp] Slot Too Large %d", slot)
	}

	if slot >= len(b.CurrEquips) {
		b.resetCurrEquip(slot)
	}
	b.CurrEquips[slot] = id
}

func (b *Equips) UnEquipImp(slot int) {
	if slot >= len(b.CurrEquips) {
		return
	}
	b.CurrEquips[slot] = 0
}

func (b *Equips) Upgrade(slot int, lv uint32) {
	if slot >= len(b.LvUpgrade) {
		logs.Error("Upgrade Unequip Slot")
		return
	}
	b.LvUpgrade[slot] += lv
}

func (b *Equips) GetUpgrade(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.LvUpgrade) {
		return 0 // 默认0级
	}
	return b.LvUpgrade[slot]
}

func (b *Equips) Evolution(slot int, lv uint32) {
	if slot >= len(b.LvEvolution) {
		logs.Error("Evolution Unequip Slot")
		return
	}
	b.LvEvolution[slot] += lv
}

func (b *Equips) GetEvolution(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.LvEvolution) {
		return 0 // 默认0级
	}
	return b.LvEvolution[slot]
}

func (b *Equips) IsHasEquip(bagid uint32) bool {
	// TODO by FanYang in [AvatarEquip) IsHasEquip 优化性能]
	for _, e := range b.CurrEquips {
		if e == bagid {
			return true
		}
	}
	return false
}

// 装备穿脱结束

// 装备升星

func (b *Equips) DebugSetAllStarLv(star uint32) {
	for i := 0; i < len(b.LvStar); i++ {
		b.LvStar[i] = star
	}
}

func (b *Equips) SetStarLv(slot int, star uint32) {
	if slot >= len(b.LvStar) {
		logs.Error("SetStarLv Unequip Slot")
		return
	}
	b.LvStar[slot] = star
}

func (b *Equips) StarLvUp(slot int) {
	if slot >= len(b.LvStar) {
		logs.Error("StarLvUp Unequip Slot")
		return
	}
	b.LvStar[slot] += 1
}

func (b *Equips) StarLvDownTo(slot int, star uint32) {
	if slot >= len(b.LvStar) {
		logs.Error("StarLvDown Unequip Slot")
		return
	}
	logs.Trace("StarLvDown %d", b.LvStar[slot])
	b.LvStar[slot] = star
	logs.Trace("StarLvDown After %d", b.LvStar[slot])
}

func (b *Equips) GetStarLv(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.LvStar) {
		return 0 // 默认0级
	}
	return b.LvStar[slot]
}

func (b *Equips) SetStarXP(slot int, value uint32) {
	if slot >= len(b.LvStarXp) {
		logs.Error("SetStarBlessLv Unequip Slot")
		return
	}
	b.LvStarXp[slot] = value
}

func (b *Equips) AddStarXP(slot int, value uint32) {
	if slot >= len(b.LvStarXp) {
		logs.Error("AddStarBlessLv Unequip Slot")
		return
	}
	b.LvStarXp[slot] += value
}

func (b *Equips) GetStarXP(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("GetStarBlessLv Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.LvStarXp) {
		return 0 // 默认0级
	}
	return b.LvStarXp[slot]
}

func (b *Equips) GetMatEnhLv(slot int) uint32 {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("GetMatEnhLv Slot Too Large %d", slot)
		return 0
	}
	if slot >= len(b.LvMaterialEnhance) {
		return 0 // 默认0级
	}
	return b.LvMaterialEnhance[slot]
}

func (b *Equips) LvlUpMatEnh(slot int) {
	if slot >= len(b.LvMaterialEnhance) {
		logs.Error("LvlUpMatEnh no Slot %d", slot)
		return
	}
	b.LvMaterialEnhance[slot] += 1
	b.MaterialEnhance[slot] = make([]bool, EQUIP_MAT_ENHANCE_MAT)
}

func (b *Equips) GetMatEnhSlotInfo(slot int) []bool {
	if slot >= EQUIP_SLOT_MAX {
		logs.Error("GetMatEnhSlotInfo Slot Too Large %d", slot)
		return make([]bool, EQUIP_MAT_ENHANCE_MAT)
	}
	if slot >= len(b.MaterialEnhance) {
		return make([]bool, EQUIP_MAT_ENHANCE_MAT)
	}
	return b.MaterialEnhance[slot]
}

func (b *Equips) SetMatEnhSlotInfo(slot int, matIndex int) {
	if slot >= len(b.MaterialEnhance) {
		logs.Error("SetMatEnhSlotInfo no Slot %d", slot)
		return
	}
	mats := b.MaterialEnhance[slot]
	if matIndex >= len(mats) {
		nmats := make([]bool, helper.EQUIP_MAT_ENHANCE_MAT)
		copy(nmats, mats)
		b.MaterialEnhance[slot] = nmats
		mats = nmats
	}
	mats[matIndex] = true
	return
}

func (b *Equips) CheckEquipInBag(acid string, p Profile, bag PlayerBag) {
	blog := false
	for i, eq := range b.CurrEquips {
		if eq > 0 {
			if bag.GetItem(eq) == nil {
				logs.Warn("%s CheckEquipInBag when login, equip %d not found->reset",
					acid, eq)
				b.CurrEquips[i] = 0
				blog = true
			}
		}
	}
	if blog {
		logiclog.LogiclogDebug(acid, p.GetCurrAvatar(),
			p.GetCorp().GetLvlInfo(), p.ChannelId)
	}
}
