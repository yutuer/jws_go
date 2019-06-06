package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) EquipMatEnhanceAdd(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Slot     int `codec:"slot"`
		MatIndex int `codec:"mat_idx"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/EquipMatEnhanceAddResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Full_Lvl
		Err_Cost
	)

	if req.Slot >= helper.PartEquipCount {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	equip := p.Profile.GetEquips()
	// 本级材料没满，查找包裹里材料并放入
	curLv := equip.GetMatEnhLv(req.Slot)
	slotMat := equip.GetMatEnhSlotInfo(req.Slot)
	if req.MatIndex >= len(slotMat) {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	cfg := gamedata.GetEquipMatEnhCfg(req.Slot, curLv+1)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_Full_Lvl, "Err_Cfg_Not_found")
	}
	cms := cfg.GetMaterials_Table()
	cost := &account.CostGroup{}
	if !slotMat[req.MatIndex] {
		nm := cms[req.MatIndex]
		// 材料够
		if cost.AddItemById(p.Account, nm.GetMaterialsID(), nm.GetMaterialsCount()) &&
			cost.CostBySync(p.Account, resp, "EquipMatEnhanceAdd") {
			equip.SetMatEnhSlotInfo(req.Slot, req.MatIndex)

			p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 2.装备

			resp.OnChangeEquip()
			resp.mkInfo(p)

			// log
			logiclog.LogEquipMatEnhAdd(p.AccountID.String(), p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, req.Slot, curLv, slotMat, equip.GetMatEnhSlotInfo(req.Slot), p.Profile.GetData().CorpCurrGS,
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}
	return rpcSuccess(resp)
}

func (p *Account) EquipMatEnhanceAutoAdd(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Slot int `codec:"slot"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/EquipMatEnhanceAutoAddResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Full_Lvl
		Err_Cost
	)

	if req.Slot >= helper.PartEquipCount {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	equip := p.Profile.GetEquips()
	// 本级材料没满，查找包裹里材料并放入
	curLv := equip.GetMatEnhLv(req.Slot)
	slotMat := equip.GetMatEnhSlotInfo(req.Slot)
	cfg := gamedata.GetEquipMatEnhCfg(req.Slot, curLv+1)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_Full_Lvl, "Err_Cfg_Not_found")
	}
	cms := cfg.GetMaterials_Table()
	cost := &account.CostGroup{}
	isUp := false
	for i := 0; i < len(cms); i++ {
		// 某个未放材料的位
		if !slotMat[i] {
			nm := cms[i]
			// 材料够
			if cost.AddItemById(p.Account, nm.GetMaterialsID(), nm.GetMaterialsCount()) {
				equip.SetMatEnhSlotInfo(req.Slot, i)
				isUp = true
			}
		}
	}

	if isUp {
		if !cost.CostBySync(p.Account, resp, "EquipMatEnhanceAutoAdd") {
			return rpcErrorWithMsg(resp, Err_Cost, "Err_Cost")
		}

		p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 2.装备

		resp.OnChangeEquip()
		resp.mkInfo(p)

		// log
		logiclog.LogEquipMatEnhAdd(p.AccountID.String(), p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, req.Slot, curLv, slotMat, equip.GetMatEnhSlotInfo(req.Slot),
			p.Profile.GetData().CorpCurrGS,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}
	return rpcSuccess(resp)
}

func (p *Account) EquipMatEnhanceLvlUp(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Slot int `codec:"slot"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/EquipMatEnhanceLvlUpResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Full_Lvl
		Err_Cost
		Err_Mat_Not_Full
	)

	if req.Slot >= helper.PartEquipCount {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	equip := p.Profile.GetEquips()
	curLv := equip.GetMatEnhLv(req.Slot)
	slotMat := equip.GetMatEnhSlotInfo(req.Slot)
	cfg := gamedata.GetEquipMatEnhCfg(req.Slot, curLv+1)
	if cfg == nil {
		logs.Warn("EquipMatEnhanceLvlUp Err_Full_Lvl")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	cms := cfg.GetMaterials_Table()
	// 材料是否满
	for i := 0; i < len(cms); i++ {
		if !slotMat[i] {
			logs.Warn("EquipMatEnhanceLvlUp Err_Mat_Not_Full")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
	}
	// 扣钱
	cost := &account.CostGroup{}
	if !cost.AddSc(p.Account, helper.SC_Money, int64(cfg.GetSC())) ||
		!cost.CostBySync(p.Account, resp, "EquipMatEnhanceLvlUp") {
		logs.Warn("EquipMatEnhanceLvlUp Err_Cost")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	// 升级
	equip.LvlUpMatEnh(req.Slot)

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 2.装备

	// 52.将N件装备升阶N次, P1装备数, P2升阶等级
	p.updateCondition(account.COND_TYP_Equip_Mat_Enh,
		0, 0, "", "", resp)

	resp.OnChangeEquip()
	resp.mkInfo(p)

	// log
	logiclog.LogEquipMatEnhLvlUp(p.AccountID.String(), p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, req.Slot, curLv, equip.GetMatEnhLv(req.Slot),
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}
