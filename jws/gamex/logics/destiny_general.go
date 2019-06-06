package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/dest_gen_first"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) UnlockDestinyGeneral(r servers.Request) *servers.Response {
	req := &struct {
		Req
		DestinyGeneralID int `codec:"id"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/UnlockDestinyGlRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                   = iota
		CODE_No_Data_Err    // 失败:数据错误
		CODE_Cost_Err       // 失败:没有物品
		CODE_Corp_Lv_Err    // 失败:角色等级不足
		CODE_Destiny_Lv_Err // 失败:神将等级不足
	)

	// 有些是有花费的
	data := gamedata.GetDestinyGeneralUnlockData(req.DestinyGeneralID)
	if data == nil {
		return rpcError(resp, CODE_No_Data_Err)
	}

	distiny := p.Profile.GetDestinyGeneral()

	// 1. 角色等级限制
	corpLv, _ := p.Profile.GetCorp().GetXpInfo()
	if corpLv < data.UnlockLevel {
		return rpcError(resp, CODE_Corp_Lv_Err)
	}

	// 2. 前置的神将等级限制
	dg := distiny.GetGeneral(data.UnlockGeneralID)
	if data.UnlockGeneralLevel > 0 && (dg == nil || dg.LevelIndex < data.UnlockGeneralLevel) {
		return rpcError(resp, CODE_Destiny_Lv_Err)
	}

	// 3. 消耗解锁道具
	if !account.CostBySync(p.Account, &data.UnlockCost, resp, "DistinyGeneralUnlock") {
		return rpcError(resp, CODE_Cost_Err)
	}

	distiny.AddNewGeneral(req.DestinyGeneralID)

	// 首次激活记录
	if data.Cfg.GetUnlockGeneralMaterial() != "" {
		p._firstDest(req.DestinyGeneralID)
	}

	// MaxGS可能变化 10.神将
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.mkInfo(p)

	logiclog.LogDestinyGeneralAct(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		req.DestinyGeneralID,
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	// sysnotice
	cfg := gamedata.ActDestingGeneralSysNotice(uint32(req.DestinyGeneralID))
	if cfg != nil {
		sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfg.GetServerMsgID())).
			AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
	}
	return rpcSuccess(resp)
}

func (p *Account) SetDestinyGeneralSkill(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Skills []int `codec:"skill"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/SetDestinyGlSkillRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                  = iota
		CODE_No_Data_Err   // 失败:
		CODE_ID_Err        // 失败:
		CODE_Avatar_Lv_Err // 失败:
	)

	distiny := p.Profile.GetDestinyGeneral()
	if req.Skills == nil || len(req.Skills) == 0 || len(req.Skills) > helper.DestinyGeneralSkillMax {
		logs.Warn("SetDestinyGeneralSkill CODE_No_Data_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	if !distiny.SetSkills(req.Skills) {
		return rpcError(resp, CODE_ID_Err)
	}

	// 强制刷新任务条件支持红点
	p.updateCondition(account.COND_TYP_DestinyGeneralLv,
		0, 0, "", "", resp)
	// 检查神兽任务是否完成
	resp.OnChangeQuestAll()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// NewAddDestinyGeneralLv : 新神将升级协议
// 新神将升级协议

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgNewAddDestinyGeneralLv 新神将升级协议请求消息定义
type reqMsgNewAddDestinyGeneralLv struct {
	Req
	DestinyGeneralID int64 `codec:"id"`       // 神兽ID
	LvlUpTyp         int64 `codec:"lvluptyp"` // 升级类型，0:普通，1:高级, 2:连续高级
}

// rspMsgNewAddDestinyGeneralLv 新神将升级协议回复消息定义
type rspMsgNewAddDestinyGeneralLv struct {
	SyncRespWithRewards
	AddExp []int64 `codec:"addexp"` // 每次增加的经验
	IsCrit []int64 `codec:"iscrit"` // 每次是否暴击，0：不暴击，1：小暴击，2：大暴击
}

// NewAddDestinyGeneralLv 新神将升级协议: 新神将升级协议
func (p *Account) NewAddDestinyGeneralLv(r servers.Request) *servers.Response {
	req := new(reqMsgNewAddDestinyGeneralLv)
	rsp := new(rspMsgNewAddDestinyGeneralLv)

	initReqRsp(
		"Attr/NewAddDestinyGeneralLvRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
		Err_Cfg
		CODE_ID_Err
		Err_Cost
		Err_Vip
	)

	distiny := p.Profile.GetDestinyGeneral()
	distiny.UpdateDGTimes(int(p.Profile.GetVipLevel()), p.Profile.GetProfileNowTime())

	destId := int(req.DestinyGeneralID)
	dg := distiny.GetGeneral(destId)
	if dg == nil {
		return rpcErrorWithMsg(rsp, CODE_ID_Err, "CODE_ID_Err")
	}

	dCfg := gamedata.GetDestinyConfig()
	ulCfg := gamedata.GetDestinyGeneralUnlockData(destId)
	lvlCfg := gamedata.GetNewDestinyGeneralLevelData(destId, dg.LevelIndex+1)
	if ulCfg == nil || lvlCfg == nil {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	rsp.AddExp = make([]int64, 0, 20)
	rsp.IsCrit = make([]int64, 0, 20)
	curLvl := dg.LevelIndex
	curExp := dg.Exp
	norm_coin := p.Profile.GetSC().GetSC(helper.SCId(ulCfg.Cfg.GetGeneralCoin()))
	hc := p.Profile.GetHC().GetHC()
	bef_norm_coin := norm_coin
	bef_hc := hc
	var reason string
	switch req.LvlUpTyp {
	case 0: //  普通
		reason = "AddDestinyGeneralLv-Norm"
		for i := 0; i < int(dCfg.GetTrainTimes()); i++ {
			// 检查等级，防止超过满级
			lvlCfg = gamedata.GetNewDestinyGeneralLevelData(destId, curLvl+1)
			if lvlCfg == nil {
				break
			}
			if norm_coin < int64(ulCfg.Cfg.GetGeneralCost()) {
				break // 只扣当前dc够的那么多次数
			}
			norm_coin -= int64(ulCfg.Cfg.GetGeneralCost())
			var add_exp uint32
			if p.GetRand().Float32() <= lvlCfg.Cfg.GetDGLittleBonusRate()+
				lvlCfg.Cfg.GetDGBigBonusRate() { // 暴击
				add_exp = ulCfg.Cfg.GetExpUnit() * dCfg.GetDGLittleBonus()
				rsp.IsCrit = append(rsp.IsCrit, 1)
			} else {
				add_exp = ulCfg.Cfg.GetExpUnit()
				rsp.IsCrit = append(rsp.IsCrit, 0)
			}
			rsp.AddExp = append(rsp.AddExp, int64(add_exp))
			curLvl, curExp = _simDGAddExp(destId, curLvl, curExp, add_exp, lvlCfg)
		}
	case 1: // 高级
		// 次数检查
		reason = "AddDestinyGeneralLv-Vip"
		if distiny.VipTimes > 0 { // 优先每日免费次数
			distiny.VipTimes--
		} else { // 消耗
			if hc < int64(ulCfg.Cfg.GetVIPCost()) {
				logs.Warn("NewAddDestinyGeneralLv Err_Cost1")
				return rpcWarn(rsp, errCode.ClickTooQuickly)
			}
			hc -= int64(ulCfg.Cfg.GetVIPCost())
		}
		rf := p.GetRand().Float32()
		if rf <= lvlCfg.Cfg.GetDGBigBonusRate() { // 大暴击
			var add_exp uint32
			curLvl, curExp, add_exp = _simDGAddLv(curLvl, curExp, lvlCfg)
			rsp.AddExp = append(rsp.AddExp, int64(add_exp))
			rsp.IsCrit = append(rsp.IsCrit, 2)
		} else if rf <= lvlCfg.Cfg.GetDGLittleBonusRate() { // 小暴击
			add_exp := ulCfg.Cfg.GetExpVIPUnit() * dCfg.GetDGLittleBonus()
			curLvl, curExp = _simDGAddExp(destId, curLvl, curExp, add_exp, lvlCfg)
			rsp.AddExp = append(rsp.AddExp, int64(add_exp))
			rsp.IsCrit = append(rsp.IsCrit, 1)
		} else {
			add_exp := ulCfg.Cfg.GetExpVIPUnit()
			curLvl, curExp = _simDGAddExp(destId, curLvl, curExp, add_exp, lvlCfg)
			rsp.AddExp = append(rsp.AddExp, int64(add_exp))
			rsp.IsCrit = append(rsp.IsCrit, 0)
		}
	case 2: // 连续高级
		reason = "AddDestinyGeneralLv-VipContnu"
		// vip 检查
		vipCfg := gamedata.GetVIPCfg(int(p.Profile.GetVipLevel()))
		if vipCfg == nil || !vipCfg.DGVipAdv {
			return rpcErrorWithMsg(rsp, Err_Vip, "Err_Vip")
		}
		for i := 0; i < int(dCfg.GetTrainContinuityTimes()); i++ {
			// 检查等级，防止因为暴击超过满级
			lvlCfg = gamedata.GetNewDestinyGeneralLevelData(destId, curLvl+1)
			if lvlCfg == nil {
				break
			}
			if hc < int64(ulCfg.Cfg.GetVIPCost()) {
				break
			}
			hc -= int64(ulCfg.Cfg.GetVIPCost())

			rf := p.GetRand().Float32()
			if rf <= lvlCfg.Cfg.GetDGBigBonusRate() { // 大暴击
				var add_exp uint32
				curLvl, curExp, add_exp = _simDGAddLv(curLvl, curExp, lvlCfg)
				rsp.AddExp = append(rsp.AddExp, int64(add_exp))
				rsp.IsCrit = append(rsp.IsCrit, 2)
			} else if rf <= lvlCfg.Cfg.GetDGLittleBonusRate() {
				add_exp := ulCfg.Cfg.GetExpVIPUnit() * dCfg.GetDGLittleBonus()
				curLvl, curExp = _simDGAddExp(destId, curLvl, curExp, add_exp, lvlCfg)
				rsp.AddExp = append(rsp.AddExp, int64(add_exp))
				rsp.IsCrit = append(rsp.IsCrit, 1)
			} else {
				add_exp := ulCfg.Cfg.GetExpVIPUnit()
				curLvl, curExp = _simDGAddExp(destId, curLvl, curExp, add_exp, lvlCfg)
				rsp.AddExp = append(rsp.AddExp, int64(add_exp))
				rsp.IsCrit = append(rsp.IsCrit, 0)
			}
		}
	default:
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}

	dg.LevelIndex = curLvl
	dg.Exp = curExp

	data := &gamedata.CostData{}
	data.AddItem(ulCfg.Cfg.GetGeneralCoin(), uint32(bef_norm_coin-norm_coin))
	data.AddItem(ulCfg.Cfg.GetVIPCoin(), uint32(bef_hc-hc))
	if !account.CostBySync(p.Account, data, rsp, reason) {
		logs.Warn("NewAddDestinyGeneralLv Err_Cost2")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	// 首次满级记录
	p._firstDest(destId)

	// MaxGS可能变化 10.神将
	p.Profile.GetData().SetNeedCheckMaxGS()

	// 更新神兽等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByDestiny.Add(&info)

	rsp.OnChangeDestinyGeneral()
	// 检查神兽任务是否完成
	rsp.OnChangeQuestAll()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func _simDGAddExp(destId, lvl int, bef_exp, add_xp uint32,
	lvlCfg *gamedata.DestinyGeneralLevelData) (
	aft_lvl int, aft_exp uint32) {

	exp := bef_exp + add_xp
	for lvlCfg != nil && exp >= lvlCfg.Cfg.GetDestinyGeneralExp() {
		lvl++
		exp -= lvlCfg.Cfg.GetDestinyGeneralExp()
		lvlCfg = gamedata.GetNewDestinyGeneralLevelData(destId, lvl+1)
	}
	return lvl, exp
}

func _simDGAddLv(lvl int, bef_exp uint32,
	lvlCfg *gamedata.DestinyGeneralLevelData) (
	aft_lvl int, aft_exp, add_exp uint32) {

	add_exp = lvlCfg.Cfg.GetDestinyGeneralExp() - bef_exp
	aft_lvl = lvl + 1
	aft_exp = 0
	return
}

func (p *Account) _firstDest(destId int) {
	dg := p.Profile.GetDestinyGeneral().GetGeneral(destId)
	cfgData := gamedata.GetNewDestinyGeneralLevelDatas(destId)
	maxLv := len(cfgData) - 1
	if maxLv == dg.LevelIndex {

		if dest_gen_first.GetModule(p.AccountID.ShardId).
			TryAddFirstDestGen(destId, p.Profile.Name,
				p.Profile.CurrAvatar) {
			// 跑马灯, 全服首个
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_SHENSHOUDIYI).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_DGId, fmt.Sprintf("%d", destId)).Send()
		} else {
			// 跑马灯, 自己满级
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_SHENSHOUMANJI).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_DGId, fmt.Sprintf("%d", destId)).Send()
		}
	}
}

