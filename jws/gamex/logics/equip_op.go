package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	equipOp_Upgrade           int = 3
	equipOp_Evolution         int = 4
	equipOp_StarLevelUp       int = 5
	equipOp_StarBlessLevelUp  int = 6
	equipOp_StarLevelUpWithHc int = 7
)

const (
	// 每强化5次可以精炼一次，意即：不可超过武器的“强化等级/5取整”
	equip_Evolution_Pre_Upgrade_Lv = 5
)

type RequestEquipOp struct {
	Req
	Op   int `codec:"op"`
	Slot int `codec:"slot"`
	P1   int `codec:"p1"`
	//P2        int `codec:"p2"`
	//P3        int `codec:"p3"`
	//P4 int `codec:"p4"`
	Ext []string `codec:"ext"`
}

type ResponseEquipOp struct {
	SyncResp
	R1 int   `codec:"r1"`
	R2 int   `codec:"r2"`
	R3 []int `codec:"r3"`
}

func (p *Account) EquipOp(r servers.Request) *servers.Response {
	req := &RequestEquipOp{}
	resp := &ResponseEquipOp{}

	initReqRsp(
		"PlayerAttr/EquipOpRsp",
		r.RawBytes,
		req, resp, p)

	var code uint32 = 0
	switch req.Op {
	case equipOp_Upgrade:
		code = p.Upgrade(req.Slot, req.P1, resp)
	case equipOp_Evolution: // 强化
		code = p.Evolution(req.Slot, req.P1, resp)
	case equipOp_StarLevelUp: // 升星
		code, resp.R3 = p.EquipStarLevelUp(req.Slot, req.P1, false, resp)
	case equipOp_StarLevelUpWithHc: // HC升星
		code, resp.R3 = p.EquipStarLevelUp(req.Slot, req.P1, true, resp)
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 4. 装备 强化 精炼

	if code != 0 {
		if code > 200 {
			logs.SentryLogicCritical(p.AccountID.String(), "EquipOpFail op %d code %d",
				req.Op, code)
		}
		resp.MsgOK = "failed"
		resp.Code = code
	}

	resp.OnChangeEquip()
	resp.OnChangeAvatarEquip()
	logs.Trace("resp %v", resp)
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) Upgrade(slot, lv_add int, sync helper.ISyncRsp) uint32 {
	logs.Trace("[%s]EquipUpgrade:%d,%d",
		p.AccountID, slot, lv_add)

	const (
		_                    = iota
		CODE_Limit_Avatar_Lv //失败:达到上限，不能超过角色等级
		CODE_Limit_Equip_Max //失败:达到上限，不能超过装备最大等级
		CODE_No_Enough_Cost  //失败:没有足够的物品
		CODE_No_Equip        //失败:升级精炼位置错误，没有装备
		CODE_No_Equip_Info   //失败:装备信息缺失
	)

	equips := p.Profile.GetEquips()
	from := equips.GetUpgrade(slot)
	to := uint32(lv_add) + from

	//当前装备最大强化等级判断
	// 一个装备的最大强化等级依赖于已下
	//   装备本身最大强化等级
	//   当前装备角色的等级

	// 先判断装备本身最大强化等级 --> 目前这个最大强化等级都被配置成100了,暂时不依赖这个功能
	equip_id := equips.GetEquip(slot)
	equip_data, data_ok := p.BagProfile.GetItemData(equip_id)
	if !data_ok {
		logs.Error("equip data nil %d", equip_id)
		return mkCode(CODE_ERR, CODE_No_Equip_Info)
	}
	max_lv := equip_data.GetFuseLevelLimit()
	if to > uint32(max_lv) {
		logs.Error("equip Upgrade lv cannot be max level %d!", max_lv)
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	// 再判断武将等级
	avatar_lv, _ := p.Profile.GetCorp().GetXpInfo()
	if to > uint32(avatar_lv) {
		logs.Error("equip Upgrade lv cannot max then avatar_lv %d!", avatar_lv)
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	cost_data := gamedata.GetEquipUpgradeNeed(from, to)
	logs.Trace("cost_data %v", cost_data)

	cost_group := account.CostGroup{}

	is_has := cost_group.AddCostData(p.Account, cost_data)
	if !is_has {
		logs.Trace("No Has Cost!")
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	is_cost := cost_group.CostBySync(p.Account, sync, "EquipUpgrade")
	if !is_cost {
		logs.Trace("Cost Error!")
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	equips.Upgrade(slot, uint32(lv_add))

	// 条件更新
	p.updateCondition(account.COND_TYP_Upgrade, 1, 0, "", "", sync)

	return 0
}

func (p *Account) Evolution(slot, lv_add int, sync helper.ISyncRsp) uint32 {
	logs.Trace("[%s]EquipEvolution:%d,%d",
		p.AccountID, slot, lv_add)
	const (
		_                    = iota
		CODE_Limit_Avatar_Lv //失败:达到上限，不能超过角色等级
		CODE_Limit_Equip_Max //失败:达到上限，不能超过装备最大等级
		CODE_No_Enough_Cost  //失败:没有足够的物品
		CODE_No_Equip        //失败:升级精炼位置错误，没有装备
		CODE_No_Equip_Info   //失败:装备信息缺失
	)

	equips := p.Profile.GetEquips()
	from := equips.GetEvolution(slot)
	to := uint32(lv_add) + from

	// 再判断武将等级
	avatar_lv, _ := p.Profile.GetCorp().GetXpInfo()
	if to > uint32(avatar_lv) {
		logs.Warn("equip Evolution lv cannot max then avatar_lv %d!", avatar_lv)
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	//当前装备最大精炼等级判断
	// 每强化5次可以精炼一次，意即：不可超过武器的“强化等级/5取整”

	//max_lv := aequip.GetUpgrade(avatar_id, slot) / equip_Evolution_Pre_Upgrade_Lv
	//if to > uint32(max_lv) {
	//	logs.Error("equip evolution lv cannot be max level %d!", max_lv)
	//	return mkCode(CODE_ERR, CODE_Limit_Equip_Max)
	//}

	cost_data := gamedata.GetEquipEvolutionNeed(slot, from, to)
	if cost_data == nil {
		return mkCode(CODE_ERR, CODE_No_Equip_Info)
	}
	logs.Trace("cost_data %v", cost_data)

	cost_group := account.CostGroup{}

	is_has := cost_group.AddCostData(p.Account, cost_data)
	if !is_has {
		logs.Trace("No Has Cost!")
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	is_cost := cost_group.CostBySync(p.Account, sync, "EquipEvolution")
	if !is_cost {
		logs.Trace("Cost Error!")
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	equips.Evolution(slot, uint32(lv_add))

	// 条件更新
	p.updateCondition(account.COND_TYP_Evolution, 1, 0, "", "", sync)

	return 0
}
