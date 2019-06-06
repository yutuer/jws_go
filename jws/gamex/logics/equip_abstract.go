package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/planx/servers/game"

	"math"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util"
	logiclogTyp "vcs.taiyouxi.net/platform/planx/util/logiclog"
)

func bestTrickIn(t []string) int {
	if len(t) == 0 {
		return -1
	}
	var (
		r int
		f float32
	)
	for i, trick := range t {
		data := gamedata.GetEquipTrickData(trick)
		if data.GetValueScore() > f {
			r = i
			f = data.GetValueScore()
		}
	}

	return r
}

func costScByEquipAbstract(p *Account, sync helper.ISyncRsp, aim_tier, aim_rare int) bool {

	setting_data := gamedata.GetEquipTrickSettingData(aim_tier)
	if setting_data == nil {
		return false
	}
	var sc uint32 = 0

	switch aim_rare {
	case gamedata.RareLv_White:
		sc = setting_data.GetPrice0()
		break
	case gamedata.RareLv_Green:
		sc = setting_data.GetPrice1()
		break
	case gamedata.RareLv_Blue:
		sc = setting_data.GetPrice2()
		break
	case gamedata.RareLv_Purple:
		sc = setting_data.GetPrice3()
		break
	case gamedata.RareLv_Gold:
		sc = setting_data.GetPrice4()
		break
	case gamedata.RareLv_Red:
		sc = setting_data.GetPrice5()
		break
	}

	if sc == 0 {
		return true
	}

	cost := &gamedata.CostData{}
	cost.AddItem(gamedata.VI_Sc0, sc)
	return account.CostBySync(p.Account, cost, sync, "EquipAbstract")
}

func mkMatEquipTrickRandPool(tier int, data *gamedata.BagItemData) (util.RandIntSet, bool) {
	// 随机选择技能
	// 首先需要确定技能的颜色
	// EUIPTRICKDETAIL表中用TrickID和ColorX确定所谓的Color
	//

	res := util.RandIntSet{}
	res.Init(len(data.TrickGroup))

	for idx, trick := range data.TrickGroup {
		detail_data := gamedata.GetEquipTrickData(trick)

		if detail_data == nil {
			return res, false
		}

		// FIXME 喊罗凯去改一下表结构 by FanYang
		color := 0
		switch tier {
		case 0:
			color = int(detail_data.GetColor0())
			break
		case 1:
			color = int(detail_data.GetColor1())
			break
		case 2:
			color = int(detail_data.GetColor2())
			break
		case 3:
			color = int(detail_data.GetColor3())
			break
		case 4:
			color = int(detail_data.GetColor4())
			break
		case 5:
			color = int(detail_data.GetColor5())
			break
		case 6:
			color = int(detail_data.GetColor6())
			break
		case 7:
			color = int(detail_data.GetColor7())
			break
		case 8:
			color = int(detail_data.GetColor8())
			break
		case 9:
			color = int(detail_data.GetColor9())
			break
		case 10:
			color = int(detail_data.GetColor10())
			break
		}

		setting_data := gamedata.GetEquipTrickSettingData(tier)
		if setting_data == nil {
			return res, false
		}
		var power uint32 = 1

		switch color {
		case 0:
			power = setting_data.GetWeight0()
			break
		case 1:
			power = setting_data.GetWeight1()
			break
		case 2:
			power = setting_data.GetWeight2()
			break
		case 3:
			power = setting_data.GetWeight3()
			break
		case 4:
			power = setting_data.GetWeight4()
			break
		}

		logs.Trace("add trick %d %s--> color %d weight %d", tier, trick, color, power)

		res.Add(idx, power)
	}

	if !res.Make() {
		logs.Error("mkMatEquipTrickRandPool make err by %v", data.TrickGroup)
		return res, false
	}
	return res, true
}

