package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 手动镶嵌和卸下龙玉
func (p *Account) ChgJade(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ObjTyp int      `codec:"objtyp"` // 0 装备slot，1 神将
		ObjId  int      `codec:"objid"`
		Jades  []uint32 `codec:"jades"` // 所有龙玉，0表示没有（卸下）
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/ChgJadeResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Item_Jade_No_In_Bag
		Err_Item_Type_Not_Jade
		Err_Jade_Slot_Not_Found
		Err_Jade_Slot_Not_Match
	)

	jadeImp := account.GetJadeImp(p.Account, req.ObjTyp)
	if jadeImp == nil {
		return rpcErrorWithMsg(resp, Err_Param, fmt.Sprintf("Err_Param ObjTyp %d", req.ObjTyp))
	}
	for slot, jadeId := range req.Jades {
		if jadeId > 0 { // 穿上
			// 物品是否在
			itemCfg, ok := p.Profile.GetJadeBag().GetJadeData(jadeId)
			if !ok {
				logs.Warn("ChgJade CODE_Item_Jade_No_In_Bag item %s", itemCfg.GetID())
				return rpcWarn(resp, errCode.ClickTooQuickly)
			}
			if itemCfg.GetType() != "JADE" {
				return rpcErrorWithMsg(resp, Err_Item_Type_Not_Jade, fmt.Sprintf("CODE_Item_Type_Not_Jade item %s", itemCfg.GetID()))
			}
			// slot check
			req_slot := gamedata.GetJadeSlot(itemCfg.GetPart())
			if req_slot < 0 {
				return rpcErrorWithMsg(resp, Err_Jade_Slot_Not_Found, fmt.Sprintf("[AddJade] Err_Jade_Slot_Not_Found item %s", itemCfg.GetID()))
			}
			if slot != req_slot {
				return rpcErrorWithMsg(resp, Err_Jade_Slot_Not_Match, fmt.Sprintf("[AddJade] Err_Jade_Slot_Not_Match item %s", itemCfg.GetID()))
			}
			// 放上
			if ok, errcode, errmsg, warnCode := jadeImp.EquipJade(p.Account, req.ObjId,
				slot, jadeId, itemCfg.GetID(), resp); !ok {
				if warnCode > 0 {
					return rpcWarn(resp, uint32(warnCode))
				}
				return rpcErrorWithMsg(resp, uint32(errcode+20), errmsg)
			}
		} else { // 脱下
			jadeImp.UnEquipJade(p.Account, req.ObjId, slot, resp)
		}
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 9.宝石

	// 强制刷新任务条件支持红点
	p.updateCondition(account.COND_TYP_EquipedJade,
		0, 0, "", "", resp)
	// 更新宝石等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&info)
	// 更新第二周宝石等级排行榜
	rank.GetModule(p.AccountID.ShardId).RankByHeroJadeTwo.AddScoreCanZero(&info)
	jadeImp.SyncUpdate(resp)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 一键镶嵌龙玉
func (p *Account) AutoAddJade(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ObjTyp int `codec:"objtyp"` // 0 装备slot，1 神将
		ObjId  int `codec:"objid"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/AutoAddJadeResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
	)

	jadeImp := account.GetJadeImp(p.Account, req.ObjTyp)
	if jadeImp == nil {
		return rpcErrorWithMsg(resp, Err_Param, fmt.Sprintf("Err_Param ObjTyp %d", req.ObjTyp))
	}
	// 对象是否解锁了
	if ok, errcode, errmsg := jadeImp.IsObjUnlock(p.Account, req.ObjId); !ok {
		return rpcErrorWithMsg(resp, uint32(errcode+20), errmsg)
	}
	if jadeImp.AutoEquip(p.Account, req.ObjId, resp) {
		// 强制刷新任务条件支持红点
		p.updateCondition(account.COND_TYP_EquipedJade,
			0, 0, "", "", resp)

		jadeImp.SyncUpdate(resp)
		resp.mkInfo(p)

		p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 9.宝石
	}
	// 更新宝石等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&info)
	// 更新第二周宝石等级排行榜
	rank.GetModule(p.AccountID.ShardId).RankByHeroJadeTwo.Add(&info)
	return rpcSuccess(resp)
}

// 提升
const jade_lvlUp_max_cost_count = 3

func (p *Account) LvlUpJade(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ObjTyp        int      `codec:"objtyp"` // 0 装备slot，1 神将
		ObjId         int      `codec:"objid"`
		Jade          uint32   `codec:"jade"`      // 身上的龙玉
		JadeCost      []uint32 `codec:"jadecost"`  // 消耗的龙玉
		JadeCostCount []int    `codec:"jadecostc"` // 消耗的龙玉的数量
	}{}
	resp := &struct {
		SyncResp
		ResJadeId uint32 `codec:"jade"` // 身上的龙玉
	}{}

	initReqRsp(
		"PlayerAttr/LvlUpJadeResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Item_Jade_No_In_Bag
		Err_Item_Type_Not_Jade
		Err_Jade_Slot_Not_Found
		Err_Jade_Slot_Not_Match
	)

	// 作弊检查
	for _, n := range req.JadeCostCount {
		if n < 0 || n > uutil.CHEAT_INT_MAX {
			return rpcErrorWithMsg(resp, 99, "LvlUpJade JadeCostCount cheat")
		}
	}

	if len(req.JadeCost) <= 0 || len(req.JadeCost) > jade_lvlUp_max_cost_count || len(req.JadeCostCount) != len(req.JadeCost) {
		return rpcErrorWithMsg(resp, Err_Param, fmt.Sprintf("Err_Param JadeCost %v JadeCostCount %v", req.JadeCost, req.JadeCostCount))
	}

	jadeImp := account.GetJadeImp(p.Account, req.ObjTyp)
	if jadeImp == nil {
		return rpcErrorWithMsg(resp, Err_Param, fmt.Sprintf("Err_Param ObjTyp %d", req.ObjTyp))
	}
	// 目标宝石检查
	// 物品是否在
	jadeCfg, ok := p.Profile.GetJadeBag().GetJadeData(req.Jade)
	if !ok {
		logs.Warn("LvlUpJade CODE_Item_Jade_No_In_Bag item %s", jadeCfg.GetID())
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	if jadeCfg.GetType() != "JADE" {
		return rpcErrorWithMsg(resp, Err_Item_Type_Not_Jade, fmt.Sprintf("CODE_Item_Type_Not_Jade item %s", jadeCfg.GetID()))
	}
	// slot check
	req_slot := gamedata.GetJadeSlot(jadeCfg.GetPart())
	if req_slot < 0 {
		return rpcErrorWithMsg(resp, Err_Jade_Slot_Not_Found, fmt.Sprintf("[AddJade] Err_Jade_Slot_Not_Found item %s", jadeCfg.GetID()))
	}
	// 消耗宝石检查
	for _, cjid := range req.JadeCost {
		// 物品是否在
		itemCfg, ok := p.Profile.GetJadeBag().GetJadeData(cjid)
		if !ok {
			return rpcErrorWithMsg(resp, Err_Item_Jade_No_In_Bag, fmt.Sprintf("CODE_Item_Jade_No_In_Bag or be deleted item bagid %d", cjid))
		}
		if itemCfg.GetType() != "JADE" {
			return rpcErrorWithMsg(resp, Err_Item_Type_Not_Jade, fmt.Sprintf("CODE_Item_Type_Not_Jade item %s", itemCfg.GetID()))
		}
		// slot check
		cost_slot := gamedata.GetJadeSlot(itemCfg.GetPart())
		if cost_slot < 0 {
			return rpcErrorWithMsg(resp, Err_Jade_Slot_Not_Found, fmt.Sprintf("[AddJade] Err_Jade_Slot_Not_Found item %s", itemCfg.GetID()))
		}
		if cost_slot != req_slot {
			return rpcErrorWithMsg(resp, Err_Jade_Slot_Not_Match, fmt.Sprintf("[AddJade] Err_Jade_Slot_Not_Match item %s", itemCfg.GetID()))
		}
	}
	ok, resJadeId, errcode, errmsg, warnCode := jadeImp.JadeLvlUp(p.Account, req.ObjId, req_slot, req.Jade, req.JadeCost, req.JadeCostCount, resp)
	if !ok {
		if warnCode > 0 {
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		return rpcErrorWithMsg(resp, uint32(errcode+20), errmsg)
	}
	resp.ResJadeId = resJadeId
	// 强制刷新任务条件支持红点
	p.updateCondition(account.COND_TYP_EquipedJade,
		0, 0, "", "", resp)
	// 更新宝石等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&info)
	// 更新第二周宝石等级排行榜
	rank.GetModule(p.AccountID.ShardId).RankByHeroJadeTwo.Add(&info)
	jadeImp.SyncUpdate(resp)
	resp.mkInfo(p)

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 9.宝石
	return rpcSuccess(resp)
}

func (p *Account) LvlUpJadeInBag(r servers.Request) *servers.Response {
	req := &struct {
		Req
		JadeInBag uint32 `codec:"bagjade"` // 消耗的龙玉
		IsAll     bool   `codec:"isall"`
	}{}
	resp := &struct {
		SyncResp
		ResJadeId uint32 `codec:"resjade"` // 合成后的龙玉
	}{}

	initReqRsp(
		"PlayerAttr/LvlUpJadeInBagResp",
		r.RawBytes,
		req, resp, p)

	ok, res, errcode, errmsg := account.JadeLvlUpInBag(p.Account,
		req.JadeInBag, req.IsAll, resp)
	if !ok {
		logs.Warn("LvlUpJadeInBag JadeLvlUpInBag err %d %s", errcode, errmsg)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	resp.ResJadeId = res
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