// 老的逻辑，现在不用
func (p *Account) AddDestinyGeneralLv(r servers.Request) *servers.Response {
	req := &struct {
		Req
		DestinyGeneralID int `codec:"id"`
		LvAdd            int `codec:"ladd"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"Attr/AddDestinyGlLvRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                = iota
		CODE_No_Data_Err // 失败:数据错误
		CODE_Cost_Err    // 失败:没有物品
		CODE_ID_Err      // 失败:神将ID错误
	)

	distiny := p.Profile.GetDestinyGeneral()

	dg := distiny.GetGeneral(req.DestinyGeneralID)
	if dg == nil {
		return rpcError(resp, CODE_ID_Err)
	}

	costs := account.CostGroup{}

	for i := 0; i < req.LvAdd; i++ {
		lvdata := gamedata.GetDestinyGeneralLevelData(dg.Id, dg.LevelIndex+i+1)
		if lvdata == nil {
			return rpcError(resp, CODE_No_Data_Err)
		}
		if !costs.AddCostData(p.Account, &lvdata.LevelUpCost.Cost) {
			return rpcError(resp, CODE_Cost_Err)
		}
	}

	if !costs.CostBySync(p.Account, resp, "AddDestinyGeneralLv") {
		return rpcError(resp, CODE_Cost_Err)
	}

	distiny.AddGeneralLevel(dg.Id, req.LvAdd)
	// MaxGS可能变化 10.神将
	p.Profile.GetData().SetNeedCheckMaxGS()

	// 强制刷新任务条件支持红点
	p.updateCondition(account.COND_TYP_DestinyGeneralLv,
		0, 0, "", "", resp)
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
