package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Equip struct {
	Id                 uint32 `codec:"id"`
	UpLv               uint32 `codec:"lv_upgrade"`
	EvoLv              uint32 `codec:"lv_evolution"`
	StarLv             uint32 `codec:"lv_star"`
	StarXp             uint32 `codec:"star_xp"`
	Lv_MaterialEnhance uint32 `codec:"lv_mat_enh"`
}

type RequestChangeEquip struct {
	Req
	Equips []uint32 `codec:"avatar_equip"`
}

type ResponseChangeEquip struct {
	SyncResp
}

func (p *Account) ChangeEquip(r servers.Request) *servers.Response {
	req := &RequestChangeEquip{}
	resp := &ResponseChangeEquip{}

	initReqRsp(
		"PlayerAttr/ChangeEquipResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Avatar_Param
	)

	res, warnCode, isChgGs := p.changeEquipImp(req.Equips)
	if isChgGs {
		p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 3. 装备新装备
	}
	if warnCode > 0 {
		logs.Warn("ChangeEquip err %d", res)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	if res != 0 {
		return rpcErrorWithMsg(resp, uint32(res), "equip change err")
	}

	// 50.穿戴N件N品质的装备, P1装备数, P2品质
	p.updateCondition(account.COND_TYP_Equip_Wear,
		0, 0, "", "", resp)

	resp.OnChangeEquip()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// need_update
// 表示是否需要全量更新道具和软通信息到客户端，装备穿脱会引发产出
func (p *Account) changeEquipImp(equips []uint32) (res, warnCode int, isChgGs bool) {
	const (
		_ = iota
		CODE_Avatar_Param_Err
		CODE_Item_to_Equip_No_Exist //失败:装备不存在
		CODE_Equip_State_Err        //失败:装备状态错误：同一件装备不能装备两次
		CODE_Equip_Slot_Err         //失败:装备位置与类型不一致
		CODE_Equip_Info_Err         //失败:装备类型信息无效
	)

	for i := 0; i < len(equips) && i < helper.EQUIP_SLOT_MAX; i++ {
		curr_equip := p.Profile.GetEquips().GetEquip(i)
		if equips[i] == curr_equip {
			continue
		}
		slot := i

		pEquips := p.Profile.GetEquips()
		if equips[i] != 0 {
			// 如果在穿脱过程中有产出
			// 这里应该通知后面将软通、背包信息全量更新给客户端
			id := equips[i]
			// 先检查装备是否可以装备到这个位置上
			is_can_res, warnCode := p.IsCanEquip(slot, id)
			if is_can_res == 0 {
				// 同一件装备不能重复装备
				if pEquips.IsHasEquip(id) {
					logs.Warn("changeEquipImp CODE_Equip_State_Err")
					return CODE_Equip_State_Err, errCode.ClickTooQuickly, isChgGs
				}

				old_equip := pEquips.GetEquip(slot)
				if old_equip != 0 {
					pEquips.UnEquipImp(slot)
				}
				pEquips.EquipImp(slot, id)
				isChgGs = true
			} else {
				return is_can_res + 20, warnCode, isChgGs
			}
		}

		if equips[i] == 0 && curr_equip != 0 {
			pEquips.UnEquipImp(slot)
		}
	}

	return
}

func (p *Account) IsCanEquip(slot int, id uint32) (int, int) {
	// 各种穿装备时的限制
	const (
		_                           = iota
		CODE_Item_to_Equip_No_Exist //失败:装备不存在
		CODE_Equip_State_Err        //失败:装备状态错误：同一件装备不能装备两次
		CODE_Equip_Slot_Err         //失败:装备位置与类型不一致
		CODE_Equip_Info_Err         //失败:装备类型信息无效
		CODE_No_UnEquip_Null        //警告:装备不能被脱下
		CODE_Lvl_Not_Enough         //失败：等级不足
	)

	is_has := p.BagProfile.IsHasBagId(id)
	if !is_has {
		return CODE_Item_to_Equip_No_Exist, errCode.ClickTooQuickly
	}
	item_data, data_ok := p.BagProfile.GetItemData(id)
	if !data_ok {
		return CODE_Equip_Info_Err, 0
	}

	part := item_data.GetPart()
	slot_should := gamedata.GetEquipSlot(part)
	a_slot := gamedata.GetAvatarEquipSlot(part)
	if slot_should < 0 || a_slot > 0 || slot_should != slot {
		return CODE_Equip_Slot_Err, 0
	}

	lvl, _ := p.Profile.GetCorp().GetXpInfo()
	if lvl < uint32(item_data.GetEnableLevel()) {
		logs.Warn("IsCanEquip CODE_Lvl_Not_Enough %s", item_data.GetID())
		return CODE_Lvl_Not_Enough, errCode.ClickTooQuickly
	}
	return 0, 0
}