func (p *Account) EquipAbstract(r servers.Request) *servers.Response {
	req := &struct {
		Req
		AimEquipId      uint32 `codec:"aid"`
		AimTrickIdx     int    `codec:"atidx"`
		MaterialEquipId uint32 `codec:"mid"`
	}{}

	resp := &struct {
		SyncRespWithRewards
		MaterialTrickIdx int    `codec:"mtidx"`
		MaterialTrick    string `codec:"mt"`
	}{}

	acid := p.AccountID.String()
	initReqRsp(
		"PlayerAttr/EquipAbstractRsp",
		r.RawBytes, req, resp, p)

	const (
		_                          = iota
		CODE_AimEquipId_Err        // 失败 : AimEquipId错误
		CODE_AimEquipData_Err      // 失败 : AimEquip技能数据错误
		CODE_AimTrickIdx_Err       // 失败 : AimTrickIdx错误
		CODE_MaterialEquipId_Err   // 失败 : MaterialEquipId错误
		CODE_MaterialEquipData_Err // 失败 : MaterialEquip技能数据错误
		CODE_Equip_Data_Err        // 失败 : 装备数据缺失
		CODE_Part_Err              // 失败 : 装备位置不一致
		CODE_AimRareErr            // 失败 : Aim装备级别不足
		CODE_MatRandErr            // 失败 : 随机过程出错
		CODE_NoMoney               // 失败 : 没钱
	)

	if req.AimTrickIdx < 0 || req.AimTrickIdx >= gamedata.EquipTrickMaxCount {
		return rpcError(resp, CODE_AimTrickIdx_Err)
	}

	if req.AimEquipId == req.MaterialEquipId {
		logs.Warn("EquipAbstract CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	aim_equip := p.BagProfile.GetItem(req.AimEquipId)
	if aim_equip == nil || aim_equip.IsFixedID() {
		logs.Warn("EquipAbstract CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	oldCount := uint32(aim_equip.Count)

	mat_equip := p.BagProfile.GetItem(req.MaterialEquipId)
	if mat_equip == nil || mat_equip.IsFixedID() {
		return rpcWarn(resp, errCode.MaterialNotEnough)
	}

	aim_itemdata := aim_equip.GetItemData()
	if aim_itemdata == nil {
		return rpcError(resp, CODE_AimEquipData_Err)
	}
	aim_trick_len := len(aim_itemdata.TrickGroup)

	mat_itemdata := mat_equip.GetItemData()
	if mat_itemdata == nil {
		return rpcError(resp, CODE_MaterialEquipData_Err)
	}

	trick_len := len(mat_itemdata.TrickGroup)
	if trick_len <= 0 {
		return rpcError(resp, CODE_MaterialEquipData_Err)
	}

	// 同部位判断
	aim_data, aim_data_ok := gamedata.GetProtoItem(aim_equip.TableID)
	mat_data, mat_data_ok := gamedata.GetProtoItem(mat_equip.TableID)

	if !aim_data_ok || !mat_data_ok {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	if aim_data == nil || mat_data == nil {
		return rpcError(resp, CODE_Equip_Data_Err)
	}

	if aim_data.GetPart() != mat_data.GetPart() {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 蓝装以上可以洗练
	if aim_data.GetRareLevel() < gamedata.RareLv_Blue {
		return rpcError(resp, CODE_AimRareErr)
	}

	if aim_trick_len >= gamedata.EquipTrickMaxCount && req.AimTrickIdx >= aim_trick_len {
		return rpcError(resp, CODE_AimTrickIdx_Err)
	}

	var tidx int

	playerAbstract := p.Profile.GetAbstractCancelInfo()

	if playerAbstract.MN.IsNowNeedNewTurn() {
		playerAbstract.MN.Reset(
			playerAbstract.GetNumAndSpace())
	}

	isSelected := playerAbstract.MN.Selector(p.GetRand())
	playerAbstract.MN.LogicLog(acid, logiclogTyp.LogType_AbstractMN, "")

	if isSelected {
		tidx = bestTrickIn(mat_itemdata.TrickGroup[:])
	} else {
		rander, ok := mkMatEquipTrickRandPool(int(mat_data.GetTier()), mat_itemdata)
		if !ok {
			return rpcError(resp, CODE_MatRandErr)
		}
		tidx = rander.Rand(p.GetRand())
	}
	tid := mat_itemdata.TrickGroup[tidx]
	playerAbstract.AddCount()

	tier := int(math.Max(float64(aim_data.GetTier()), float64(mat_data.GetTier())))
	rareLvl := int(math.Max(float64(aim_data.GetRareLevel()), float64(mat_data.GetRareLevel())))
	sc_cost_ok := costScByEquipAbstract(p, resp, tier, rareLvl)

	if !sc_cost_ok {
		return rpcError(resp, CODE_NoMoney)
	}

	oldIron := p.Profile.GetSC().GetSC(helper.SC_FineIron)
	// 销毁材料装备
	resolve_code := p.EquipResolve(req.MaterialEquipId, resp, "EquipAbstract")
	if resolve_code != 0 {
		logs.Warn("EquipAbstract err %s %d MaterialEquipId %d", acid,
			CODE_NoMoney+resolve_code, req.MaterialEquipId)
		if resolve_code < 200 {
			return rpcWarn(resp, resolve_code)
		}
		return rpcError(resp, resolve_code)
	}

	logs.Trace("[%s]EquipAbstract %v to %d %s by %v",
		acid, aim_equip, tidx, tid, mat_equip)

	p.updateCondition(account.COND_TYP_EquipAbstract,
		1, 0, "", "", resp)

	oldAimTrick := aim_itemdata.TrickGroup[:]
	if aim_trick_len < gamedata.EquipTrickMaxCount {
		// 目标技能中有空槽, 直接追加
		aim_itemdata.TrickGroup = append(aim_itemdata.TrickGroup, tid)
	} else {

		// 目标技能中没有空槽, 这才按照请求的槽位覆盖
		old_tid := aim_itemdata.TrickGroup[req.AimTrickIdx]
		playerAbstract.Add(req.AimEquipId, old_tid, req.AimTrickIdx)
		aim_itemdata.TrickGroup[req.AimTrickIdx] = tid
	}
	aim_equip.SetItemData(aim_itemdata)
	p.BagProfile.UpdateItem(aim_equip)

	resp.MaterialTrick = tid
	resp.MaterialTrickIdx = tidx

	p.Profile.GetData().SetNeedCheckMaxGS()

	resp.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, req.AimEquipId, int64(oldCount), "EquipAbstract")
	resp.mkInfo(p)

	// log
	logiclog.LogEquipAbstract(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		aim_equip.TableID, mat_equip.TableID,
		oldAimTrick, aim_itemdata.TrickGroup,
		p.Profile.GetSC().GetSC(helper.SC_FineIron)-oldIron,
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

func (p *Account) EquipAbstractCancel(r servers.Request) *servers.Response {
	req := &struct {
		Req
		AimEquipId uint32 `codec:"aid"`
	}{}
	resp := &struct {
		SyncResp
		AimTrickIdx int    `codec:"atidx"`
		AimTrick    string `codec:"at"`
	}{}

	acid := p.AccountID.String()
	const (
		_                     = iota
		CODE_AimEquipId_Err   // 失败 : AimEquipId错误
		CODE_AimEquipData_Err // 失败 : AimEquip技能数据错误
		CODE_AimTrickIdx_Err  // 失败 : AimEquip技能数据错误
		CODE_CostHC_Err       // 失败 : hc不足
	)

	initReqRsp(
		"PlayerAttr/EquipAbstractCancelRsp",
		r.RawBytes, req, resp, p)

	abinfo := p.Profile.GetAbstractCancelInfo()

	info := abinfo.Get(req.AimEquipId)
	if info == nil {
		logs.Warn("EquipAbstractCancel CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	logs.Trace("[%s]EquipAbstractCancel %v", acid, info)

	aim_equip := p.BagProfile.GetItem(req.AimEquipId)
	if aim_equip == nil || aim_equip.IsFixedID() {
		logs.Warn("EquipAbstractCancel CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	oldCount := uint32(aim_equip.Count)

	aim_itemdata := aim_equip.GetItemData()
	if aim_itemdata == nil {
		return rpcError(resp, CODE_AimEquipData_Err)
	}
	aim_trick_len := len(aim_itemdata.TrickGroup)
	if info.AimEquipIdx < 0 || info.AimEquipIdx >= aim_trick_len {
		return rpcError(resp, CODE_AimTrickIdx_Err)
	}

	hccost := gamedata.GetAbstractCancelCost()
	if !account.CostBySync(p.Account, hccost, resp, "AbstractCancelCost") {
		return rpcError(resp, CODE_CostHC_Err)
	}

	oldTricks := aim_itemdata.TrickGroup[:]
	aim_itemdata.TrickGroup[info.AimEquipIdx] = info.AimEquipTrick
	aim_equip.SetItemData(aim_itemdata)
	p.BagProfile.UpdateItem(aim_equip)

	resp.AimTrick = info.AimEquipTrick
	resp.AimTrickIdx = info.AimEquipIdx
	abinfo.Clean(req.AimEquipId)

	p.Profile.GetData().SetNeedCheckMaxGS()

	resp.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, req.AimEquipId, int64(oldCount), "EquipAbstractCancel")
	resp.mkInfo(p)
	// log
	logiclog.LogEquipAbstractCancel(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		aim_equip.TableID, oldTricks, aim_itemdata.TrickGroup,
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) EquipTrickSwap(r servers.Request) *servers.Response {
	req := &struct {
		Req
		AimEquipId      uint32 `codec:"aid"`
		MaterialEquipId uint32 `codec:"mid"`
	}{}

	resp := &struct {
		SyncResp
	}{}

	acid := p.AccountID.String()
	initReqRsp(
		"PlayerAttr/EquipTrickSwapRsp",
		r.RawBytes, req, resp, p)

	const (
		_                          = iota
		CODE_AimEquipId_Err        // 失败 : AimEquipId错误
		CODE_AimEquipData_Err      // 失败 : AimEquip技能数据错误
		CODE_AimTrickIdx_Err       // 失败 : AimTrickIdx错误
		CODE_MaterialEquipId_Err   // 失败 : MaterialEquipId错误
		CODE_MaterialEquipData_Err // 失败 : MaterialEquip技能数据错误
		CODE_Equip_Data_Err        // 失败 : 装备数据缺失
		CODE_Part_Err              // 失败 : 装备位置不一致
		CODE_AimRareErr            // 失败 : Aim装备级别不足
		CODE_MatRandErr            // 失败 : 随机过程出错
		CODE_NoMoney               // 失败 : 没钱
	)

	if req.AimEquipId == req.MaterialEquipId {
		logs.Warn("EquipTrickSwap CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	aim_equip := p.BagProfile.GetItem(req.AimEquipId)
	if aim_equip == nil || aim_equip.IsFixedID() {
		logs.Warn("EquipTrickSwap CODE_AimEquipId_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	oldEquipCount := uint32(aim_equip.Count)

	mat_equip := p.BagProfile.GetItem(req.MaterialEquipId)
	if mat_equip == nil || mat_equip.IsFixedID() {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}
	oldMatCount := uint32(mat_equip.Count)

	aim_itemdata := aim_equip.GetItemData()
	if aim_itemdata == nil {
		return rpcError(resp, CODE_AimEquipData_Err)
	}
	mat_itemdata := mat_equip.GetItemData()
	if mat_itemdata == nil {
		return rpcError(resp, CODE_MaterialEquipData_Err)
	}

	// 同部位判断
	aim_data, aim_data_ok := gamedata.GetProtoItem(aim_equip.TableID)
	mat_data, mat_data_ok := gamedata.GetProtoItem(mat_equip.TableID)

	if !aim_data_ok || !mat_data_ok {
		return rpcError(resp, CODE_Equip_Data_Err)
	}

	if aim_data == nil || mat_data == nil {
		return rpcError(resp, CODE_Equip_Data_Err)
	}

	if aim_data.GetPart() != mat_data.GetPart() {
		logs.Warn("EquipTrickSwap CODE_Part_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 蓝装以上
	if aim_data.GetRareLevel() < gamedata.RareLv_Blue {
		return rpcError(resp, CODE_AimRareErr)
	}

	if mat_data.GetRareLevel() < gamedata.RareLv_Blue {
		return rpcError(resp, CODE_AimRareErr)
	}

	cost := gamedata.GetEquipTrickSwapCost()
	sc_cost_ok := account.CostBySync(p.Account, cost, resp, "EquipTrickSwap")

	if !sc_cost_ok {
		return rpcError(resp, CODE_NoMoney)
	}

	logs.Trace("[%s]EquipTrickSwap %v by %v",
		acid, aim_equip, mat_equip)

	t := mat_equip.GetItemData().TrickGroup[:]
	mat_itemdata.TrickGroup = aim_itemdata.TrickGroup[:]
	aim_itemdata.TrickGroup = t
	aim_equip.SetItemData(aim_itemdata)
	p.BagProfile.UpdateItem(aim_equip)
	mat_equip.SetItemData(mat_itemdata)
	p.BagProfile.UpdateItem(mat_equip)

	p.Profile.GetData().SetNeedCheckMaxGS()

	resp.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, req.AimEquipId, int64(oldEquipCount), "EquipTrickSwap")
	resp.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, req.MaterialEquipId, int64(oldMatCount), "EquipTrickSwap")
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
