package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type AvatarEquip struct {
	Id uint32 `codec:"id"`
}

type RequestChangeAvatarEquip struct {
	Req
	AvatarID int      `codec:"avatar_id"`
	Equips   []uint32 `codec:"avatar_equip"`
}

type ResponseChangeAvatarEquip struct {
	SyncResp
}

func (p *Account) ChangeAvatarEquip(r servers.Request) *servers.Response {
	req := &RequestChangeAvatarEquip{}
	resp := &ResponseChangeAvatarEquip{}

	initReqRsp(
		"PlayerAttr/ChangeAvatarEquipResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Avatar_Param
		Err_Avatar_No_Unlock
	)

	p.FashionRefresh(resp)

	// 角色是否解锁
	if !p.Account.IsAvatarUnblock(req.AvatarID) {
		return rpcErrorWithMsg(resp, uint32(Err_Avatar_No_Unlock),
			fmt.Sprintf("Err_Avatar_No_Unlock AvatarID %d", req.AvatarID))
	}

	res := p.changeAvatarEquipImp(req.AvatarID, req.Equips)
	if res != 0 {
		return rpcErrorWithMsg(resp, uint32(res), "equip change err")
	}

	resp.OnChangeAvatarEquip()

	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) changeAvatarEquipImp(avatar_id int, equips []uint32) (res int) {
	const (
		_ = iota
		CODE_Avatar_Param_Err
		CODE_Item_to_Equip_No_Exist //失败:装备不存在
		CODE_Equip_State_Err        //失败:装备状态错误：同一件装备不能装备两次
		CODE_Equip_Slot_Err         //失败:装备位置与类型不一致
		CODE_Equip_Info_Err         //失败:装备类型信息无效
	)

	if avatar_id >= helper.AVATAR_NUM_MAX || avatar_id < 0 {
		return CODE_Avatar_Param_Err
	}

	for i := 0; i < len(equips) && i < helper.AVATAR_EQUIP_SLOT_MAX; i++ {
		curr_equip := p.Profile.GetAvatarEquips().GetEquip(avatar_id, i)
		if equips[i] == curr_equip {
			continue
		}
		slot := i

		aEquips := p.Profile.GetAvatarEquips()
		if equips[i] != 0 {
			// 如果在穿脱过程中有产出
			// 这里应该通知后面将软通、背包信息全量更新给客户端
			id := equips[i]
			// 先检查装备是否可以装备到这个位置上
			is_can_res := p.IsCanAvatarEquip(avatar_id, slot, id)
			if is_can_res == 0 {
				// 同一件装备不能重复装备
				if aEquips.IsHasEquip(id) {
					return CODE_Equip_State_Err
				}

				old_equip := aEquips.GetEquip(avatar_id, slot)
				if old_equip != 0 {
					aEquips.UnEquipImp(avatar_id, slot)
				}
				aEquips.EquipImp(avatar_id, slot, id)
			} else {
				return is_can_res + 20
			}
		}

		if equips[i] == 0 && curr_equip != 0 {
			aEquips.UnEquipImp(avatar_id, slot)
		}
	}

	return
}

func (p *Account) IsCanAvatarEquip(avatar_id, slot int, id uint32) int {
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

	is_has := false
	var item_data *ProtobufGen.Item
	if gamedata.IsSlotFashion(slot) {
		has, info := p.Profile.GetFashionBag().GetFashionInfo(id)
		is_has = has
		if is_has {
			_, item_data = gamedata.IsFashion(info.TableID)
		}
	}
	if !is_has || item_data == nil {
		return CODE_Item_to_Equip_No_Exist
	}

	part := item_data.GetPart()
	slot_should := gamedata.GetAvatarEquipSlot(part)
	if slot_should < 0 || slot_should != slot {
		return CODE_Equip_Slot_Err
	}

	logs.Trace("item_data.GetRoleOnly() %d", item_data.GetRoleOnly())
	if item_data.GetRoleOnly() > 0 && int32(avatar_id) != item_data.GetRoleOnly() {
		return CODE_Equip_Slot_Err
	}

	lvl, _ := p.Profile.GetCorp().GetXpInfo()
	if lvl < uint32(item_data.GetEnableLevel()) {
		return CODE_Lvl_Not_Enough
	}
	return 0
}
